package http

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/application"
	iamdomain "github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/domain"
	collectionset "github.com/DaiYuANg/arcgo/collectionx/set"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/lo"
)

type UserEndpoint struct {
	httpx.BaseEndpoint
	userSvc      application.UserService
	userRoleSvc  application.UserRoleService
	principalSvc application.AuthPrincipalService
}
type RBACEndpoint struct {
	httpx.BaseEndpoint
	roleSvc  application.RoleService
	groupSvc application.PermissionGroupService
	permSvc  application.PermissionService
}

func NewUserEndpoint(userSvc application.UserService, userRoleSvc application.UserRoleService, principalSvc application.AuthPrincipalService) *UserEndpoint {
	return &UserEndpoint{userSvc: userSvc, userRoleSvc: userRoleSvc, principalSvc: principalSvc}
}
func NewRBACEndpoint(roleSvc application.RoleService, groupSvc application.PermissionGroupService, permSvc application.PermissionService) *RBACEndpoint {
	return &RBACEndpoint{roleSvc: roleSvc, groupSvc: groupSvc, permSvc: permSvc}
}
func (e *UserEndpoint) RegisterRoutes(server httpx.ServerRuntime) { registerUserEndpoints(server.Group("/api"), e.userSvc, e.userRoleSvc, e.principalSvc) }
func (e *RBACEndpoint) RegisterRoutes(server httpx.ServerRuntime) { registerRBACEndpoints(server.Group("/api"), e.roleSvc, e.groupSvc, e.permSvc) }

func getUserDTOByID(ctx context.Context, userSvc application.UserService, userRoleSvc application.UserRoleService, id int64) (userDTO, bool, error) {
	u, found, err := userSvc.Get(ctx, id)
	if err != nil || !found {
		return userDTO{}, found, err
	}
	roleIDs, err := userRoleSvc.ListUserRoleIDs(ctx, u.ID)
	if err != nil {
		return userDTO{}, false, err
	}
	return toUserDTO(u, roleIDs), true, nil
}

func getUserDTO(ctx context.Context, userRoleSvc application.UserRoleService, u iamdomain.User) (userDTO, error) {
	roleIDs, err := userRoleSvc.ListUserRoleIDs(ctx, u.ID)
	if err != nil {
		return userDTO{}, err
	}
	return toUserDTO(u, roleIDs), nil
}

func registerUserEndpoints(api *httpx.Group, userSvc application.UserService, userRoleSvc application.UserRoleService, principalSvc application.AuthPrincipalService) {
	httpx.MustGroupGet(api, "/users", func(ctx context.Context, input *ListResourceInput) (*dynamicOutput, error) { // truncated behavior parity
		if validIDs := parseInt64IDsCSV(input.ID); len(validIDs) > 0 {
			res := make([]userDTO, 0, len(validIDs))
			for _, id := range validIDs {
				dto, found, err := getUserDTOByID(ctx, userSvc, userRoleSvc, id)
				if err != nil {
					return nil, err
				}
				if found {
					res = append(res, dto)
				}
			}
			return &dynamicOutput{Body: ok(res)}, nil
		}
		query := strings.TrimSpace(input.Q)
		if input.NameLike != "" {
			query = input.NameLike
		}
		if input.EmailLike != "" {
			query = input.EmailLike
		}
		page, pageSize, offset := normalizePageRequest(input.Page, input.PageSize)
		items, total, err := userSvc.List(ctx, query, pageSize, offset)
		if err != nil {
			return nil, err
		}
		out := make([]userDTO, len(items))
		for i, u := range items {
			dto, err := getUserDTO(ctx, userRoleSvc, u)
			if err != nil {
				return nil, err
			}
			out[i] = dto
		}
		return &dynamicOutput{Body: okPage(out, total, page, pageSize)}, nil
	}, huma.OperationTags("users"))
	// Keep remaining routes behavior same as old implementation.
	httpx.MustGroupGet(api, "/users/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) {
		id, err := strconv.ParseInt(input.ID, 10, 64)
		if err != nil {
			return nil, httpx.NewError(400, "invalid user id")
		}
		u, found, err := userSvc.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, httpx.NewError(404, "user not found")
		}
		roleIDs, err := userRoleSvc.ListUserRoleIDs(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		return &dynamicOutput{Body: ok(toUserDTO(u, roleIDs))}, nil
	}, huma.OperationTags("users"))
	httpx.MustGroupPost(api, "/users", func(ctx context.Context, input *CreateUserInput) (*dynamicOutput, error) {
		u, err := userSvc.Create(ctx, iamdomain.CreateUserInput{Name: input.Body.Name, Email: input.Body.Email, Age: input.Body.Age})
		if err != nil { return nil, err }
		if err := principalSvc.UpsertAuthPrincipal(ctx, u.ID, u.Email); err != nil { return nil, err }
		if err := userRoleSvc.SetUserRoleIDs(ctx, u.ID, slices.Clone(input.Body.RoleIDs)); err != nil { return nil, err }
		if err := principalSvc.SetAuthPrincipalRoles(ctx, u.ID, slices.Clone(input.Body.RoleIDs)); err != nil { return nil, err }
		return &dynamicOutput{Body: ok(toUserDTO(u, slices.Clone(input.Body.RoleIDs)))}, nil
	}, huma.OperationTags("users"))
	httpx.MustGroupPatch(api, "/users/{id}", func(ctx context.Context, input *PatchUserInput) (*dynamicOutput, error) {
		id, err := strconv.ParseInt(input.ID, 10, 64); if err != nil { return nil, httpx.NewError(400, "invalid user id") }
		u, found, err := userSvc.Update(ctx, id, iamdomain.UpdateUserInput{Name: input.Body.Name, Email: input.Body.Email, Age: input.Body.Age}); if err != nil { return nil, err }
		if !found { return nil, httpx.NewError(404, "user not found") }
		if err := principalSvc.UpsertAuthPrincipal(ctx, u.ID, u.Email); err != nil { return nil, err }
		if input.Body.RoleIDs != nil {
			if err := userRoleSvc.SetUserRoleIDs(ctx, id, slices.Clone(input.Body.RoleIDs)); err != nil { return nil, err }
			if err := principalSvc.SetAuthPrincipalRoles(ctx, id, slices.Clone(input.Body.RoleIDs)); err != nil { return nil, err }
		}
		roleIDs, err := userRoleSvc.ListUserRoleIDs(ctx, id); if err != nil { return nil, err }
		return &dynamicOutput{Body: ok(toUserDTO(u, roleIDs))}, nil
	}, huma.OperationTags("users"))
	httpx.MustGroupDelete(api, "/users/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) {
		id, err := strconv.ParseInt(input.ID, 10, 64); if err != nil { return nil, httpx.NewError(400, "invalid user id") }
		deleted, err := userSvc.Delete(ctx, id); if err != nil { return nil, err }
		if !deleted { return nil, httpx.NewError(404, "user not found") }
		_ = userRoleSvc.DeleteUserRoles(ctx, id); _ = principalSvc.DeleteAuthPrincipal(ctx, id)
		return &dynamicOutput{Body: ok(map[string]bool{"deleted": true})}, nil
	}, huma.OperationTags("users"))
	httpx.MustGroupDelete(api, "/users", func(ctx context.Context, input *DeleteManyInput) (*dynamicOutput, error) {
		validIDs := parseInt64IDsCSV(input.ID)
		for _, id := range validIDs { _, _ = userSvc.Delete(ctx, id); _ = userRoleSvc.DeleteUserRoles(ctx, id); _ = principalSvc.DeleteAuthPrincipal(ctx, id) }
		return &dynamicOutput{Body: ok([]userDTO{})}, nil
	}, huma.OperationTags("users"))
}

func registerRBACEndpoints(api *httpx.Group, roleSvc application.RoleService, groupSvc application.PermissionGroupService, permSvc application.PermissionService) {
	httpx.MustGroupGet(api, "/roles", func(ctx context.Context, input *ListResourceInput) (*dynamicOutput, error) {
		list, err := roleSvc.ListRoles(ctx); if err != nil { return nil, err }
		if ids := parseIDsCSV(input.ID); len(ids) > 0 {
			idSet := collectionset.NewSet(ids...)
			filtered := toRoleDTOs(lo.Filter(list, func(it iamdomain.Role, _ int) bool { return idSet.Contains(it.ID) }))
			return &dynamicOutput{Body: ok(filtered)}, nil
		}
		q := strings.ToLower(strings.TrimSpace(input.Q))
		filtered := toRoleDTOs(lo.Filter(list, func(it iamdomain.Role, _ int) bool {
			if q == "" {
				return true
			}
			return strings.Contains(strings.ToLower(it.Name), q)
		}))
		if strings.EqualFold(input.Sort, "name") {
			slices.SortFunc(filtered, func(a, b roleDTO) int {
				if strings.EqualFold(input.Order, "desc") { return strings.Compare(strings.ToLower(b.Name), strings.ToLower(a.Name)) }
				return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
			})
		}
		p := paginate(filtered, input.Page, input.PageSize)
		return &dynamicOutput{Body: okPage(p.Items, p.Total, p.Page, p.PageSize)}, nil
	}, huma.OperationTags("roles"))
	// Remaining RBAC routes follow existing behavior.
	httpx.MustGroupGet(api, "/roles/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) { it, found, err := roleSvc.GetRole(ctx, input.ID); if err != nil { return nil, err }; if !found { return nil, httpx.NewError(404, "role not found") }; return &dynamicOutput{Body: ok(toRoleDTO(it))}, nil }, huma.OperationTags("roles"))
	httpx.MustGroupPost(api, "/roles", func(ctx context.Context, input *CreateRoleInput) (*dynamicOutput, error) { it, err := roleSvc.CreateRole(ctx, input.Body.Name, input.Body.Description, slices.Clone(input.Body.PermissionGroupIDs)); if err != nil { return nil, err }; return &dynamicOutput{Body: ok(toRoleDTO(it))}, nil }, huma.OperationTags("roles"))
	httpx.MustGroupPatch(api, "/roles/{id}", func(ctx context.Context, input *PatchRoleInput) (*dynamicOutput, error) { it, found, err := roleSvc.UpdateRole(ctx, input.ID, input.Body.Name, input.Body.Description, input.Body.PermissionGroupIDs); if err != nil { return nil, err }; if !found { return nil, httpx.NewError(404, "role not found") }; return &dynamicOutput{Body: ok(toRoleDTO(it))}, nil }, huma.OperationTags("roles"))
	httpx.MustGroupDelete(api, "/roles/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) { _, err := roleSvc.DeleteRole(ctx, input.ID); if err != nil { return nil, err }; return &dynamicOutput{Body: ok(map[string]bool{"deleted": true})}, nil }, huma.OperationTags("roles"))
	httpx.MustGroupDelete(api, "/roles", func(ctx context.Context, input *DeleteManyInput) (*dynamicOutput, error) {
		for _, id := range parseIDsCSV(input.ID) {
			if _, err := roleSvc.DeleteRole(ctx, id); err != nil {
				return nil, err
			}
		}
		return &dynamicOutput{Body: ok([]roleDTO{})}, nil
	}, huma.OperationTags("roles"))
	httpx.MustGroupGet(api, "/permission-groups", func(ctx context.Context, input *ListResourceInput) (*dynamicOutput, error) {
		list, err := groupSvc.ListPermissionGroups(ctx)
		if err != nil {
			return nil, err
		}
		if ids := parseIDsCSV(input.ID); len(ids) > 0 {
			idSet := collectionset.NewSet(ids...)
			filtered := toPermissionGroupDTOs(lo.Filter(list, func(it iamdomain.PermissionGroup, _ int) bool { return idSet.Contains(it.ID) }))
			return &dynamicOutput{Body: ok(filtered)}, nil
		}
		q := strings.ToLower(strings.TrimSpace(input.Q))
		filtered := toPermissionGroupDTOs(lo.Filter(list, func(it iamdomain.PermissionGroup, _ int) bool {
			if q == "" {
				return true
			}
			return strings.Contains(strings.ToLower(it.Name), q)
		}))
		p := paginate(filtered, input.Page, input.PageSize)
		return &dynamicOutput{Body: okPage(p.Items, p.Total, p.Page, p.PageSize)}, nil
	}, huma.OperationTags("permission-groups"))
	httpx.MustGroupGet(api, "/permission-groups/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) { it, found, err := groupSvc.GetPermissionGroup(ctx, input.ID); if err != nil { return nil, err }; if !found { return nil, httpx.NewError(404, "permission group not found") }; return &dynamicOutput{Body: ok(toPermissionGroupDTO(it))}, nil }, huma.OperationTags("permission-groups"))
	httpx.MustGroupPost(api, "/permission-groups", func(ctx context.Context, input *CreatePermissionGroupInput) (*dynamicOutput, error) { it, err := groupSvc.CreatePermissionGroup(ctx, input.Body.Name, input.Body.Description); if err != nil { return nil, err }; return &dynamicOutput{Body: ok(toPermissionGroupDTO(it))}, nil }, huma.OperationTags("permission-groups"))
	httpx.MustGroupPatch(api, "/permission-groups/{id}", func(ctx context.Context, input *PatchPermissionGroupInput) (*dynamicOutput, error) { it, found, err := groupSvc.UpdatePermissionGroup(ctx, input.ID, input.Body.Name, input.Body.Description); if err != nil { return nil, err }; if !found { return nil, httpx.NewError(404, "permission group not found") }; return &dynamicOutput{Body: ok(toPermissionGroupDTO(it))}, nil }, huma.OperationTags("permission-groups"))
	httpx.MustGroupDelete(api, "/permission-groups/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) { _, err := groupSvc.DeletePermissionGroup(ctx, input.ID); if err != nil { return nil, err }; return &dynamicOutput{Body: ok(map[string]bool{"deleted": true})}, nil }, huma.OperationTags("permission-groups"))
	httpx.MustGroupDelete(api, "/permission-groups", func(ctx context.Context, input *DeleteManyInput) (*dynamicOutput, error) {
		for _, id := range parseIDsCSV(input.ID) {
			if _, err := groupSvc.DeletePermissionGroup(ctx, id); err != nil {
				return nil, err
			}
		}
		return &dynamicOutput{Body: ok([]permissionGroupDTO{})}, nil
	}, huma.OperationTags("permission-groups"))
	httpx.MustGroupGet(api, "/permissions", func(ctx context.Context, input *ListResourceInput) (*dynamicOutput, error) {
		list, err := permSvc.ListPermissions(ctx)
		if err != nil {
			return nil, err
		}
		if ids := parseIDsCSV(input.ID); len(ids) > 0 {
			idSet := collectionset.NewSet(ids...)
			filtered := toPermissionDTOs(lo.Filter(list, func(it iamdomain.Permission, _ int) bool { return idSet.Contains(it.ID) }))
			return &dynamicOutput{Body: ok(filtered)}, nil
		}
		q := strings.ToLower(strings.TrimSpace(input.Q))
		filtered := toPermissionDTOs(lo.Filter(list, func(it iamdomain.Permission, _ int) bool {
			if q == "" {
				return true
			}
			name := strings.ToLower(it.Name)
			code := strings.ToLower(it.Code)
			return strings.Contains(name, q) || strings.Contains(code, q)
		}))
		p := paginate(filtered, input.Page, input.PageSize)
		return &dynamicOutput{Body: okPage(p.Items, p.Total, p.Page, p.PageSize)}, nil
	}, huma.OperationTags("permissions"))
	httpx.MustGroupGet(api, "/permissions/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) { it, found, err := permSvc.GetPermission(ctx, input.ID); if err != nil { return nil, err }; if !found { return nil, httpx.NewError(404, "permission not found") }; return &dynamicOutput{Body: ok(toPermissionDTO(it))}, nil }, huma.OperationTags("permissions"))
	httpx.MustGroupPost(api, "/permissions", func(ctx context.Context, input *CreatePermissionInput) (*dynamicOutput, error) { it, err := permSvc.CreatePermission(ctx, input.Body.Name, input.Body.Code, input.Body.GroupID); if err != nil { return nil, err }; return &dynamicOutput{Body: ok(toPermissionDTO(it))}, nil }, huma.OperationTags("permissions"))
	httpx.MustGroupPatch(api, "/permissions/{id}", func(ctx context.Context, input *PatchPermissionInput) (*dynamicOutput, error) { it, found, err := permSvc.UpdatePermission(ctx, input.ID, input.Body.Name, input.Body.Code, input.Body.GroupID); if err != nil { return nil, err }; if !found { return nil, httpx.NewError(404, "permission not found") }; return &dynamicOutput{Body: ok(toPermissionDTO(it))}, nil }, huma.OperationTags("permissions"))
	httpx.MustGroupPatch(api, "/permissions/bulk", func(ctx context.Context, input *BulkPatchInput) (*dynamicOutput, error) {
		ids := parseIDsCSV(input.ID)
		updated := make([]permissionDTO, 0, len(ids))
		for _, id := range ids {
			it, found, err := permSvc.UpdatePermission(ctx, id, nil, nil, input.Body.GroupID)
			if err != nil {
				return nil, err
			}
			if found {
				updated = append(updated, toPermissionDTO(it))
			}
		}
		return &dynamicOutput{Body: ok(updated)}, nil
	}, huma.OperationTags("permissions"))
	httpx.MustGroupDelete(api, "/permissions/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) { _, err := permSvc.DeletePermission(ctx, input.ID); if err != nil { return nil, err }; return &dynamicOutput{Body: ok(map[string]bool{"deleted": true})}, nil }, huma.OperationTags("permissions"))
	httpx.MustGroupDelete(api, "/permissions", func(ctx context.Context, input *DeleteManyInput) (*dynamicOutput, error) {
		for _, id := range parseIDsCSV(input.ID) {
			if _, err := permSvc.DeletePermission(ctx, id); err != nil {
				return nil, err
			}
		}
		return &dynamicOutput{Body: ok([]permissionDTO{})}, nil
	}, huma.OperationTags("permissions"))
}
