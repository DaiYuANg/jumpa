package http

import (
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/jumpa/internal/httpendpoint"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion"
)

var Module = dix.NewModule("bastion-http",
	dix.WithModuleImports(bastion.Module),
	dix.WithModuleProviders(
		httpendpoint.Provider1("bastion.overview", NewOverviewEndpoint),
		httpendpoint.Provider1("bastion.asset", NewAssetEndpoint),
		httpendpoint.Provider2("bastion.access", NewAccessEndpoint),
		httpendpoint.Provider1("bastion.session", NewSessionEndpoint),
	),
)
