package authtoken

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/DaiYuANg/arcgo/authx"
	authjwt "github.com/DaiYuANg/arcgo/authx/jwt"
	"github.com/golang-jwt/jwt/v5"
)

const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

type Config struct {
	Secret string
	Issuer string
}

type Service struct {
	cfg      Config
	provider *authjwt.Provider
}

type Claims struct {
	Email     string
	Type      string
	ID        string
	ExpiresAt time.Time
}

type tokenClaims struct {
	Email string `json:"email,omitempty"`
	Type  string `json:"typ,omitempty"`
	authjwt.Claims
}

func NewService(cfg Config) *Service {
	return &Service{
		cfg: cfg,
		provider: authjwt.NewProvider(
			authjwt.WithHMACSecret([]byte(cfg.Secret)),
			authjwt.WithParserOptions(jwt.WithIssuer(cfg.Issuer)),
			authjwt.WithClaimsMapper(func(_ context.Context, claims *authjwt.Claims) (authx.AuthenticationResult, error) {
				if claims == nil {
					return authx.AuthenticationResult{}, authx.ErrInvalidAuthenticationCredential
				}
				return authx.AuthenticationResult{Principal: claims}, nil
			}),
		),
	}
}

func (s *Service) Issue(email, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	normalizedEmail := normalizeEmail(email)
	jti, err := randomJTI()
	if err != nil {
		return "", err
	}
	claims := tokenClaims{
		Email: normalizedEmail,
		Type:  tokenType,
		Claims: authjwt.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    s.cfg.Issuer,
				Subject:   normalizedEmail,
				Audience:  audience(tokenType),
				ID:        jti,
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			},
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.Secret))
}

func (s *Service) Parse(ctx context.Context, token string) (*Claims, error) {
	result, err := s.provider.Authenticate(ctx, authjwt.NewTokenCredential(token))
	if err != nil {
		return nil, err
	}
	verifiedClaims, ok := result.Principal.(*authjwt.Claims)
	if !ok || verifiedClaims == nil {
		return nil, authx.ErrInvalidAuthenticationCredential
	}

	// The signature and registered claims are already verified; this only reads
	// legacy custom fields that authx/jwt intentionally does not model.
	legacyClaims := parseLegacyClaims(token)
	return normalizeClaims(verifiedClaims, legacyClaims)
}

func randomJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func audience(tokenType string) jwt.ClaimStrings {
	tokenType = strings.TrimSpace(tokenType)
	if tokenType == "" {
		return nil
	}
	return jwt.ClaimStrings{tokenType}
}

func parseLegacyClaims(token string) *tokenClaims {
	claims := &tokenClaims{}
	_, _, _ = jwt.NewParser().ParseUnverified(token, claims)
	return claims
}

func normalizeClaims(verified *authjwt.Claims, legacy *tokenClaims) (*Claims, error) {
	email, err := mergeEmail(verified.Subject, legacy.Email)
	if err != nil {
		return nil, err
	}
	tokenType, err := mergeType(firstAudience(verified.Audience), legacy.Type)
	if err != nil {
		return nil, err
	}
	if email == "" || tokenType == "" {
		return nil, authx.ErrInvalidAuthenticationCredential
	}

	expiresAt := time.Time{}
	if verified.ExpiresAt != nil {
		expiresAt = verified.ExpiresAt.Time
	} else if legacy.ExpiresAt != nil {
		expiresAt = legacy.ExpiresAt.Time
	}

	id := verified.ID
	if id == "" {
		id = legacy.ID
	}

	return &Claims{
		Email:     email,
		Type:      tokenType,
		ID:        id,
		ExpiresAt: expiresAt,
	}, nil
}

func mergeEmail(subject, legacyEmail string) (string, error) {
	subject = normalizeEmail(subject)
	legacyEmail = normalizeEmail(legacyEmail)
	if subject != "" && legacyEmail != "" && subject != legacyEmail {
		return "", authx.ErrInvalidAuthenticationCredential
	}
	if subject != "" {
		return subject, nil
	}
	return legacyEmail, nil
}

func mergeType(audienceType, legacyType string) (string, error) {
	audienceType = strings.TrimSpace(audienceType)
	legacyType = strings.TrimSpace(legacyType)
	if audienceType != "" && legacyType != "" && audienceType != legacyType {
		return "", authx.ErrInvalidAuthenticationCredential
	}
	if audienceType != "" {
		return audienceType, nil
	}
	return legacyType, nil
}

func firstAudience(audience jwt.ClaimStrings) string {
	for _, item := range audience {
		if item = strings.TrimSpace(item); item != "" {
			return item
		}
	}
	return ""
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
