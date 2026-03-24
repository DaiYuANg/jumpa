package repo

import (
	"context"
	"time"

	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/samber/lo"
)

func (r *permissionRepo) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := dbx.QueryAll[permissionRow](ctx, r.db, dbx.Select(r.ps.AllColumns()...).From(r.ps).OrderBy(r.ps.ID.Asc()), dbx.MustMapper[permissionRow](r.ps))
	if err != nil {
		return nil, err
	}
	return lo.Map(rows, func(row permissionRow, _ int) Permission {
		return Permission{ID: row.ID, Name: row.Name, Code: row.Code, GroupID: row.GroupID, CreatedAt: row.CreatedAt}
	}), nil
}

func (r *permissionRepo) GetPermission(ctx context.Context, id string) (Permission, bool, error) {
	rows, err := dbx.QueryAll[permissionRow](ctx, r.db, dbx.Select(r.ps.AllColumns()...).From(r.ps).Where(r.ps.ID.Eq(id)), dbx.MustMapper[permissionRow](r.ps))
	if err != nil {
		return Permission{}, false, err
	}
	if len(rows) == 0 {
		return Permission{}, false, nil
	}
	row := rows[0]
	return Permission{ID: row.ID, Name: row.Name, Code: row.Code, GroupID: row.GroupID, CreatedAt: row.CreatedAt}, true, nil
}

func (r *permissionRepo) CreatePermission(ctx context.Context, in CreatePermissionInput) (Permission, error) {
	now := time.Now().UTC()
	_, err := dbx.Exec(ctx, r.db, dbx.InsertInto(r.ps).Columns(r.ps.ID, r.ps.Name, r.ps.Code, r.ps.GroupID, r.ps.CreatedAt).Values(
		r.ps.ID.Set(in.ID), r.ps.Name.Set(in.Name), r.ps.Code.Set(in.Code), r.ps.GroupID.Set(in.GroupID), r.ps.CreatedAt.Set(now),
	))
	if err != nil {
		return Permission{}, err
	}
	it, _, err := r.GetPermission(ctx, in.ID)
	return it, err
}

func (r *permissionRepo) UpdatePermission(ctx context.Context, id string, in PatchPermissionInput) (Permission, bool, error) {
	assignments := []dbx.Assignment{}
	if in.Name != nil {
		assignments = append(assignments, r.ps.Name.Set(*in.Name))
	}
	if in.Code != nil {
		assignments = append(assignments, r.ps.Code.Set(*in.Code))
	}
	assignments = append(assignments, r.ps.GroupID.Set(in.GroupID))
	res, err := dbx.Exec(ctx, r.db, dbx.Update(r.ps).Set(assignments...).Where(r.ps.ID.Eq(id)))
	if err != nil {
		return Permission{}, false, err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return Permission{}, false, nil
	}
	it, _, err := r.GetPermission(ctx, id)
	return it, true, err
}

func (r *permissionRepo) DeletePermission(ctx context.Context, id string) (bool, error) {
	res, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.ps).Where(r.ps.ID.Eq(id)))
	if err != nil {
		return false, err
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}
