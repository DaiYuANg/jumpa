package http

import (
	"github.com/arcgolabs/httpx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
)

type SessionEndpoint struct {
	sessionSvc application.SessionService
}

func NewSessionEndpoint(sessionSvc application.SessionService) *SessionEndpoint {
	return &SessionEndpoint{sessionSvc: sessionSvc}
}

func (e *SessionEndpoint) EndpointSpec() httpx.EndpointSpec {
	return httpx.EndpointSpec{Prefix: "/api"}
}

func (e *SessionEndpoint) Register(registrar httpx.Registrar) {
	registerSessionRoutes(registrar.Scope(), e.sessionSvc)
}
