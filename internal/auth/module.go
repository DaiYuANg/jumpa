package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DaiYuANg/arcgo/authx"
	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/kvx"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
	db2 "github.com/DaiYuANg/jumpa/internal/db"
	"github.com/DaiYuANg/jumpa/internal/kv"
	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/mo"
)

type jwtClaims struct {
	Email string `json:"email"`
	Type  string `json:"typ"`
	jwt.RegisteredClaims
}

type authPrincipalRow struct {
	ID    string `dbx:"id"`
	Email string `dbx:"email"`
}

type authPrincipalSchema struct {
	dbx.Schema[authPrincipalRow]
	ID    dbx.Column[authPrincipalRow, string] `dbx:"id,pk"`
	Email dbx.Column[authPrincipalRow, string] `dbx:"email"`
}

type authPrincipalRoleRow struct {
	PrincipalID string `dbx:"principal_id"`
	Role        string `dbx:"role"`
}
type authPrincipalRoleSchema struct {
	dbx.Schema[authPrincipalRoleRow]
	PrincipalID dbx.Column[authPrincipalRoleRow, string] `dbx:"principal_id"`
	Role        dbx.Column[authPrincipalRoleRow, string] `dbx:"role"`
}

type authPrincipalPermissionRow struct {
	PrincipalID string `dbx:"principal_id"`
	Permission  string `dbx:"permission"`
}
type authPrincipalPermissionSchema struct {
	dbx.Schema[authPrincipalPermissionRow]
	PrincipalID dbx.Column[authPrincipalPermissionRow, string] `dbx:"principal_id"`
	Permission  dbx.Column[authPrincipalPermissionRow, string] `dbx:"permission"`
}

func revokedKey(jti string) string { return "auth:revoked:" + jti }

func parseAccessToken(secret, issuer, token string) (*jwtClaims, error) {
	t, err := jwt.ParseWithClaims(token, &jwtClaims{}, func(_ *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithIssuer(issuer))
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*jwtClaims)
	if !ok || !t.Valid || claims.Type != "access" {
		return nil, authx.ErrInvalidAuthenticationCredential
	}
	return claims, nil
}

func principalFromDB(ctx context.Context, db *dbx.DB, email string) (mo.Option[authx.Principal], error) {
	ps := dbx.MustSchema("app_auth_principals", authPrincipalSchema{})
	prs := dbx.MustSchema("app_auth_principal_roles", authPrincipalRoleSchema{})
	pps := dbx.MustSchema("app_auth_principal_permissions", authPrincipalPermissionSchema{})
	rows, err := dbx.QueryAll[authPrincipalRow](ctx, db,
		dbx.Select(ps.AllColumns().Values()...).From(ps).Where(ps.Email.Eq(strings.ToLower(strings.TrimSpace(email)))),
		dbx.MustMapper[authPrincipalRow](ps),
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
		dbx.Select(prs.AllColumns().Values()...).From(prs).Where(prs.PrincipalID.Eq(row.ID)),
		dbx.MustMapper[authPrincipalRoleRow](prs),
	)
	if err != nil {
		return mo.None[authx.Principal](), err
	}
	permRows, err := dbx.QueryAll[authPrincipalPermissionRow](ctx, db,
		dbx.Select(pps.AllColumns().Values()...).From(pps).Where(pps.PrincipalID.Eq(row.ID)),
		dbx.MustMapper[authPrincipalPermissionRow](pps),
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
				authx.WithAuthenticationManager(authx.AuthenticationManagerFunc(func(ctx context.Context, credential any) (authx.AuthenticationResult, error) {
					token, ok := credential.(string)
					if !ok || strings.TrimSpace(token) == "" {
						return authx.AuthenticationResult{}, authx.ErrUnauthenticated
					}
					claims, err := parseAccessToken(cfg.JWT.Secret, cfg.JWT.Issuer, token)
					if err != nil {
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
					principal.Attributes = collectionx.NewMapFrom(map[string]any{
						"email": claims.Email,
						"exp":   claims.ExpiresAt.Time.Format(time.RFC3339),
					})
					return authx.AuthenticationResult{Principal: principal}, nil
				})),
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
