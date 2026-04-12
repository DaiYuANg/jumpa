package gateway

import (
	"context"
	"log/slog"

	"github.com/DaiYuANg/arcgo/dix"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/identity"
	"github.com/DaiYuANg/jumpa/internal/modules/audit"
	auditapp "github.com/DaiYuANg/jumpa/internal/modules/audit/application"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
	"github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry"
	registryapp "github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/application"
)

type serviceDeps struct {
	Authenticator identity.Authenticator
	TargetSvc     application.TargetService
	AccessSvc     application.AccessService
	SessionSvc    application.SessionRuntimeService
	AuditSvc      auditapp.SessionEventService
	RegistrySvc   registryapp.GatewayService
}

var Module = dix.NewModule("gateway",
	dix.WithModuleImports(config2.Module, identity.Module, bastion.Module, audit.Module, gatewayregistry.Module),
	dix.WithModuleProviders(
		dix.Provider6(func(authenticator identity.Authenticator, targetSvc application.TargetService, accessSvc application.AccessService, sessionSvc application.SessionRuntimeService, auditSvc auditapp.SessionEventService, registrySvc registryapp.GatewayService) serviceDeps {
			return serviceDeps{
				Authenticator: authenticator,
				TargetSvc:     targetSvc,
				AccessSvc:     accessSvc,
				SessionSvc:    sessionSvc,
				AuditSvc:      auditSvc,
				RegistrySvc:   registrySvc,
			}
		}),
		dix.Provider3(func(cfg config2.AppConfig, log *slog.Logger, deps serviceDeps) *Service {
			return NewService(cfg, log, deps.Authenticator, deps.TargetSvc, deps.AccessSvc, deps.SessionSvc, deps.AuditSvc, deps.RegistrySvc)
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
