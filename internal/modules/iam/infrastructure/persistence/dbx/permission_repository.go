package dbx

import (
	"context"
	"errors"
	"time"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/DaiYuANg/arcgo/dbx"
	columnx "github.com/DaiYuANg/arcgo/dbx/column"
	"github.com/DaiYuANg/arcgo/dbx/querydsl"
	"github.com/DaiYuANg/arcgo/dbx/repository"
	schemax "github.com/DaiYuANg/arcgo/dbx/schema"
	"github.com/DaiYuANg/jumpa/internal/modules/iam/ports"
	"github.com/samber/mo"
)

type permissionRow struct {
	ID        string    `dbx:"id"`
	Name      string    `dbx:"name"`
	Code      string    `dbx:"code"`
	GroupID   *string   `dbx:"group_id"`
	CreatedAt time.Time `dbx:"created_at,codec=rfc3339_time"`
}

type permissionSchema struct {
	schemax.Schema[permissionRow]
	ID        columnx.Column[permissionRow, string]    `dbx:"id,pk"`
	Name      columnx.Column[permissionRow, string]    `dbx:"name"`
	Code      columnx.Column[permissionRow, string]    `dbx:"code"`
	GroupID   columnx.Column[permissionRow, *string]   `dbx:"group_id"`
	CreatedAt columnx.Column[permissionRow, time.Time] `dbx:"created_at,codec=rfc3339_time"`
}

type permissionRepo struct {
	ps       permissionSchema
	permRepo *repository.Base[permissionRow, permissionSchema]
}

func NewPermissionRepository(db *dbx.DB) ports.PermissionRepository {
	ps := schemax.MustSchema("app_permissions", permissionSchema{})
	return &permissionRepo{
		ps:       ps,
		permRepo: repository.New[permissionRow](db, ps),
	}
}

func (r *permissionRepo) ListPermissions(ctx context.Context) ([]ports.Permission, error) {
	rows, err := r.permRepo.ListSpec(ctx, repository.OrderBy(r.ps.ID.Asc()))
	if err != nil {
		return nil, err
	}
	return collectionx.MapList(rows, func(_ int, row permissionRow) ports.Permission {
		return ports.Permission{ID: row.ID, Name: row.Name, Code: row.Code, GroupID: row.GroupID, CreatedAt: row.CreatedAt}
	}).Values(), nil
}

func (r *permissionRepo) GetPermission(ctx context.Context, id string) (mo.Option[ports.Permission], error) {
	row, err := r.permRepo.FirstSpec(ctx, repository.Where(r.ps.ID.Eq(id)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return mo.None[ports.Permission](), nil
		}
		return mo.None[ports.Permission](), err
	}
	return mo.Some(ports.Permission{ID: row.ID, Name: row.Name, Code: row.Code, GroupID: row.GroupID, CreatedAt: row.CreatedAt}), nil
}

func (r *permissionRepo) CreatePermission(ctx context.Context, in ports.CreatePermissionInput) (ports.Permission, error) {
	now := time.Now().UTC()
	row := permissionRow{ID: in.ID, Name: in.Name, Code: in.Code, GroupID: in.GroupID, CreatedAt: now}
	if err := r.permRepo.Create(ctx, &row); err != nil {
		return ports.Permission{}, err
	}
	it, err := r.GetPermission(ctx, in.ID)
	if err != nil {
		return ports.Permission{}, err
	}
	return it.MustGet(), nil
}

func (r *permissionRepo) UpdatePermission(ctx context.Context, id string, in ports.PatchPermissionInput) (mo.Option[ports.Permission], error) {
	assignments := []querydsl.Assignment{}
	if in.Name != nil {
		assignments = append(assignments, r.ps.Name.Set(*in.Name))
	}
	if in.Code != nil {
		assignments = append(assignments, r.ps.Code.Set(*in.Code))
	}
	assignments = append(assignments, r.ps.GroupID.Set(in.GroupID))
	res, err := r.permRepo.UpdateByID(ctx, id, assignments...)
	if err != nil {
		return mo.None[ports.Permission](), err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return mo.None[ports.Permission](), nil
	}
	return r.GetPermission(ctx, id)
}

func (r *permissionRepo) DeletePermission(ctx context.Context, id string) (bool, error) {
	res, err := r.permRepo.DeleteByID(ctx, id)
	if err != nil {
		return false, err
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}
