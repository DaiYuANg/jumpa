package kv

import (
	"context"
	"log/slog"

	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/kvx"
	adaptervalkey "github.com/DaiYuANg/arcgo/kvx/adapter/valkey"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
)

func toKVXClientOptions(cfg config2.AppConfig) kvx.ClientOptions {
	return kvx.ClientOptions{
		Addrs:    []string{cfg.Valkey.Addr},
		Password: cfg.Valkey.Password,
		DB:       cfg.Valkey.DB,
		UseTLS:   cfg.Valkey.UseTLS,
	}
}

var Module = dix.NewModule("kv",
	dix.WithModuleImports(config2.Module),
	dix.WithModuleProviders(
		dix.Provider2(func(cfg config2.AppConfig, log *slog.Logger) kvx.Client {
			if !cfg.Valkey.Enabled {
				log.Info("valkey disabled; using noop kv client")
				return newNoopClient()
			}

			opts := toKVXClientOptions(cfg)
			client, err := adaptervalkey.New(opts)
			if err != nil {
				panic(err)
			}
			if _, err := client.Exists(context.Background(), "health:ping"); err != nil {
				panic(err)
			}
			log.Info("valkey connected", slog.String("addr", opts.Addrs[0]), slog.Int("db", opts.DB))
			return client
		}),
	),
	dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
		client, _ := dix.ResolveAs[kvx.Client](c)
		lc.OnStop(func(ctx context.Context) error {
			return client.Close()
		})
		return nil
	}),
)
