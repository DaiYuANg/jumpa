package api

import (
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/kvx"
	apiendpoints "github.com/DaiYuANg/jumpa/internal/api/endpoints"
	"github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/kv"
	bastionhttp "github.com/DaiYuANg/jumpa/internal/modules/bastion/interfaces/http"
	iamhttp "github.com/DaiYuANg/jumpa/internal/modules/iam/interfaces/http"
)

var Module = dix.NewModule("api",
	dix.WithModuleImports(iamhttp.Module, bastionhttp.Module, kv.Module, config.Module),
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
		dix.Provider6(func(
			system *apiendpoints.SystemEndpoint,
			auth *apiendpoints.AuthEndpoint,
			dashboard *apiendpoints.DashboardEndpoint,
			user *iamhttp.UserEndpoint,
			rbac *iamhttp.RBACEndpoint,
			bastion *bastionhttp.BastionEndpoint,
		) []httpx.Endpoint {
			return []httpx.Endpoint{system, auth, dashboard, user, rbac, bastion}
		}),
	),
)
