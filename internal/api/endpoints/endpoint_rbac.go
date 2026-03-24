package endpoints

import (
	"context"
	"slices"
	"strings"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
	collectionset "github.com/DaiYuANg/arcgo/collectionx/set"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/lo"
)

func registerRBACEndpoints(api *httpx.Group, roleSvc RoleService, groupSvc PermissionGroupService, permSvc PermissionService) {
	// roles
	httpx.MustGroupGet(api, "/roles", func(ctx context.Context, input *ListResourceInput) (*dynamicOutput, error) {
		list, err := roleSvc.ListRoles(ctx)
		if err != nil {
			return nil, err
		}
		filtered := make([]roleDTO, 0, len(list))
		if ids := parseIDsCSV(input.ID); len(ids) > 0 {
			idSet := collectionset.NewSet(ids...)
			filtered = toRoleDTOs(lo.Filter(list, func(it repo.Role, _ int) bool {
				return idSet.Contains(it.ID)
			}))
			return &dynamicOutput{Body: ok(filtered)}, nil
		}
		for _, it := range list {
			if input.Q == "" || containsFold(it.Name, input.Q) {
				filtered = append(filtered, toRoleDTO(it))
			}
		}
		if strings.EqualFold(input.Sort, "name") {
			slices.SortFunc(filtered, func(a, b roleDTO) int {
				if strings.EqualFold(input.Order, "desc") {
					return strings.Compare(strings.ToLower(b.Name), strings.ToLower(a.Name))
				}
				return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
			})
		}
		p := paginate(filtered, input.Page, input.PageSize)
		return &dynamicOutput{Body: okPage(p.Items, p.Total, p.Page, p.PageSize)}, nil
	}, huma.OperationTags("roles"))

	httpx.MustGroupGet(api, "/roles/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) {
		it, found, err := roleSvc.GetRole(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, httpx.NewError(404, "role not found")
		}
		return &dynamicOutput{Body: ok(toRoleDTO(it))}, nil
	}, huma.OperationTags("roles"))

	httpx.MustGroupPost(api, "/roles", func(ctx context.Context, input *CreateRoleInput) (*dynamicOutput, error) {
		it, err := roleSvc.CreateRole(ctx, input.Body.Name, input.Body.Description, slices.Clone(input.Body.PermissionGroupIDs))
		if err != nil {
			return nil, err
		}
		return &dynamicOutput{Body: ok(toRoleDTO(it))}, nil
	}, huma.OperationTags("roles"))

	httpx.MustGroupPatch(api, "/roles/{id}", func(ctx context.Context, input *PatchRoleInput) (*dynamicOutput, error) {
		it, found, err := roleSvc.UpdateRole(ctx, input.ID, input.Body.Name, input.Body.Description, input.Body.PermissionGroupIDs)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, httpx.NewError(404, "role not found")
		}
		return &dynamicOutput{Body: ok(toRoleDTO(it))}, nil
	}, huma.OperationTags("roles"))

	httpx.MustGroupDelete(api, "/roles/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) {
		_, err := roleSvc.DeleteRole(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		return &dynamicOutput{Body: ok(map[string]bool{"deleted": true})}, nil
	}, huma.OperationTags("roles"))

	httpx.MustGroupDelete(api, "/roles", func(ctx context.Context, input *DeleteManyInput) (*dynamicOutput, error) {
		for _, id := range parseIDsCSV(input.ID) {
			if _, err := roleSvc.DeleteRole(ctx, id); err != nil {
				return nil, err
			}
		}
		return &dynamicOutput{Body: ok([]roleDTO{})}, nil
	}, huma.OperationTags("roles"))

	// permission groups
	httpx.MustGroupGet(api, "/permission-groups", func(ctx context.Context, input *ListResourceInput) (*dynamicOutput, error) {
		list, err := groupSvc.ListPermissionGroups(ctx)
		if err != nil {
			return nil, err
		}
		filtered := make([]permissionGroupDTO, 0, len(list))
		if ids := parseIDsCSV(input.ID); len(ids) > 0 {
			idSet := collectionset.NewSet(ids...)
			filtered = toPermissionGroupDTOs(lo.Filter(list, func(it repo.PermissionGroup, _ int) bool {
				return idSet.Contains(it.ID)
			}))
			return &dynamicOutput{Body: ok(filtered)}, nil
		}
		for _, it := range list {
			if input.Q == "" || containsFold(it.Name, input.Q) {
				filtered = append(filtered, toPermissionGroupDTO(it))
			}
		}
		p := paginate(filtered, input.Page, input.PageSize)
		return &dynamicOutput{Body: okPage(p.Items, p.Total, p.Page, p.PageSize)}, nil
	}, huma.OperationTags("permission-groups"))

	httpx.MustGroupGet(api, "/permission-groups/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) {
		it, found, err := groupSvc.GetPermissionGroup(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, httpx.NewError(404, "permission group not found")
		}
		return &dynamicOutput{Body: ok(toPermissionGroupDTO(it))}, nil
	}, huma.OperationTags("permission-groups"))

	httpx.MustGroupPost(api, "/permission-groups", func(ctx context.Context, input *CreatePermissionGroupInput) (*dynamicOutput, error) {
		it, err := groupSvc.CreatePermissionGroup(ctx, input.Body.Name, input.Body.Description)
		if err != nil {
			return nil, err
		}
		return &dynamicOutput{Body: ok(toPermissionGroupDTO(it))}, nil
	}, huma.OperationTags("permission-groups"))

	httpx.MustGroupPatch(api, "/permission-groups/{id}", func(ctx context.Context, input *PatchPermissionGroupInput) (*dynamicOutput, error) {
		it, found, err := groupSvc.UpdatePermissionGroup(ctx, input.ID, input.Body.Name, input.Body.Description)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, httpx.NewError(404, "permission group not found")
		}
		return &dynamicOutput{Body: ok(toPermissionGroupDTO(it))}, nil
	}, huma.OperationTags("permission-groups"))

	httpx.MustGroupDelete(api, "/permission-groups/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) {
		_, err := groupSvc.DeletePermissionGroup(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		return &dynamicOutput{Body: ok(map[string]bool{"deleted": true})}, nil
	}, huma.OperationTags("permission-groups"))

	httpx.MustGroupDelete(api, "/permission-groups", func(ctx context.Context, input *DeleteManyInput) (*dynamicOutput, error) {
		for _, id := range parseIDsCSV(input.ID) {
			if _, err := groupSvc.DeletePermissionGroup(ctx, id); err != nil {
				return nil, err
			}
		}
		return &dynamicOutput{Body: ok([]permissionGroupDTO{})}, nil
	}, huma.OperationTags("permission-groups"))

	// permissions
	httpx.MustGroupGet(api, "/permissions", func(ctx context.Context, input *ListResourceInput) (*dynamicOutput, error) {
		list, err := permSvc.ListPermissions(ctx)
		if err != nil {
			return nil, err
		}
		filtered := make([]permissionDTO, 0, len(list))
		if ids := parseIDsCSV(input.ID); len(ids) > 0 {
			idSet := collectionset.NewSet(ids...)
			filtered = toPermissionDTOs(lo.Filter(list, func(it repo.Permission, _ int) bool {
				return idSet.Contains(it.ID)
			}))
			return &dynamicOutput{Body: ok(filtered)}, nil
		}
		for _, it := range list {
			if input.Q == "" || containsFold(it.Name, input.Q) || containsFold(it.Code, input.Q) {
				filtered = append(filtered, toPermissionDTO(it))
			}
		}
		p := paginate(filtered, input.Page, input.PageSize)
		return &dynamicOutput{Body: okPage(p.Items, p.Total, p.Page, p.PageSize)}, nil
	}, huma.OperationTags("permissions"))

	httpx.MustGroupGet(api, "/permissions/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) {
		it, found, err := permSvc.GetPermission(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, httpx.NewError(404, "permission not found")
		}
		return &dynamicOutput{Body: ok(toPermissionDTO(it))}, nil
	}, huma.OperationTags("permissions"))

	httpx.MustGroupPost(api, "/permissions", func(ctx context.Context, input *CreatePermissionInput) (*dynamicOutput, error) {
		it, err := permSvc.CreatePermission(ctx, input.Body.Name, input.Body.Code, input.Body.GroupID)
		if err != nil {
			return nil, err
		}
		return &dynamicOutput{Body: ok(toPermissionDTO(it))}, nil
	}, huma.OperationTags("permissions"))

	httpx.MustGroupPatch(api, "/permissions/{id}", func(ctx context.Context, input *PatchPermissionInput) (*dynamicOutput, error) {
		it, found, err := permSvc.UpdatePermission(ctx, input.ID, input.Body.Name, input.Body.Code, input.Body.GroupID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, httpx.NewError(404, "permission not found")
		}
		return &dynamicOutput{Body: ok(toPermissionDTO(it))}, nil
	}, huma.OperationTags("permissions"))

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

	httpx.MustGroupDelete(api, "/permissions/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) {
		_, err := permSvc.DeletePermission(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		return &dynamicOutput{Body: ok(map[string]bool{"deleted": true})}, nil
	}, huma.OperationTags("permissions"))

	httpx.MustGroupDelete(api, "/permissions", func(ctx context.Context, input *DeleteManyInput) (*dynamicOutput, error) {
		for _, id := range parseIDsCSV(input.ID) {
			if _, err := permSvc.DeletePermission(ctx, id); err != nil {
				return nil, err
			}
		}
		return &dynamicOutput{Body: ok([]permissionDTO{})}, nil
	}, huma.OperationTags("permissions"))

	httpx.MustGroupPatch(api, "/roles/bulk", func(ctx context.Context, _ *struct{}) (*dynamicOutput, error) {
		return nil, httpx.NewError(501, "bulk patch not implemented")
	}, huma.OperationTags("roles"))
}
