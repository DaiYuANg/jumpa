package endpoints

import "github.com/DaiYuANg/arcgo/httpx"
import "github.com/DaiYuANg/arcgo/kvx"

type SystemEndpoint struct{ httpx.BaseEndpoint }
type AuthEndpoint struct {
	httpx.BaseEndpoint
	cfg      AuthConfig
	kvClient kvx.Client
}
type DashboardEndpoint struct{ httpx.BaseEndpoint }

func NewSystemEndpoint() *SystemEndpoint       { return &SystemEndpoint{} }
func NewAuthEndpoint(cfg AuthConfig, kvClient kvx.Client) *AuthEndpoint {
	return &AuthEndpoint{cfg: cfg, kvClient: kvClient}
}
func NewDashboardEndpoint() *DashboardEndpoint { return &DashboardEndpoint{} }

func (e *SystemEndpoint) RegisterRoutes(server httpx.ServerRuntime)    { registerSystemEndpoints(server) }
func (e *AuthEndpoint) RegisterRoutes(server httpx.ServerRuntime) {
	registerAuthEndpoints(server.Group("/api"), e.cfg, e.kvClient)
}
func (e *DashboardEndpoint) RegisterRoutes(server httpx.ServerRuntime) { registerDashboardEndpoints(server.Group("/api")) }
