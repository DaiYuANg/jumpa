package iam

import (
	"log/slog"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/event"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/application"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/infrastructure/persistence/wire"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/ports"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/eventx"
)

var Module = dix.NewModule("iam",
	dix.WithModuleImports(wire.Module, event.Module),
	dix.WithModuleProviders(
		dix.Provider3(func(r ports.UserRepository, bus eventx.BusRuntime, log *slog.Logger) application.UserService {
			return application.NewUserService(r, bus, log)
		}),
		dix.Provider3(func(uow ports.UnitOfWork, r ports.RoleRepository, rpg ports.RolePermissionGroupRepository) application.RoleService {
			return application.NewRoleService(uow, r, rpg)
		}),
		dix.Provider1(func(r ports.PermissionGroupRepository) application.PermissionGroupService {
			return application.NewPermissionGroupService(r)
		}),
		dix.Provider1(func(r ports.PermissionRepository) application.PermissionService {
			return application.NewPermissionService(r)
		}),
		dix.Provider1(func(r ports.UserRoleRepository) application.UserRoleService {
			return application.NewUserRoleService(r)
		}),
		dix.Provider1(func(r ports.AuthPrincipalRepository) application.AuthPrincipalService {
			return application.NewAuthPrincipalService(r)
		}),
	),
)
