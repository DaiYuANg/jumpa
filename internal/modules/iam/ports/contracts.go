package ports

import (
	"context"
	"database/sql"
	"time"

	iamdomain "github.com/DaiYuANg/jumpa/internal/modules/iam/domain"
	"github.com/samber/mo"
)

type UserRepository interface {
	List(ctx context.Context, search string, limit, offset int) ([]iamdomain.User, int, error)
	GetByID(ctx context.Context, id int64) (mo.Option[iamdomain.User], error)
	Create(ctx context.Context, in iamdomain.CreateUserInput) (iamdomain.User, error)
	Update(ctx context.Context, id int64, in iamdomain.UpdateUserInput) (mo.Option[iamdomain.User], error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type RoleRecord struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}

type CreateRoleInput struct {
	ID          string
	Name        string
	Description string
}

type PatchRoleInput struct {
	Name        *string
	Description *string
}

type PermissionGroup struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}

type Permission struct {
	ID        string
	Name      string
	Code      string
	GroupID   *string
	CreatedAt time.Time
}

type CreatePermissionGroupInput struct{ ID, Name, Description string }
type PatchPermissionGroupInput struct{ Name, Description *string }

type CreatePermissionInput struct {
	ID      string
	Name    string
	Code    string
	GroupID *string
}

type PatchPermissionInput struct {
	Name    *string
	Code    *string
	GroupID *string
}

type RoleRepository interface {
	ListRoles(ctx context.Context) ([]RoleRecord, error)
	GetRole(ctx context.Context, id string) (mo.Option[RoleRecord], error)
	CreateRole(ctx context.Context, in CreateRoleInput) (RoleRecord, error)
	UpdateRole(ctx context.Context, id string, in PatchRoleInput) (mo.Option[RoleRecord], error)
	DeleteRole(ctx context.Context, id string) (bool, error)
}

type RolePermissionGroupPair struct {
	RoleID            string
	PermissionGroupID string
}

type RolePermissionGroupRepository interface {
	ListPairs(ctx context.Context) ([]RolePermissionGroupPair, error)
	ListPairsByRoleIDs(ctx context.Context, roleIDs []string) ([]RolePermissionGroupPair, error)
	ListPermissionGroupIDsByRoleID(ctx context.Context, roleID string) ([]string, error)
	ReplacePermissionGroupIDs(ctx context.Context, roleID string, permissionGroupIDs []string) error
	DeleteByRoleID(ctx context.Context, roleID string) error
}

type UnitOfWork interface {
	InTx(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context, tx UnitOfWorkTx) error) error
}

type UnitOfWorkTx interface {
	Roles() RoleRepository
	RolePermissionGroups() RolePermissionGroupRepository
}

type PermissionGroupRepository interface {
	ListPermissionGroups(ctx context.Context) ([]PermissionGroup, error)
	GetPermissionGroup(ctx context.Context, id string) (mo.Option[PermissionGroup], error)
	CreatePermissionGroup(ctx context.Context, in CreatePermissionGroupInput) (PermissionGroup, error)
	UpdatePermissionGroup(ctx context.Context, id string, in PatchPermissionGroupInput) (mo.Option[PermissionGroup], error)
	DeletePermissionGroup(ctx context.Context, id string) (bool, error)
}

type PermissionRepository interface {
	ListPermissions(ctx context.Context) ([]Permission, error)
	GetPermission(ctx context.Context, id string) (mo.Option[Permission], error)
	CreatePermission(ctx context.Context, in CreatePermissionInput) (Permission, error)
	UpdatePermission(ctx context.Context, id string, in PatchPermissionInput) (mo.Option[Permission], error)
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
