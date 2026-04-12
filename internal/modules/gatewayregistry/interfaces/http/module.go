package http

import (
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry"
	"github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/application"
)

var Module = dix.NewModule("gateway-registry-http",
	dix.WithModuleImports(gatewayregistry.Module),
	dix.WithModuleProviders(
		dix.Provider1(func(service application.GatewayService) *GatewayEndpoint {
			return NewGatewayEndpoint(service)
		}),
	),
)
