package dbx

import (
	"context"

	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/samber/lo"
)

type rolePermissionGroupRepo struct {
	rpg    rolePermissionGroupSchema
	mapper dbx.Mapper[rolePermissionGroupRow]
}

func NewRolePermissionGroupRepository(_ *dbx.DB) RolePermissionGroupRepository {
	rpg := dbx.MustSchema("app_role_permission_groups", rolePermissionGroupSchema{})
	return &rolePermissionGroupRepo{
		rpg:    rpg,
		mapper: dbx.MustMapper[rolePermissionGroupRow](rpg),
	}
}

func (r *rolePermissionGroupRepo) ListPairs(ctx context.Context, session dbx.Session) ([]RolePermissionGroupPair, error) {
	rows, err := dbx.QueryAll[rolePermissionGroupRow](ctx, session, dbx.Select(r.rpg.AllColumns()...).From(r.rpg), r.mapper)
	if err != nil {
		return nil, err
	}
	return lo.Map(rows, func(row rolePermissionGroupRow, _ int) RolePermissionGroupPair {
		return RolePermissionGroupPair{RoleID: row.RoleID, PermissionGroupID: row.PermissionGroupID}
	}), nil
}

func (r *rolePermissionGroupRepo) ListPairsByRoleIDs(ctx context.Context, session dbx.Session, roleIDs []string) ([]RolePermissionGroupPair, error) {
	ids := normalizeIDs(roleIDs)
	if len(ids) == 0 {
		return []RolePermissionGroupPair{}, nil
	}
	rows, err := dbx.QueryAll[rolePermissionGroupRow](
		ctx,
		session,
		dbx.Select(r.rpg.AllColumns()...).From(r.rpg).Where(r.rpg.RoleID.In(ids...)),
		r.mapper,
	)
	if err != nil {
		return nil, err
	}
	return lo.Map(rows, func(row rolePermissionGroupRow, _ int) RolePermissionGroupPair {
		return RolePermissionGroupPair{RoleID: row.RoleID, PermissionGroupID: row.PermissionGroupID}
	}), nil
}

func (r *rolePermissionGroupRepo) ListPermissionGroupIDsByRoleID(ctx context.Context, session dbx.Session, roleID string) ([]string, error) {
	rows, err := dbx.QueryAll[rolePermissionGroupRow](ctx, session, dbx.Select(r.rpg.AllColumns()...).From(r.rpg).Where(r.rpg.RoleID.Eq(roleID)), r.mapper)
	if err != nil {
		return nil, err
	}
	return lo.Map(rows, func(row rolePermissionGroupRow, _ int) string { return row.PermissionGroupID }), nil
}

func (r *rolePermissionGroupRepo) ReplacePermissionGroupIDs(ctx context.Context, session dbx.Session, roleID string, permissionGroupIDs []string) error {
	if _, err := dbx.Exec(ctx, session, dbx.DeleteFrom(r.rpg).Where(r.rpg.RoleID.Eq(roleID))); err != nil {
		return err
	}
	ids := normalizeIDs(permissionGroupIDs)
	if len(ids) == 0 {
		return nil
	}
	insert := dbx.InsertInto(r.rpg).Columns(r.rpg.RoleID, r.rpg.PermissionGroupID)
	for _, gid := range ids {
		insert = insert.Values(r.rpg.RoleID.Set(roleID), r.rpg.PermissionGroupID.Set(gid))
	}
	_, err := dbx.Exec(ctx, session, insert)
	return err
}

func (r *rolePermissionGroupRepo) DeleteByRoleID(ctx context.Context, session dbx.Session, roleID string) error {
	_, err := dbx.Exec(ctx, session, dbx.DeleteFrom(r.rpg).Where(r.rpg.RoleID.Eq(roleID)))
	return err
}

