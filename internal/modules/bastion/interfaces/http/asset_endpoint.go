package http

import (
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
)

type AssetEndpoint struct {
	assetSvc application.AssetService
}

func NewAssetEndpoint(assetSvc application.AssetService) *AssetEndpoint {
	return &AssetEndpoint{assetSvc: assetSvc}
}

func (e *AssetEndpoint) EndpointSpec() httpx.EndpointSpec {
	return httpx.EndpointSpec{Prefix: "/api"}
}

func (e *AssetEndpoint) Register(registrar httpx.Registrar) {
	registerAssetRoutes(registrar.Scope(), e.assetSvc)
}
