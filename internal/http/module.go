package http

import (
	"context"
	"log/slog"
	"strings"

	"github.com/DaiYuANg/jumpa/internal/api"
	auth2 "github.com/DaiYuANg/jumpa/internal/auth"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/pkg"
	"github.com/arcgolabs/authx"
	authhttp "github.com/arcgolabs/authx/http"
	authjwt "github.com/arcgolabs/authx/jwt"
	"github.com/arcgolabs/collectionx"
	collectionset "github.com/arcgolabs/collectionx/set"
	"github.com/arcgolabs/dix"
	"github.com/arcgolabs/httpx"
	"github.com/arcgolabs/httpx/adapter"
	"github.com/arcgolabs/httpx/adapter/fiber"
	"github.com/go-playground/validator/v10"
	fiberapp "github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	fiberrequestid "github.com/gofiber/fiber/v2/middleware/requestid"
)

type AuthzPolicyConfig struct {
	ProtectedPrefix     string
	PublicPaths         []string
	AuthOnlyResources   []string
	MethodActionMapping map[string]string
}

func DefaultAuthzPolicyConfig() AuthzPolicyConfig {
	return AuthzPolicyConfig{
		ProtectedPrefix: "/api",
		PublicPaths:     []string{"/api/auth/login", "/api/auth/refresh", "/api/health"},
		AuthOnlyResources: []string{
			"me",
			"auth",
		},
		MethodActionMapping: map[string]string{
			fiberapp.MethodGet:     "read",
			fiberapp.MethodHead:    "read",
			fiberapp.MethodOptions: "read",
			fiberapp.MethodPost:    "write",
			fiberapp.MethodPut:     "write",
			fiberapp.MethodPatch:   "write",
			fiberapp.MethodDelete:  "delete",
		},
	}
}

func authzPolicyFromConfig(cfg config2.AppConfig) AuthzPolicyConfig {
	p := DefaultAuthzPolicyConfig()
	if strings.TrimSpace(cfg.Authz.ProtectedPrefix) != "" {
		p.ProtectedPrefix = strings.TrimSpace(cfg.Authz.ProtectedPrefix)
	}
	if strings.TrimSpace(cfg.Authz.PublicPathsCSV) != "" {
		p.PublicPaths = pkg.ParseCSVList(cfg.Authz.PublicPathsCSV)
	}
	if strings.TrimSpace(cfg.Authz.AuthOnlyResourcesCSV) != "" {
		p.AuthOnlyResources = pkg.ParseCSVList(cfg.Authz.AuthOnlyResourcesCSV)
	}
	return p
}

func newFiberApp() *fiberapp.App {
	return fiberapp.New()
}

func setupFiberMiddleware(app *fiberapp.App, guard *authhttp.Guard, policy AuthzPolicyConfig) {
	app.Use(fiberlogger.New(), fiberrecover.New(), fiberrequestid.New())
	app.Use(buildAuthMiddleware(guard, policy))
}

func newAuthGuard(engine *authx.Engine, policy AuthzPolicyConfig) *authhttp.Guard {
	return authhttp.NewGuard(
		engine,
		authhttp.WithCredentialResolverFunc(resolveAuthCredential),
		authhttp.WithAuthorizationResolverFunc(func(_ context.Context, req authhttp.RequestInfo, principal any) (authx.AuthorizationModel, error) {
			return authx.AuthorizationModel{
				Principal: principal,
				Resource:  protectedResource(req.Path, policy.ProtectedPrefix),
				Action:    resolveAuthAction(req.Method, policy.MethodActionMapping),
				Context: collectionx.NewMapFrom(map[string]any{
					"path":          req.Path,
					"method":        req.Method,
					"route_pattern": req.RoutePattern,
				}),
			}, nil
		}),
	)
}

func buildAuthMiddleware(guard *authhttp.Guard, policy AuthzPolicyConfig) fiberapp.Handler {
	publicPathSet := collectionset.NewSet(policy.PublicPaths...)
	authOnlySet := collectionset.NewSet(policy.AuthOnlyResources...)
	return func(c *fiberapp.Ctx) error {
		path := c.Path()
		if !strings.HasPrefix(path, policy.ProtectedPrefix) {
			return c.Next()
		}
		if publicPathSet.Contains(path) {
			return c.Next()
		}

		req := authRequestInfo(c)
		resource := protectedResource(path, policy.ProtectedPrefix)
		if authOnlySet.Contains(resource) {
			result, err := guard.Check(c.UserContext(), req)
			if err != nil {
				return c.Status(authhttp.StatusCodeFromError(err)).JSON(map[string]string{"message": authhttp.ErrorMessage(err)})
			}
			c.SetUserContext(authx.WithPrincipal(c.UserContext(), result.Principal))
			return c.Next()
		}

		result, decision, err := guard.Require(c.UserContext(), req)
		if err != nil {
			return c.Status(authhttp.StatusCodeFromError(err)).JSON(map[string]string{"message": authhttp.ErrorMessage(err)})
		}
		if !decision.Allowed {
			return c.Status(403).JSON(map[string]string{"message": authhttp.DeniedMessage(decision)})
		}
		c.SetUserContext(authx.WithPrincipal(c.UserContext(), result.Principal))
		return c.Next()
	}
}

func resolveAuthCredential(_ context.Context, req authhttp.RequestInfo) (any, error) {
	token, ok := parseBearerToken(req.Header("Authorization"))
	if !ok {
		return nil, authx.ErrInvalidAuthenticationCredential
	}
	return authjwt.NewTokenCredential(token), nil
}

func resolveAuthAction(method string, methodActionMapping map[string]string) string {
	if action, ok := methodActionMapping[method]; ok {
		return action
	}
	return "read"
}

func parseBearerToken(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if !strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return "", false
	}
	token := strings.TrimSpace(value[7:])
	return token, token != ""
}

func protectedResource(path, prefix string) string {
	if path == prefix {
		return ""
	}
	resource := strings.TrimPrefix(path, prefix+"/")
	if idx := strings.IndexByte(resource, '/'); idx >= 0 {
		resource = resource[:idx]
	}
	return resource
}

func authRequestInfo(c *fiberapp.Ctx) authhttp.RequestInfo {
	routePattern := c.Path()
	var pathParams map[string]string
	if route := c.Route(); route != nil {
		if pattern := strings.TrimSpace(route.Path); pattern != "" {
			routePattern = pattern
		}
		if len(route.Params) > 0 {
			pathParams = make(map[string]string, len(route.Params))
			for _, param := range route.Params {
				pathParams[param] = c.Params(param)
			}
		}
	}
	return authhttp.RequestInfo{
		Method:       c.Method(),
		Path:         c.Path(),
		RoutePattern: routePattern,
		PathParams:   pathParams,
		Native:       c,
	}
}

func buildHTTPServer(app *fiberapp.App, endpoints collectionx.List[httpx.Endpoint], log *slog.Logger) httpx.ServerRuntime {
	ad := fiber.New(app, adapter.HumaOptions{
		Title:       "ArcGo Backend API",
		Version:     "1.0.0",
		Description: "configx + logx + eventx + httpx(fiber) + dix + dbx",
		DocsPath:    "/docs",
		OpenAPIPath: "/openapi.json",
	})
	server := httpx.New(
		httpx.WithAdapter(ad),
		httpx.WithLogger(log),
		httpx.WithPrintRoutes(true),
		httpx.WithValidator(validator.New(validator.WithRequiredStructEnabled())),
		httpx.WithValidation(),
	)
	endpoints.Range(func(_ int, endpoint httpx.Endpoint) bool {
		server.RegisterOnly(endpoint)
		return true
	})
	return server
}

var Module = dix.NewModule("http",
	dix.WithModuleImports(config2.Module, api.Module, auth2.Module),
	dix.WithModuleProviders(
		dix.Provider1(func(cfg config2.AppConfig) AuthzPolicyConfig { return authzPolicyFromConfig(cfg) }),
		dix.Provider2(func(engine *authx.Engine, policy AuthzPolicyConfig) *authhttp.Guard {
			return newAuthGuard(engine, policy)
		}),
		dix.Provider4(func(endpoints collectionx.List[httpx.Endpoint], log *slog.Logger, guard *authhttp.Guard, policy AuthzPolicyConfig) httpx.ServerRuntime {
			app := newFiberApp()
			setupFiberMiddleware(app, guard, policy)
			return buildHTTPServer(app, endpoints, log)
		}),
	),
	dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
		server, _ := dix.ResolveAs[httpx.ServerRuntime](c)
		cfg, _ := dix.ResolveAs[config2.AppConfig](c)
		p := cfg.Server.Port
		lc.OnStart(func(ctx context.Context) error {
			go func() { _ = server.ListenPort(p) }()
			return nil
		})
		lc.OnStop(func(ctx context.Context) error { return server.Shutdown() })
		return nil
	}),
)
