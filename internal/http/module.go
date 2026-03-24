package http

import (
	"context"
	"log/slog"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/api"
	config2 "github.com/DaiYuANg/arcgo-rbac-template/internal/config"
	service2 "github.com/DaiYuANg/arcgo-rbac-template/internal/service"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/DaiYuANg/arcgo/httpx/adapter/fiber"
	"github.com/go-playground/validator/v10"
	fiberapp "github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	fiberrequestid "github.com/gofiber/fiber/v2/middleware/requestid"
)

var Module = dix.NewModule("http",
	dix.WithModuleImports(config2.Module, service2.Module),
	dix.WithModuleProviders(
		dix.Provider3(func(cfg config2.AppConfig, svc service2.UserService, log *slog.Logger) httpx.ServerRuntime {
			app := fiberapp.New()
			app.Use(fiberlogger.New(), fiberrecover.New(), fiberrequestid.New())
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
			api.RegisterRoutes(server, svc)
			return server
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
