package session

import (
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
)

var Module = dix.NewModule("bastion-session",
	dix.WithModuleProviders(
		dix.Provider1(func(sessionRepo ports.SessionRepository) application.SessionService {
			return application.NewSessionService(sessionRepo)
		}),
		dix.Provider2(func(sessionRepo ports.SessionRepository, eventRepo ports.SessionEventRepository) application.SessionRuntimeService {
			return application.NewSessionRuntimeService(sessionRepo, eventRepo)
		}),
	),
)
