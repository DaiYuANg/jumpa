package http

import (
	"context"

	"github.com/DaiYuANg/arcgo/httpx"
	apiendpoints "github.com/DaiYuANg/jumpa/internal/api/endpoints"
	"github.com/danielgtaylor/huma/v2"
)

func (e *BastionEndpoint) registerOverviewRoutes(api *httpx.Group) {
	httpx.MustGroupGet(api, "/bastion/overview", func(ctx context.Context, _ *struct{}) (*apiendpoints.DynamicOutput, error) {
		data, err := e.overviewSvc.Get(ctx)
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toOverviewDTO(data))}, nil
	}, huma.OperationTags("bastion"))
}
