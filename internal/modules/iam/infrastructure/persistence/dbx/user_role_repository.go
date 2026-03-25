package dbx

import (
	"context"

	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dbx/repository"
	"github.com/samber/lo"
)

type userRoleRow struct {
	UserID int64  `dbx:"user_id"`
	RoleID string `dbx:"role_id"`
}

type userRoleSchema struct {
	dbx.Schema[userRoleRow]
	UserID dbx.Column[userRoleRow, int64]  `dbx:"user_id"`
	RoleID dbx.Column[userRoleRow, string] `dbx:"role_id"`
}

type userRoleRepo struct {
	urs          userRoleSchema
	userRoleRepo *repository.Base[userRoleRow, userRoleSchema]
}

func NewUserRoleRepository(db *dbx.DB) UserRoleRepository {
	urs := dbx.MustSchema("app_user_roles", userRoleSchema{})
	return &userRoleRepo{
		urs:          urs,
		userRoleRepo: repository.New[userRoleRow](db, urs),
	}
}

func (r *userRoleRepo) ListUserRoleIDs(ctx context.Context, userID int64) ([]string, error) {
	rows, err := r.userRoleRepo.ListSpec(ctx, repository.Where(r.urs.UserID.Eq(userID)))
	if err != nil {
		return nil, err
	}
	return lo.Map(rows, func(row userRoleRow, _ int) string { return row.RoleID }), nil
}

func (r *userRoleRepo) SetUserRoleIDs(ctx context.Context, userID int64, roleIDs []string) error {
	if _, err := r.userRoleRepo.Delete(ctx, dbx.DeleteFrom(r.urs).Where(r.urs.UserID.Eq(userID))); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if roleID == "" {
			continue
		}
		row := userRoleRow{UserID: userID, RoleID: roleID}
		if err := r.userRoleRepo.Create(ctx, &row); err != nil {
			return err
		}
	}
	return nil
}

func (r *userRoleRepo) DeleteUserRoles(ctx context.Context, userID int64) error {
	_, err := r.userRoleRepo.Delete(ctx, dbx.DeleteFrom(r.urs).Where(r.urs.UserID.Eq(userID)))
	return err
}

