package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/response"
)

func RequireNoAssistance() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAssistanceSession(c) {
			response.Abort(c, http.StatusForbidden, apperrors.CodeForbidden)
			return
		}
		c.Next()
	}
}
func RestrictAssistanceToFamilyWork() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAssistanceSession(c) {
			c.Next()
			return
		}
		allowed := map[string]bool{
			"GET /api/families": true, "GET /api/families/:familyId": true, "GET /api/families/:familyId/members": true, "GET /api/families/:familyId/members/:memberId": true, "GET /api/families/:familyId/tree": true,
			"POST /api/families/:familyId/members": true, "PUT /api/families/:familyId/members/:memberId": true, "POST /api/families/:familyId/relationships": true, "POST /api/families/:familyId/relationships/place-existing": true, "PUT /api/families/:familyId/relationships/:relationshipId": true,
		}
		if !allowed[c.Request.Method+" "+c.FullPath()] {
			response.Abort(c, http.StatusForbidden, apperrors.CodeForbidden)
			return
		}
		c.Next()
	}
}
