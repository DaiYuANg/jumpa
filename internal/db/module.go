package db

import (
	"context"
	"log/slog"

	config2 "github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/schema"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dix"
)

var Module = dix.NewModule("db",
	dix.WithModuleImports(config2.Module),
	dix.WithModuleProviders(
		dix.Provider2(func(cfg config2.AppConfig, log *slog.Logger) *dbx.DB {
			opts := DefaultOpts(log)
			if cfg.DB.NodeID != 0 {
				opts = append(opts, dbx.WithNodeID(cfg.DB.NodeID))
			}
			database, err := Open(cfg.DB.Driver, cfg.DB.DSN, opts...)
			if err != nil {
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
		database, err := dix.ResolveAs[*dbx.DB](c)
		if err != nil {
			return err
		}
		lc.OnStop(func(ctx context.Context) error { return database.Close() })
		return nil
	}),
)
