package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID        = "currentUserID"
	ContextAdminID       = "currentAdminID"
	ContextAdminRole     = "currentAdminRole"
	ContextJWTID         = "currentJWTID"
	ContextJWTExpiresAt  = "currentJWTExpiresAt"
	ContextPhoneVerified = "currentPhoneVerified"
)

var ErrContextValueMissing = errors.New("context value missing")

func CurrentUserID(ctx *gin.Context) (uint64, error) {
	return uint64Value(ctx, ContextUserID)
}

func CurrentAdminID(ctx *gin.Context) (uint64, error) {
	return uint64Value(ctx, ContextAdminID)
}

func CurrentAdminRole(ctx *gin.Context) (string, error) {
	value, ok := ctx.Get(ContextAdminRole)
	if !ok {
		return "", ErrContextValueMissing
	}
	role, ok := value.(string)
	if !ok || role == "" {
		return "", ErrContextValueMissing
	}
	return role, nil
}

func CurrentJWTID(ctx *gin.Context) (string, error) {
	value, ok := ctx.Get(ContextJWTID)
	if !ok {
		return "", ErrContextValueMissing
	}
	id, ok := value.(string)
	if !ok || id == "" {
		return "", ErrContextValueMissing
	}
	return id, nil
}

func CurrentJWTExpiresAt(ctx *gin.Context) (int64, error) {
	value, ok := ctx.Get(ContextJWTExpiresAt)
	if !ok {
		return 0, ErrContextValueMissing
	}
	expiresAt, ok := value.(int64)
	if !ok || expiresAt <= 0 {
		return 0, ErrContextValueMissing
	}
	return expiresAt, nil
}

func uint64Value(ctx *gin.Context, key string) (uint64, error) {
	value, ok := ctx.Get(key)
	if !ok {
		return 0, ErrContextValueMissing
	}
	id, ok := value.(uint64)
	if !ok || id == 0 {
		return 0, ErrContextValueMissing
	}
	return id, nil
}
