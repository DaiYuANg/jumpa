package service

import (
	"context"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
)

type RoleService interface {
	ListRoles(ctx context.Context) ([]repo.Role, error)
	GetRole(ctx context.Context, id string) (repo.Role, bool, error)
	CreateRole(ctx context.Context, name, description string, permissionGroupIDs []string) (repo.Role, error)
	UpdateRole(ctx context.Context, id string, name, description *string, permissionGroupIDs []string) (repo.Role, bool, error)
	DeleteRole(ctx context.Context, id string) (bool, error)
}
