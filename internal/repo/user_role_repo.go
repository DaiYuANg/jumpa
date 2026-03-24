package repo

import (
	"context"

	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/samber/lo"
)

func (r *userRoleRepo) ListUserRoleIDs(ctx context.Context, userID int64) ([]string, error) {
	rows, err := dbx.QueryAll[userRoleRow](ctx, r.db, dbx.Select(r.urs.AllColumns()...).From(r.urs).Where(r.urs.UserID.Eq(userID)), dbx.MustMapper[userRoleRow](r.urs))
	if err != nil {
		return nil, err
	}
	return lo.Map(rows, func(row userRoleRow, _ int) string { return row.RoleID }), nil
}

func (r *userRoleRepo) SetUserRoleIDs(ctx context.Context, userID int64, roleIDs []string) error {
	if _, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.urs).Where(r.urs.UserID.Eq(userID))); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if roleID == "" {
			continue
		}
		if _, err := dbx.Exec(ctx, r.db, dbx.InsertInto(r.urs).Columns(r.urs.UserID, r.urs.RoleID).Values(r.urs.UserID.Set(userID), r.urs.RoleID.Set(roleID))); err != nil {
			return err
		}
	}
	return nil
}

func (r *userRoleRepo) DeleteUserRoles(ctx context.Context, userID int64) error {
	_, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.urs).Where(r.urs.UserID.Eq(userID)))
	return err
}
