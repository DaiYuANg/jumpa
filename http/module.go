package http

import (
	"context"
	"log/slog"

	"github.com/DaiYuANg/arcgo-rbac-template/api"
	"github.com/DaiYuANg/arcgo-rbac-template/config"
	"github.com/DaiYuANg/arcgo-rbac-template/service"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/DaiYuANg/arcgo/httpx/adapter/std"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

var Module = dix.NewModule("http",
	dix.WithModuleImports(config.Module, service.Module),
	dix.WithModuleProviders(
		dix.Provider3(func(cfg config.AppConfig, svc service.UserService, log *slog.Logger) httpx.ServerRuntime {
			router := chi.NewMux()
			router.Use(middleware.Logger, middleware.Recoverer, middleware.RequestID)
			ad := std.New(router, adapter.HumaOptions{
				Title:       "ArcGo Backend API",
				Version:     "1.0.0",
				Description: "configx + logx + eventx + httpx + dix + dbx",
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
			api.RegisterRoutes(server, svc)
			return server
		}),
	),
	dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
		server, _ := dix.ResolveAs[httpx.ServerRuntime](c)
		cfg, _ := dix.ResolveAs[config.AppConfig](c)
		p := cfg.Server.Port
		lc.OnStart(func(ctx context.Context) error {
			go func() { _ = server.ListenPort(p) }()
			return nil
		})
		lc.OnStop(func(ctx context.Context) error { return server.Shutdown() })
		return nil
	}),
)
