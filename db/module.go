package db

import (
	"context"
	"log/slog"

	"github.com/DaiYuANg/arcgo-rbac-template/config"
	"github.com/DaiYuANg/arcgo-rbac-template/schema"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dix"
)

var Module = dix.NewModule("db",
	dix.WithModuleImports(config.Module),
	dix.WithModuleProviders(
		dix.Provider2(func(cfg config.AppConfig, log *slog.Logger) *dbx.DB {
			database, err := Open(cfg.DB.Driver, cfg.DB.DSN, DefaultOpts(log)...)
			if err != nil {
				panic(err)
			}
			userSchema := schema.UserSchema{}
			users := dbx.MustSchema("users", userSchema)
			if _, err := database.AutoMigrate(context.Background(), users); err != nil {
				panic(err)
			}
			return database
		}),
		dix.Provider0(func() schema.UserSchema {
			s := schema.UserSchema{}
			return dbx.MustSchema("users", s)
		}),
	),
	dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
		database, _ := dix.ResolveAs[*dbx.DB](c)
		lc.OnStop(func(ctx context.Context) error { return database.Close() })
		return nil
	}),
)
