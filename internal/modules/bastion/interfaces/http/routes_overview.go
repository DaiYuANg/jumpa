package http

import (
	"context"

	"github.com/arcgolabs/httpx"
	apiendpoints "github.com/DaiYuANg/jumpa/internal/api/endpoints"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
	"github.com/danielgtaylor/huma/v2"
)

func registerOverviewRoutes(api *httpx.Group, overviewSvc application.OverviewService) {
	httpx.MustGroupGet(api, "/bastion/overview", func(ctx context.Context, _ *struct{}) (*apiendpoints.DynamicOutput, error) {
		data, err := overviewSvc.Get(ctx)
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toOverviewDTO(data))}, nil
	}, huma.OperationTags("bastion"))
}
