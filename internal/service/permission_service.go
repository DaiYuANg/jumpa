package service

import (
	"context"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
)

type permissionService struct {
	repo repo.PermissionRepository
}

func NewPermissionService(r repo.PermissionRepository) PermissionService {
	return &permissionService{repo: r}
}

func (s *permissionService) ListPermissions(ctx context.Context) ([]repo.Permission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *permissionService) GetPermission(ctx context.Context, id string) (repo.Permission, bool, error) {
	return s.repo.GetPermission(ctx, id)
}

func (s *permissionService) CreatePermission(ctx context.Context, name, code string, groupID *string) (repo.Permission, error) {
	return s.repo.CreatePermission(ctx, repo.CreatePermissionInput{
		ID:      makeID("perm"),
		Name:    name,
		Code:    code,
		GroupID: groupID,
	})
}

func (s *permissionService) UpdatePermission(ctx context.Context, id string, name, code *string, groupID *string) (repo.Permission, bool, error) {
	return s.repo.UpdatePermission(ctx, id, repo.PatchPermissionInput{
		Name:    name,
		Code:    code,
		GroupID: groupID,
	})
}

func (s *permissionService) DeletePermission(ctx context.Context, id string) (bool, error) {
	return s.repo.DeletePermission(ctx, id)
}
