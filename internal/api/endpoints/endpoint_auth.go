package endpoints

import (
	"context"
	"strings"
	"time"

	"github.com/DaiYuANg/jumpa/internal/authtoken"
	"github.com/arcgolabs/authx"
	"github.com/arcgolabs/httpx"
	"github.com/arcgolabs/kvx"
)

type AuthConfig struct {
	Secret         string
	Issuer         string
	AccessTTLMin   int
	RefreshTTLHour int
	UseValkey      bool
	RevokedPrefix  string
}

func revokedKey(cfg AuthConfig, jti string) string {
	prefix := cfg.RevokedPrefix
	if strings.TrimSpace(prefix) == "" {
		prefix = "auth:revoked"
	}
	return prefix + ":" + jti
}

func isRevoked(ctx context.Context, cfg AuthConfig, client kvx.Client, claims *authtoken.Claims) bool {
	if !cfg.UseValkey || claims == nil || claims.ID == "" {
		return false
	}
	exists, err := client.Exists(ctx, revokedKey(cfg, claims.ID))
	return err == nil && exists
}

func revokeToken(ctx context.Context, cfg AuthConfig, client kvx.Client, claims *authtoken.Claims) {
	if !cfg.UseValkey || claims == nil || claims.ID == "" || claims.ExpiresAt.IsZero() {
		return
	}
	ttl := time.Until(claims.ExpiresAt)
	if ttl <= 0 {
		return
	}
	_ = client.Set(ctx, revokedKey(cfg, claims.ID), []byte("1"), ttl)
}

func parseBearerToken(ctx context.Context) string {
	if hctx, ok := ctx.(interface{ Header(string) string }); ok {
		authz := strings.TrimSpace(hctx.Header("Authorization"))
		if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			return strings.TrimSpace(authz[7:])
		}
	}
	return ""
}

func parseRefreshCookie(ctx context.Context) string {
	if hctx, ok := ctx.(interface{ Header(string) string }); ok {
		cookieHeader := hctx.Header("Cookie")
		parts := strings.Split(cookieHeader, ";")
		for _, p := range parts {
			kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
			if len(kv) == 2 && kv[0] == "refresh_token" {
				return kv[1]
			}
		}
	}
	return ""
}

func toMeByEmail(email string) meResponse {
	e := strings.ToLower(strings.TrimSpace(email))
	switch e {
	case "readonly@example.com":
		return meResponse{ID: "u-readonly", Name: "Readonly", Email: e, Roles: []idName{{ID: "2", Name: "readonly"}}, Permissions: []string{"users:read", "roles:read", "permissions:read", "permission-groups:read"}}
	case "users@example.com":
		return meResponse{ID: "u-users", Name: "Users Manager", Email: e, Roles: []idName{{ID: "3", Name: "users-manager"}}, Permissions: []string{"users:read", "users:write", "roles:read"}}
	case "roles@example.com":
		return meResponse{ID: "u-roles", Name: "Roles Manager", Email: e, Roles: []idName{{ID: "4", Name: "roles-manager"}}, Permissions: []string{"roles:read", "roles:write"}}
	case "guest@example.com":
		return meResponse{ID: "u-guest", Name: "Guest", Email: e, Roles: []idName{{ID: "5", Name: "guest"}}, Permissions: []string{}}
	default:
		return meResponse{ID: "u-admin", Name: "Admin", Email: "admin@example.com", Roles: []idName{{ID: "1", Name: "admin"}}, Permissions: []string{"users:read", "users:write", "roles:read", "roles:write", "permissions:read", "permissions:write", "permission-groups:read", "permission-groups:write"}}
	}
}

func (e *AuthEndpoint) tokenService() *authtoken.Service {
	return authtoken.NewService(authtoken.Config{Secret: e.cfg.Secret, Issuer: e.cfg.Issuer})
}

func accessTTL(cfg AuthConfig) time.Duration {
	ttl := time.Duration(cfg.AccessTTLMin) * time.Minute
	if ttl <= 0 {
		return 30 * time.Minute
	}
	return ttl
}

func refreshTTL(cfg AuthConfig) time.Duration {
	ttl := time.Duration(cfg.RefreshTTLHour) * time.Hour
	if ttl <= 0 {
		return 7 * 24 * time.Hour
	}
	return ttl
}

func principalEmail(principal authx.Principal) (string, bool) {
	if principal.Attributes == nil {
		return "", false
	}
	value, ok := principal.Attributes.Get("email")
	if !ok {
		return "", false
	}
	email, ok := value.(string)
	if !ok {
		return "", false
	}
	email = strings.TrimSpace(email)
	return email, email != ""
}

func (e *AuthEndpoint) CreateAuthLogin(_ context.Context, input *LoginInput) (*dynamicOutput, error) {
	if strings.TrimSpace(input.Body.Password) == "" {
		return nil, httpx.NewError(401, "invalid credentials")
	}

	tokens := e.tokenService()
	access, err := tokens.Issue(input.Body.Email, authtoken.TypeAccess, accessTTL(e.cfg))
	if err != nil {
		return nil, httpx.NewError(500, "failed to issue access token")
	}
	refresh, err := tokens.Issue(input.Body.Email, authtoken.TypeRefresh, refreshTTL(e.cfg))
	if err != nil {
		return nil, httpx.NewError(500, "failed to issue refresh token")
	}

	return &dynamicOutput{Body: ok(map[string]any{
		"accessToken":  access,
		"refreshToken": refresh,
		"user":         toMeByEmail(input.Body.Email),
	})}, nil
}

func (e *AuthEndpoint) CreateAuthRefresh(ctx context.Context, input *RefreshInput) (*dynamicOutput, error) {
	rt := strings.TrimSpace(input.Body.RefreshToken)
	if rt == "" {
		rt = parseRefreshCookie(ctx)
	}
	if rt == "" {
		return nil, httpx.NewError(401, "missing refresh token")
	}

	tokens := e.tokenService()
	claims, err := tokens.Parse(ctx, rt)
	if err != nil || claims.Type != authtoken.TypeRefresh {
		return nil, httpx.NewError(401, "invalid refresh token")
	}
	if isRevoked(ctx, e.cfg, e.kvClient, claims) {
		return nil, httpx.NewError(401, "refresh token revoked")
	}

	access, err := tokens.Issue(claims.Email, authtoken.TypeAccess, accessTTL(e.cfg))
	if err != nil {
		return nil, httpx.NewError(500, "failed to issue access token")
	}
	refresh, err := tokens.Issue(claims.Email, authtoken.TypeRefresh, refreshTTL(e.cfg))
	if err != nil {
		return nil, httpx.NewError(500, "failed to issue refresh token")
	}

	revokeToken(ctx, e.cfg, e.kvClient, claims)
	return &dynamicOutput{Body: ok(map[string]string{"accessToken": access, "refreshToken": refresh})}, nil
}

func (e *AuthEndpoint) CreateAuthLogout(ctx context.Context, input *LogoutInput) (*dynamicOutput, error) {
	tokens := e.tokenService()
	if token := parseBearerToken(ctx); token != "" {
		if claims, err := tokens.Parse(ctx, token); err == nil && claims.Type == authtoken.TypeAccess {
			revokeToken(ctx, e.cfg, e.kvClient, claims)
		}
	}

	rt := strings.TrimSpace(input.Body.RefreshToken)
	if rt == "" {
		rt = parseRefreshCookie(ctx)
	}
	if rt != "" {
		if claims, err := tokens.Parse(ctx, rt); err == nil && claims.Type == authtoken.TypeRefresh {
			revokeToken(ctx, e.cfg, e.kvClient, claims)
		}
	}

	return &dynamicOutput{Body: ok(map[string]bool{"ok": true})}, nil
}

func (e *AuthEndpoint) GetMe(ctx context.Context, _ *struct{}) (*dynamicOutput, error) {
	if principal, principalOK := authx.PrincipalFromContextAs[authx.Principal](ctx); principalOK {
		if email, emailOK := principalEmail(principal); emailOK {
			return &dynamicOutput{Body: ok(toMeByEmail(email))}, nil
		}
	}

	tokens := e.tokenService()
	if token := parseBearerToken(ctx); token != "" {
		claims, err := tokens.Parse(ctx, token)
		if err == nil && claims.Type == authtoken.TypeAccess {
			if isRevoked(ctx, e.cfg, e.kvClient, claims) {
				return nil, httpx.NewError(401, "access token revoked")
			}
			return &dynamicOutput{Body: ok(toMeByEmail(claims.Email))}, nil
		}
		return nil, httpx.NewError(401, "invalid access token")
	}

	email := "admin@example.com"
	if hctx, ok := ctx.(interface{ Header(string) string }); ok && strings.TrimSpace(hctx.Header("X-User-Email")) != "" {
		email = strings.TrimSpace(hctx.Header("X-User-Email"))
	}
	return &dynamicOutput{Body: ok(toMeByEmail(email))}, nil
}
