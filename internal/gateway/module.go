package gateway

import (
	"context"
	"log/slog"

	"github.com/DaiYuANg/arcgo/dix"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/identity"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
)

var Module = dix.NewModule("gateway",
	dix.WithModuleImports(config2.Module, identity.Module, bastion.Module),
	dix.WithModuleProviders(
		dix.Provider6(func(cfg config2.AppConfig, log *slog.Logger, authenticator identity.Authenticator, targetSvc application.TargetService, accessSvc application.AccessService, sessionSvc application.SessionRuntimeService) *Service {
			return NewService(cfg, log, authenticator, targetSvc, accessSvc, sessionSvc)
		}),
	),
	dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
		svc, err := dix.ResolveAs[*Service](c)
		if err != nil {
			return err
		}

		lc.OnStart(func(ctx context.Context) error { return svc.Start(ctx) })
		lc.OnStop(func(ctx context.Context) error { return svc.Shutdown(ctx) })
		return nil
	}),
)
