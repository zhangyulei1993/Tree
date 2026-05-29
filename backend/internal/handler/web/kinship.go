package web

import (
	"tree/backend/internal/kinship"
	"tree/backend/internal/response"

	"github.com/gin-gonic/gin"
)

func KinshipResolve(c *gin.Context) {
	var req kinship.ResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	response.Success(c, kinship.Resolve(req))
}
