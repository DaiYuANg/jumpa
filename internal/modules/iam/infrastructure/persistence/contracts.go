package persistence

import (
	"context"

	legacydomain "github.com/DaiYuANg/arcgo-rbac-template/internal/domain"
	dbxrepo "github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/infrastructure/persistence/dbx"
)

type UserRepository interface {
	List(ctx context.Context, search string, limit, offset int) ([]legacydomain.User, int, error)
	GetByID(ctx context.Context, id int64) (legacydomain.User, bool, error)
	Create(ctx context.Context, in legacydomain.CreateUserInput) (legacydomain.User, error)
	Update(ctx context.Context, id int64, in legacydomain.UpdateUserInput) (legacydomain.User, bool, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type Role = dbxrepo.Role
type PermissionGroup = dbxrepo.PermissionGroup
type Permission = dbxrepo.Permission
type CreateRoleInput = dbxrepo.CreateRoleInput
type PatchRoleInput = dbxrepo.PatchRoleInput
type CreatePermissionGroupInput = dbxrepo.CreatePermissionGroupInput
type PatchPermissionGroupInput = dbxrepo.PatchPermissionGroupInput
type CreatePermissionInput = dbxrepo.CreatePermissionInput
type PatchPermissionInput = dbxrepo.PatchPermissionInput

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
