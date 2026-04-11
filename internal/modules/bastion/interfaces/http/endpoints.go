package http

import (
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
)

type BastionEndpoint struct {
	httpx.BaseEndpoint
	overviewSvc application.OverviewService
	assetSvc    application.AssetService
	policySvc   application.PolicyService
	requestSvc  application.AccessRequestService
	sessionSvc  application.SessionService
}

func NewBastionEndpoint(
	overviewSvc application.OverviewService,
	assetSvc application.AssetService,
	policySvc application.PolicyService,
	requestSvc application.AccessRequestService,
	sessionSvc application.SessionService,
) *BastionEndpoint {
	return &BastionEndpoint{
		overviewSvc: overviewSvc,
		assetSvc:    assetSvc,
		policySvc:   policySvc,
		requestSvc:  requestSvc,
		sessionSvc:  sessionSvc,
	}
}

func (e *BastionEndpoint) RegisterRoutes(server httpx.ServerRuntime) {
	api := server.Group("/api")
	e.registerOverviewRoutes(api)
	e.registerAssetRoutes(api)
	e.registerAccessRoutes(api)
	e.registerSessionRoutes(api)
}

func boolOr(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func normalizePageRequest(page, pageSize int) (normalizedPage int, normalizedPageSize int, offset int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize, (page - 1) * pageSize
}
