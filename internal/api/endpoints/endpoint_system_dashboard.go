package endpoints

import (
	"context"

	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/danielgtaylor/huma/v2"
)

func registerSystemEndpoints(server httpx.ServerRuntime) {
	httpx.MustGet(server, "/health", func(ctx context.Context, _ *struct{}) (*HealthOutput, error) {
		out := &HealthOutput{}
		out.Body.Status = "UP"
		return out, nil
	}, huma.OperationTags("system"))
}

func registerDashboardEndpoints(api *httpx.Group) {
	httpx.MustGroupGet(api, "/dashboard/stats", func(ctx context.Context, _ *struct{}) (*dynamicOutput, error) {
		return &dynamicOutput{
			Body: ok(map[string]any{
				"statCards": []map[string]any{
					{"key": "users", "value": 12, "labelKey": "dashboard.totalUsers"},
					{"key": "roles", "value": 3, "labelKey": "dashboard.totalRoles"},
				},
				"userActivity": []map[string]any{
					{"month": "Jan", "users": 10, "logins": 40},
					{"month": "Feb", "users": 12, "logins": 48},
				},
				"roleDistribution": []map[string]any{
					{"name": "admin", "value": 2, "color": "var(--chart-1)"},
					{"name": "user", "value": 10, "color": "var(--chart-2)"},
				},
				"permissionGroups": []map[string]any{{"name": "core", "count": 8}},
			}),
		}, nil
	}, huma.OperationTags("dashboard"))

	httpx.MustGroupGet(api, "/health", func(ctx context.Context, _ *struct{}) (*dynamicOutput, error) {
		return &dynamicOutput{Body: ok(map[string]string{"status": "UP"})}, nil
	}, huma.OperationTags("system"))

	httpx.MustGroupPost(api, "/debug/reset-rbac", func(ctx context.Context, _ *struct{}) (*dynamicOutput, error) {
		return nil, httpx.NewError(410, "in-memory rbac store removed")
	}, huma.OperationTags("debug"))
}
