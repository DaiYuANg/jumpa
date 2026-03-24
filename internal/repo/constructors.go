package repo

import "github.com/DaiYuANg/arcgo/dbx"

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
func NewPermissionRepository(db *dbx.DB) PermissionRepository {
	return &permissionRepo{repoSchemas: newRepoSchemas(db)}
}
func NewUserRoleRepository(db *dbx.DB) UserRoleRepository {
	return &userRoleRepo{repoSchemas: newRepoSchemas(db)}
}
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
