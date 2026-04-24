package gatewayregistry

import (
	"github.com/arcgolabs/dix"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/application"
	"github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/infrastructure/persistence/wire"
	"github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/ports"
)

var Module = dix.NewModule("gateway-registry",
	dix.WithModuleImports(config2.Module, wire.Module),
	dix.WithModuleProviders(
		dix.Provider2(func(cfg config2.AppConfig, repo ports.GatewayRepository) application.GatewayService {
			return application.NewGatewayService(cfg, repo)
		}),
	),
)
