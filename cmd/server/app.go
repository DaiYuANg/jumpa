package main

import (
	"log/slog"
	"os"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/config"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/db"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/event"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/http"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/kv"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/scheduler"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/service"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/logx"
)

func Run() {
	logger := logx.MustNew(logx.WithConsole(true), logx.WithTraceLevel())
	defer func() { _ = logx.Close(logger) }()

	a := dix.New(
		"backend",
		dix.WithVersion("0.1.0"),
		dix.WithLogger(logger),
		dix.WithModules(
			config.Module,
			event.Module,
			db.Module,
			kv.Module,
			repo.Module,
			service.Module,
			scheduler.Module,
			http.Module,
		),
	)

	if err := a.Run(); err != nil {
		logger.Error("backend exited", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
