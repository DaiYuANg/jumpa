package http

import (
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
)

type OverviewEndpoint struct {
	httpx.BaseEndpoint
	overviewSvc application.OverviewService
}

func NewOverviewEndpoint(overviewSvc application.OverviewService) *OverviewEndpoint {
	return &OverviewEndpoint{overviewSvc: overviewSvc}
}

func (e *OverviewEndpoint) RegisterRoutes(server httpx.ServerRuntime) {
	registerOverviewRoutes(server.Group("/api"), e.overviewSvc)
}
