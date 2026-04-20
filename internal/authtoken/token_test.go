package authtoken

import (
	"context"
	"testing"
	"time"

	authjwt "github.com/DaiYuANg/arcgo/authx/jwt"
	"github.com/golang-jwt/jwt/v5"
)

func TestServiceIssueAndParse(t *testing.T) {
	service := NewService(Config{Secret: "test-secret", Issuer: "jumpa-test"})

	token, err := service.Issue(" Admin@Example.com ", TypeAccess, time.Hour)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	claims, err := service.Parse(context.Background(), token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.Email != "admin@example.com" {
		t.Fatalf("email = %q, want admin@example.com", claims.Email)
	}
	if claims.Type != TypeAccess {
		t.Fatalf("type = %q, want %q", claims.Type, TypeAccess)
	}
	if claims.ID == "" {
		t.Fatal("expected jti")
	}
	if claims.ExpiresAt.IsZero() {
		t.Fatal("expected expires_at")
	}
}

func TestServiceParseLegacyTypeClaim(t *testing.T) {
	cfg := Config{Secret: "test-secret", Issuer: "jumpa-test"}
	service := NewService(cfg)
	now := time.Now().UTC()
	legacyClaims := tokenClaims{
		Email: "Legacy@Example.com",
		Type:  TypeRefresh,
		Claims: authjwt.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    cfg.Issuer,
				Subject:   "legacy@example.com",
				ID:        "legacy-jti",
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			},
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, legacyClaims).SignedString([]byte(cfg.Secret))
	if err != nil {
		t.Fatalf("sign legacy token: %v", err)
	}

	claims, err := service.Parse(context.Background(), token)
	if err != nil {
		t.Fatalf("parse legacy token: %v", err)
	}
	if claims.Type != TypeRefresh {
		t.Fatalf("type = %q, want %q", claims.Type, TypeRefresh)
	}
	if claims.Email != "legacy@example.com" {
		t.Fatalf("email = %q, want legacy@example.com", claims.Email)
	}
}

func TestServiceRejectsInconsistentTokenTypeClaims(t *testing.T) {
	cfg := Config{Secret: "test-secret", Issuer: "jumpa-test"}
	service := NewService(cfg)
	now := time.Now().UTC()
	claims := tokenClaims{
		Email: "user@example.com",
		Type:  TypeRefresh,
		Claims: authjwt.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    cfg.Issuer,
				Subject:   "user@example.com",
				Audience:  jwt.ClaimStrings{TypeAccess},
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			},
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.Secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := service.Parse(context.Background(), token); err == nil {
		t.Fatal("expected inconsistent token type claims to be rejected")
	}
}

func TestAuthenticationProviderUsesJWTTokenCredential(t *testing.T) {
	cfg := Config{Secret: "test-secret", Issuer: "jumpa-test"}
	service := NewService(cfg)
	token, err := service.Issue("user@example.com", TypeAccess, time.Hour)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	provider := NewAuthenticationProvider(cfg, nil)
	result, err := provider.AuthenticateAny(context.Background(), authjwt.NewTokenCredential(token))
	if err != nil {
		t.Fatalf("authenticate token credential: %v", err)
	}

	claims, ok := result.Principal.(*Claims)
	if !ok {
		t.Fatalf("principal type = %T, want *Claims", result.Principal)
	}
	if claims.Email != "user@example.com" || claims.Type != TypeAccess {
		t.Fatalf("claims = %#v", claims)
	}
}
