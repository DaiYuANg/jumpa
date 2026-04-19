package dbx

import (
	"context"
	"strings"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/DaiYuANg/arcgo/dbx"
	columnx "github.com/DaiYuANg/arcgo/dbx/column"
	mapperx "github.com/DaiYuANg/arcgo/dbx/mapper"
	"github.com/DaiYuANg/arcgo/dbx/querydsl"
	schemax "github.com/DaiYuANg/arcgo/dbx/schema"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
)

type principalAccessPrincipalRow struct {
	ID    string `dbx:"id"`
	Email string `dbx:"email"`
}

type principalAccessPrincipalSchema struct {
	schemax.Schema[principalAccessPrincipalRow]
	ID    columnx.Column[principalAccessPrincipalRow, string] `dbx:"id,pk"`
	Email columnx.Column[principalAccessPrincipalRow, string] `dbx:"email"`
}

type principalAccessRoleRow struct {
	PrincipalID string `dbx:"principal_id"`
	Role        string `dbx:"role"`
}

type principalAccessRoleSchema struct {
	schemax.Schema[principalAccessRoleRow]
	PrincipalID columnx.Column[principalAccessRoleRow, string] `dbx:"principal_id"`
	Role        columnx.Column[principalAccessRoleRow, string] `dbx:"role"`
}

type principalAccessRepo struct {
	db  *dbx.DB
	ps  principalAccessPrincipalSchema
	prs principalAccessRoleSchema
}

func NewPrincipalAccessRepository(db *dbx.DB) ports.PrincipalAccessRepository {
	return &principalAccessRepo{
		db:  db,
		ps:  schemax.MustSchema("app_auth_principals", principalAccessPrincipalSchema{}),
		prs: schemax.MustSchema("app_auth_principal_roles", principalAccessRoleSchema{}),
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
		querydsl.Select(querydsl.AllColumns(r.ps).Values()...).From(r.ps).Where(r.ps.Email.Eq(value)).Limit(1),
		mapperx.MustMapper[principalAccessPrincipalRow](r.ps),
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
		querydsl.Select(querydsl.AllColumns(r.prs).Values()...).From(r.prs).Where(r.prs.PrincipalID.Eq(principal.MustGet().ID)),
		mapperx.MustMapper[principalAccessRoleRow](r.prs),
	)
	if err != nil {
		return nil, err
	}

	return collectionx.MapList(rows, func(_ int, row principalAccessRoleRow) string {
		return row.Role
	}).Values(), nil
}
