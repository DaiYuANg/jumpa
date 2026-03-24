package event

import (
	"context"
	"log/slog"

	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/eventx"
)

var Module = dix.NewModule("event",
		dix.WithModuleProviders(
			dix.Provider0(func() eventx.BusRuntime {
				return eventx.New(
					eventx.WithAntsPool(4),
					eventx.WithParallelDispatch(true),
				)
			}),
		),
		dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
			bus, _ := dix.ResolveAs[eventx.BusRuntime](c)
			lc.OnStop(func(ctx context.Context) error { return bus.Close() })
			return nil
		}),
		dix.WithModuleInvokes(
			dix.Invoke2(func(bus eventx.BusRuntime, log *slog.Logger) {
				_, _ = eventx.Subscribe[UserCreatedEvent](bus, func(ctx context.Context, e UserCreatedEvent) error {
					log.Info("user created (event)",
						slog.Int64("user_id", e.UserID),
						slog.String("name", e.UserName),
						slog.String("email", e.Email),
					)
					return nil
				})
			}),
		),
)
