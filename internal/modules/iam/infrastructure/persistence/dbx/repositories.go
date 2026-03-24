package dbx

import (
	"context"
	"fmt"
	"strings"
	"time"

	legacydomain "github.com/DaiYuANg/arcgo-rbac-template/internal/domain"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/schema"
	collectionmap "github.com/DaiYuANg/arcgo/collectionx/mapping"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/samber/lo"
	"slices"
)

type UserRow = schema.UserRow
type UserSchema = schema.UserSchema

type UserRepository interface {
	List(ctx context.Context, search string, limit, offset int) ([]legacydomain.User, int, error)
	GetByID(ctx context.Context, id int64) (legacydomain.User, bool, error)
	Create(ctx context.Context, in legacydomain.CreateUserInput) (legacydomain.User, error)
	Update(ctx context.Context, id int64, in legacydomain.UpdateUserInput) (legacydomain.User, bool, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type userRepo struct {
	db     *dbx.DB
	schema UserSchema
}

func NewUserRepository(db *dbx.DB, s UserSchema) UserRepository { return &userRepo{db: db, schema: s} }

func (r *userRepo) List(ctx context.Context, search string, limit, offset int) ([]legacydomain.User, int, error) {
	s := r.schema
	mapper := dbx.MustMapper[UserRow](s)
	q := dbx.Select(s.AllColumns()...).From(s)
	if search != "" {
		pattern := "%" + strings.TrimSpace(search) + "%"
		q = q.Where(dbx.Or(dbx.Like(s.Name, pattern), dbx.Like(s.Email, pattern)))
	}
	q = q.OrderBy(s.ID.Asc())
	all, err := dbx.QueryAll[UserRow](ctx, r.db, q, mapper)
	if err != nil {
		return nil, 0, err
	}
	total := len(all)
	if offset >= total {
		return []legacydomain.User{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	page := all[offset:end]
	users := make([]legacydomain.User, len(page))
	for i, row := range page {
		users[i] = legacydomain.User{ID: row.ID, Name: row.Name, Email: row.Email, Age: row.Age, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
	}
	return users, total, nil
}
func (r *userRepo) GetByID(ctx context.Context, id int64) (legacydomain.User, bool, error) {
	s := r.schema
	rows, err := dbx.QueryAll[UserRow](ctx, r.db, dbx.Select(s.AllColumns()...).From(s).Where(s.ID.Eq(id)), dbx.MustMapper[UserRow](s))
	if err != nil {
		return legacydomain.User{}, false, err
	}
	if len(rows) == 0 {
		return legacydomain.User{}, false, nil
	}
	row := rows[0]
	return legacydomain.User{ID: row.ID, Name: row.Name, Email: row.Email, Age: row.Age, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, true, nil
}
func (r *userRepo) Create(ctx context.Context, in legacydomain.CreateUserInput) (legacydomain.User, error) {
	s := r.schema
	now := time.Now().UTC()
	rows, err := dbx.QueryAll[UserRow](ctx, r.db, dbx.InsertInto(s).Columns(s.Name, s.Email, s.Age, s.CreatedAt, s.UpdatedAt).
		Values(s.Name.Set(in.Name), s.Email.Set(in.Email), s.Age.Set(in.Age), s.CreatedAt.Set(now), s.UpdatedAt.Set(now)).Returning(s.AllColumns()...), dbx.MustMapper[UserRow](s))
	if err != nil {
		return legacydomain.User{}, err
	}
	row := rows[0]
	return legacydomain.User{ID: row.ID, Name: row.Name, Email: row.Email, Age: row.Age, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}
func (r *userRepo) Update(ctx context.Context, id int64, in legacydomain.UpdateUserInput) (legacydomain.User, bool, error) {
	s := r.schema
	assignments := []dbx.Assignment{s.UpdatedAt.Set(time.Now().UTC())}
	if in.Name != nil {
		assignments = append(assignments, s.Name.Set(*in.Name))
	}
	if in.Email != nil {
		assignments = append(assignments, s.Email.Set(*in.Email))
	}
	if in.Age != nil {
		assignments = append(assignments, s.Age.Set(*in.Age))
	}
	rows, err := dbx.QueryAll[UserRow](ctx, r.db, dbx.Update(s).Set(assignments...).Where(s.ID.Eq(id)).Returning(s.AllColumns()...), dbx.MustMapper[UserRow](s))
	if err != nil {
		return legacydomain.User{}, false, err
	}
	if len(rows) == 0 {
		return legacydomain.User{}, false, nil
	}
	row := rows[0]
	return legacydomain.User{ID: row.ID, Name: row.Name, Email: row.Email, Age: row.Age, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, true, nil
}
func (r *userRepo) Delete(ctx context.Context, id int64) (bool, error) {
	res, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.schema).Where(r.schema.ID.Eq(id)))
	if err != nil {
		return false, err
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}

type Role struct {
	ID                 string
	Name               string
	Description        string
	PermissionGroupIDs []string
	CreatedAt          time.Time
}
type PermissionGroup struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}
type Permission struct {
	ID        string
	Name      string
	Code      string
	GroupID   *string
	CreatedAt time.Time
}
type CreateRoleInput struct {
	ID                 string
	Name               string
	Description        string
	PermissionGroupIDs []string
}
type PatchRoleInput struct {
	Name               *string
	Description        *string
	PermissionGroupIDs []string
}
type CreatePermissionGroupInput struct{ ID, Name, Description string }
type PatchPermissionGroupInput struct{ Name, Description *string }
type CreatePermissionInput struct {
	ID      string
	Name    string
	Code    string
	GroupID *string
}
type PatchPermissionInput struct {
	Name    *string
	Code    *string
	GroupID *string
}

type RoleRepository interface {
	ListRoles(ctx context.Context) ([]Role, error)
	GetRole(ctx context.Context, id string) (Role, bool, error)
	CreateRole(ctx context.Context, in CreateRoleInput) (Role, error)
	UpdateRole(ctx context.Context, id string, in PatchRoleInput) (Role, bool, error)
	DeleteRole(ctx context.Context, id string) (bool, error)
}
type PermissionGroupRepository interface {
	ListPermissionGroups(ctx context.Context) ([]PermissionGroup, error)
	GetPermissionGroup(ctx context.Context, id string) (PermissionGroup, bool, error)
	CreatePermissionGroup(ctx context.Context, in CreatePermissionGroupInput) (PermissionGroup, error)
	UpdatePermissionGroup(ctx context.Context, id string, in PatchPermissionGroupInput) (PermissionGroup, bool, error)
	DeletePermissionGroup(ctx context.Context, id string) (bool, error)
}
type PermissionRepository interface {
	ListPermissions(ctx context.Context) ([]Permission, error)
	GetPermission(ctx context.Context, id string) (Permission, bool, error)
	CreatePermission(ctx context.Context, in CreatePermissionInput) (Permission, error)
	UpdatePermission(ctx context.Context, id string, in PatchPermissionInput) (Permission, bool, error)
	DeletePermission(ctx context.Context, id string) (bool, error)
}
type UserRoleRepository interface {
	ListUserRoleIDs(ctx context.Context, userID int64) ([]string, error)
	SetUserRoleIDs(ctx context.Context, userID int64, roleIDs []string) error
	DeleteUserRoles(ctx context.Context, userID int64) error
}
type AuthPrincipalRepository interface {
	UpsertAuthPrincipal(ctx context.Context, userID int64, email string) error
	DeleteAuthPrincipal(ctx context.Context, userID int64) error
	SetAuthPrincipalRoles(ctx context.Context, userID int64, roleIDs []string) error
}

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
type userRoleRow struct {
	UserID int64  `dbx:"user_id"`
	RoleID string `dbx:"role_id"`
}
type userRoleSchema struct {
	dbx.Schema[userRoleRow]
	UserID dbx.Column[userRoleRow, int64]  `dbx:"user_id"`
	RoleID dbx.Column[userRoleRow, string] `dbx:"role_id"`
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
type authPrincipalRow struct {
	ID    string `dbx:"id"`
	Email string `dbx:"email"`
}
type authPrincipalSchema struct {
	dbx.Schema[authPrincipalRow]
	ID    dbx.Column[authPrincipalRow, string] `dbx:"id,pk"`
	Email dbx.Column[authPrincipalRow, string] `dbx:"email"`
}
type authPrincipalRoleRow struct {
	PrincipalID string `dbx:"principal_id"`
	Role        string `dbx:"role"`
}
type authPrincipalRoleSchema struct {
	dbx.Schema[authPrincipalRoleRow]
	PrincipalID dbx.Column[authPrincipalRoleRow, string] `dbx:"principal_id"`
	Role        dbx.Column[authPrincipalRoleRow, string] `dbx:"role"`
}
type repoSchemas struct {
	db  *dbx.DB
	rs  roleSchema
	pgs permissionGroupSchema
	ps  permissionSchema
	urs userRoleSchema
	rpg rolePermissionGroupSchema
	aps authPrincipalSchema
	apr authPrincipalRoleSchema
}
type roleRepo struct{ *repoSchemas }
type permissionGroupRepo struct{ *repoSchemas }
type permissionRepo struct{ *repoSchemas }
type userRoleRepo struct{ *repoSchemas }
type authPrincipalRepo struct{ *repoSchemas }

func NewRoleRepository(db *dbx.DB) RoleRepository { return &roleRepo{repoSchemas: newRepoSchemas(db)} }
func NewPermissionGroupRepository(db *dbx.DB) PermissionGroupRepository {
	return &permissionGroupRepo{repoSchemas: newRepoSchemas(db)}
}
func NewPermissionRepository(db *dbx.DB) PermissionRepository { return &permissionRepo{repoSchemas: newRepoSchemas(db)} }
func NewUserRoleRepository(db *dbx.DB) UserRoleRepository     { return &userRoleRepo{repoSchemas: newRepoSchemas(db)} }
func NewAuthPrincipalRepository(db *dbx.DB) AuthPrincipalRepository {
	return &authPrincipalRepo{repoSchemas: newRepoSchemas(db)}
}
func newRepoSchemas(db *dbx.DB) *repoSchemas {
	return &repoSchemas{
		db:  db,
		rs:  dbx.MustSchema("app_roles", roleSchema{}),
		pgs: dbx.MustSchema("app_permission_groups", permissionGroupSchema{}),
		ps:  dbx.MustSchema("app_permissions", permissionSchema{}),
		urs: dbx.MustSchema("app_user_roles", userRoleSchema{}),
		rpg: dbx.MustSchema("app_role_permission_groups", rolePermissionGroupSchema{}),
		aps: dbx.MustSchema("app_auth_principals", authPrincipalSchema{}),
		apr: dbx.MustSchema("app_auth_principal_roles", authPrincipalRoleSchema{}),
	}
}
func normalizeIDs(ids []string) []string {
	return lo.Uniq(lo.FilterMap(ids, func(id string, _ int) (string, bool) {
		v := strings.TrimSpace(id)
		return v, v != ""
	}))
}
func principalIDByUser(userID int64) string { return fmt.Sprintf("user:%d", userID) }

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
	return lo.Map(rows, func(row roleRow, _ int) Role {
		groupIDs, _ := gm.Get(row.ID)
		return Role{ID: row.ID, Name: row.Name, Description: row.Description, PermissionGroupIDs: slices.Clone(groupIDs), CreatedAt: row.CreatedAt}
	}), nil
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
	defer func() { if !committed { _ = tx.Rollback() } }()
	_, err = dbx.Exec(ctx, tx, dbx.InsertInto(r.rs).Columns(r.rs.ID, r.rs.Name, r.rs.Description, r.rs.CreatedAt).Values(r.rs.ID.Set(in.ID), r.rs.Name.Set(in.Name), r.rs.Description.Set(in.Description), r.rs.CreatedAt.Set(now)))
	if err != nil {
		return Role{}, err
	}
	groupIDs := normalizeIDs(in.PermissionGroupIDs)
	if len(groupIDs) > 0 {
		insert := dbx.InsertInto(r.rpg).Columns(r.rpg.RoleID, r.rpg.PermissionGroupID)
		for _, gid := range groupIDs { insert = insert.Values(r.rpg.RoleID.Set(in.ID), r.rpg.PermissionGroupID.Set(gid)) }
		if _, err = dbx.Exec(ctx, tx, insert); err != nil { return Role{}, err }
	}
	if err := tx.Commit(); err != nil { return Role{}, err }
	committed = true
	it, _, err := r.GetRole(ctx, in.ID)
	return it, err
}
func (r *roleRepo) UpdateRole(ctx context.Context, id string, in PatchRoleInput) (Role, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return Role{}, false, err }
	committed := false
	defer func() { if !committed { _ = tx.Rollback() } }()
	assignments := []dbx.Assignment{}
	if in.Name != nil { assignments = append(assignments, r.rs.Name.Set(*in.Name)) }
	if in.Description != nil { assignments = append(assignments, r.rs.Description.Set(*in.Description)) }
	if len(assignments) > 0 {
		res, err := dbx.Exec(ctx, tx, dbx.Update(r.rs).Set(assignments...).Where(r.rs.ID.Eq(id)))
		if err != nil { return Role{}, false, err }
		affected, _ := res.RowsAffected()
		if affected == 0 { return Role{}, false, nil }
	}
	if in.PermissionGroupIDs != nil {
		if _, err := dbx.Exec(ctx, tx, dbx.DeleteFrom(r.rpg).Where(r.rpg.RoleID.Eq(id))); err != nil { return Role{}, false, err }
		groupIDs := normalizeIDs(in.PermissionGroupIDs)
		if len(groupIDs) > 0 {
			insert := dbx.InsertInto(r.rpg).Columns(r.rpg.RoleID, r.rpg.PermissionGroupID)
			for _, gid := range groupIDs { insert = insert.Values(r.rpg.RoleID.Set(id), r.rpg.PermissionGroupID.Set(gid)) }
			if _, err := dbx.Exec(ctx, tx, insert); err != nil { return Role{}, false, err }
		}
	}
	if err := tx.Commit(); err != nil { return Role{}, false, err }
	committed = true
	role, ok, err := r.GetRole(ctx, id)
	return role, ok, err
}
func (r *roleRepo) DeleteRole(ctx context.Context, id string) (bool, error) {
	_, _ = dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.rpg).Where(r.rpg.RoleID.Eq(id)))
	res, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.rs).Where(r.rs.ID.Eq(id)))
	if err != nil { return false, err }
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}
func (r *permissionGroupRepo) ListPermissionGroups(ctx context.Context) ([]PermissionGroup, error) {
	rows, err := dbx.QueryAll[permissionGroupRow](ctx, r.db, dbx.Select(r.pgs.AllColumns()...).From(r.pgs).OrderBy(r.pgs.ID.Asc()), dbx.MustMapper[permissionGroupRow](r.pgs))
	if err != nil { return nil, err }
	return lo.Map(rows, func(row permissionGroupRow, _ int) PermissionGroup { return PermissionGroup{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt} }), nil
}
func (r *permissionGroupRepo) GetPermissionGroup(ctx context.Context, id string) (PermissionGroup, bool, error) {
	rows, err := dbx.QueryAll[permissionGroupRow](ctx, r.db, dbx.Select(r.pgs.AllColumns()...).From(r.pgs).Where(r.pgs.ID.Eq(id)), dbx.MustMapper[permissionGroupRow](r.pgs))
	if err != nil { return PermissionGroup{}, false, err }
	if len(rows) == 0 { return PermissionGroup{}, false, nil }
	row := rows[0]
	return PermissionGroup{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}, true, nil
}
func (r *permissionGroupRepo) CreatePermissionGroup(ctx context.Context, in CreatePermissionGroupInput) (PermissionGroup, error) {
	now := time.Now().UTC()
	_, err := dbx.Exec(ctx, r.db, dbx.InsertInto(r.pgs).Columns(r.pgs.ID, r.pgs.Name, r.pgs.Description, r.pgs.CreatedAt).Values(r.pgs.ID.Set(in.ID), r.pgs.Name.Set(in.Name), r.pgs.Description.Set(in.Description), r.pgs.CreatedAt.Set(now)))
	if err != nil { return PermissionGroup{}, err }
	it, _, err := r.GetPermissionGroup(ctx, in.ID)
	return it, err
}
func (r *permissionGroupRepo) UpdatePermissionGroup(ctx context.Context, id string, in PatchPermissionGroupInput) (PermissionGroup, bool, error) {
	assignments := []dbx.Assignment{}
	if in.Name != nil { assignments = append(assignments, r.pgs.Name.Set(*in.Name)) }
	if in.Description != nil { assignments = append(assignments, r.pgs.Description.Set(*in.Description)) }
	if len(assignments) > 0 {
		res, err := dbx.Exec(ctx, r.db, dbx.Update(r.pgs).Set(assignments...).Where(r.pgs.ID.Eq(id)))
		if err != nil { return PermissionGroup{}, false, err }
		ra, _ := res.RowsAffected()
		if ra == 0 { return PermissionGroup{}, false, nil }
	}
	it, ok, err := r.GetPermissionGroup(ctx, id)
	return it, ok, err
}
func (r *permissionGroupRepo) DeletePermissionGroup(ctx context.Context, id string) (bool, error) {
	res, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.pgs).Where(r.pgs.ID.Eq(id)))
	if err != nil { return false, err }
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}
func (r *permissionRepo) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := dbx.QueryAll[permissionRow](ctx, r.db, dbx.Select(r.ps.AllColumns()...).From(r.ps).OrderBy(r.ps.ID.Asc()), dbx.MustMapper[permissionRow](r.ps))
	if err != nil { return nil, err }
	return lo.Map(rows, func(row permissionRow, _ int) Permission { return Permission{ID: row.ID, Name: row.Name, Code: row.Code, GroupID: row.GroupID, CreatedAt: row.CreatedAt} }), nil
}
func (r *permissionRepo) GetPermission(ctx context.Context, id string) (Permission, bool, error) {
	rows, err := dbx.QueryAll[permissionRow](ctx, r.db, dbx.Select(r.ps.AllColumns()...).From(r.ps).Where(r.ps.ID.Eq(id)), dbx.MustMapper[permissionRow](r.ps))
	if err != nil { return Permission{}, false, err }
	if len(rows) == 0 { return Permission{}, false, nil }
	row := rows[0]
	return Permission{ID: row.ID, Name: row.Name, Code: row.Code, GroupID: row.GroupID, CreatedAt: row.CreatedAt}, true, nil
}
func (r *permissionRepo) CreatePermission(ctx context.Context, in CreatePermissionInput) (Permission, error) {
	now := time.Now().UTC()
	_, err := dbx.Exec(ctx, r.db, dbx.InsertInto(r.ps).Columns(r.ps.ID, r.ps.Name, r.ps.Code, r.ps.GroupID, r.ps.CreatedAt).Values(r.ps.ID.Set(in.ID), r.ps.Name.Set(in.Name), r.ps.Code.Set(in.Code), r.ps.GroupID.Set(in.GroupID), r.ps.CreatedAt.Set(now)))
	if err != nil { return Permission{}, err }
	it, _, err := r.GetPermission(ctx, in.ID)
	return it, err
}
func (r *permissionRepo) UpdatePermission(ctx context.Context, id string, in PatchPermissionInput) (Permission, bool, error) {
	assignments := []dbx.Assignment{}
	if in.Name != nil { assignments = append(assignments, r.ps.Name.Set(*in.Name)) }
	if in.Code != nil { assignments = append(assignments, r.ps.Code.Set(*in.Code)) }
	assignments = append(assignments, r.ps.GroupID.Set(in.GroupID))
	res, err := dbx.Exec(ctx, r.db, dbx.Update(r.ps).Set(assignments...).Where(r.ps.ID.Eq(id)))
	if err != nil { return Permission{}, false, err }
	ra, _ := res.RowsAffected()
	if ra == 0 { return Permission{}, false, nil }
	it, _, err := r.GetPermission(ctx, id)
	return it, true, err
}
func (r *permissionRepo) DeletePermission(ctx context.Context, id string) (bool, error) {
	res, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.ps).Where(r.ps.ID.Eq(id)))
	if err != nil { return false, err }
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}
func (r *userRoleRepo) ListUserRoleIDs(ctx context.Context, userID int64) ([]string, error) {
	rows, err := dbx.QueryAll[userRoleRow](ctx, r.db, dbx.Select(r.urs.AllColumns()...).From(r.urs).Where(r.urs.UserID.Eq(userID)), dbx.MustMapper[userRoleRow](r.urs))
	if err != nil { return nil, err }
	return lo.Map(rows, func(row userRoleRow, _ int) string { return row.RoleID }), nil
}
func (r *userRoleRepo) SetUserRoleIDs(ctx context.Context, userID int64, roleIDs []string) error {
	if _, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.urs).Where(r.urs.UserID.Eq(userID))); err != nil { return err }
	for _, roleID := range roleIDs {
		if roleID == "" { continue }
		if _, err := dbx.Exec(ctx, r.db, dbx.InsertInto(r.urs).Columns(r.urs.UserID, r.urs.RoleID).Values(r.urs.UserID.Set(userID), r.urs.RoleID.Set(roleID))); err != nil { return err }
	}
	return nil
}
func (r *userRoleRepo) DeleteUserRoles(ctx context.Context, userID int64) error {
	_, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.urs).Where(r.urs.UserID.Eq(userID)))
	return err
}
func (r *authPrincipalRepo) UpsertAuthPrincipal(ctx context.Context, userID int64, email string) error {
	id := principalIDByUser(userID)
	rows, err := dbx.QueryAll[authPrincipalRow](ctx, r.db, dbx.Select(r.aps.AllColumns()...).From(r.aps).Where(r.aps.ID.Eq(id)), dbx.MustMapper[authPrincipalRow](r.aps))
	if err != nil { return err }
	if len(rows) == 0 {
		_, err = dbx.Exec(ctx, r.db, dbx.InsertInto(r.aps).Columns(r.aps.ID, r.aps.Email).Values(r.aps.ID.Set(id), r.aps.Email.Set(email)))
		return err
	}
	_, err = dbx.Exec(ctx, r.db, dbx.Update(r.aps).Set(r.aps.Email.Set(email)).Where(r.aps.ID.Eq(id)))
	return err
}
func (r *authPrincipalRepo) DeleteAuthPrincipal(ctx context.Context, userID int64) error {
	id := principalIDByUser(userID)
	_, _ = dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.apr).Where(r.apr.PrincipalID.Eq(id)))
	_, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.aps).Where(r.aps.ID.Eq(id)))
	return err
}
func (r *authPrincipalRepo) SetAuthPrincipalRoles(ctx context.Context, userID int64, roleIDs []string) error {
	id := principalIDByUser(userID)
	if _, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.apr).Where(r.apr.PrincipalID.Eq(id))); err != nil { return err }
	for _, roleID := range roleIDs {
		if roleID == "" { continue }
		if _, err := dbx.Exec(ctx, r.db, dbx.InsertInto(r.apr).Columns(r.apr.PrincipalID, r.apr.Role).Values(r.apr.PrincipalID.Set(id), r.apr.Role.Set(roleID))); err != nil { return err }
	}
	return nil
}
