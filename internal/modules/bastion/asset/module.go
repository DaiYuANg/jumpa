package asset

import (
	"github.com/arcgolabs/dix"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
)

var Module = dix.NewModule("bastion-asset",
	dix.WithModuleProviders(
		dix.Provider2(func(hostRepo ports.HostRepository, hostAccountRepo ports.HostAccountRepository) application.AssetService {
			return application.NewAssetService(hostRepo, hostAccountRepo)
		}),
		dix.Provider2(func(hostRepo ports.HostRepository, hostAccountRepo ports.HostAccountRepository) application.TargetService {
			return application.NewTargetService(hostRepo, hostAccountRepo)
		}),
	),
)
