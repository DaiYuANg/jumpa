package endpoints

import (
	"context"

	"github.com/arcgolabs/httpx"
)

func (e *SystemEndpoint) GetHealth(context.Context, *struct{}) (*HealthOutput, error) {
	out := &HealthOutput{}
	out.Body.Status = "UP"
	return out, nil
}

func (e *DashboardEndpoint) GetDashboardStats(context.Context, *struct{}) (*dynamicOutput, error) {
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
}

func (e *DashboardEndpoint) GetHealth(context.Context, *struct{}) (*dynamicOutput, error) {
	return &dynamicOutput{Body: ok(map[string]string{"status": "UP"})}, nil
}

func (e *DashboardEndpoint) CreateDebugResetRBAC(context.Context, *struct{}) (*dynamicOutput, error) {
	return nil, httpx.NewError(410, "in-memory rbac store removed")
}
