package service

import (
	"context"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
)

type authPrincipalService struct {
	repo repo.AuthPrincipalRepository
}

func NewAuthPrincipalService(r repo.AuthPrincipalRepository) AuthPrincipalService {
	return &authPrincipalService{repo: r}
}

func (s *authPrincipalService) UpsertAuthPrincipal(ctx context.Context, userID int64, email string) error {
	return s.repo.UpsertAuthPrincipal(ctx, userID, email)
}

func (s *authPrincipalService) DeleteAuthPrincipal(ctx context.Context, userID int64) error {
	return s.repo.DeleteAuthPrincipal(ctx, userID)
}

func (s *authPrincipalService) SetAuthPrincipalRoles(ctx context.Context, userID int64, roleIDs []string) error {
	return s.repo.SetAuthPrincipalRoles(ctx, userID, roleIDs)
}
