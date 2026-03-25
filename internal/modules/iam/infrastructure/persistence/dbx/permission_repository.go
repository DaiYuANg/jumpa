package dbx

import (
	"context"
	"errors"
	"time"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/infrastructure/persistence"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dbx/repository"
	"github.com/samber/lo"
)

type permissionRow struct {
	ID        string    `dbx:"id"`
	Name      string    `dbx:"name"`
	Code      string    `dbx:"code"`
	GroupID   *string   `dbx:"group_id"`
	CreatedAt time.Time `dbx:"created_at,codec=rfc3339_time"`
}

type permissionSchema struct {
	dbx.Schema[permissionRow]
	ID        dbx.Column[permissionRow, string]    `dbx:"id,pk"`
	Name      dbx.Column[permissionRow, string]    `dbx:"name"`
	Code      dbx.Column[permissionRow, string]    `dbx:"code"`
	GroupID   dbx.Column[permissionRow, *string]   `dbx:"group_id"`
	CreatedAt dbx.Column[permissionRow, time.Time] `dbx:"created_at,codec=rfc3339_time"`
}

type permissionRepo struct {
	ps       permissionSchema
	permRepo *repository.Base[permissionRow, permissionSchema]
}

func NewPermissionRepository(db *dbx.DB) persistence.PermissionRepository {
	ps := dbx.MustSchema("app_permissions", permissionSchema{})
	return &permissionRepo{
		ps:       ps,
		permRepo: repository.New[permissionRow](db, ps),
	}
}

func (r *permissionRepo) ListPermissions(ctx context.Context) ([]persistence.Permission, error) {
	rows, err := r.permRepo.ListSpec(ctx, repository.OrderBy(r.ps.ID.Asc()))
	if err != nil {
		return nil, err
	}
	return lo.Map(rows, func(row permissionRow, _ int) persistence.Permission {
		return persistence.Permission{ID: row.ID, Name: row.Name, Code: row.Code, GroupID: row.GroupID, CreatedAt: row.CreatedAt}
	}), nil
}

func (r *permissionRepo) GetPermission(ctx context.Context, id string) (persistence.Permission, bool, error) {
	row, err := r.permRepo.FirstSpec(ctx, repository.Where(r.ps.ID.Eq(id)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return persistence.Permission{}, false, nil
		}
		return persistence.Permission{}, false, err
	}
	return persistence.Permission{ID: row.ID, Name: row.Name, Code: row.Code, GroupID: row.GroupID, CreatedAt: row.CreatedAt}, true, nil
}

func (r *permissionRepo) CreatePermission(ctx context.Context, in persistence.CreatePermissionInput) (persistence.Permission, error) {
	now := time.Now().UTC()
	row := permissionRow{ID: in.ID, Name: in.Name, Code: in.Code, GroupID: in.GroupID, CreatedAt: now}
	if err := r.permRepo.Create(ctx, &row); err != nil {
		return persistence.Permission{}, err
	}
	it, _, err := r.GetPermission(ctx, in.ID)
	return it, err
}

func (r *permissionRepo) UpdatePermission(ctx context.Context, id string, in persistence.PatchPermissionInput) (persistence.Permission, bool, error) {
	assignments := []dbx.Assignment{}
	if in.Name != nil {
		assignments = append(assignments, r.ps.Name.Set(*in.Name))
	}
	if in.Code != nil {
		assignments = append(assignments, r.ps.Code.Set(*in.Code))
	}
	assignments = append(assignments, r.ps.GroupID.Set(in.GroupID))
	res, err := r.permRepo.UpdateByID(ctx, id, assignments...)
	if err != nil {
		return persistence.Permission{}, false, err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return persistence.Permission{}, false, nil
	}
	it, _, err := r.GetPermission(ctx, id)
	return it, true, err
}

func (r *permissionRepo) DeletePermission(ctx context.Context, id string) (bool, error) {
	res, err := r.permRepo.DeleteByID(ctx, id)
	if err != nil {
		return false, err
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}

