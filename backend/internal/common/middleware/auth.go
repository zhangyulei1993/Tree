package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	commonjwt "tree/backend/internal/common/jwt"
	"tree/backend/internal/common/redis"
	"tree/backend/internal/common/response"
)

type UserAuthState struct {
	Status        string
	PhoneVerified bool
}

type UserAuthStateLookup func(context.Context, uint64) (UserAuthState, error)

func UserAuth(manager *commonjwt.Manager, blacklist redis.TokenBlacklist) gin.HandlerFunc {
	return auth(manager, blacklist, commonjwt.TokenTypeUser, nil)
}

func UserAuthWithStateLookup(manager *commonjwt.Manager, blacklist redis.TokenBlacklist, lookup UserAuthStateLookup) gin.HandlerFunc {
	return auth(manager, blacklist, commonjwt.TokenTypeUser, lookup)
}

func AdminAuth(manager *commonjwt.Manager, blacklist redis.TokenBlacklist) gin.HandlerFunc {
	return auth(manager, blacklist, commonjwt.TokenTypeAdmin, nil)
}

func RequirePhoneVerified() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		verified, _ := ctx.Get(ContextPhoneVerified)
		if ok, _ := verified.(bool); !ok {
			response.Abort(ctx, http.StatusForbidden, apperrors.CodePhoneBindRequired)
			return
		}
		ctx.Next()
	}
}

func auth(manager *commonjwt.Manager, blacklist redis.TokenBlacklist, tokenType commonjwt.TokenType, lookup UserAuthStateLookup) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := commonjwt.BearerToken(ctx.GetHeader("Authorization"))
		if err != nil {
			response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
			return
		}

		claims, err := manager.Parse(token, tokenType)
		if err != nil {
			response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeLoginExpired)
			return
		}

		if blacklist != nil {
			revoked, err := blacklist.IsRevoked(ctx.Request.Context(), claims.TokenID)
			if err != nil || revoked {
				response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeLoginExpired)
				return
			}
			if tokenType == commonjwt.TokenTypeUser {
				userRevoked, err := blacklist.IsUserTokenRevoked(ctx.Request.Context(), claims.Subject, claims.IssuedAt)
				if err != nil || userRevoked {
					response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeLoginExpired)
					return
				}
			}
		}

		ctx.Set(ContextJWTID, claims.TokenID)
		ctx.Set(ContextJWTExpiresAt, claims.ExpiresAt)
		switch tokenType {
		case commonjwt.TokenTypeUser:
			if lookup != nil {
				state, err := lookup(ctx.Request.Context(), claims.Subject)
				if err != nil {
					response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeLoginExpired)
					return
				}
				if state.Status != "ACTIVE" && state.Status != "PENDING_BIND" {
					response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeLoginDisabled)
					return
				}
				ctx.Set(ContextPhoneVerified, state.PhoneVerified)
				ctx.Set(ContextUserStatus, state.Status)
			} else {
				ctx.Set(ContextPhoneVerified, false)
			}
			ctx.Set(ContextUserID, claims.Subject)
			ctx.Set(ContextUserRole, claims.Role)
		case commonjwt.TokenTypeAdmin:
			ctx.Set(ContextAdminID, claims.Subject)
			ctx.Set(ContextAdminRole, claims.Role)
		}

		ctx.Next()
	}
}
