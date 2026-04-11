package main

import (
	"log/slog"
	"os"

	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/logx"
	"github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/db"
	"github.com/DaiYuANg/jumpa/internal/event"
	"github.com/DaiYuANg/jumpa/internal/http"
	"github.com/DaiYuANg/jumpa/internal/kv"
	"github.com/DaiYuANg/jumpa/internal/modules/iam"
	"github.com/DaiYuANg/jumpa/internal/scheduler"
)

func Run() {
	logger := logx.MustNew(logx.WithConsole(true), logx.WithTraceLevel())
	defer func() { _ = logx.Close(logger) }()

	a := dix.New(
		"jumpa",
		dix.WithVersion("0.1.0"),
		dix.WithLogger(logger),
		dix.WithModules(
			config.Module,
			event.Module,
			db.Module,
			kv.Module,
			iam.Module,
			scheduler.Module,
			http.Module,
		),
	)

	if err := a.Run(); err != nil {
		logger.Error("jumpa exited", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
