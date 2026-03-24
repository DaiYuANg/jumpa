package api

import (
	apiendpoints "github.com/DaiYuANg/arcgo-rbac-template/internal/api/endpoints"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/config"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/kv"
	iamhttp "github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/interfaces/http"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/kvx"
)

var Module = dix.NewModule("api",
	dix.WithModuleImports(iamhttp.Module, kv.Module, config.Module),
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
		dix.Provider4(func(
			system *apiendpoints.SystemEndpoint,
			auth *apiendpoints.AuthEndpoint,
			dashboard *apiendpoints.DashboardEndpoint,
			iamEndpoints []httpx.Endpoint,
		) []httpx.Endpoint {
			return append([]httpx.Endpoint{system, auth, dashboard}, iamEndpoints...)
		}),
	),
)
