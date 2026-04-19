package api

import (
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/kvx"
	apiendpoints "github.com/DaiYuANg/jumpa/internal/api/endpoints"
	"github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/httpendpoint"
	"github.com/DaiYuANg/jumpa/internal/kv"
	bastionhttp "github.com/DaiYuANg/jumpa/internal/modules/bastion/interfaces/http"
	gatewayregistryhttp "github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/interfaces/http"
	iamhttp "github.com/DaiYuANg/jumpa/internal/modules/iam/interfaces/http"
)

var Module = dix.NewModule("api",
	dix.WithModuleImports(iamhttp.Module, bastionhttp.Module, gatewayregistryhttp.Module, kv.Module, config.Module),
	dix.WithModuleProviders(
		httpendpoint.Provider0("api.system", apiendpoints.NewSystemEndpoint),
		httpendpoint.Provider2("api.auth", func(cfg config.AppConfig, kvClient kvx.Client) *apiendpoints.AuthEndpoint {
			return apiendpoints.NewAuthEndpoint(apiendpoints.AuthConfig{
				Secret:         cfg.JWT.Secret,
				Issuer:         cfg.JWT.Issuer,
				AccessTTLMin:   cfg.JWT.AccessTTLMin,
				RefreshTTLHour: cfg.JWT.RefreshTTLHour,
				UseValkey:      cfg.Valkey.Enabled,
				RevokedPrefix:  "auth:revoked",
			}, kvClient)
		}),
		httpendpoint.Provider0("api.dashboard", apiendpoints.NewDashboardEndpoint),
	),
)
