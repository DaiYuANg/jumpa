package dbx

import (
	"context"
	"time"

	"github.com/DaiYuANg/arcgo/dbx"
)

type Role struct {
	ID                 string
	Name               string
	Description        string
	PermissionGroupIDs []string
	CreatedAt          time.Time
}

// RoleRecord is the roles table row mapped to domain-relevant fields.
// Aggregations (e.g. PermissionGroupIDs) are composed in the service layer.
type RoleRecord struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
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

type CreateRoleInput struct {
	ID          string
	Name        string
	Description string
}

type PatchRoleInput struct {
	Name        *string
	Description *string
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
	ListRoles(ctx context.Context, session dbx.Session) ([]RoleRecord, error)
	GetRole(ctx context.Context, session dbx.Session, id string) (RoleRecord, bool, error)
	CreateRole(ctx context.Context, session dbx.Session, in CreateRoleInput) (RoleRecord, error)
	UpdateRole(ctx context.Context, session dbx.Session, id string, in PatchRoleInput) (RoleRecord, bool, error)
	DeleteRole(ctx context.Context, session dbx.Session, id string) (bool, error)
}

type RolePermissionGroupPair struct {
	RoleID            string
	PermissionGroupID string
}

type RolePermissionGroupRepository interface {
	ListPairs(ctx context.Context, session dbx.Session) ([]RolePermissionGroupPair, error)
	ListPairsByRoleIDs(ctx context.Context, session dbx.Session, roleIDs []string) ([]RolePermissionGroupPair, error)
	ListPermissionGroupIDsByRoleID(ctx context.Context, session dbx.Session, roleID string) ([]string, error)
	ReplacePermissionGroupIDs(ctx context.Context, session dbx.Session, roleID string, permissionGroupIDs []string) error
	DeleteByRoleID(ctx context.Context, session dbx.Session, roleID string) error
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

