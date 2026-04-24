package http

import (
	"github.com/arcgolabs/httpx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
)

type OverviewEndpoint struct {
	overviewSvc application.OverviewService
}

func NewOverviewEndpoint(overviewSvc application.OverviewService) *OverviewEndpoint {
	return &OverviewEndpoint{overviewSvc: overviewSvc}
}

func (e *OverviewEndpoint) EndpointSpec() httpx.EndpointSpec {
	return httpx.EndpointSpec{Prefix: "/api"}
}

func (e *OverviewEndpoint) Register(registrar httpx.Registrar) {
	registerOverviewRoutes(registrar.Scope(), e.overviewSvc)
}
