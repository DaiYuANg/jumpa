package http

import (
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
)

type SessionEndpoint struct {
	httpx.BaseEndpoint
	sessionSvc application.SessionService
}

func NewSessionEndpoint(sessionSvc application.SessionService) *SessionEndpoint {
	return &SessionEndpoint{sessionSvc: sessionSvc}
}

func (e *SessionEndpoint) RegisterRoutes(server httpx.ServerRuntime) {
	registerSessionRoutes(server.Group("/api"), e.sessionSvc)
}
