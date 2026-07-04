package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"tree/backend/internal/common/config"
)

type TokenType string

const (
	TokenTypeUser  TokenType = "USER"
	TokenTypeAdmin TokenType = "ADMIN"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenTypeInvalid = errors.New("token type invalid")
)

type Claims struct {
	Subject   uint64    `json:"sub"`
	TokenType TokenType `json:"typ"`
	Issuer    string    `json:"iss"`
	IssuedAt  int64     `json:"iat"`
	ExpiresAt int64     `json:"exp"`
	TokenID   string    `json:"jti,omitempty"`
	Role      string    `json:"role,omitempty"`
}

type Manager struct {
	userSecret  []byte
	adminSecret []byte
	issuer      string
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

const maxIssuedAtFutureSkew = time.Minute

func NewManager(cfg config.JWTConfig, appName string) (*Manager, error) {
	if cfg.UserSecret == "" || cfg.AdminSecret == "" {
		return nil, errors.New("jwt secrets are required")
	}
	if cfg.UserSecret == cfg.AdminSecret {
		return nil, errors.New("jwt user secret and admin secret must be different")
	}

	accessTTL := time.Duration(cfg.AccessTokenExpireMinutes) * time.Minute
	if accessTTL <= 0 {
		accessTTL = 2 * time.Hour
	}
	refreshTTL := time.Duration(cfg.RefreshTokenExpireDays) * 24 * time.Hour
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	if appName == "" {
		appName = "Tree"
	}

	return &Manager{
		userSecret:  []byte(cfg.UserSecret),
		adminSecret: []byte(cfg.AdminSecret),
		issuer:      appName,
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
	}, nil
}

func (m *Manager) GenerateAccessToken(subject uint64, tokenType TokenType, role string, tokenID string) (string, error) {
	return m.GenerateAccessTokenWithIssuedAt(subject, tokenType, role, tokenID, time.Now().Unix())
}

func (m *Manager) GenerateAccessTokenWithIssuedAt(subject uint64, tokenType TokenType, role string, tokenID string, issuedAt int64) (string, error) {
	if subject == 0 {
		return "", errors.New("jwt subject is required")
	}
	if issuedAt <= 0 {
		return "", errors.New("jwt issuedAt is required")
	}
	return m.generateSignedToken(subject, tokenType, role, tokenID, issuedAt, int64(m.accessTTL.Seconds()))
}

func (m *Manager) GenerateToken(subject uint64, tokenType TokenType, role string, tokenID string, ttl time.Duration) (string, error) {
	if subject == 0 {
		return "", errors.New("jwt subject is required")
	}
	if ttl <= 0 {
		return "", errors.New("jwt ttl must be positive")
	}
	return m.generateSignedToken(subject, tokenType, role, tokenID, time.Now().Unix(), int64(ttl.Seconds()))
}

func (m *Manager) generateSignedToken(subject uint64, tokenType TokenType, role string, tokenID string, issuedAt int64, ttlSeconds int64) (string, error) {
	if ttlSeconds <= 0 {
		return "", errors.New("jwt ttl must be positive")
	}
	claims := Claims{
		Subject:   subject,
		TokenType: tokenType,
		Issuer:    m.issuer,
		IssuedAt:  issuedAt,
		ExpiresAt: issuedAt + ttlSeconds,
		TokenID:   tokenID,
		Role:      role,
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	signature, err := m.sign(unsigned, tokenType)
	if err != nil {
		return "", err
	}

	return unsigned + "." + signature, nil
}

func (m *Manager) GenerateRefreshToken(subject uint64, tokenType TokenType, role string, tokenID string) (string, error) {
	return m.GenerateToken(subject, tokenType, role, tokenID, m.refreshTTL)
}

func (m *Manager) Parse(token string, expectedType TokenType) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	unsigned := parts[0] + "." + parts[1]
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if claims.TokenType != expectedType {
		return nil, ErrTokenTypeInvalid
	}
	now := time.Now()
	if claims.Subject == 0 ||
		claims.Issuer == "" ||
		claims.Issuer != m.issuer ||
		claims.IssuedAt <= 0 ||
		claims.ExpiresAt <= 0 ||
		claims.ExpiresAt <= claims.IssuedAt ||
		claims.IssuedAt > now.Add(maxIssuedAtFutureSkew).Unix() {
		return nil, ErrInvalidToken
	}

	expectedSignature, err := m.sign(unsigned, claims.TokenType)
	if err != nil {
		return nil, err
	}
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return nil, ErrInvalidToken
	}
	if claims.ExpiresAt <= now.Unix() {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}

func (m *Manager) sign(unsigned string, tokenType TokenType) (string, error) {
	var secret []byte
	switch tokenType {
	case TokenTypeUser:
		secret = m.userSecret
	case TokenTypeAdmin:
		secret = m.adminSecret
	default:
		return "", ErrTokenTypeInvalid
	}

	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write([]byte(unsigned)); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func BearerToken(authorization string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(authorization, prefix) {
		return "", ErrInvalidToken
	}
	token := strings.TrimSpace(strings.TrimPrefix(authorization, prefix))
	if token == "" {
		return "", ErrInvalidToken
	}
	return token, nil
}

func SubjectString(subject uint64) string {
	return strconv.FormatUint(subject, 10)
}

func ParseSubject(subject string) (uint64, error) {
	id, err := strconv.ParseUint(subject, 10, 64)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid jwt subject: %w", err)
	}
	return id, nil
}
