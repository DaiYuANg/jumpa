package service

import "context"

type UserRoleService interface {
	ListUserRoleIDs(ctx context.Context, userID int64) ([]string, error)
	SetUserRoleIDs(ctx context.Context, userID int64, roleIDs []string) error
	DeleteUserRoles(ctx context.Context, userID int64) error
}
