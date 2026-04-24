package http

import (
	"context"

	"github.com/arcgolabs/httpx"
	apiendpoints "github.com/DaiYuANg/jumpa/internal/api/endpoints"
	"github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/application"
	gatewaydomain "github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/domain"
	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/lo"
)

type GatewayEndpoint struct {
	service application.GatewayService
}

func NewGatewayEndpoint(service application.GatewayService) *GatewayEndpoint {
	return &GatewayEndpoint{service: service}
}

func (e *GatewayEndpoint) EndpointSpec() httpx.EndpointSpec {
	return httpx.EndpointSpec{Prefix: "/api"}
}

func (e *GatewayEndpoint) Register(registrar httpx.Registrar) {
	api := registrar.Scope()
	httpx.MustGroupGet(api, "/gateways", func(ctx context.Context, _ *struct{}) (*apiendpoints.DynamicOutput, error) {
		items, err := e.service.List(ctx)
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toGatewayDTOs(items))}, nil
	}, huma.OperationTags("gateways"))

	httpx.MustGroupGet(api, "/gateways/{id}", func(ctx context.Context, input *apiendpoints.ByIDInput) (*apiendpoints.DynamicOutput, error) {
		item, err := e.service.Get(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "gateway not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toGatewayDTOs([]gatewaydomain.Gateway{item.MustGet()})[0])}, nil
	}, huma.OperationTags("gateways"))
}

func toGatewayDTOs(items []gatewaydomain.Gateway) []gatewayDTO {
	return lo.Map(items, func(it gatewaydomain.Gateway, _ int) gatewayDTO {
		return gatewayDTO{
			ID:              it.ID,
			NodeKey:         it.NodeKey,
			NodeName:        it.NodeName,
			RuntimeType:     it.RuntimeType,
			AdvertiseAddr:   it.AdvertiseAddr,
			SSHListenAddr:   it.SSHListenAddr,
			Zone:            it.Zone,
			Tags:            it.Tags,
			State:           it.State,
			EffectiveStatus: it.EffectiveStatus,
			RegisteredAt:    it.RegisteredAt,
			LastSeenAt:      it.LastSeenAt,
			UpdatedAt:       it.UpdatedAt,
		}
	})
}
