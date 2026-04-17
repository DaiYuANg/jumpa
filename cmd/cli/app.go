package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/DaiYuANg/arcgo/logx"
)

func Run() {
	logger := logx.MustNew(logx.WithConsole(true), logx.WithTraceLevel())
	defer func() { _ = logx.Close(logger) }()

	root := newRootCommand(logger)
	root.SetContext(context.Background())

	if err := root.Execute(); err != nil {
		logger.Error("jumpa cli exited", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
