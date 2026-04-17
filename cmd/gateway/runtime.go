package main

import (
	"log/slog"

	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/logx"
	"github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/gateway"
	"github.com/DaiYuANg/jumpa/internal/identity"
)

const (
	gatewayRuntimeName    = "jumpa-gateway"
	gatewayRuntimeVersion = "0.1.0"
)

func runGateway() error {
	logger := newGatewayLogger()
	defer func() { _ = logx.Close(logger) }()

	app := dix.New(
		gatewayRuntimeName,
		dix.WithVersion(gatewayRuntimeVersion),
		dix.WithLogger(logger),
		dix.WithModules(gatewayModules()...),
	)

	if err := app.Run(); err != nil {
		logger.Error("jumpa gateway exited", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func newGatewayLogger() *slog.Logger {
	return logx.MustNew(logx.WithConsole(true), logx.WithTraceLevel())
}

func gatewayModules() []dix.Module {
	return []dix.Module{
		config.Module,
		identity.Module,
		gateway.Module,
	}
}
