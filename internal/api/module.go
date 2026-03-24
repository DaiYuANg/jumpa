package api

import (
	apiendpoints "github.com/DaiYuANg/arcgo-rbac-template/internal/api/endpoints"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/config"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/kv"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/service"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/kvx"
)

var Module = dix.NewModule("api",
	dix.WithModuleImports(service.Module, kv.Module, config.Module),
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
		dix.Provider3(func(
			userSvc service.UserService,
			userRoleSvc service.UserRoleService,
			principalSvc service.AuthPrincipalService,
		) *apiendpoints.UserEndpoint {
			return apiendpoints.NewUserEndpoint(userSvc, userRoleSvc, principalSvc)
		}),
		dix.Provider3(func(
			roleSvc service.RoleService,
			groupSvc service.PermissionGroupService,
			permSvc service.PermissionService,
		) *apiendpoints.RBACEndpoint {
			return apiendpoints.NewRBACEndpoint(roleSvc, groupSvc, permSvc)
		}),
		dix.Provider5(func(
			system *apiendpoints.SystemEndpoint,
			auth *apiendpoints.AuthEndpoint,
			dashboard *apiendpoints.DashboardEndpoint,
			user *apiendpoints.UserEndpoint,
			rbac *apiendpoints.RBACEndpoint,
		) []httpx.Endpoint {
			return []httpx.Endpoint{system, auth, dashboard, user, rbac}
		}),
	),
)
