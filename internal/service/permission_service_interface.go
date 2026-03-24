package service

import (
	"context"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
)

type PermissionService interface {
	ListPermissions(ctx context.Context) ([]repo.Permission, error)
	GetPermission(ctx context.Context, id string) (repo.Permission, bool, error)
	CreatePermission(ctx context.Context, name, code string, groupID *string) (repo.Permission, error)
	UpdatePermission(ctx context.Context, id string, name, code *string, groupID *string) (repo.Permission, bool, error)
	DeletePermission(ctx context.Context, id string) (bool, error)
}
