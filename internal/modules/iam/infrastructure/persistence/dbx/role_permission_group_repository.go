package dbx

import (
	"context"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/jumpa/internal/modules/iam/ports"
)

type rolePermissionGroupRepo struct {
	session dbx.Session
	rpg     rolePermissionGroupSchema
	mapper  dbx.Mapper[rolePermissionGroupRow]
}

func NewRolePermissionGroupRepository(session dbx.Session) ports.RolePermissionGroupRepository {
	rpg := dbx.MustSchema("app_role_permission_groups", rolePermissionGroupSchema{})
	return &rolePermissionGroupRepo{
		session: session,
		rpg:     rpg,
		mapper:  dbx.MustMapper[rolePermissionGroupRow](rpg),
	}
}

func (r *rolePermissionGroupRepo) ListPairs(ctx context.Context) ([]ports.RolePermissionGroupPair, error) {
	rows, err := dbx.QueryAll[rolePermissionGroupRow](ctx, r.session, dbx.Select(r.rpg.AllColumns().Values()...).From(r.rpg), r.mapper)
	if err != nil {
		return nil, err
	}
	return collectionx.MapList(rows, func(_ int, row rolePermissionGroupRow) ports.RolePermissionGroupPair {
		return ports.RolePermissionGroupPair{RoleID: row.RoleID, PermissionGroupID: row.PermissionGroupID}
	}).Values(), nil
}

func (r *rolePermissionGroupRepo) ListPairsByRoleIDs(ctx context.Context, roleIDs []string) ([]ports.RolePermissionGroupPair, error) {
	ids := normalizeIDs(roleIDs)
	if len(ids) == 0 {
		return []ports.RolePermissionGroupPair{}, nil
	}
	rows, err := dbx.QueryAll[rolePermissionGroupRow](
		ctx,
		r.session,
		dbx.Select(r.rpg.AllColumns().Values()...).From(r.rpg).Where(r.rpg.RoleID.In(ids...)),
		r.mapper,
	)
	if err != nil {
		return nil, err
	}
	return collectionx.MapList(rows, func(_ int, row rolePermissionGroupRow) ports.RolePermissionGroupPair {
		return ports.RolePermissionGroupPair{RoleID: row.RoleID, PermissionGroupID: row.PermissionGroupID}
	}).Values(), nil
}

func (r *rolePermissionGroupRepo) ListPermissionGroupIDsByRoleID(ctx context.Context, roleID string) ([]string, error) {
	rows, err := dbx.QueryAll[rolePermissionGroupRow](ctx, r.session, dbx.Select(r.rpg.AllColumns().Values()...).From(r.rpg).Where(r.rpg.RoleID.Eq(roleID)), r.mapper)
	if err != nil {
		return nil, err
	}
	return collectionx.MapList(rows, func(_ int, row rolePermissionGroupRow) string { return row.PermissionGroupID }).Values(), nil
}

func (r *rolePermissionGroupRepo) ReplacePermissionGroupIDs(ctx context.Context, roleID string, permissionGroupIDs []string) error {
	if _, err := dbx.Exec(ctx, r.session, dbx.DeleteFrom(r.rpg).Where(r.rpg.RoleID.Eq(roleID))); err != nil {
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
	_, err := dbx.Exec(ctx, r.session, insert)
	return err
}

func (r *rolePermissionGroupRepo) DeleteByRoleID(ctx context.Context, roleID string) error {
	_, err := dbx.Exec(ctx, r.session, dbx.DeleteFrom(r.rpg).Where(r.rpg.RoleID.Eq(roleID)))
	return err
}
