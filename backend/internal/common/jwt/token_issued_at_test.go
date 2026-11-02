package jwt

import (
	"testing"
	"time"

	"tree/backend/internal/common/config"
)

func TestGenerateAccessTokenWithIssuedAtAfterRevokeCutoff(t *testing.T) {
	manager, err := NewManager(config.JWTConfig{
		UserSecret:               "user-secret-test",
		AdminSecret:              "admin-secret-test",
		AccessTokenExpireMinutes: 30,
	}, "TreeTest")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	revokedAt := time.Now().Unix()
	token, err := manager.GenerateAccessTokenWithIssuedAt(42, TokenTypeUser, "", "jti-1", revokedAt+1)
	if err != nil {
		t.Fatalf("GenerateAccessTokenWithIssuedAt: %v", err)
	}
	claims, err := manager.Parse(token, TokenTypeUser)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.IssuedAt != revokedAt+1 {
		t.Fatalf("unexpected iat: %d", claims.IssuedAt)
	}
}
