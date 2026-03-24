package endpoints

import (
	"strconv"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/domain"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
	"github.com/samber/lo"
)

func toUserDTO(u domain.User, roleIDs []string) userDTO {
	return userDTO{
		ID:        strconv.FormatInt(u.ID, 10),
		Name:      u.Name,
		Email:     u.Email,
		Age:       u.Age,
		RoleIDs:   roleIDs,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toRoleDTO(r repo.Role) roleDTO {
	return roleDTO{
		ID:                 r.ID,
		Name:               r.Name,
		Description:        r.Description,
		PermissionGroupIDs: r.PermissionGroupIDs,
		CreatedAt:          r.CreatedAt,
	}
}

func toPermissionGroupDTO(g repo.PermissionGroup) permissionGroupDTO {
	return permissionGroupDTO{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		CreatedAt:   g.CreatedAt,
	}
}

func toPermissionDTO(p repo.Permission) permissionDTO {
	return permissionDTO{
		ID:        p.ID,
		Name:      p.Name,
		Code:      p.Code,
		GroupID:   p.GroupID,
		CreatedAt: p.CreatedAt,
	}
}

func toRoleDTOs(items []repo.Role) []roleDTO {
	return lo.Map(items, func(it repo.Role, _ int) roleDTO { return toRoleDTO(it) })
}

func toPermissionGroupDTOs(items []repo.PermissionGroup) []permissionGroupDTO {
	return lo.Map(items, func(it repo.PermissionGroup, _ int) permissionGroupDTO { return toPermissionGroupDTO(it) })
}

func toPermissionDTOs(items []repo.Permission) []permissionDTO {
	return lo.Map(items, func(it repo.Permission, _ int) permissionDTO { return toPermissionDTO(it) })
}
