package wire

import (
	"github.com/DaiYuANg/arcgo-rbac-template/internal/db"
	dbxrepo "github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/infrastructure/persistence/dbx"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/ports"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/schema"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/dbx"
)

var Module = dix.NewModule("iam.infrastructure.persistence.wire",
	dix.WithModuleImports(db.Module),
	dix.WithModuleProviders(
		dix.Provider1(func(database *dbx.DB) ports.UnitOfWork {
			return dbxrepo.NewUnitOfWork(database)
		}),

		dix.Provider1(func(database *dbx.DB) ports.RoleRepository {
			return dbxrepo.NewRoleRepository(database)
		}),
		dix.Provider1(func(database *dbx.DB) ports.RolePermissionGroupRepository {
			return dbxrepo.NewRolePermissionGroupRepository(database)
		}),
		dix.Provider1(func(database *dbx.DB) ports.PermissionGroupRepository {
			return dbxrepo.NewPermissionGroupRepository(database)
		}),
		dix.Provider1(func(database *dbx.DB) ports.PermissionRepository {
			return dbxrepo.NewPermissionRepository(database)
		}),

		dix.Provider2(func(database *dbx.DB, s schema.UserSchema) ports.UserRepository {
			return dbxrepo.NewUserRepository(database, s)
		}),
		dix.Provider1(func(database *dbx.DB) ports.UserRoleRepository {
			return dbxrepo.NewUserRoleRepository(database)
		}),
		dix.Provider1(func(database *dbx.DB) ports.AuthPrincipalRepository {
			return dbxrepo.NewAuthPrincipalRepository(database)
		}),
	),
)

