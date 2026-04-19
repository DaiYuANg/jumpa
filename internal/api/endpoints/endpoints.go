package endpoints

import (
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/kvx"
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
	registerSystemEndpoints(registrar.Scope())
}

func (e *AuthEndpoint) EndpointSpec() httpx.EndpointSpec {
	return httpx.EndpointSpec{Prefix: "/api"}
}

func (e *AuthEndpoint) Register(registrar httpx.Registrar) {
	registerAuthEndpoints(registrar.Scope(), e.cfg, e.kvClient)
}

func (e *DashboardEndpoint) EndpointSpec() httpx.EndpointSpec {
	return httpx.EndpointSpec{Prefix: "/api"}
}

func (e *DashboardEndpoint) Register(registrar httpx.Registrar) {
	registerDashboardEndpoints(registrar.Scope())
}
