package main

import (
	"log/slog"
	"os"

	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/logx"
	"github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/gateway"
	"github.com/DaiYuANg/jumpa/internal/identity"
)

func Run() {
	logger := logx.MustNew(logx.WithConsole(true), logx.WithTraceLevel())
	defer func() { _ = logx.Close(logger) }()

	a := dix.New(
		"jumpa-gateway",
		dix.WithVersion("0.1.0"),
		dix.WithLogger(logger),
		dix.WithModules(
			config.Module,
			identity.Module,
			gateway.Module,
		),
	)

	if err := a.Run(); err != nil {
		logger.Error("jumpa gateway exited", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
