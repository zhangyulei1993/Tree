package logger

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"go.uber.org/zap"
)

var sensitiveFieldNames = map[string]struct{}{
	"access_token":  {},
	"refresh_token": {},
	"token":         {},
	"code":          {},
	"password":      {},
	"password_hash": {},
	"session_key":   {},
	"appsecret":     {},
	"app_secret":    {},
	"openid":        {},
	"unionid":       {},
}

func New(level string, env string) (*zap.Logger, error) {
	var cfg zap.Config
	if env == "prod" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	if level != "" {
		if err := cfg.Level.UnmarshalText([]byte(level)); err != nil {
			return nil, err
		}
	}

	return cfg.Build()
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

func SafeString(key string, value string) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key+"_masked", Mask(value))
	}
	return zap.String(key, value)
}

func Mask(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "***" + value[len(value)-4:]
}

func Hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	_, ok := sensitiveFieldNames[normalized]
	return ok
}
