package http

import (
	"context"
	"log/slog"
	"strings"

	"github.com/DaiYuANg/arcgo/authx"
	authhttp "github.com/DaiYuANg/arcgo/authx/http"
	"github.com/DaiYuANg/arcgo/collectionx"
	collectionset "github.com/DaiYuANg/arcgo/collectionx/set"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/DaiYuANg/arcgo/httpx/adapter/fiber"
	"github.com/DaiYuANg/jumpa/internal/api"
	auth2 "github.com/DaiYuANg/jumpa/internal/auth"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/httpendpoint"
	"github.com/DaiYuANg/jumpa/pkg"
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

func setupFiberMiddleware(app *fiberapp.App, engine *authx.Engine, policy AuthzPolicyConfig) {
	app.Use(fiberlogger.New(), fiberrecover.New(), fiberrequestid.New())
	app.Use(buildAuthMiddleware(engine, policy))
}

func buildAuthMiddleware(engine *authx.Engine, policy AuthzPolicyConfig) fiberapp.Handler {
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
		token := ""
		authz := strings.TrimSpace(c.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			token = strings.TrimSpace(authz[7:])
		}
		if token == "" {
			return c.Status(401).JSON(map[string]string{"message": "missing bearer token"})
		}
		check, err := engine.Check(c.UserContext(), token)
		if err != nil {
			return c.Status(401).JSON(map[string]string{"message": authhttp.ErrorMessage(err)})
		}
		resource := strings.TrimPrefix(path, policy.ProtectedPrefix+"/")
		if idx := strings.IndexByte(resource, '/'); idx >= 0 {
			resource = resource[:idx]
		}
		if authOnlySet.Contains(resource) {
			c.SetUserContext(authx.WithPrincipal(c.UserContext(), check.Principal))
			return c.Next()
		}
		action, ok := policy.MethodActionMapping[c.Method()]
		if !ok {
			action = "read"
		}
		decision, canErr := engine.Can(c.UserContext(), authx.AuthorizationModel{
			Principal: check.Principal,
			Resource:  resource,
			Action:    action,
			Context: collectionx.NewMapFrom(map[string]any{
				"path":   path,
				"method": c.Method(),
			}),
		})
		if canErr != nil {
			return c.Status(500).JSON(map[string]string{"message": authhttp.ErrorMessage(canErr)})
		}
		if !decision.Allowed {
			return c.Status(403).JSON(map[string]string{"message": authhttp.DeniedMessage(decision)})
		}
		c.SetUserContext(authx.WithPrincipal(c.UserContext(), check.Principal))
		return c.Next()
	}
}

func buildHTTPServer(app *fiberapp.App, endpoints []httpx.Endpoint, log *slog.Logger) httpx.ServerRuntime {
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
	for _, endpoint := range endpoints {
		server.RegisterOnly(endpoint)
	}
	return server
}

var Module = dix.NewModule("http",
	dix.WithModuleImports(config2.Module, api.Module, auth2.Module),
	dix.WithModuleProviders(
		httpendpoint.SliceProvider(),
		dix.Provider1(func(cfg config2.AppConfig) AuthzPolicyConfig { return authzPolicyFromConfig(cfg) }),
		dix.Provider4(func(endpoints []httpx.Endpoint, log *slog.Logger, engine *authx.Engine, policy AuthzPolicyConfig) httpx.ServerRuntime {
			app := newFiberApp()
			setupFiberMiddleware(app, engine, policy)
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
