package dbx

import (
	"context"
	"time"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/DaiYuANg/arcgo/dbx"
	columnx "github.com/DaiYuANg/arcgo/dbx/column"
	mapperx "github.com/DaiYuANg/arcgo/dbx/mapper"
	"github.com/DaiYuANg/arcgo/dbx/querydsl"
	schemax "github.com/DaiYuANg/arcgo/dbx/schema"
	"github.com/DaiYuANg/jumpa/internal/modules/iam/ports"
	"github.com/samber/mo"
)

type roleRow struct {
	ID          string    `dbx:"id"`
	Name        string    `dbx:"name"`
	Description string    `dbx:"description"`
	CreatedAt   time.Time `dbx:"created_at,codec=rfc3339_time"`
}

type roleSchema struct {
	schemax.Schema[roleRow]
	ID          columnx.Column[roleRow, string]    `dbx:"id,pk"`
	Name        columnx.Column[roleRow, string]    `dbx:"name"`
	Description columnx.Column[roleRow, string]    `dbx:"description"`
	CreatedAt   columnx.Column[roleRow, time.Time] `dbx:"created_at,codec=rfc3339_time"`
}

type rolePermissionGroupRow struct {
	RoleID            string `dbx:"role_id"`
	PermissionGroupID string `dbx:"permission_group_id"`
}

type rolePermissionGroupSchema struct {
	schemax.Schema[rolePermissionGroupRow]
	RoleID            columnx.Column[rolePermissionGroupRow, string] `dbx:"role_id"`
	PermissionGroupID columnx.Column[rolePermissionGroupRow, string] `dbx:"permission_group_id"`
}

type roleRepo struct {
	session dbx.Session
	rs      roleSchema
	mapper  mapperx.Mapper[roleRow]
}

func NewRoleRepository(session dbx.Session) ports.RoleRepository {
	rs := schemax.MustSchema("app_roles", roleSchema{})
	return &roleRepo{
		session: session,
		rs:      rs,
		mapper:  mapperx.MustMapper[roleRow](rs),
	}
}

func (r *roleRepo) ListRoles(ctx context.Context) ([]ports.RoleRecord, error) {
	rows, err := dbx.QueryAll[roleRow](ctx, r.session, querydsl.Select(querydsl.AllColumns(r.rs).Values()...).From(r.rs).OrderBy(r.rs.ID.Asc()), r.mapper)
	if err != nil {
		return nil, err
	}
	return collectionx.MapList(rows, func(_ int, row roleRow) ports.RoleRecord {
		return ports.RoleRecord{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}
	}).Values(), nil
}

func (r *roleRepo) GetRole(ctx context.Context, id string) (mo.Option[ports.RoleRecord], error) {
	rows, err := dbx.QueryAll[roleRow](ctx, r.session, querydsl.Select(querydsl.AllColumns(r.rs).Values()...).From(r.rs).Where(r.rs.ID.Eq(id)), r.mapper)
	if err != nil {
		return mo.None[ports.RoleRecord](), err
	}
	rowOpt := rows.GetFirstOption()
	if rowOpt.IsAbsent() {
		return mo.None[ports.RoleRecord](), nil
	}
	row := rowOpt.MustGet()
	return mo.Some(ports.RoleRecord{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}), nil
}

func (r *roleRepo) CreateRole(ctx context.Context, in ports.CreateRoleInput) (ports.RoleRecord, error) {
	now := time.Now().UTC()
	row := roleRow{ID: in.ID, Name: in.Name, Description: in.Description, CreatedAt: now}
	_, err := dbx.Exec(ctx, r.session, querydsl.InsertInto(r.rs).
		Columns(r.rs.ID, r.rs.Name, r.rs.Description, r.rs.CreatedAt).
		Values(r.rs.ID.Set(row.ID), r.rs.Name.Set(row.Name), r.rs.Description.Set(row.Description), r.rs.CreatedAt.Set(row.CreatedAt)))
	if err != nil {
		return ports.RoleRecord{}, err
	}
	return ports.RoleRecord{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}, nil
}

func (r *roleRepo) UpdateRole(ctx context.Context, id string, in ports.PatchRoleInput) (mo.Option[ports.RoleRecord], error) {
	assignments := []querydsl.Assignment{}
	if in.Name != nil {
		assignments = append(assignments, r.rs.Name.Set(*in.Name))
	}
	if in.Description != nil {
		assignments = append(assignments, r.rs.Description.Set(*in.Description))
	}
	if len(assignments) == 0 {
		return r.GetRole(ctx, id)
	}
	res, err := dbx.Exec(ctx, r.session, querydsl.Update(r.rs).Set(assignments...).Where(r.rs.ID.Eq(id)))
	if err != nil {
		return mo.None[ports.RoleRecord](), err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return mo.None[ports.RoleRecord](), nil
	}
	return r.GetRole(ctx, id)
}

func (r *roleRepo) DeleteRole(ctx context.Context, id string) (bool, error) {
	res, err := dbx.Exec(ctx, r.session, querydsl.DeleteFrom(r.rs).Where(r.rs.ID.Eq(id)))
	if err != nil {
		return false, err
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}
