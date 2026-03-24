package repo

import "context"

type RoleRepository interface {
	ListRoles(ctx context.Context) ([]Role, error)
	GetRole(ctx context.Context, id string) (Role, bool, error)
	CreateRole(ctx context.Context, in CreateRoleInput) (Role, error)
	UpdateRole(ctx context.Context, id string, in PatchRoleInput) (Role, bool, error)
	DeleteRole(ctx context.Context, id string) (bool, error)
}

type PermissionGroupRepository interface {
	ListPermissionGroups(ctx context.Context) ([]PermissionGroup, error)
	GetPermissionGroup(ctx context.Context, id string) (PermissionGroup, bool, error)
	CreatePermissionGroup(ctx context.Context, in CreatePermissionGroupInput) (PermissionGroup, error)
	UpdatePermissionGroup(ctx context.Context, id string, in PatchPermissionGroupInput) (PermissionGroup, bool, error)
	DeletePermissionGroup(ctx context.Context, id string) (bool, error)
}

type PermissionRepository interface {
	ListPermissions(ctx context.Context) ([]Permission, error)
	GetPermission(ctx context.Context, id string) (Permission, bool, error)
	CreatePermission(ctx context.Context, in CreatePermissionInput) (Permission, error)
	UpdatePermission(ctx context.Context, id string, in PatchPermissionInput) (Permission, bool, error)
	DeletePermission(ctx context.Context, id string) (bool, error)
}

type UserRoleRepository interface {
	ListUserRoleIDs(ctx context.Context, userID int64) ([]string, error)
	SetUserRoleIDs(ctx context.Context, userID int64, roleIDs []string) error
	DeleteUserRoles(ctx context.Context, userID int64) error
}

type AuthPrincipalRepository interface {
	UpsertAuthPrincipal(ctx context.Context, userID int64, email string) error
	DeleteAuthPrincipal(ctx context.Context, userID int64) error
	SetAuthPrincipalRoles(ctx context.Context, userID int64, roleIDs []string) error
}
