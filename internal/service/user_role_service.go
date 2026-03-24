package service

import (
	"context"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
)

type userRoleService struct {
	repo repo.UserRoleRepository
}

func NewUserRoleService(r repo.UserRoleRepository) UserRoleService {
	return &userRoleService{repo: r}
}

func (s *userRoleService) ListUserRoleIDs(ctx context.Context, userID int64) ([]string, error) {
	return s.repo.ListUserRoleIDs(ctx, userID)
}

func (s *userRoleService) SetUserRoleIDs(ctx context.Context, userID int64, roleIDs []string) error {
	return s.repo.SetUserRoleIDs(ctx, userID, roleIDs)
}

func (s *userRoleService) DeleteUserRoles(ctx context.Context, userID int64) error {
	return s.repo.DeleteUserRoles(ctx, userID)
}
