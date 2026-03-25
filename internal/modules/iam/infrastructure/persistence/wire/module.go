package wire

import (
	"github.com/DaiYuANg/arcgo-rbac-template/internal/db"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/infrastructure/persistence"
	dbxrepo "github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/infrastructure/persistence/dbx"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/schema"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/dbx"
)

var Module = dix.NewModule("iam.infrastructure.persistence.wire",
	dix.WithModuleImports(db.Module),
	dix.WithModuleProviders(
		dix.Provider1(func(database *dbx.DB) persistence.UnitOfWork {
			return dbxrepo.NewUnitOfWork(database)
		}),

		dix.Provider1(func(database *dbx.DB) persistence.RoleRepository {
			return dbxrepo.NewRoleRepository(database)
		}),
		dix.Provider1(func(database *dbx.DB) persistence.RolePermissionGroupRepository {
			return dbxrepo.NewRolePermissionGroupRepository(database)
		}),
		dix.Provider1(func(database *dbx.DB) persistence.PermissionGroupRepository {
			return dbxrepo.NewPermissionGroupRepository(database)
		}),
		dix.Provider1(func(database *dbx.DB) persistence.PermissionRepository {
			return dbxrepo.NewPermissionRepository(database)
		}),

		dix.Provider2(func(database *dbx.DB, s schema.UserSchema) persistence.UserRepository {
			return dbxrepo.NewUserRepository(database, s)
		}),
		dix.Provider1(func(database *dbx.DB) persistence.UserRoleRepository {
			return dbxrepo.NewUserRoleRepository(database)
		}),
		dix.Provider1(func(database *dbx.DB) persistence.AuthPrincipalRepository {
			return dbxrepo.NewAuthPrincipalRepository(database)
		}),
	),
)

