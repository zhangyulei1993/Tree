package middleware

import "github.com/gin-gonic/gin"

func UserAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
	}
}

func AdminAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
	}
}
