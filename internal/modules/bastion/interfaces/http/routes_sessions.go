package http

import (
	"context"

	"github.com/DaiYuANg/arcgo/httpx"
	apiendpoints "github.com/DaiYuANg/jumpa/internal/api/endpoints"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
	bastiondomain "github.com/DaiYuANg/jumpa/internal/modules/bastion/domain"
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

	httpx.MustGroupGet(api, "/sessions/{id}", func(ctx context.Context, input *apiendpoints.ByIDInput) (*apiendpoints.DynamicOutput, error) {
		item, err := sessionSvc.GetSession(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "session not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toSessionDTOs([]bastiondomain.Session{item.MustGet()})[0])}, nil
	}, huma.OperationTags("sessions"))
}
