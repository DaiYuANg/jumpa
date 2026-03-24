package repo

import (
	"context"
	"time"

	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/samber/lo"
)

func (r *permissionGroupRepo) ListPermissionGroups(ctx context.Context) ([]PermissionGroup, error) {
	rows, err := dbx.QueryAll[permissionGroupRow](ctx, r.db, dbx.Select(r.pgs.AllColumns()...).From(r.pgs).OrderBy(r.pgs.ID.Asc()), dbx.MustMapper[permissionGroupRow](r.pgs))
	if err != nil {
		return nil, err
	}
	return lo.Map(rows, func(row permissionGroupRow, _ int) PermissionGroup {
		return PermissionGroup{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}
	}), nil
}

func (r *permissionGroupRepo) GetPermissionGroup(ctx context.Context, id string) (PermissionGroup, bool, error) {
	rows, err := dbx.QueryAll[permissionGroupRow](ctx, r.db, dbx.Select(r.pgs.AllColumns()...).From(r.pgs).Where(r.pgs.ID.Eq(id)), dbx.MustMapper[permissionGroupRow](r.pgs))
	if err != nil {
		return PermissionGroup{}, false, err
	}
	if len(rows) == 0 {
		return PermissionGroup{}, false, nil
	}
	row := rows[0]
	return PermissionGroup{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}, true, nil
}

func (r *permissionGroupRepo) CreatePermissionGroup(ctx context.Context, in CreatePermissionGroupInput) (PermissionGroup, error) {
	now := time.Now().UTC()
	_, err := dbx.Exec(ctx, r.db, dbx.InsertInto(r.pgs).Columns(r.pgs.ID, r.pgs.Name, r.pgs.Description, r.pgs.CreatedAt).Values(
		r.pgs.ID.Set(in.ID), r.pgs.Name.Set(in.Name), r.pgs.Description.Set(in.Description), r.pgs.CreatedAt.Set(now),
	))
	if err != nil {
		return PermissionGroup{}, err
	}
	it, _, err := r.GetPermissionGroup(ctx, in.ID)
	return it, err
}

func (r *permissionGroupRepo) UpdatePermissionGroup(ctx context.Context, id string, in PatchPermissionGroupInput) (PermissionGroup, bool, error) {
	assignments := []dbx.Assignment{}
	if in.Name != nil {
		assignments = append(assignments, r.pgs.Name.Set(*in.Name))
	}
	if in.Description != nil {
		assignments = append(assignments, r.pgs.Description.Set(*in.Description))
	}
	if len(assignments) > 0 {
		res, err := dbx.Exec(ctx, r.db, dbx.Update(r.pgs).Set(assignments...).Where(r.pgs.ID.Eq(id)))
		if err != nil {
			return PermissionGroup{}, false, err
		}
		ra, _ := res.RowsAffected()
		if ra == 0 {
			return PermissionGroup{}, false, nil
		}
	}
	it, ok, err := r.GetPermissionGroup(ctx, id)
	return it, ok, err
}

func (r *permissionGroupRepo) DeletePermissionGroup(ctx context.Context, id string) (bool, error) {
	res, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.pgs).Where(r.pgs.ID.Eq(id)))
	if err != nil {
		return false, err
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}
