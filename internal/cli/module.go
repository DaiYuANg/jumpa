package cli

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/jumpa/internal/cli/api"
)

var Module = dix.NewModule("cli",
	dix.WithModuleProviders(
		dix.Provider1(func(log *slog.Logger) Config {
			return loadConfig(log)
		}),
		dix.Provider0(func() *http.Client {
			return &http.Client{Timeout: 15 * time.Second}
		}),
		dix.Provider2(func(cfg Config, httpClient *http.Client) *api.Client {
			return api.NewClient(cfg.APIURL, api.WithHTTPClient(httpClient))
		}),
		dix.Provider3(func(cfg Config, log *slog.Logger, client *api.Client) *Runner {
			return NewRunner(cfg, log, client)
		}),
	),
)
