package repo

import "time"

type Role struct {
	ID                 string
	Name               string
	Description        string
	PermissionGroupIDs []string
	CreatedAt          time.Time
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
	ID                 string
	Name               string
	Description        string
	PermissionGroupIDs []string
}

type PatchRoleInput struct {
	Name               *string
	Description        *string
	PermissionGroupIDs []string
}

type CreatePermissionGroupInput struct {
	ID          string
	Name        string
	Description string
}

type PatchPermissionGroupInput struct {
	Name        *string
	Description *string
}

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
