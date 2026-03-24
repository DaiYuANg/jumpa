package config

import (
	"fmt"
	"log/slog"

	"github.com/DaiYuANg/arcgo/configx"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/pkg/randomport"
)

var Module = dix.NewModule("config",
	dix.WithModuleProviders(
		dix.Provider1(func(log *slog.Logger) AppConfig {
			var cfg AppConfig
			loader := configx.New(
				configx.WithDefaults(map[string]any{
					"server.port":             8080,
					"db.driver":               "sqlite",
					"db.dsn":                  "file:backend?mode=memory&cache=shared",
					"scheduler.enabled":       true,
					"scheduler.heartbeat_sec": 60,
					"scheduler.distributed.enabled":    false,
					"scheduler.distributed.key_prefix": "gocron:lock",
					"scheduler.distributed.ttl_sec":    30,
					"valkey.enabled":          false,
					"valkey.addr":             "127.0.0.1:6379",
					"valkey.password":         "",
					"valkey.db":               0,
					"valkey.use_tls":          false,
				}),
				configx.WithEnvPrefix("APP_"),
			)
			if err := loader.Load(&cfg); err != nil {
				log.Error("config load failed", slog.String("error", err.Error()))
				panic(err)
			}
			if cfg.Server.Port == 0 {
				cfg.Server.Port = randomport.MustFind()
			}
			return cfg
		}),
	),
	dix.WithModuleInvokes(
		dix.Invoke2(func(cfg AppConfig, log *slog.Logger) {
			addr := fmt.Sprintf(":%d", cfg.Server.Port)
			log.Info("backend starting",
				slog.String("address", addr),
				slog.String("docs", "http://localhost"+addr+"/docs"),
				slog.String("openapi", "http://localhost"+addr+"/openapi.json"),
			)
		}),
	),
)
