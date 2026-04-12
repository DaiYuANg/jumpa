package http

import (
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
	"github.com/samber/lo"
)

var Module = dix.NewModule("bastion-http",
	dix.WithModuleImports(bastion.Module),
	dix.WithModuleProviders(
		dix.Provider1(func(overviewSvc application.OverviewService) *OverviewEndpoint {
			return NewOverviewEndpoint(overviewSvc)
		}),
		dix.Provider1(func(assetSvc application.AssetService) *AssetEndpoint {
			return NewAssetEndpoint(assetSvc)
		}),
		dix.Provider2(func(
			policySvc application.PolicyService,
			requestSvc application.AccessRequestService,
		) *AccessEndpoint {
			return NewAccessEndpoint(policySvc, requestSvc)
		}),
		dix.Provider1(func(sessionSvc application.SessionService) *SessionEndpoint {
			return NewSessionEndpoint(sessionSvc)
		}),
		dix.Provider4(func(
			overview *OverviewEndpoint,
			asset *AssetEndpoint,
			access *AccessEndpoint,
			session *SessionEndpoint,
		) Endpoints {
			return Endpoints(lo.Map([]httpx.Endpoint{overview, asset, access, session}, func(it httpx.Endpoint, _ int) httpx.Endpoint {
				return it
			}))
		}),
	),
)
