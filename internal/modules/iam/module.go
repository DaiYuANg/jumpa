package iam

import (
	"log/slog"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/event"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/application"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/infrastructure/persistence"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/eventx"
)

var Module = dix.NewModule("iam",
	dix.WithModuleImports(persistence.Module, event.Module),
	dix.WithModuleProviders(
		dix.Provider3(func(r persistence.UserRepository, bus eventx.BusRuntime, log *slog.Logger) application.UserService {
			return application.NewUserService(r, bus, log)
		}),
		dix.Provider1(func(r persistence.RoleRepository) application.RoleService { return application.NewRoleService(r) }),
		dix.Provider1(func(r persistence.PermissionGroupRepository) application.PermissionGroupService {
			return application.NewPermissionGroupService(r)
		}),
		dix.Provider1(func(r persistence.PermissionRepository) application.PermissionService {
			return application.NewPermissionService(r)
		}),
		dix.Provider1(func(r persistence.UserRoleRepository) application.UserRoleService {
			return application.NewUserRoleService(r)
		}),
		dix.Provider1(func(r persistence.AuthPrincipalRepository) application.AuthPrincipalService {
			return application.NewAuthPrincipalService(r)
		}),
	),
)
