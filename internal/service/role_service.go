package service

import (
	"context"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
)

type roleService struct {
	repo repo.RoleRepository
}

func NewRoleService(r repo.RoleRepository) RoleService {
	return &roleService{repo: r}
}

func (s *roleService) ListRoles(ctx context.Context) ([]repo.Role, error) {
	return s.repo.ListRoles(ctx)
}

func (s *roleService) GetRole(ctx context.Context, id string) (repo.Role, bool, error) {
	return s.repo.GetRole(ctx, id)
}

func (s *roleService) CreateRole(ctx context.Context, name, description string, permissionGroupIDs []string) (repo.Role, error) {
	return s.repo.CreateRole(ctx, repo.CreateRoleInput{
		ID:                 makeID("role"),
		Name:               name,
		Description:        description,
		PermissionGroupIDs: permissionGroupIDs,
	})
}

func (s *roleService) UpdateRole(ctx context.Context, id string, name, description *string, permissionGroupIDs []string) (repo.Role, bool, error) {
	return s.repo.UpdateRole(ctx, id, repo.PatchRoleInput{
		Name:               name,
		Description:        description,
		PermissionGroupIDs: permissionGroupIDs,
	})
}

func (s *roleService) DeleteRole(ctx context.Context, id string) (bool, error) {
	return s.repo.DeleteRole(ctx, id)
}
