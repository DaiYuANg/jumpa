package application

import (
	"context"
	"log/slog"
	"time"

	legacydomain "github.com/DaiYuANg/arcgo-rbac-template/internal/domain"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/event"
	iamdomain "github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/domain"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/infrastructure/persistence"
	"github.com/DaiYuANg/arcgo/eventx"
)

type userAppService struct {
	repo persistence.UserRepository
	bus  eventx.BusRuntime
	log  *slog.Logger
}

func NewUserService(repo persistence.UserRepository, bus eventx.BusRuntime, log *slog.Logger) UserService {
	return &userAppService{repo: repo, bus: bus, log: log}
}

func (s *userAppService) List(ctx context.Context, search string, limit, offset int) ([]iamdomain.User, int, error) {
	items, total, err := s.repo.List(ctx, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]iamdomain.User, len(items))
	for i, it := range items {
		out[i] = iamdomain.User(it)
	}
	return out, total, nil
}

func (s *userAppService) Get(ctx context.Context, id int64) (iamdomain.User, bool, error) {
	it, ok, err := s.repo.GetByID(ctx, id)
	return iamdomain.User(it), ok, err
}

func (s *userAppService) Create(ctx context.Context, in iamdomain.CreateUserInput) (iamdomain.User, error) {
	user, err := s.repo.Create(ctx, legacydomain.CreateUserInput{Name: in.Name, Email: in.Email, Age: in.Age})
	if err != nil {
		return iamdomain.User{}, err
	}
	_ = s.bus.PublishAsync(ctx, event.UserCreatedEvent{
		UserID: user.ID, UserName: user.Name, Email: user.Email, CreatedAt: user.CreatedAt,
	})
	return iamdomain.User(user), nil
}

func (s *userAppService) Update(ctx context.Context, id int64, in iamdomain.UpdateUserInput) (iamdomain.User, bool, error) {
	user, ok, err := s.repo.Update(ctx, id, legacydomain.UpdateUserInput{Name: in.Name, Email: in.Email, Age: in.Age})
	return iamdomain.User(user), ok, err
}

func (s *userAppService) Delete(ctx context.Context, id int64) (bool, error) {
	return s.repo.Delete(ctx, id)
}

type roleAppService struct{ repo persistence.RoleRepository }

func NewRoleService(r persistence.RoleRepository) RoleService { return &roleAppService{repo: r} }

func (s *roleAppService) ListRoles(ctx context.Context) ([]iamdomain.Role, error) {
	items, err := s.repo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]iamdomain.Role, len(items))
	for i, it := range items {
		out[i] = toDomainRole(it)
	}
	return out, nil
}

func (s *roleAppService) GetRole(ctx context.Context, id string) (iamdomain.Role, bool, error) {
	it, ok, err := s.repo.GetRole(ctx, id)
	return toDomainRole(it), ok, err
}
func (s *roleAppService) CreateRole(ctx context.Context, name, description string, permissionGroupIDs []string) (iamdomain.Role, error) {
	it, err := s.repo.CreateRole(ctx, persistence.CreateRoleInput{
		ID:                 makeID("role"),
		Name:               name,
		Description:        description,
		PermissionGroupIDs: permissionGroupIDs,
	})
	if err != nil {
		return iamdomain.Role{}, err
	}
	return toDomainRole(it), nil
}
func (s *roleAppService) UpdateRole(ctx context.Context, id string, name, description *string, permissionGroupIDs []string) (iamdomain.Role, bool, error) {
	it, ok, err := s.repo.UpdateRole(ctx, id, persistence.PatchRoleInput{Name: name, Description: description, PermissionGroupIDs: permissionGroupIDs})
	return toDomainRole(it), ok, err
}
func (s *roleAppService) DeleteRole(ctx context.Context, id string) (bool, error) { return s.repo.DeleteRole(ctx, id) }

type permissionGroupAppService struct{ repo persistence.PermissionGroupRepository }

func NewPermissionGroupService(r persistence.PermissionGroupRepository) PermissionGroupService {
	return &permissionGroupAppService{repo: r}
}
func (s *permissionGroupAppService) ListPermissionGroups(ctx context.Context) ([]iamdomain.PermissionGroup, error) {
	items, err := s.repo.ListPermissionGroups(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]iamdomain.PermissionGroup, len(items))
	for i, it := range items {
		out[i] = toDomainPermissionGroup(it)
	}
	return out, nil
}
func (s *permissionGroupAppService) GetPermissionGroup(ctx context.Context, id string) (iamdomain.PermissionGroup, bool, error) {
	it, ok, err := s.repo.GetPermissionGroup(ctx, id)
	return toDomainPermissionGroup(it), ok, err
}
func (s *permissionGroupAppService) CreatePermissionGroup(ctx context.Context, name, description string) (iamdomain.PermissionGroup, error) {
	it, err := s.repo.CreatePermissionGroup(ctx, persistence.CreatePermissionGroupInput{ID: makeID("pg"), Name: name, Description: description})
	if err != nil {
		return iamdomain.PermissionGroup{}, err
	}
	return toDomainPermissionGroup(it), nil
}
func (s *permissionGroupAppService) UpdatePermissionGroup(ctx context.Context, id string, name, description *string) (iamdomain.PermissionGroup, bool, error) {
	it, ok, err := s.repo.UpdatePermissionGroup(ctx, id, persistence.PatchPermissionGroupInput{Name: name, Description: description})
	return toDomainPermissionGroup(it), ok, err
}
func (s *permissionGroupAppService) DeletePermissionGroup(ctx context.Context, id string) (bool, error) {
	return s.repo.DeletePermissionGroup(ctx, id)
}

type permissionAppService struct{ repo persistence.PermissionRepository }

func NewPermissionService(r persistence.PermissionRepository) PermissionService { return &permissionAppService{repo: r} }
func (s *permissionAppService) ListPermissions(ctx context.Context) ([]iamdomain.Permission, error) {
	items, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]iamdomain.Permission, len(items))
	for i, it := range items {
		out[i] = toDomainPermission(it)
	}
	return out, nil
}
func (s *permissionAppService) GetPermission(ctx context.Context, id string) (iamdomain.Permission, bool, error) {
	it, ok, err := s.repo.GetPermission(ctx, id)
	return toDomainPermission(it), ok, err
}
func (s *permissionAppService) CreatePermission(ctx context.Context, name, code string, groupID *string) (iamdomain.Permission, error) {
	it, err := s.repo.CreatePermission(ctx, persistence.CreatePermissionInput{ID: makeID("perm"), Name: name, Code: code, GroupID: groupID})
	if err != nil {
		return iamdomain.Permission{}, err
	}
	return toDomainPermission(it), nil
}
func (s *permissionAppService) UpdatePermission(ctx context.Context, id string, name, code *string, groupID *string) (iamdomain.Permission, bool, error) {
	it, ok, err := s.repo.UpdatePermission(ctx, id, persistence.PatchPermissionInput{Name: name, Code: code, GroupID: groupID})
	return toDomainPermission(it), ok, err
}
func (s *permissionAppService) DeletePermission(ctx context.Context, id string) (bool, error) {
	return s.repo.DeletePermission(ctx, id)
}

type userRoleAppService struct{ repo persistence.UserRoleRepository }

func NewUserRoleService(r persistence.UserRoleRepository) UserRoleService { return &userRoleAppService{repo: r} }
func (s *userRoleAppService) ListUserRoleIDs(ctx context.Context, userID int64) ([]string, error) {
	return s.repo.ListUserRoleIDs(ctx, userID)
}
func (s *userRoleAppService) SetUserRoleIDs(ctx context.Context, userID int64, roleIDs []string) error {
	return s.repo.SetUserRoleIDs(ctx, userID, roleIDs)
}
func (s *userRoleAppService) DeleteUserRoles(ctx context.Context, userID int64) error {
	return s.repo.DeleteUserRoles(ctx, userID)
}

type authPrincipalAppService struct{ repo persistence.AuthPrincipalRepository }

func NewAuthPrincipalService(r persistence.AuthPrincipalRepository) AuthPrincipalService {
	return &authPrincipalAppService{repo: r}
}
func (s *authPrincipalAppService) UpsertAuthPrincipal(ctx context.Context, userID int64, email string) error {
	return s.repo.UpsertAuthPrincipal(ctx, userID, email)
}
func (s *authPrincipalAppService) DeleteAuthPrincipal(ctx context.Context, userID int64) error {
	return s.repo.DeleteAuthPrincipal(ctx, userID)
}
func (s *authPrincipalAppService) SetAuthPrincipalRoles(ctx context.Context, userID int64, roleIDs []string) error {
	return s.repo.SetAuthPrincipalRoles(ctx, userID, roleIDs)
}

func toDomainRole(it persistence.Role) iamdomain.Role {
	return iamdomain.Role{
		ID:                 it.ID,
		Name:               it.Name,
		Description:        it.Description,
		PermissionGroupIDs: it.PermissionGroupIDs,
		CreatedAt:          it.CreatedAt,
	}
}
func toDomainPermissionGroup(it persistence.PermissionGroup) iamdomain.PermissionGroup {
	return iamdomain.PermissionGroup{ID: it.ID, Name: it.Name, Description: it.Description, CreatedAt: it.CreatedAt}
}
func toDomainPermission(it persistence.Permission) iamdomain.Permission {
	return iamdomain.Permission{ID: it.ID, Name: it.Name, Code: it.Code, GroupID: it.GroupID, CreatedAt: it.CreatedAt}
}

func makeID(prefix string) string {
	return prefix + "_" + time.Now().UTC().Format("20060102150405.000000000")
}
