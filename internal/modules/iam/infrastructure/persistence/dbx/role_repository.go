package dbx

import (
	"context"
	"time"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/infrastructure/persistence"
	"github.com/DaiYuANg/arcgo/dbx"
)

type roleRow struct {
	ID          string    `dbx:"id"`
	Name        string    `dbx:"name"`
	Description string    `dbx:"description"`
	CreatedAt   time.Time `dbx:"created_at,codec=rfc3339_time"`
}

type roleSchema struct {
	dbx.Schema[roleRow]
	ID          dbx.Column[roleRow, string]    `dbx:"id,pk"`
	Name        dbx.Column[roleRow, string]    `dbx:"name"`
	Description dbx.Column[roleRow, string]    `dbx:"description"`
	CreatedAt   dbx.Column[roleRow, time.Time] `dbx:"created_at,codec=rfc3339_time"`
}

type rolePermissionGroupRow struct {
	RoleID            string `dbx:"role_id"`
	PermissionGroupID string `dbx:"permission_group_id"`
}

type rolePermissionGroupSchema struct {
	dbx.Schema[rolePermissionGroupRow]
	RoleID            dbx.Column[rolePermissionGroupRow, string] `dbx:"role_id"`
	PermissionGroupID dbx.Column[rolePermissionGroupRow, string] `dbx:"permission_group_id"`
}

type roleRepo struct {
	session  dbx.Session
	rs       roleSchema
	mapper   dbx.Mapper[roleRow]
}

func NewRoleRepository(session dbx.Session) persistence.RoleRepository {
	rs := dbx.MustSchema("app_roles", roleSchema{})
	return &roleRepo{
		session: session,
		rs:     rs,
		mapper: dbx.MustMapper[roleRow](rs),
	}
}

func (r *roleRepo) ListRoles(ctx context.Context) ([]persistence.RoleRecord, error) {
	rows, err := dbx.QueryAll[roleRow](ctx, r.session, dbx.Select(r.rs.AllColumns()...).From(r.rs).OrderBy(r.rs.ID.Asc()), r.mapper)
	if err != nil {
		return nil, err
	}
	out := make([]persistence.RoleRecord, len(rows))
	for i, row := range rows {
		out[i] = persistence.RoleRecord{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}
	}
	return out, nil
}

func (r *roleRepo) GetRole(ctx context.Context, id string) (persistence.RoleRecord, bool, error) {
	rows, err := dbx.QueryAll[roleRow](ctx, r.session, dbx.Select(r.rs.AllColumns()...).From(r.rs).Where(r.rs.ID.Eq(id)), r.mapper)
	if err != nil {
		return persistence.RoleRecord{}, false, err
	}
	if len(rows) == 0 {
		return persistence.RoleRecord{}, false, nil
	}
	row := rows[0]
	return persistence.RoleRecord{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}, true, nil
}

func (r *roleRepo) CreateRole(ctx context.Context, in persistence.CreateRoleInput) (persistence.RoleRecord, error) {
	now := time.Now().UTC()
	row := roleRow{ID: in.ID, Name: in.Name, Description: in.Description, CreatedAt: now}
	_, err := dbx.Exec(ctx, r.session, dbx.InsertInto(r.rs).
		Columns(r.rs.ID, r.rs.Name, r.rs.Description, r.rs.CreatedAt).
		Values(r.rs.ID.Set(row.ID), r.rs.Name.Set(row.Name), r.rs.Description.Set(row.Description), r.rs.CreatedAt.Set(row.CreatedAt)))
	if err != nil {
		return persistence.RoleRecord{}, err
	}
	return persistence.RoleRecord{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}, nil
}

func (r *roleRepo) UpdateRole(ctx context.Context, id string, in persistence.PatchRoleInput) (persistence.RoleRecord, bool, error) {
	assignments := []dbx.Assignment{}
	if in.Name != nil {
		assignments = append(assignments, r.rs.Name.Set(*in.Name))
	}
	if in.Description != nil {
		assignments = append(assignments, r.rs.Description.Set(*in.Description))
	}
	if len(assignments) == 0 {
		it, ok, err := r.GetRole(ctx, id)
		return it, ok, err
	}
	res, err := dbx.Exec(ctx, r.session, dbx.Update(r.rs).Set(assignments...).Where(r.rs.ID.Eq(id)))
	if err != nil {
		return persistence.RoleRecord{}, false, err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return persistence.RoleRecord{}, false, nil
	}
	it, ok, err := r.GetRole(ctx, id)
	if err != nil {
		return persistence.RoleRecord{}, false, err
	}
	if !ok {
		return persistence.RoleRecord{}, false, nil
	}
	return it, true, nil
}

func (r *roleRepo) DeleteRole(ctx context.Context, id string) (bool, error) {
	res, err := dbx.Exec(ctx, r.session, dbx.DeleteFrom(r.rs).Where(r.rs.ID.Eq(id)))
	if err != nil {
		return false, err
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}

