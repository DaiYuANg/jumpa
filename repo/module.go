package repo

import (
	"github.com/DaiYuANg/arcgo-rbac-template/db"
	"github.com/DaiYuANg/arcgo-rbac-template/schema"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dix"
)

var Module = dix.NewModule("repo",
	dix.WithModuleImports(db.Module),
	dix.WithModuleProviders(
		dix.Provider2(func(database *dbx.DB, s schema.UserSchema) UserRepository {
			return NewUserRepository(database, s)
		}),
	),
)
