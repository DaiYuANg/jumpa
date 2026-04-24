package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/arcgolabs/eventx"
	"github.com/DaiYuANg/jumpa/internal/event"
	iamdomain "github.com/DaiYuANg/jumpa/internal/modules/iam/domain"
	"github.com/DaiYuANg/jumpa/internal/modules/iam/ports"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

type userAppService struct {
	repo ports.UserRepository
	bus  eventx.BusRuntime
	log  *slog.Logger
}

func NewUserService(repo ports.UserRepository, bus eventx.BusRuntime, log *slog.Logger) UserService {
	return &userAppService{repo: repo, bus: bus, log: log}
}

func (s *userAppService) List(ctx context.Context, search string, limit, offset int) ([]iamdomain.User, int, error) {
	items, total, err := s.repo.List(ctx, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *userAppService) Get(ctx context.Context, id int64) (mo.Option[iamdomain.User], error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userAppService) Create(ctx context.Context, in iamdomain.CreateUserInput) (iamdomain.User, error) {
	user, err := s.repo.Create(ctx, in)
	if err != nil {
		return iamdomain.User{}, err
	}
	_ = s.bus.PublishAsync(ctx, event.UserCreatedEvent{
		UserID: user.ID, UserName: user.Name, Email: user.Email, CreatedAt: user.CreatedAt,
	})
	return user, nil
}

func (s *userAppService) Update(ctx context.Context, id int64, in iamdomain.UpdateUserInput) (mo.Option[iamdomain.User], error) {
	return s.repo.Update(ctx, id, in)
}

func (s *userAppService) Delete(ctx context.Context, id int64) (bool, error) {
	return s.repo.Delete(ctx, id)
}

type roleAppService struct {
	uow     ports.UnitOfWork
	repo    ports.RoleRepository
	rpgRepo ports.RolePermissionGroupRepository
}

func NewRoleService(uow ports.UnitOfWork, r ports.RoleRepository, rpg ports.RolePermissionGroupRepository) RoleService {
	return &roleAppService{uow: uow, repo: r, rpgRepo: rpg}
}

func (s *roleAppService) ListRoles(ctx context.Context) ([]iamdomain.Role, error) {
	items, err := s.repo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	roleIDs := lo.Map(items, func(it ports.RoleRecord, _ int) string { return it.ID })
	pairs, err := s.rpgRepo.ListPairsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	groupIDsByRole := lo.MapValues(
		lo.GroupBy(pairs, func(p ports.RolePermissionGroupPair) string { return p.RoleID }),
		func(ps []ports.RolePermissionGroupPair, _ string) []string {
			return lo.Map(ps, func(p ports.RolePermissionGroupPair, _ int) string { return p.PermissionGroupID })
		},
	)
	out := make([]iamdomain.Role, len(items))
	for i, it := range items {
		out[i] = iamdomain.Role{
			ID:                 it.ID,
			Name:               it.Name,
			Description:        it.Description,
			PermissionGroupIDs: groupIDsByRole[it.ID],
			CreatedAt:          it.CreatedAt,
		}
	}
	return out, nil
}

func (s *roleAppService) GetRole(ctx context.Context, id string) (mo.Option[iamdomain.Role], error) {
	it, err := s.repo.GetRole(ctx, id)
	if err != nil || it.IsAbsent() {
		return mo.None[iamdomain.Role](), err
	}
	role := it.MustGet()
	groupIDs, err := s.rpgRepo.ListPermissionGroupIDsByRoleID(ctx, id)
	if err != nil {
		return mo.None[iamdomain.Role](), err
	}
	return mo.Some(iamdomain.Role{
		ID:                 role.ID,
		Name:               role.Name,
		Description:        role.Description,
		PermissionGroupIDs: groupIDs,
		CreatedAt:          role.CreatedAt,
	}), nil
}
func (s *roleAppService) CreateRole(ctx context.Context, name, description string, permissionGroupIDs []string) (iamdomain.Role, error) {
	id := makeID("role")
	var out iamdomain.Role
	err := s.uow.InTx(ctx, nil, func(ctx context.Context, tx ports.UnitOfWorkTx) error {
		it, err := tx.Roles().CreateRole(ctx, ports.CreateRoleInput{ID: id, Name: name, Description: description})
		if err != nil {
			return err
		}
		if err := tx.RolePermissionGroups().ReplacePermissionGroupIDs(ctx, id, permissionGroupIDs); err != nil {
			return err
		}
		out = iamdomain.Role{ID: it.ID, Name: it.Name, Description: it.Description, PermissionGroupIDs: permissionGroupIDs, CreatedAt: it.CreatedAt}
		return nil
	})
	return out, err
}
func (s *roleAppService) UpdateRole(ctx context.Context, id string, name, description *string, permissionGroupIDs []string) (mo.Option[iamdomain.Role], error) {
	out := mo.None[iamdomain.Role]()
	err := s.uow.InTx(ctx, nil, func(ctx context.Context, tx ports.UnitOfWorkTx) error {
		it, err := tx.Roles().UpdateRole(ctx, id, ports.PatchRoleInput{Name: name, Description: description})
		if err != nil {
			return err
		}
		if it.IsAbsent() {
			out = mo.None[iamdomain.Role]()
			return nil
		}
		if err := tx.RolePermissionGroups().ReplacePermissionGroupIDs(ctx, id, permissionGroupIDs); err != nil {
			return err
		}
		groupIDs, err := tx.RolePermissionGroups().ListPermissionGroupIDsByRoleID(ctx, id)
		if err != nil {
			return err
		}
		role := it.MustGet()
		out = mo.Some(iamdomain.Role{ID: role.ID, Name: role.Name, Description: role.Description, PermissionGroupIDs: groupIDs, CreatedAt: role.CreatedAt})
		return nil
	})
	return out, err
}
func (s *roleAppService) DeleteRole(ctx context.Context, id string) (bool, error) {
	var deleted bool
	err := s.uow.InTx(ctx, nil, func(ctx context.Context, tx ports.UnitOfWorkTx) error {
		if err := tx.RolePermissionGroups().DeleteByRoleID(ctx, id); err != nil {
			return err
		}
		ok, err := tx.Roles().DeleteRole(ctx, id)
		if err != nil {
			return err
		}
		deleted = ok
		return nil
	})
	return deleted, err
}

type permissionGroupAppService struct {
	repo ports.PermissionGroupRepository
}

func NewPermissionGroupService(r ports.PermissionGroupRepository) PermissionGroupService {
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
func (s *permissionGroupAppService) GetPermissionGroup(ctx context.Context, id string) (mo.Option[iamdomain.PermissionGroup], error) {
	it, err := s.repo.GetPermissionGroup(ctx, id)
	if err != nil || it.IsAbsent() {
		return mo.None[iamdomain.PermissionGroup](), err
	}
	return mo.Some(toDomainPermissionGroup(it.MustGet())), nil
}
func (s *permissionGroupAppService) CreatePermissionGroup(ctx context.Context, name, description string) (iamdomain.PermissionGroup, error) {
	it, err := s.repo.CreatePermissionGroup(ctx, ports.CreatePermissionGroupInput{ID: makeID("pg"), Name: name, Description: description})
	if err != nil {
		return iamdomain.PermissionGroup{}, err
	}
	return toDomainPermissionGroup(it), nil
}
func (s *permissionGroupAppService) UpdatePermissionGroup(ctx context.Context, id string, name, description *string) (mo.Option[iamdomain.PermissionGroup], error) {
	it, err := s.repo.UpdatePermissionGroup(ctx, id, ports.PatchPermissionGroupInput{Name: name, Description: description})
	if err != nil || it.IsAbsent() {
		return mo.None[iamdomain.PermissionGroup](), err
	}
	return mo.Some(toDomainPermissionGroup(it.MustGet())), nil
}
func (s *permissionGroupAppService) DeletePermissionGroup(ctx context.Context, id string) (bool, error) {
	return s.repo.DeletePermissionGroup(ctx, id)
}

type permissionAppService struct{ repo ports.PermissionRepository }

func NewPermissionService(r ports.PermissionRepository) PermissionService {
	return &permissionAppService{repo: r}
}
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
func (s *permissionAppService) GetPermission(ctx context.Context, id string) (mo.Option[iamdomain.Permission], error) {
	it, err := s.repo.GetPermission(ctx, id)
	if err != nil || it.IsAbsent() {
		return mo.None[iamdomain.Permission](), err
	}
	return mo.Some(toDomainPermission(it.MustGet())), nil
}
func (s *permissionAppService) CreatePermission(ctx context.Context, name, code string, groupID *string) (iamdomain.Permission, error) {
	it, err := s.repo.CreatePermission(ctx, ports.CreatePermissionInput{ID: makeID("perm"), Name: name, Code: code, GroupID: groupID})
	if err != nil {
		return iamdomain.Permission{}, err
	}
	return toDomainPermission(it), nil
}
func (s *permissionAppService) UpdatePermission(ctx context.Context, id string, name, code *string, groupID *string) (mo.Option[iamdomain.Permission], error) {
	it, err := s.repo.UpdatePermission(ctx, id, ports.PatchPermissionInput{Name: name, Code: code, GroupID: groupID})
	if err != nil || it.IsAbsent() {
		return mo.None[iamdomain.Permission](), err
	}
	return mo.Some(toDomainPermission(it.MustGet())), nil
}
func (s *permissionAppService) DeletePermission(ctx context.Context, id string) (bool, error) {
	return s.repo.DeletePermission(ctx, id)
}

type userRoleAppService struct{ repo ports.UserRoleRepository }

func NewUserRoleService(r ports.UserRoleRepository) UserRoleService {
	return &userRoleAppService{repo: r}
}
func (s *userRoleAppService) ListUserRoleIDs(ctx context.Context, userID int64) ([]string, error) {
	return s.repo.ListUserRoleIDs(ctx, userID)
}
func (s *userRoleAppService) SetUserRoleIDs(ctx context.Context, userID int64, roleIDs []string) error {
	return s.repo.SetUserRoleIDs(ctx, userID, roleIDs)
}
func (s *userRoleAppService) DeleteUserRoles(ctx context.Context, userID int64) error {
	return s.repo.DeleteUserRoles(ctx, userID)
}

type authPrincipalAppService struct{ repo ports.AuthPrincipalRepository }

func NewAuthPrincipalService(r ports.AuthPrincipalRepository) AuthPrincipalService {
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

func toDomainPermissionGroup(it ports.PermissionGroup) iamdomain.PermissionGroup {
	return iamdomain.PermissionGroup{ID: it.ID, Name: it.Name, Description: it.Description, CreatedAt: it.CreatedAt}
}
func toDomainPermission(it ports.Permission) iamdomain.Permission {
	return iamdomain.Permission{ID: it.ID, Name: it.Name, Code: it.Code, GroupID: it.GroupID, CreatedAt: it.CreatedAt}
}

func makeID(prefix string) string {
	return prefix + "_" + time.Now().UTC().Format("20060102150405.000000000")
}
