package endpoints

import (
	"context"
	"strconv"
	"strings"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/domain"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/lo"
	"slices"
)

func registerUserEndpoints(api *httpx.Group, userSvc UserService, userRoleSvc UserRoleService, principalSvc AuthPrincipalService) {
	httpx.MustGroupGet(api, "/users", func(ctx context.Context, input *ListResourceInput) (*dynamicOutput, error) {
		if ids := parseIDsCSV(input.ID); len(ids) > 0 {
			res := make([]userDTO, 0, len(ids))
			validIDs := lo.FilterMap(ids, func(idStr string, _ int) (int64, bool) {
				id, err := strconv.ParseInt(idStr, 10, 64)
				return id, err == nil
			})
			for _, id := range validIDs {
				u, found, err := userSvc.Get(ctx, id)
				if err != nil {
					return nil, err
				}
				if !found {
					continue
				}
				roleIDs, err := userRoleSvc.ListUserRoleIDs(ctx, u.ID)
				if err != nil {
					return nil, err
				}
				res = append(res, toUserDTO(u, roleIDs))
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
		page, pageSize := input.Page, input.PageSize
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}
		items, total, err := userSvc.List(ctx, query, pageSize, (page-1)*pageSize)
		if err != nil {
			return nil, err
		}
		out := make([]userDTO, len(items))
		for i, u := range items {
			roleIDs, roleErr := userRoleSvc.ListUserRoleIDs(ctx, u.ID)
			if roleErr != nil {
				return nil, roleErr
			}
			out[i] = toUserDTO(u, roleIDs)
		}
		return &dynamicOutput{Body: okPage(out, total, page, pageSize)}, nil
	}, huma.OperationTags("users"))

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
		u, err := userSvc.Create(ctx, domain.CreateUserInput{Name: input.Body.Name, Email: input.Body.Email, Age: input.Body.Age})
		if err != nil {
			return nil, err
		}
		if err := principalSvc.UpsertAuthPrincipal(ctx, u.ID, u.Email); err != nil {
			return nil, err
		}
		if err := userRoleSvc.SetUserRoleIDs(ctx, u.ID, slices.Clone(input.Body.RoleIDs)); err != nil {
			return nil, err
		}
		if err := principalSvc.SetAuthPrincipalRoles(ctx, u.ID, slices.Clone(input.Body.RoleIDs)); err != nil {
			return nil, err
		}
		return &dynamicOutput{Body: ok(toUserDTO(u, slices.Clone(input.Body.RoleIDs)))}, nil
	}, huma.OperationTags("users"))

	httpx.MustGroupPatch(api, "/users/{id}", func(ctx context.Context, input *PatchUserInput) (*dynamicOutput, error) {
		id, err := strconv.ParseInt(input.ID, 10, 64)
		if err != nil {
			return nil, httpx.NewError(400, "invalid user id")
		}
		u, found, err := userSvc.Update(ctx, id, domain.UpdateUserInput{Name: input.Body.Name, Email: input.Body.Email, Age: input.Body.Age})
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, httpx.NewError(404, "user not found")
		}
		if err := principalSvc.UpsertAuthPrincipal(ctx, u.ID, u.Email); err != nil {
			return nil, err
		}
		if input.Body.RoleIDs != nil {
			if err := userRoleSvc.SetUserRoleIDs(ctx, id, slices.Clone(input.Body.RoleIDs)); err != nil {
				return nil, err
			}
			if err := principalSvc.SetAuthPrincipalRoles(ctx, id, slices.Clone(input.Body.RoleIDs)); err != nil {
				return nil, err
			}
		}
		roleIDs, err := userRoleSvc.ListUserRoleIDs(ctx, id)
		if err != nil {
			return nil, err
		}
		return &dynamicOutput{Body: ok(toUserDTO(u, roleIDs))}, nil
	}, huma.OperationTags("users"))

	httpx.MustGroupDelete(api, "/users/{id}", func(ctx context.Context, input *ByIDInput) (*dynamicOutput, error) {
		id, err := strconv.ParseInt(input.ID, 10, 64)
		if err != nil {
			return nil, httpx.NewError(400, "invalid user id")
		}
		deleted, err := userSvc.Delete(ctx, id)
		if err != nil {
			return nil, err
		}
		if !deleted {
			return nil, httpx.NewError(404, "user not found")
		}
		_ = userRoleSvc.DeleteUserRoles(ctx, id)
		_ = principalSvc.DeleteAuthPrincipal(ctx, id)
		return &dynamicOutput{Body: ok(map[string]bool{"deleted": true})}, nil
	}, huma.OperationTags("users"))

	httpx.MustGroupDelete(api, "/users", func(ctx context.Context, input *DeleteManyInput) (*dynamicOutput, error) {
		validIDs := lo.FilterMap(parseIDsCSV(input.ID), func(idStr string, _ int) (int64, bool) {
			id, err := strconv.ParseInt(idStr, 10, 64)
			return id, err == nil
		})
		for _, id := range validIDs {
			_, _ = userSvc.Delete(ctx, id)
			_ = userRoleSvc.DeleteUserRoles(ctx, id)
			_ = principalSvc.DeleteAuthPrincipal(ctx, id)
		}
		return &dynamicOutput{Body: ok([]userDTO{})}, nil
	}, huma.OperationTags("users"))

	httpx.MustGroupPost(api, "/users/bulk", func(ctx context.Context, _ *struct{}) (*dynamicOutput, error) {
		return nil, httpx.NewError(501, "bulk create not implemented")
	}, huma.OperationTags("users"))

	httpx.MustGroupPatch(api, "/users/bulk", func(ctx context.Context, _ *struct{}) (*dynamicOutput, error) {
		return nil, httpx.NewError(501, "bulk patch not implemented")
	}, huma.OperationTags("users"))
}
