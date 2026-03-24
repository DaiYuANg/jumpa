package app

import (
	"log/slog"
	"os"

	"github.com/DaiYuANg/arcgo-rbac-template/config"
	"github.com/DaiYuANg/arcgo-rbac-template/db"
	"github.com/DaiYuANg/arcgo-rbac-template/event"
	"github.com/DaiYuANg/arcgo-rbac-template/http"
	"github.com/DaiYuANg/arcgo-rbac-template/repo"
	"github.com/DaiYuANg/arcgo-rbac-template/service"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/logx"
)

func Run() {
	logger := logx.MustNew(logx.WithConsole(true), logx.WithDebugLevel())
	defer func() { _ = logx.Close(logger) }()

	a := dix.New(
		"backend",
		dix.WithVersion("0.1.0"),
		dix.WithLogger(logger),
		dix.WithModules(
			config.Module,
			event.Module,
			db.Module,
			repo.Module,
			service.Module,
			http.Module,
		),
	)

	if err := a.Run(); err != nil {
		logger.Error("backend exited", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
