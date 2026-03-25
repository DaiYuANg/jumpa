package http

import (
	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/application"
	"github.com/DaiYuANg/arcgo/dix"
)

var Module = dix.NewModule("iam-http",
	dix.WithModuleImports(iam.Module),
	dix.WithModuleProviders(
		dix.Provider3(func(
			userSvc application.UserService,
			userRoleSvc application.UserRoleService,
			principalSvc application.AuthPrincipalService,
		) *UserEndpoint {
			return NewUserEndpoint(userSvc, userRoleSvc, principalSvc)
		}),
		dix.Provider3(func(
			roleSvc application.RoleService,
			groupSvc application.PermissionGroupService,
			permSvc application.PermissionService,
		) *RBACEndpoint {
			return NewRBACEndpoint(roleSvc, groupSvc, permSvc)
		}),
	),
)
