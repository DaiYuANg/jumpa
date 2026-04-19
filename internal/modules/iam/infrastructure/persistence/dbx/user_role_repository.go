package dbx

import (
	"context"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/DaiYuANg/arcgo/dbx"
	columnx "github.com/DaiYuANg/arcgo/dbx/column"
	"github.com/DaiYuANg/arcgo/dbx/querydsl"
	"github.com/DaiYuANg/arcgo/dbx/repository"
	schemax "github.com/DaiYuANg/arcgo/dbx/schema"
	"github.com/DaiYuANg/jumpa/internal/modules/iam/ports"
)

type userRoleRow struct {
	UserID int64  `dbx:"user_id"`
	RoleID string `dbx:"role_id"`
}

type userRoleSchema struct {
	schemax.Schema[userRoleRow]
	UserID columnx.Column[userRoleRow, int64]  `dbx:"user_id"`
	RoleID columnx.Column[userRoleRow, string] `dbx:"role_id"`
}

type userRoleRepo struct {
	urs          userRoleSchema
	userRoleRepo *repository.Base[userRoleRow, userRoleSchema]
}

func NewUserRoleRepository(db *dbx.DB) ports.UserRoleRepository {
	urs := schemax.MustSchema("app_user_roles", userRoleSchema{})
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
	return collectionx.MapList(rows, func(_ int, row userRoleRow) string { return row.RoleID }).Values(), nil
}

func (r *userRoleRepo) SetUserRoleIDs(ctx context.Context, userID int64, roleIDs []string) error {
	if _, err := r.userRoleRepo.Delete(ctx, querydsl.DeleteFrom(r.urs).Where(r.urs.UserID.Eq(userID))); err != nil {
		return err
	}
	for _, roleID := range normalizeIDs(roleIDs) {
		row := userRoleRow{UserID: userID, RoleID: roleID}
		if err := r.userRoleRepo.Create(ctx, &row); err != nil {
			return err
		}
	}
	return nil
}

func (r *userRoleRepo) DeleteUserRoles(ctx context.Context, userID int64) error {
	_, err := r.userRoleRepo.Delete(ctx, querydsl.DeleteFrom(r.urs).Where(r.urs.UserID.Eq(userID)))
	return err
}
