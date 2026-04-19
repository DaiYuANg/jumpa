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

type permissionGroupRow struct {
	ID          string    `dbx:"id"`
	Name        string    `dbx:"name"`
	Description string    `dbx:"description"`
	CreatedAt   time.Time `dbx:"created_at,codec=rfc3339_time"`
}

type permissionGroupSchema struct {
	schemax.Schema[permissionGroupRow]
	ID          columnx.Column[permissionGroupRow, string]    `dbx:"id,pk"`
	Name        columnx.Column[permissionGroupRow, string]    `dbx:"name"`
	Description columnx.Column[permissionGroupRow, string]    `dbx:"description"`
	CreatedAt   columnx.Column[permissionGroupRow, time.Time] `dbx:"created_at,codec=rfc3339_time"`
}

type permissionGroupRepo struct {
	pgs       permissionGroupSchema
	groupRepo *repository.Base[permissionGroupRow, permissionGroupSchema]
}

func NewPermissionGroupRepository(db *dbx.DB) ports.PermissionGroupRepository {
	pgs := schemax.MustSchema("app_permission_groups", permissionGroupSchema{})
	return &permissionGroupRepo{
		pgs:       pgs,
		groupRepo: repository.New[permissionGroupRow](db, pgs),
	}
}

func (r *permissionGroupRepo) ListPermissionGroups(ctx context.Context) ([]ports.PermissionGroup, error) {
	rows, err := r.groupRepo.ListSpec(ctx, repository.OrderBy(r.pgs.ID.Asc()))
	if err != nil {
		return nil, err
	}
	return collectionx.MapList(rows, func(_ int, row permissionGroupRow) ports.PermissionGroup {
		return ports.PermissionGroup{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}
	}).Values(), nil
}

func (r *permissionGroupRepo) GetPermissionGroup(ctx context.Context, id string) (mo.Option[ports.PermissionGroup], error) {
	row, err := r.groupRepo.FirstSpec(ctx, repository.Where(r.pgs.ID.Eq(id)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return mo.None[ports.PermissionGroup](), nil
		}
		return mo.None[ports.PermissionGroup](), err
	}
	return mo.Some(ports.PermissionGroup{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}), nil
}

func (r *permissionGroupRepo) CreatePermissionGroup(ctx context.Context, in ports.CreatePermissionGroupInput) (ports.PermissionGroup, error) {
	now := time.Now().UTC()
	row := permissionGroupRow{ID: in.ID, Name: in.Name, Description: in.Description, CreatedAt: now}
	if err := r.groupRepo.Create(ctx, &row); err != nil {
		return ports.PermissionGroup{}, err
	}
	it, err := r.GetPermissionGroup(ctx, in.ID)
	if err != nil {
		return ports.PermissionGroup{}, err
	}
	return it.MustGet(), nil
}

func (r *permissionGroupRepo) UpdatePermissionGroup(ctx context.Context, id string, in ports.PatchPermissionGroupInput) (mo.Option[ports.PermissionGroup], error) {
	var assignments []querydsl.Assignment
	if in.Name != nil {
		assignments = append(assignments, r.pgs.Name.Set(*in.Name))
	}
	if in.Description != nil {
		assignments = append(assignments, r.pgs.Description.Set(*in.Description))
	}
	if len(assignments) > 0 {
		res, err := r.groupRepo.UpdateByID(ctx, id, assignments...)
		if err != nil {
			return mo.None[ports.PermissionGroup](), err
		}
		ra, _ := res.RowsAffected()
		if ra == 0 {
			return mo.None[ports.PermissionGroup](), nil
		}
	}
	return r.GetPermissionGroup(ctx, id)
}

func (r *permissionGroupRepo) DeletePermissionGroup(ctx context.Context, id string) (bool, error) {
	res, err := r.groupRepo.DeleteByID(ctx, id)
	if err != nil {
		return false, err
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}
