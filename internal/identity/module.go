package identity

import (
	"github.com/arcgolabs/dbx"
	"github.com/arcgolabs/dix"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
	db2 "github.com/DaiYuANg/jumpa/internal/db"
)

var Module = dix.NewModule("identity",
	dix.WithModuleImports(config2.Module, db2.Module),
	dix.WithModuleProviders(
		dix.Provider1(func(cfg config2.AppConfig) ProviderDescriptor {
			return CurrentProvider(cfg)
		}),
		dix.Provider1(func(cfg config2.AppConfig) osBackendConfig {
			return toOSBackendConfig(cfg)
		}),
		dix.Provider3(func(provider ProviderDescriptor, backendConfig osBackendConfig, database *dbx.DB) Authenticator {
			if provider.Kind == "os" {
				return NewOSAuthenticator(provider, backendConfig)
			}
			return NewLocalAuthenticator(provider, database)
		}),
	),
)
