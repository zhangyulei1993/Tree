package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	commonjwt "tree/backend/internal/common/jwt"
	"tree/backend/internal/common/redis"
	"tree/backend/internal/common/response"
)

func UserAuth(manager *commonjwt.Manager, blacklist redis.TokenBlacklist) gin.HandlerFunc {
	return auth(manager, blacklist, commonjwt.TokenTypeUser)
}

func AdminAuth(manager *commonjwt.Manager, blacklist redis.TokenBlacklist) gin.HandlerFunc {
	return auth(manager, blacklist, commonjwt.TokenTypeAdmin)
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

func auth(manager *commonjwt.Manager, blacklist redis.TokenBlacklist, tokenType commonjwt.TokenType) gin.HandlerFunc {
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
			ctx.Set(ContextUserID, claims.Subject)
			// TODO(M5): load user status and phone_verified from database/cache.
			ctx.Set(ContextPhoneVerified, false)
		case commonjwt.TokenTypeAdmin:
			ctx.Set(ContextAdminID, claims.Subject)
			ctx.Set(ContextAdminRole, claims.Role)
		}

		ctx.Next()
	}
}
