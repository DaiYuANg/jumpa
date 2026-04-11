package http

import (
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
)

var Module = dix.NewModule("bastion-http",
	dix.WithModuleImports(bastion.Module),
	dix.WithModuleProviders(
		dix.Provider5(func(
			overviewSvc application.OverviewService,
			assetSvc application.AssetService,
			policySvc application.PolicyService,
			requestSvc application.AccessRequestService,
			sessionSvc application.SessionService,
		) *BastionEndpoint {
			return NewBastionEndpoint(overviewSvc, assetSvc, policySvc, requestSvc, sessionSvc)
		}),
	),
)
