package repo

import (
	"time"

	"github.com/DaiYuANg/arcgo/dbx"
)

type roleRow struct {
	ID          string    `dbx:"id"`
	Name        string    `dbx:"name"`
	Description string    `dbx:"description"`
	CreatedAt   time.Time `dbx:"created_at,codec=rfc3339_time"`
}
type roleSchema struct {
	dbx.Schema[roleRow]
	ID          dbx.Column[roleRow, string]    `dbx:"id,pk"`
	Name        dbx.Column[roleRow, string]    `dbx:"name"`
	Description dbx.Column[roleRow, string]    `dbx:"description"`
	CreatedAt   dbx.Column[roleRow, time.Time] `dbx:"created_at,codec=rfc3339_time"`
}

type permissionGroupRow struct {
	ID          string    `dbx:"id"`
	Name        string    `dbx:"name"`
	Description string    `dbx:"description"`
	CreatedAt   time.Time `dbx:"created_at,codec=rfc3339_time"`
}
type permissionGroupSchema struct {
	dbx.Schema[permissionGroupRow]
	ID          dbx.Column[permissionGroupRow, string]    `dbx:"id,pk"`
	Name        dbx.Column[permissionGroupRow, string]    `dbx:"name"`
	Description dbx.Column[permissionGroupRow, string]    `dbx:"description"`
	CreatedAt   dbx.Column[permissionGroupRow, time.Time] `dbx:"created_at,codec=rfc3339_time"`
}

type permissionRow struct {
	ID        string    `dbx:"id"`
	Name      string    `dbx:"name"`
	Code      string    `dbx:"code"`
	GroupID   *string   `dbx:"group_id"`
	CreatedAt time.Time `dbx:"created_at,codec=rfc3339_time"`
}
type permissionSchema struct {
	dbx.Schema[permissionRow]
	ID        dbx.Column[permissionRow, string]    `dbx:"id,pk"`
	Name      dbx.Column[permissionRow, string]    `dbx:"name"`
	Code      dbx.Column[permissionRow, string]    `dbx:"code"`
	GroupID   dbx.Column[permissionRow, *string]   `dbx:"group_id"`
	CreatedAt dbx.Column[permissionRow, time.Time] `dbx:"created_at,codec=rfc3339_time"`
}

type userRoleRow struct {
	UserID int64  `dbx:"user_id"`
	RoleID string `dbx:"role_id"`
}
type userRoleSchema struct {
	dbx.Schema[userRoleRow]
	UserID dbx.Column[userRoleRow, int64]  `dbx:"user_id"`
	RoleID dbx.Column[userRoleRow, string] `dbx:"role_id"`
}

type rolePermissionGroupRow struct {
	RoleID            string `dbx:"role_id"`
	PermissionGroupID string `dbx:"permission_group_id"`
}
type rolePermissionGroupSchema struct {
	dbx.Schema[rolePermissionGroupRow]
	RoleID            dbx.Column[rolePermissionGroupRow, string] `dbx:"role_id"`
	PermissionGroupID dbx.Column[rolePermissionGroupRow, string] `dbx:"permission_group_id"`
}

type authPrincipalRow struct {
	ID    string `dbx:"id"`
	Email string `dbx:"email"`
}
type authPrincipalSchema struct {
	dbx.Schema[authPrincipalRow]
	ID    dbx.Column[authPrincipalRow, string] `dbx:"id,pk"`
	Email dbx.Column[authPrincipalRow, string] `dbx:"email"`
}

type authPrincipalRoleRow struct {
	PrincipalID string `dbx:"principal_id"`
	Role        string `dbx:"role"`
}
type authPrincipalRoleSchema struct {
	dbx.Schema[authPrincipalRoleRow]
	PrincipalID dbx.Column[authPrincipalRoleRow, string] `dbx:"principal_id"`
	Role        dbx.Column[authPrincipalRoleRow, string] `dbx:"role"`
}
