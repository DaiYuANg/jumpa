package http

import (
	"strconv"
	"strings"

	iamdomain "github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/domain"
	"github.com/samber/lo"
)

func ok[T any](data T) Result[T] { return Result[T]{Success: true, Data: data} }
func okPage[T any](items []T, total, page, pageSize int) Result[PageResult[T]] {
	return ok(PageResult[T]{Items: items, Total: total, Page: page, PageSize: pageSize})
}
func parseIDsCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
func containsFold(s, sub string) bool { return strings.Contains(strings.ToLower(s), strings.ToLower(sub)) }
func paginate[T any](items []T, page, pageSize int) PageResult[T] {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	total := len(items)
	start := (page - 1) * pageSize
	if start >= total {
		return PageResult[T]{Items: []T{}, Total: total, Page: page, PageSize: pageSize}
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return PageResult[T]{Items: items[start:end], Total: total, Page: page, PageSize: pageSize}
}
func toUserDTO(u iamdomain.User, roleIDs []string) userDTO {
	return userDTO{ID: strconv.FormatInt(u.ID, 10), Name: u.Name, Email: u.Email, Age: u.Age, RoleIDs: roleIDs, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt}
}
func toRoleDTO(r iamdomain.Role) roleDTO {
	return roleDTO{ID: r.ID, Name: r.Name, Description: r.Description, PermissionGroupIDs: r.PermissionGroupIDs, CreatedAt: r.CreatedAt}
}
func toPermissionGroupDTO(g iamdomain.PermissionGroup) permissionGroupDTO {
	return permissionGroupDTO{ID: g.ID, Name: g.Name, Description: g.Description, CreatedAt: g.CreatedAt}
}
func toPermissionDTO(p iamdomain.Permission) permissionDTO {
	return permissionDTO{ID: p.ID, Name: p.Name, Code: p.Code, GroupID: p.GroupID, CreatedAt: p.CreatedAt}
}
func toRoleDTOs(items []iamdomain.Role) []roleDTO {
	return lo.Map(items, func(it iamdomain.Role, _ int) roleDTO { return toRoleDTO(it) })
}
func toPermissionGroupDTOs(items []iamdomain.PermissionGroup) []permissionGroupDTO {
	return lo.Map(items, func(it iamdomain.PermissionGroup, _ int) permissionGroupDTO { return toPermissionGroupDTO(it) })
}
func toPermissionDTOs(items []iamdomain.Permission) []permissionDTO {
	return lo.Map(items, func(it iamdomain.Permission, _ int) permissionDTO { return toPermissionDTO(it) })
}
