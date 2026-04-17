package main

import (
	"log/slog"

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

const (
	serverRuntimeName    = "jumpa"
	serverRuntimeVersion = "0.1.0"
)

func runServer() error {
	logger := newServerLogger()
	defer func() { _ = logx.Close(logger) }()

	app := dix.New(
		serverRuntimeName,
		dix.WithVersion(serverRuntimeVersion),
		dix.WithLogger(logger),
		dix.WithModules(serverModules()...),
	)

	if err := app.Run(); err != nil {
		logger.Error("jumpa exited", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func newServerLogger() *slog.Logger {
	return logx.MustNew(logx.WithConsole(true), logx.WithTraceLevel())
}

func serverModules() []dix.Module {
	return []dix.Module{
		config.Module,
		event.Module,
		db.Module,
		kv.Module,
		iam.Module,
		scheduler.Module,
		http.Module,
	}
}
