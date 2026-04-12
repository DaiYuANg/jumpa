package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/logx"
	"github.com/DaiYuANg/jumpa/internal/cli"
)

func Run() {
	logger := logx.MustNew(logx.WithConsole(true), logx.WithTraceLevel())
	defer func() { _ = logx.Close(logger) }()

	a := dix.New(
		"jumpa-cli",
		dix.WithVersion("0.1.0"),
		dix.WithLogger(logger),
		dix.WithModules(cli.Module),
	)

	rt, err := a.Build()
	if err != nil {
		logger.Error("jumpa cli build failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	runner, err := dix.ResolveAs[*cli.Runner](rt.Container())
	if err != nil {
		logger.Error("jumpa cli resolve failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := runner.Run(context.Background()); err != nil {
		logger.Error("jumpa cli exited", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
