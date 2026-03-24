package service

import (
	"log/slog"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/event"
	repo2 "github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/eventx"
)

var Module = dix.NewModule("service",
	dix.WithModuleImports(repo2.Module, event.Module),
	dix.WithModuleProviders(
		dix.Provider3(func(r repo2.UserRepository, bus eventx.BusRuntime, log *slog.Logger) UserService {
			return NewUserService(r, bus, log)
		}),
	),
)
