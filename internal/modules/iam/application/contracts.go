package application

import (
	"context"

	iamdomain "github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/domain"
	"github.com/samber/mo"
)

type UserService interface {
	List(ctx context.Context, search string, limit, offset int) ([]iamdomain.User, int, error)
	Get(ctx context.Context, id int64) (mo.Option[iamdomain.User], error)
	Create(ctx context.Context, in iamdomain.CreateUserInput) (iamdomain.User, error)
	Update(ctx context.Context, id int64, in iamdomain.UpdateUserInput) (mo.Option[iamdomain.User], error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type RoleService interface {
	ListRoles(ctx context.Context) ([]iamdomain.Role, error)
	GetRole(ctx context.Context, id string) (mo.Option[iamdomain.Role], error)
	CreateRole(ctx context.Context, name, description string, permissionGroupIDs []string) (iamdomain.Role, error)
	UpdateRole(ctx context.Context, id string, name, description *string, permissionGroupIDs []string) (mo.Option[iamdomain.Role], error)
	DeleteRole(ctx context.Context, id string) (bool, error)
}

type PermissionGroupService interface {
	ListPermissionGroups(ctx context.Context) ([]iamdomain.PermissionGroup, error)
	GetPermissionGroup(ctx context.Context, id string) (mo.Option[iamdomain.PermissionGroup], error)
	CreatePermissionGroup(ctx context.Context, name, description string) (iamdomain.PermissionGroup, error)
	UpdatePermissionGroup(ctx context.Context, id string, name, description *string) (mo.Option[iamdomain.PermissionGroup], error)
	DeletePermissionGroup(ctx context.Context, id string) (bool, error)
}

type PermissionService interface {
	ListPermissions(ctx context.Context) ([]iamdomain.Permission, error)
	GetPermission(ctx context.Context, id string) (mo.Option[iamdomain.Permission], error)
	CreatePermission(ctx context.Context, name, code string, groupID *string) (iamdomain.Permission, error)
	UpdatePermission(ctx context.Context, id string, name, code *string, groupID *string) (mo.Option[iamdomain.Permission], error)
	DeletePermission(ctx context.Context, id string) (bool, error)
}

type UserRoleService interface {
	ListUserRoleIDs(ctx context.Context, userID int64) ([]string, error)
	SetUserRoleIDs(ctx context.Context, userID int64, roleIDs []string) error
	DeleteUserRoles(ctx context.Context, userID int64) error
}

type AuthPrincipalService interface {
	UpsertAuthPrincipal(ctx context.Context, userID int64, email string) error
	DeleteAuthPrincipal(ctx context.Context, userID int64) error
	SetAuthPrincipalRoles(ctx context.Context, userID int64, roleIDs []string) error
}
