package endpoints

import (
	"github.com/arcgolabs/httpx"
	"github.com/arcgolabs/kvx"
	"github.com/danielgtaylor/huma/v2"
)

type SystemEndpoint struct{}
type AuthEndpoint struct {
	cfg      AuthConfig
	kvClient kvx.Client
}
type DashboardEndpoint struct{}

func NewSystemEndpoint() *SystemEndpoint { return &SystemEndpoint{} }
func NewAuthEndpoint(cfg AuthConfig, kvClient kvx.Client) *AuthEndpoint {
	return &AuthEndpoint{cfg: cfg, kvClient: kvClient}
}
func NewDashboardEndpoint() *DashboardEndpoint { return &DashboardEndpoint{} }

func (e *SystemEndpoint) Register(registrar httpx.Registrar) {
	httpx.MustAuto(registrar,
		httpx.Auto(e.GetHealth, huma.OperationTags("system")),
	)
}

func (e *AuthEndpoint) EndpointSpec() httpx.EndpointSpec {
	return httpx.EndpointSpec{Prefix: "/api"}
}

func (e *AuthEndpoint) Register(registrar httpx.Registrar) {
	httpx.MustAuto(registrar,
		httpx.Auto(e.CreateAuthLogin, huma.OperationTags("auth")),
		httpx.Auto(e.CreateAuthRefresh, huma.OperationTags("auth")),
		httpx.Auto(e.CreateAuthLogout, huma.OperationTags("auth")),
		httpx.Auto(e.GetMe, huma.OperationTags("auth")),
	)
}

func (e *DashboardEndpoint) EndpointSpec() httpx.EndpointSpec {
	return httpx.EndpointSpec{Prefix: "/api"}
}

func (e *DashboardEndpoint) Register(registrar httpx.Registrar) {
	httpx.MustAuto(registrar,
		httpx.Auto(e.GetDashboardStats, huma.OperationTags("dashboard")),
		httpx.Auto(e.GetHealth, huma.OperationTags("system")),
		httpx.Auto(e.CreateDebugResetRBAC, huma.OperationTags("debug")),
	)
}
