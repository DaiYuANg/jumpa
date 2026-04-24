package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/arcgolabs/authx"
	"github.com/arcgolabs/collectionx"
	"github.com/arcgolabs/dbx"
	columnx "github.com/arcgolabs/dbx/column"
	mapperx "github.com/arcgolabs/dbx/mapper"
	"github.com/arcgolabs/dbx/querydsl"
	schemax "github.com/arcgolabs/dbx/schema"
	"github.com/arcgolabs/dix"
	"github.com/arcgolabs/kvx"
	"github.com/DaiYuANg/jumpa/internal/authtoken"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
	db2 "github.com/DaiYuANg/jumpa/internal/db"
	"github.com/DaiYuANg/jumpa/internal/kv"
	"github.com/samber/mo"
)

type authPrincipalRow struct {
	ID    string `dbx:"id"`
	Email string `dbx:"email"`
}

type authPrincipalSchema struct {
	schemax.Schema[authPrincipalRow]
	ID    columnx.Column[authPrincipalRow, string] `dbx:"id,pk"`
	Email columnx.Column[authPrincipalRow, string] `dbx:"email"`
}

type authPrincipalRoleRow struct {
	PrincipalID string `dbx:"principal_id"`
	Role        string `dbx:"role"`
}
type authPrincipalRoleSchema struct {
	schemax.Schema[authPrincipalRoleRow]
	PrincipalID columnx.Column[authPrincipalRoleRow, string] `dbx:"principal_id"`
	Role        columnx.Column[authPrincipalRoleRow, string] `dbx:"role"`
}

type authPrincipalPermissionRow struct {
	PrincipalID string `dbx:"principal_id"`
	Permission  string `dbx:"permission"`
}
type authPrincipalPermissionSchema struct {
	schemax.Schema[authPrincipalPermissionRow]
	PrincipalID columnx.Column[authPrincipalPermissionRow, string] `dbx:"principal_id"`
	Permission  columnx.Column[authPrincipalPermissionRow, string] `dbx:"permission"`
}

func revokedKey(jti string) string { return "auth:revoked:" + jti }

func principalFromDB(ctx context.Context, db *dbx.DB, email string) (mo.Option[authx.Principal], error) {
	ps := schemax.MustSchema("app_auth_principals", authPrincipalSchema{})
	prs := schemax.MustSchema("app_auth_principal_roles", authPrincipalRoleSchema{})
	pps := schemax.MustSchema("app_auth_principal_permissions", authPrincipalPermissionSchema{})
	rows, err := dbx.QueryAll[authPrincipalRow](ctx, db,
		querydsl.Select(querydsl.AllColumns(ps).Values()...).From(ps).Where(ps.Email.Eq(strings.ToLower(strings.TrimSpace(email)))),
		mapperx.MustMapper[authPrincipalRow](ps),
	)
	if err != nil {
		return mo.None[authx.Principal](), err
	}
	rowOpt := rows.GetFirstOption()
	if rowOpt.IsAbsent() {
		return mo.None[authx.Principal](), nil
	}
	row := rowOpt.MustGet()
	roleRows, err := dbx.QueryAll[authPrincipalRoleRow](ctx, db,
		querydsl.Select(querydsl.AllColumns(prs).Values()...).From(prs).Where(prs.PrincipalID.Eq(row.ID)),
		mapperx.MustMapper[authPrincipalRoleRow](prs),
	)
	if err != nil {
		return mo.None[authx.Principal](), err
	}
	permRows, err := dbx.QueryAll[authPrincipalPermissionRow](ctx, db,
		querydsl.Select(querydsl.AllColumns(pps).Values()...).From(pps).Where(pps.PrincipalID.Eq(row.ID)),
		mapperx.MustMapper[authPrincipalPermissionRow](pps),
	)
	if err != nil {
		return mo.None[authx.Principal](), err
	}
	roles := collectionx.MapList(roleRows, func(_ int, rr authPrincipalRoleRow) string { return rr.Role })
	perms := collectionx.MapList(permRows, func(_ int, pr authPrincipalPermissionRow) string { return pr.Permission })
	return mo.Some(authx.Principal{
		ID:          row.ID,
		Roles:       roles,
		Permissions: perms,
	}), nil
}

func hasPermission(pr authx.Principal, perm string) bool {
	return pr.Permissions.AnyMatch(func(_ int, item string) bool { return item == perm || item == "*" })
}

var Module = dix.NewModule("auth",
	dix.WithModuleImports(config2.Module, kv.Module, db2.Module),
	dix.WithModuleProviders(
		dix.Provider3(func(cfg config2.AppConfig, kvClient kvx.Client, db *dbx.DB) *authx.Engine {
			engine := authx.NewEngine(
				authx.WithAuthenticationManager(authx.NewProviderManager(
					authtoken.NewAuthenticationProvider(authtoken.Config{Secret: cfg.JWT.Secret, Issuer: cfg.JWT.Issuer}, func(ctx context.Context, claims *authtoken.Claims) (authx.AuthenticationResult, error) {
						if claims == nil || claims.Type != authtoken.TypeAccess {
							return authx.AuthenticationResult{}, authx.ErrInvalidAuthenticationCredential
						}
						if cfg.Valkey.Enabled {
							exists, exErr := kvClient.Exists(ctx, revokedKey(claims.ID))
							if exErr == nil && exists {
								return authx.AuthenticationResult{}, authx.ErrUnauthenticated
							}
						}
						pr, dbErr := principalFromDB(ctx, db, claims.Email)
						if dbErr != nil {
							return authx.AuthenticationResult{}, authx.ErrUnauthenticated
						}
						if pr.IsAbsent() {
							return authx.AuthenticationResult{}, authx.ErrUnauthenticated
						}
						principal := pr.MustGet()
						attributes := map[string]any{
							"email": claims.Email,
						}
						if !claims.ExpiresAt.IsZero() {
							attributes["exp"] = claims.ExpiresAt.Format(time.RFC3339)
						}
						principal.Attributes = collectionx.NewMapFrom(attributes)
						return authx.AuthenticationResult{Principal: principal}, nil
					}),
				)),
				authx.WithAuthorizer(authx.AuthorizerFunc(func(_ context.Context, input authx.AuthorizationModel) (authx.Decision, error) {
					pr, ok := input.Principal.(authx.Principal)
					if !ok {
						return authx.Decision{Allowed: false, Reason: "invalid principal"}, nil
					}
					perm := fmt.Sprintf("%s:%s", input.Resource, input.Action)
					if hasPermission(pr, perm) {
						return authx.Decision{Allowed: true}, nil
					}
					return authx.Decision{Allowed: false, Reason: "permission denied"}, nil
				})),
			)
			return engine
		}),
	),
)
