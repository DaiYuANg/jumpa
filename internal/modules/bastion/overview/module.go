package overview

import (
	"github.com/arcgolabs/dix"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/identity"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
)

var Module = dix.NewModule("bastion-overview",
	dix.WithModuleProviders(
		dix.Provider3(func(cfg config2.AppConfig, provider identity.ProviderDescriptor, authenticator identity.Authenticator) application.OverviewService {
			return application.NewOverviewService(cfg, provider, authenticator)
		}),
	),
)
