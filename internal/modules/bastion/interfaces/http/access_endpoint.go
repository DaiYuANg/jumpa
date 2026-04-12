package http

import (
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
)

type AccessEndpoint struct {
	httpx.BaseEndpoint
	policySvc  application.PolicyService
	requestSvc application.AccessRequestService
}

func NewAccessEndpoint(policySvc application.PolicyService, requestSvc application.AccessRequestService) *AccessEndpoint {
	return &AccessEndpoint{
		policySvc:  policySvc,
		requestSvc: requestSvc,
	}
}

func (e *AccessEndpoint) RegisterRoutes(server httpx.ServerRuntime) {
	registerAccessRoutes(server.Group("/api"), e.policySvc, e.requestSvc)
}
