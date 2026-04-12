package http

import (
	"context"

	"github.com/DaiYuANg/arcgo/httpx"
	apiendpoints "github.com/DaiYuANg/jumpa/internal/api/endpoints"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
	"github.com/danielgtaylor/huma/v2"
)

func registerSessionRoutes(api *httpx.Group, sessionSvc application.SessionService) {
	httpx.MustGroupGet(api, "/sessions", func(ctx context.Context, _ *struct{}) (*apiendpoints.DynamicOutput, error) {
		items, err := sessionSvc.ListSessions(ctx)
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toSessionDTOs(items))}, nil
	}, huma.OperationTags("sessions"))
}
