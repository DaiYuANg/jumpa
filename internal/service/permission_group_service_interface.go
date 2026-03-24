package service

import (
	"context"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
)

type PermissionGroupService interface {
	ListPermissionGroups(ctx context.Context) ([]repo.PermissionGroup, error)
	GetPermissionGroup(ctx context.Context, id string) (repo.PermissionGroup, bool, error)
	CreatePermissionGroup(ctx context.Context, name, description string) (repo.PermissionGroup, error)
	UpdatePermissionGroup(ctx context.Context, id string, name, description *string) (repo.PermissionGroup, bool, error)
	DeletePermissionGroup(ctx context.Context, id string) (bool, error)
}
