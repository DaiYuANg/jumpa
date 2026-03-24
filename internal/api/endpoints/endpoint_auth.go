package endpoints

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/kvx"
	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
)

type AuthConfig struct {
	Secret         string
	Issuer         string
	AccessTTLMin   int
	RefreshTTLHour int
	UseValkey      bool
	RevokedPrefix  string
}

type authClaims struct {
	Email string `json:"email"`
	Type  string `json:"typ"`
	jwt.RegisteredClaims
}

func randomJTI() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func issueToken(cfg AuthConfig, email, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := authClaims{
		Email: strings.ToLower(strings.TrimSpace(email)),
		Type:  tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			Subject:   strings.ToLower(strings.TrimSpace(email)),
			ID:        randomJTI(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.Secret))
}

func parseToken(cfg AuthConfig, token string) (*authClaims, error) {
	t, err := jwt.ParseWithClaims(token, &authClaims{}, func(_ *jwt.Token) (any, error) {
		return []byte(cfg.Secret), nil
	}, jwt.WithIssuer(cfg.Issuer))
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*authClaims)
	if !ok || !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func revokedKey(cfg AuthConfig, jti string) string {
	prefix := cfg.RevokedPrefix
	if strings.TrimSpace(prefix) == "" {
		prefix = "auth:revoked"
	}
	return prefix + ":" + jti
}

func isRevoked(ctx context.Context, cfg AuthConfig, client kvx.Client, claims *authClaims) bool {
	if !cfg.UseValkey || claims == nil || claims.ID == "" {
		return false
	}
	exists, err := client.Exists(ctx, revokedKey(cfg, claims.ID))
	return err == nil && exists
}

func revokeToken(ctx context.Context, cfg AuthConfig, client kvx.Client, claims *authClaims) {
	if !cfg.UseValkey || claims == nil || claims.ID == "" || claims.ExpiresAt == nil {
		return
	}
	ttl := time.Until(claims.ExpiresAt.Time)
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

func registerAuthEndpoints(api *httpx.Group, cfg AuthConfig, kvClient kvx.Client) {
	httpx.MustGroupPost(api, "/auth/login", func(ctx context.Context, input *LoginInput) (*dynamicOutput, error) {
		if strings.TrimSpace(input.Body.Password) == "" {
			return nil, httpx.NewError(401, "invalid credentials")
		}
		accessTTL := time.Duration(cfg.AccessTTLMin) * time.Minute
		refreshTTL := time.Duration(cfg.RefreshTTLHour) * time.Hour
		if accessTTL <= 0 {
			accessTTL = 30 * time.Minute
		}
		if refreshTTL <= 0 {
			refreshTTL = 7 * 24 * time.Hour
		}
		access, err := issueToken(cfg, input.Body.Email, "access", accessTTL)
		if err != nil {
			return nil, httpx.NewError(500, "failed to issue access token")
		}
		refresh, err := issueToken(cfg, input.Body.Email, "refresh", refreshTTL)
		if err != nil {
			return nil, httpx.NewError(500, "failed to issue refresh token")
		}
		return &dynamicOutput{Body: ok(map[string]any{"accessToken": access, "refreshToken": refresh, "user": toMeByEmail(input.Body.Email)})}, nil
	}, huma.OperationTags("auth"))

	httpx.MustGroupPost(api, "/auth/refresh", func(ctx context.Context, input *RefreshInput) (*dynamicOutput, error) {
		rt := strings.TrimSpace(input.Body.RefreshToken)
		if rt == "" {
			rt = parseRefreshCookie(ctx)
		}
		if rt == "" {
			return nil, httpx.NewError(401, "missing refresh token")
		}
		claims, err := parseToken(cfg, rt)
		if err != nil || claims.Type != "refresh" {
			return nil, httpx.NewError(401, "invalid refresh token")
		}
		if isRevoked(ctx, cfg, kvClient, claims) {
			return nil, httpx.NewError(401, "refresh token revoked")
		}
		accessTTL := time.Duration(cfg.AccessTTLMin) * time.Minute
		if accessTTL <= 0 {
			accessTTL = 30 * time.Minute
		}
		refreshTTL := time.Duration(cfg.RefreshTTLHour) * time.Hour
		if refreshTTL <= 0 {
			refreshTTL = 7 * 24 * time.Hour
		}
		access, err := issueToken(cfg, claims.Email, "access", accessTTL)
		if err != nil {
			return nil, httpx.NewError(500, "failed to issue access token")
		}
		refresh, err := issueToken(cfg, claims.Email, "refresh", refreshTTL)
		if err != nil {
			return nil, httpx.NewError(500, "failed to issue refresh token")
		}
		// rotate refresh token and invalidate the old one
		revokeToken(ctx, cfg, kvClient, claims)
		return &dynamicOutput{Body: ok(map[string]string{"accessToken": access, "refreshToken": refresh})}, nil
	}, huma.OperationTags("auth"))

	httpx.MustGroupPost(api, "/auth/logout", func(ctx context.Context, input *LogoutInput) (*dynamicOutput, error) {
		if token := parseBearerToken(ctx); token != "" {
			if claims, err := parseToken(cfg, token); err == nil && claims.Type == "access" {
				revokeToken(ctx, cfg, kvClient, claims)
			}
		}
		rt := strings.TrimSpace(input.Body.RefreshToken)
		if rt == "" {
			rt = parseRefreshCookie(ctx)
		}
		if rt != "" {
			if claims, err := parseToken(cfg, rt); err == nil && claims.Type == "refresh" {
				revokeToken(ctx, cfg, kvClient, claims)
			}
		}
		return &dynamicOutput{Body: ok(map[string]bool{"ok": true})}, nil
	}, huma.OperationTags("auth"))

	httpx.MustGroupGet(api, "/me", func(ctx context.Context, _ *struct{}) (*dynamicOutput, error) {
		token := parseBearerToken(ctx)
		if token != "" {
			claims, err := parseToken(cfg, token)
			if err == nil && claims.Type == "access" {
				if isRevoked(ctx, cfg, kvClient, claims) {
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
	}, huma.OperationTags("auth"))
}
