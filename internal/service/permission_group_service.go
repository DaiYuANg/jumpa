package service

import (
	"context"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
)

type permissionGroupService struct {
	repo repo.PermissionGroupRepository
}

func NewPermissionGroupService(r repo.PermissionGroupRepository) PermissionGroupService {
	return &permissionGroupService{repo: r}
}

func (s *permissionGroupService) ListPermissionGroups(ctx context.Context) ([]repo.PermissionGroup, error) {
	return s.repo.ListPermissionGroups(ctx)
}

func (s *permissionGroupService) GetPermissionGroup(ctx context.Context, id string) (repo.PermissionGroup, bool, error) {
	return s.repo.GetPermissionGroup(ctx, id)
}

func (s *permissionGroupService) CreatePermissionGroup(ctx context.Context, name, description string) (repo.PermissionGroup, error) {
	return s.repo.CreatePermissionGroup(ctx, repo.CreatePermissionGroupInput{
		ID:          makeID("pg"),
		Name:        name,
		Description: description,
	})
}

func (s *permissionGroupService) UpdatePermissionGroup(ctx context.Context, id string, name, description *string) (repo.PermissionGroup, bool, error) {
	return s.repo.UpdatePermissionGroup(ctx, id, repo.PatchPermissionGroupInput{
		Name:        name,
		Description: description,
	})
}

func (s *permissionGroupService) DeletePermissionGroup(ctx context.Context, id string) (bool, error) {
	return s.repo.DeletePermissionGroup(ctx, id)
}
