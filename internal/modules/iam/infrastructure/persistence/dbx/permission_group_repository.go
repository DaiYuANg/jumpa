package dbx

import (
  "context"
  "errors"
  "time"

  "github.com/DaiYuANg/arcgo/dbx"
  "github.com/DaiYuANg/arcgo/dbx/repository"
  "github.com/samber/lo"
)

type permissionGroupRow struct {
  ID          string    `dbx:"id"`
  Name        string    `dbx:"name"`
  Description string    `dbx:"description"`
  CreatedAt   time.Time `dbx:"created_at,codec=rfc3339_time"`
}

type permissionGroupSchema struct {
  dbx.Schema[permissionGroupRow]
  ID          dbx.Column[permissionGroupRow, string]    `dbx:"id,pk"`
  Name        dbx.Column[permissionGroupRow, string]    `dbx:"name"`
  Description dbx.Column[permissionGroupRow, string]    `dbx:"description"`
  CreatedAt   dbx.Column[permissionGroupRow, time.Time] `dbx:"created_at,codec=rfc3339_time"`
}

type permissionGroupRepo struct {
  pgs       permissionGroupSchema
  groupRepo *repository.Base[permissionGroupRow, permissionGroupSchema]
}

func NewPermissionGroupRepository(db *dbx.DB) PermissionGroupRepository {
  pgs := dbx.MustSchema("app_permission_groups", permissionGroupSchema{})
  return &permissionGroupRepo{
    pgs:       pgs,
    groupRepo: repository.New[permissionGroupRow](db, pgs),
  }
}

func (r *permissionGroupRepo) ListPermissionGroups(ctx context.Context) ([]PermissionGroup, error) {
  rows, err := r.groupRepo.ListSpec(ctx, repository.OrderBy(r.pgs.ID.Asc()))
  if err != nil {
    return nil, err
  }
  return lo.Map(rows, func(row permissionGroupRow, _ int) PermissionGroup {
    return PermissionGroup{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}
  }), nil
}

func (r *permissionGroupRepo) GetPermissionGroup(ctx context.Context, id string) (PermissionGroup, bool, error) {
  row, err := r.groupRepo.FirstSpec(ctx, repository.Where(r.pgs.ID.Eq(id)))
  if err != nil {
    if errors.Is(err, repository.ErrNotFound) {
      return PermissionGroup{}, false, nil
    }
    return PermissionGroup{}, false, err
  }
  return PermissionGroup{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}, true, nil
}

func (r *permissionGroupRepo) CreatePermissionGroup(ctx context.Context, in CreatePermissionGroupInput) (PermissionGroup, error) {
  now := time.Now().UTC()
  row := permissionGroupRow{ID: in.ID, Name: in.Name, Description: in.Description, CreatedAt: now}
  if err := r.groupRepo.Create(ctx, &row); err != nil {
    return PermissionGroup{}, err
  }
  it, _, err := r.GetPermissionGroup(ctx, in.ID)
  return it, err
}

func (r *permissionGroupRepo) UpdatePermissionGroup(ctx context.Context, id string, in PatchPermissionGroupInput) (PermissionGroup, bool, error) {
  var assignments []dbx.Assignment
  if in.Name != nil {
    assignments = append(assignments, r.pgs.Name.Set(*in.Name))
  }
  if in.Description != nil {
    assignments = append(assignments, r.pgs.Description.Set(*in.Description))
  }
  if len(assignments) > 0 {
    res, err := r.groupRepo.UpdateByID(ctx, id, assignments...)
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
  res, err := r.groupRepo.DeleteByID(ctx, id)
  if err != nil {
    return false, err
  }
  ra, _ := res.RowsAffected()
  return ra > 0, nil
}
