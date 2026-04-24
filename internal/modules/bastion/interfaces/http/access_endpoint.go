package http

import (
	"github.com/arcgolabs/httpx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
)

type AccessEndpoint struct {
	policySvc  application.PolicyService
	requestSvc application.AccessRequestService
}

func NewAccessEndpoint(policySvc application.PolicyService, requestSvc application.AccessRequestService) *AccessEndpoint {
	return &AccessEndpoint{
		policySvc:  policySvc,
		requestSvc: requestSvc,
	}
}

func (e *AccessEndpoint) EndpointSpec() httpx.EndpointSpec {
	return httpx.EndpointSpec{Prefix: "/api"}
}

func (e *AccessEndpoint) Register(registrar httpx.Registrar) {
	registerAccessRoutes(registrar.Scope(), e.policySvc, e.requestSvc)
}
