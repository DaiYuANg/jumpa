package bastion

import (
	"github.com/DaiYuANg/arcgo/dix"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/identity"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/infrastructure/persistence/wire"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
)

var Module = dix.NewModule("bastion",
	dix.WithModuleImports(config2.Module, identity.Module, wire.Module),
	dix.WithModuleProviders(
		dix.Provider3(func(cfg config2.AppConfig, provider identity.ProviderDescriptor, authenticator identity.Authenticator) application.OverviewService {
			return application.NewOverviewService(cfg, provider, authenticator)
		}),
		dix.Provider2(func(hostRepo ports.HostRepository, hostAccountRepo ports.HostAccountRepository) application.AssetService {
			return application.NewAssetService(hostRepo, hostAccountRepo)
		}),
		dix.Provider2(func(hostRepo ports.HostRepository, hostAccountRepo ports.HostAccountRepository) application.TargetService {
			return application.NewTargetService(hostRepo, hostAccountRepo)
		}),
		dix.Provider1(func(policyRepo ports.PolicyRepository) application.PolicyService {
			return application.NewPolicyService(policyRepo)
		}),
		dix.Provider3(func(policyRepo ports.PolicyRepository, principalRepo ports.PrincipalAccessRepository, accessRequestRepo ports.AccessRequestRepository) application.AccessService {
			return application.NewAccessService(policyRepo, principalRepo, accessRequestRepo)
		}),
		dix.Provider1(func(sessionRepo ports.SessionRepository) application.SessionService {
			return application.NewSessionService(sessionRepo)
		}),
		dix.Provider2(func(cfg config2.AppConfig, accessRequestRepo ports.AccessRequestRepository) application.AccessRequestService {
			return application.NewAccessRequestService(cfg, accessRequestRepo)
		}),
		dix.Provider2(func(sessionRepo ports.SessionRepository, eventRepo ports.SessionEventRepository) application.SessionRuntimeService {
			return application.NewSessionRuntimeService(sessionRepo, eventRepo)
		}),
	),
)
