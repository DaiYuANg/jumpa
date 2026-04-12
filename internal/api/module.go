package api

import (
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/kvx"
	apiendpoints "github.com/DaiYuANg/jumpa/internal/api/endpoints"
	"github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/kv"
	bastionhttp "github.com/DaiYuANg/jumpa/internal/modules/bastion/interfaces/http"
	gatewayregistryhttp "github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/interfaces/http"
	iamhttp "github.com/DaiYuANg/jumpa/internal/modules/iam/interfaces/http"
	"github.com/samber/lo"
)

type staticEndpoints []httpx.Endpoint

var Module = dix.NewModule("api",
	dix.WithModuleImports(iamhttp.Module, bastionhttp.Module, gatewayregistryhttp.Module, kv.Module, config.Module),
	dix.WithModuleProviders(
		dix.Provider0(func() *apiendpoints.SystemEndpoint { return apiendpoints.NewSystemEndpoint() }),
		dix.Provider2(func(cfg config.AppConfig, kvClient kvx.Client) *apiendpoints.AuthEndpoint {
			return apiendpoints.NewAuthEndpoint(apiendpoints.AuthConfig{
				Secret:         cfg.JWT.Secret,
				Issuer:         cfg.JWT.Issuer,
				AccessTTLMin:   cfg.JWT.AccessTTLMin,
				RefreshTTLHour: cfg.JWT.RefreshTTLHour,
				UseValkey:      cfg.Valkey.Enabled,
				RevokedPrefix:  "auth:revoked",
			}, kvClient)
		}),
		dix.Provider0(func() *apiendpoints.DashboardEndpoint { return apiendpoints.NewDashboardEndpoint() }),
		dix.Provider5(func(
			system *apiendpoints.SystemEndpoint,
			auth *apiendpoints.AuthEndpoint,
			dashboard *apiendpoints.DashboardEndpoint,
			user *iamhttp.UserEndpoint,
			rbac *iamhttp.RBACEndpoint,
		) staticEndpoints {
			return staticEndpoints([]httpx.Endpoint{system, auth, dashboard, user, rbac})
		}),
		dix.Provider3(func(
			static staticEndpoints,
			bastion bastionhttp.Endpoints,
			gateway *gatewayregistryhttp.GatewayEndpoint,
		) []httpx.Endpoint {
			return lo.Flatten([][]httpx.Endpoint{
				static,
				bastion,
				{gateway},
			})
		}),
	),
)
