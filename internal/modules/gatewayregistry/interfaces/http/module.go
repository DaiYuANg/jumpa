package http

import (
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/jumpa/internal/httpendpoint"
	"github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry"
)

var Module = dix.NewModule("gateway-registry-http",
	dix.WithModuleImports(gatewayregistry.Module),
	dix.WithModuleProviders(
		httpendpoint.Provider1("gateway-registry.gateway", NewGatewayEndpoint),
	),
)
