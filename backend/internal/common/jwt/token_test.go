package jwt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"tree/backend/internal/common/config"
)

func testManager(t *testing.T) *Manager {
	t.Helper()
	manager, err := NewManager(config.JWTConfig{
		UserSecret:               "unit-test-user-secret",
		AdminSecret:              "unit-test-admin-secret",
		AccessTokenExpireMinutes: 5,
	}, "TreeTest")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return manager
}

func TestTokenTypeIsolation(t *testing.T) {
	manager := testManager(t)
	userToken, err := manager.GenerateAccessToken(11, TokenTypeUser, "", "user-jti")
	if err != nil {
		t.Fatalf("GenerateAccessToken USER: %v", err)
	}
	adminToken, err := manager.GenerateAccessToken(22, TokenTypeAdmin, "ROOT_ADMIN", "admin-jti")
	if err != nil {
		t.Fatalf("GenerateAccessToken ADMIN: %v", err)
	}

	if _, err := manager.Parse(userToken, TokenTypeAdmin); !errors.Is(err, ErrTokenTypeInvalid) {
		t.Fatalf("USER token parsed as ADMIN: %v", err)
	}
	if _, err := manager.Parse(adminToken, TokenTypeUser); !errors.Is(err, ErrTokenTypeInvalid) {
		t.Fatalf("ADMIN token parsed as USER: %v", err)
	}
}

func TestExpiredAndForgedTokensAreRejected(t *testing.T) {
	manager := testManager(t)
	now := time.Now()
	expired := signedToken(t, manager, Claims{
		Subject: 11, TokenType: TokenTypeUser, Issuer: "TreeTest",
		IssuedAt: now.Add(-2 * time.Hour).Unix(), ExpiresAt: now.Add(-time.Hour).Unix(),
		TokenID: "expired-jti",
	})
	if _, err := manager.Parse(expired, TokenTypeUser); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expired token error = %v", err)
	}

	valid, err := manager.GenerateAccessToken(11, TokenTypeUser, "", "valid-jti")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	parts := strings.Split(valid, ".")
	parts[2] = base64.RawURLEncoding.EncodeToString([]byte("forged-signature"))
	if _, err := manager.Parse(strings.Join(parts, "."), TokenTypeUser); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("forged token error = %v", err)
	}
}

func TestMissingRequiredClaimsAreRejected(t *testing.T) {
	manager := testManager(t)
	now := time.Now()
	tests := []struct {
		name   string
		claims Claims
	}{
		{
			name: "missing subject",
			claims: Claims{
				TokenType: TokenTypeUser, Issuer: "TreeTest",
				IssuedAt: now.Unix(), ExpiresAt: now.Add(time.Hour).Unix(),
			},
		},
		{
			name: "missing issuer",
			claims: Claims{
				Subject: 11, TokenType: TokenTypeUser,
				IssuedAt: now.Unix(), ExpiresAt: now.Add(time.Hour).Unix(),
			},
		},
		{
			name: "missing issued at",
			claims: Claims{
				Subject: 11, TokenType: TokenTypeUser, Issuer: "TreeTest",
				ExpiresAt: now.Add(time.Hour).Unix(),
			},
		},
		{
			name: "missing expires at",
			claims: Claims{
				Subject: 11, TokenType: TokenTypeUser, Issuer: "TreeTest",
				IssuedAt: now.Unix(),
			},
		},
		{
			name: "wrong issuer",
			claims: Claims{
				Subject: 11, TokenType: TokenTypeUser, Issuer: "OtherIssuer",
				IssuedAt: now.Unix(), ExpiresAt: now.Add(time.Hour).Unix(),
			},
		},
		{
			name: "issued at too far in future",
			claims: Claims{
				Subject: 11, TokenType: TokenTypeUser, Issuer: "TreeTest",
				IssuedAt: now.Add(2 * time.Minute).Unix(), ExpiresAt: now.Add(time.Hour).Unix(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := signedToken(t, manager, tt.claims)
			if _, err := manager.Parse(token, TokenTypeUser); !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("missing required claim accepted, err = %v", err)
			}
		})
	}
}

func signedToken(t *testing.T, manager *Manager, claims Claims) string {
	t.Helper()
	headerJSON, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	claimsJSON, _ := json.Marshal(claims)
	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." +
		base64.RawURLEncoding.EncodeToString(claimsJSON)
	signature, err := manager.sign(unsigned, claims.TokenType)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return unsigned + "." + signature
}
