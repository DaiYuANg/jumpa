package service

import (
	"log/slog"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/event"
	repo2 "github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/eventx"
)

var Module = dix.NewModule("service",
	dix.WithModuleImports(repo2.Module, event.Module),
	dix.WithModuleProviders(
		dix.Provider3(func(r repo2.UserRepository, bus eventx.BusRuntime, log *slog.Logger) UserService {
			return NewUserService(r, bus, log)
		}),
		dix.Provider1(func(r repo2.RoleRepository) RoleService { return NewRoleService(r) }),
		dix.Provider1(func(r repo2.PermissionGroupRepository) PermissionGroupService { return NewPermissionGroupService(r) }),
		dix.Provider1(func(r repo2.PermissionRepository) PermissionService { return NewPermissionService(r) }),
		dix.Provider1(func(r repo2.UserRoleRepository) UserRoleService { return NewUserRoleService(r) }),
		dix.Provider1(func(r repo2.AuthPrincipalRepository) AuthPrincipalService { return NewAuthPrincipalService(r) }),
	),
)
