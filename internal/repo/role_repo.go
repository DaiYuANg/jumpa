package repo

import (
	"context"
	"slices"
	"strings"
	"time"

	collectionmap "github.com/DaiYuANg/arcgo/collectionx/mapping"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/samber/lo"
)

func normalizeIDs(ids []string) []string {
	return lo.Uniq(lo.FilterMap(ids, func(id string, _ int) (string, bool) {
		v := strings.TrimSpace(id)
		return v, v != ""
	}))
}

func (r *roleRepo) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := dbx.QueryAll[roleRow](ctx, r.db, dbx.Select(r.rs.AllColumns()...).From(r.rs).OrderBy(r.rs.ID.Asc()), dbx.MustMapper[roleRow](r.rs))
	if err != nil {
		return nil, err
	}
	pairs, err := dbx.QueryAll[rolePermissionGroupRow](ctx, r.db, dbx.Select(r.rpg.AllColumns()...).From(r.rpg), dbx.MustMapper[rolePermissionGroupRow](r.rpg))
	if err != nil {
		return nil, err
	}
	gm := collectionmap.NewMapWithCapacity[string, []string](len(rows))
	for _, p := range pairs {
		groupIDs, _ := gm.Get(p.RoleID)
		groupIDs = append(groupIDs, p.PermissionGroupID)
		gm.Set(p.RoleID, groupIDs)
	}
	out := lo.Map(rows, func(row roleRow, _ int) Role {
		groupIDs, _ := gm.Get(row.ID)
		return Role{ID: row.ID, Name: row.Name, Description: row.Description, PermissionGroupIDs: slices.Clone(groupIDs), CreatedAt: row.CreatedAt}
	})
	return out, nil
}

func (r *roleRepo) GetRole(ctx context.Context, id string) (Role, bool, error) {
	items, err := r.ListRoles(ctx)
	if err != nil {
		return Role{}, false, err
	}
	for _, it := range items {
		if it.ID == id {
			return it, true, nil
		}
	}
	return Role{}, false, nil
}

func (r *roleRepo) CreateRole(ctx context.Context, in CreateRoleInput) (Role, error) {
	now := time.Now().UTC()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Role{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	_, err = dbx.Exec(ctx, tx, dbx.InsertInto(r.rs).Columns(r.rs.ID, r.rs.Name, r.rs.Description, r.rs.CreatedAt).Values(
		r.rs.ID.Set(in.ID), r.rs.Name.Set(in.Name), r.rs.Description.Set(in.Description), r.rs.CreatedAt.Set(now),
	))
	if err != nil {
		return Role{}, err
	}

	groupIDs := normalizeIDs(in.PermissionGroupIDs)
	if len(groupIDs) > 0 {
		insert := dbx.InsertInto(r.rpg).Columns(r.rpg.RoleID, r.rpg.PermissionGroupID)
		for _, gid := range groupIDs {
			insert = insert.Values(r.rpg.RoleID.Set(in.ID), r.rpg.PermissionGroupID.Set(gid))
		}
		if _, err = dbx.Exec(ctx, tx, insert); err != nil {
			return Role{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Role{}, err
	}
	committed = true
	it, _, err := r.GetRole(ctx, in.ID)
	return it, err
}

func (r *roleRepo) UpdateRole(ctx context.Context, id string, in PatchRoleInput) (Role, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Role{}, false, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	assignments := []dbx.Assignment{}
	if in.Name != nil {
		assignments = append(assignments, r.rs.Name.Set(*in.Name))
	}
	if in.Description != nil {
		assignments = append(assignments, r.rs.Description.Set(*in.Description))
	}
	if len(assignments) > 0 {
		res, err := dbx.Exec(ctx, tx, dbx.Update(r.rs).Set(assignments...).Where(r.rs.ID.Eq(id)))
		if err != nil {
			return Role{}, false, err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return Role{}, false, nil
		}
	}
	if in.PermissionGroupIDs != nil {
		if _, err := dbx.Exec(ctx, tx, dbx.DeleteFrom(r.rpg).Where(r.rpg.RoleID.Eq(id))); err != nil {
			return Role{}, false, err
		}
		groupIDs := normalizeIDs(in.PermissionGroupIDs)
		if len(groupIDs) > 0 {
			insert := dbx.InsertInto(r.rpg).Columns(r.rpg.RoleID, r.rpg.PermissionGroupID)
			for _, gid := range groupIDs {
				insert = insert.Values(r.rpg.RoleID.Set(id), r.rpg.PermissionGroupID.Set(gid))
			}
			if _, err := dbx.Exec(ctx, tx, insert); err != nil {
				return Role{}, false, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return Role{}, false, err
	}
	committed = true
	role, ok, err := r.GetRole(ctx, id)
	return role, ok, err
}

func (r *roleRepo) DeleteRole(ctx context.Context, id string) (bool, error) {
	_, _ = dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.rpg).Where(r.rpg.RoleID.Eq(id)))
	res, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.rs).Where(r.rs.ID.Eq(id)))
	if err != nil {
		return false, err
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}
