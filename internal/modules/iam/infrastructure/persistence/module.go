package persistence

import (
	"github.com/DaiYuANg/arcgo-rbac-template/internal/db"
	dbxrepo "github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/infrastructure/persistence/dbx"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/schema"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dix"
)

var Module = dix.NewModule("iam-persistence",
	dix.WithModuleImports(db.Module),
	dix.WithModuleProviders(
		dix.Provider2(func(database *dbx.DB, s schema.UserSchema) UserRepository {
			return dbxrepo.NewUserRepository(database, s)
		}),
		dix.Provider1(func(database *dbx.DB) RoleRepository { return dbxrepo.NewRoleRepository(database) }),
		dix.Provider1(func(database *dbx.DB) RolePermissionGroupRepository {
			return dbxrepo.NewRolePermissionGroupRepository(database)
		}),
		dix.Provider1(func(database *dbx.DB) PermissionGroupRepository { return dbxrepo.NewPermissionGroupRepository(database) }),
		dix.Provider1(func(database *dbx.DB) PermissionRepository { return dbxrepo.NewPermissionRepository(database) }),
		dix.Provider1(func(database *dbx.DB) UserRoleRepository { return dbxrepo.NewUserRoleRepository(database) }),
		dix.Provider1(func(database *dbx.DB) AuthPrincipalRepository { return dbxrepo.NewAuthPrincipalRepository(database) }),
	),
)
