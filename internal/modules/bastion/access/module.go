package access

import (
	"github.com/DaiYuANg/arcgo/dix"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
)

var Module = dix.NewModule("bastion-access",
	dix.WithModuleProviders(
		dix.Provider1(func(policyRepo ports.PolicyRepository) application.PolicyService {
			return application.NewPolicyService(policyRepo)
		}),
		dix.Provider3(func(policyRepo ports.PolicyRepository, principalRepo ports.PrincipalAccessRepository, accessRequestRepo ports.AccessRequestRepository) application.AccessService {
			return application.NewAccessService(policyRepo, principalRepo, accessRequestRepo)
		}),
		dix.Provider2(func(cfg config2.AppConfig, accessRequestRepo ports.AccessRequestRepository) application.AccessRequestService {
			return application.NewAccessRequestService(cfg, accessRequestRepo)
		}),
	),
)
