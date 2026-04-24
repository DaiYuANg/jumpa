package cli

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/arcgolabs/clientx"
	clienthttp "github.com/arcgolabs/clientx/http"
	"github.com/arcgolabs/dix"
	"github.com/DaiYuANg/jumpa/internal/cli/api"
)

type CommandRunner interface {
	Run(ctx context.Context) error
}

type stdio struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

func NewCommonModule(cfg Config) dix.Module {
	return dix.NewModule("cli-common",
		dix.WithModuleProviders(
			dix.Provider0(func() stdio {
				return stdio{
					In:  os.Stdin,
					Out: os.Stdout,
					Err: os.Stderr,
				}
			}),
			dix.Provider0(func() Config {
				return cfg
			}),
			dix.Provider2(func(cfg Config, log *slog.Logger) clienthttp.Client {
				client, err := clienthttp.New(clienthttp.Config{
					BaseURL:   cfg.APIURL,
					Timeout:   15 * time.Second,
					UserAgent: "jumpa-cli/0.1.0",
					Retry: clientx.RetryConfig{
						Enabled:    true,
						MaxRetries: 1,
						WaitMin:    200 * time.Millisecond,
						WaitMax:    1 * time.Second,
					},
				}, clienthttp.WithHooks(clientx.NewLoggingHook(log, clientx.WithLoggingHookAddress(true))))
				if err != nil {
					log.Error("clientx http client init failed", slog.String("error", err.Error()))
					panic(err)
				}
				return client
			}),
			dix.Provider2(func(cfg Config, httpClient clienthttp.Client) *api.Client {
				return api.NewClient(cfg.APIURL, httpClient)
			}),
			dix.Provider4(func(cfg Config, log *slog.Logger, client *api.Client, streams stdio) *SessionResolver {
				return NewSessionResolver(cfg, log, client, streams)
			}),
			dix.Provider3(func(cfg Config, log *slog.Logger, streams stdio) *SSHLauncher {
				return NewSSHLauncher(cfg, log, streams)
			}),
		),
		dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
			client, err := dix.ResolveAs[clienthttp.Client](c)
			if err != nil {
				return err
			}
			lc.OnStop(func(ctx context.Context) error { return client.Close() })
			return nil
		}),
	)
}
