package config

import (
	"fmt"
	"log/slog"

	"github.com/arcgolabs/configx"
	"github.com/arcgolabs/dix"
	"github.com/arcgolabs/pkg/randomport"
)

var Module = dix.NewModule("config",
	dix.WithModuleProviders(
		dix.Provider1(func(log *slog.Logger) AppConfig {
			var cfg AppConfig
			loader := configx.New(
				configx.WithTypedDefaults(DefaultAppConfig()),
				configx.WithDotenv(".env"),
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
			name := cfg.App.Name
			if name == "" {
				name = "jumpa"
			}
			log.Info("service starting",
				slog.String("name", name),
				slog.String("address", addr),
				slog.String("docs", "http://localhost"+addr+"/docs"),
				slog.String("openapi", "http://localhost"+addr+"/openapi.json"),
			)
		}),
	),
)
