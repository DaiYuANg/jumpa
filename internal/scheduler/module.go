package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/arcgolabs/dix"
	"github.com/arcgolabs/kvx"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
	kv2 "github.com/DaiYuANg/jumpa/internal/kv"
	"github.com/go-co-op/gocron/v2"
)

var Module = dix.NewModule("scheduler",
	dix.WithModuleImports(config2.Module, kv2.Module),
	dix.WithModuleProviders(
		dix.Provider3(func(cfg config2.AppConfig, kvClient kvx.Client, log *slog.Logger) gocron.Scheduler {
			schedulerOptions := make([]gocron.SchedulerOption, 0, 1)
			if cfg.Scheduler.Distributed.Enabled {
				ttlSec := cfg.Scheduler.Distributed.TTLSec
				if ttlSec <= 0 {
					ttlSec = 30
				}
				keyPrefix := cfg.Scheduler.Distributed.KeyPrefix
				if keyPrefix == "" {
					keyPrefix = "gocron:lock"
				}
				locker := newDistributedLocker(kvClient, keyPrefix, time.Duration(ttlSec)*time.Second)
				schedulerOptions = append(schedulerOptions, gocron.WithDistributedLocker(locker))
			}

			s, err := gocron.NewScheduler(schedulerOptions...)
			if err != nil {
				panic(err)
			}

			if cfg.Scheduler.Enabled {
				interval := cfg.Scheduler.HeartbeatSec
				if interval <= 0 {
					interval = 60
				}
				_, err = s.NewJob(
					gocron.DurationJob(time.Duration(interval)*time.Second),
					gocron.NewTask(func() {
						log.Info("scheduler heartbeat tick", slog.Int("interval_sec", interval))
					}),
				)
				if err != nil {
					panic(err)
				}
			}

			if cfg.Scheduler.Distributed.Enabled {
				log.Info("scheduler distributed locker enabled",
					slog.String("key_prefix", cfg.Scheduler.Distributed.KeyPrefix),
					slog.Int("ttl_sec", cfg.Scheduler.Distributed.TTLSec),
				)
			}

			return s
		}),
	),
	dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
		s := dix.MustResolveAs[gocron.Scheduler](c)

		cfg := dix.MustResolveAs[config2.AppConfig](c)

		lc.OnStart(func(ctx context.Context) error {
			if cfg.Scheduler.Enabled {
				s.Start()
			}
			return nil
		})
		lc.OnStop(func(ctx context.Context) error {
			return s.Shutdown()
		})
		return nil
	}),
)
