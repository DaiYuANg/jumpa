package http

import (
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
)

type AssetEndpoint struct {
	httpx.BaseEndpoint
	assetSvc application.AssetService
}

func NewAssetEndpoint(assetSvc application.AssetService) *AssetEndpoint {
	return &AssetEndpoint{assetSvc: assetSvc}
}

func (e *AssetEndpoint) RegisterRoutes(server httpx.ServerRuntime) {
	registerAssetRoutes(server.Group("/api"), e.assetSvc)
}
