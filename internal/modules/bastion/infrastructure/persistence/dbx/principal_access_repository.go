package dbx

import (
	"context"
	"strings"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
)

type principalAccessPrincipalRow struct {
	ID    string `dbx:"id"`
	Email string `dbx:"email"`
}

type principalAccessPrincipalSchema struct {
	dbx.Schema[principalAccessPrincipalRow]
	ID    dbx.Column[principalAccessPrincipalRow, string] `dbx:"id,pk"`
	Email dbx.Column[principalAccessPrincipalRow, string] `dbx:"email"`
}

type principalAccessRoleRow struct {
	PrincipalID string `dbx:"principal_id"`
	Role        string `dbx:"role"`
}

type principalAccessRoleSchema struct {
	dbx.Schema[principalAccessRoleRow]
	PrincipalID dbx.Column[principalAccessRoleRow, string] `dbx:"principal_id"`
	Role        dbx.Column[principalAccessRoleRow, string] `dbx:"role"`
}

type principalAccessRepo struct {
	db  *dbx.DB
	ps  principalAccessPrincipalSchema
	prs principalAccessRoleSchema
}

func NewPrincipalAccessRepository(db *dbx.DB) ports.PrincipalAccessRepository {
	return &principalAccessRepo{
		db:  db,
		ps:  dbx.MustSchema("app_auth_principals", principalAccessPrincipalSchema{}),
		prs: dbx.MustSchema("app_auth_principal_roles", principalAccessRoleSchema{}),
	}
}

func (r *principalAccessRepo) ListRoleIDsByEmail(ctx context.Context, email string) ([]string, error) {
	value := strings.ToLower(strings.TrimSpace(email))
	if value == "" {
		return []string{}, nil
	}

	principals, err := dbx.QueryAll[principalAccessPrincipalRow](
		ctx,
		r.db,
		dbx.Select(r.ps.AllColumns().Values()...).From(r.ps).Where(r.ps.Email.Eq(value)).Limit(1),
		dbx.MustMapper[principalAccessPrincipalRow](r.ps),
	)
	if err != nil {
		return nil, err
	}
	principal := principals.GetFirstOption()
	if principal.IsAbsent() {
		return []string{}, nil
	}

	rows, err := dbx.QueryAll[principalAccessRoleRow](
		ctx,
		r.db,
		dbx.Select(r.prs.AllColumns().Values()...).From(r.prs).Where(r.prs.PrincipalID.Eq(principal.MustGet().ID)),
		dbx.MustMapper[principalAccessRoleRow](r.prs),
	)
	if err != nil {
		return nil, err
	}

	return collectionx.MapList(rows, func(_ int, row principalAccessRoleRow) string {
		return row.Role
	}).Values(), nil
}
