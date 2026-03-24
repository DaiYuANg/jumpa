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
type RBACEndpoint struct {
	httpx.BaseEndpoint
	roleSvc  RoleService
	groupSvc PermissionGroupService
	permSvc  PermissionService
}
type UserEndpoint struct {
	httpx.BaseEndpoint
	userSvc      UserService
	userRoleSvc  UserRoleService
	principalSvc AuthPrincipalService
}

func NewSystemEndpoint() *SystemEndpoint       { return &SystemEndpoint{} }
func NewAuthEndpoint(cfg AuthConfig, kvClient kvx.Client) *AuthEndpoint {
	return &AuthEndpoint{cfg: cfg, kvClient: kvClient}
}
func NewDashboardEndpoint() *DashboardEndpoint { return &DashboardEndpoint{} }
func NewRBACEndpoint(roleSvc RoleService, groupSvc PermissionGroupService, permSvc PermissionService) *RBACEndpoint {
	return &RBACEndpoint{roleSvc: roleSvc, groupSvc: groupSvc, permSvc: permSvc}
}
func NewUserEndpoint(userSvc UserService, userRoleSvc UserRoleService, principalSvc AuthPrincipalService) *UserEndpoint {
	return &UserEndpoint{userSvc: userSvc, userRoleSvc: userRoleSvc, principalSvc: principalSvc}
}

func (e *SystemEndpoint) RegisterRoutes(server httpx.ServerRuntime)    { registerSystemEndpoints(server) }
func (e *AuthEndpoint) RegisterRoutes(server httpx.ServerRuntime) {
	registerAuthEndpoints(server.Group("/api"), e.cfg, e.kvClient)
}
func (e *DashboardEndpoint) RegisterRoutes(server httpx.ServerRuntime) { registerDashboardEndpoints(server.Group("/api")) }
func (e *RBACEndpoint) RegisterRoutes(server httpx.ServerRuntime) {
	registerRBACEndpoints(server.Group("/api"), e.roleSvc, e.groupSvc, e.permSvc)
}
func (e *UserEndpoint) RegisterRoutes(server httpx.ServerRuntime) {
	registerUserEndpoints(server.Group("/api"), e.userSvc, e.userRoleSvc, e.principalSvc)
}
