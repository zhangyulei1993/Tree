package router

import (
	"tree/backend/internal/config"
	"tree/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.Cors())

	r.Static("/uploads", cfg.UploadDir)

	api := r.Group("/api")

	RegisterWebRoutes(api.Group("/web"))
	RegisterAdminRoutes(api.Group("/admin"), cfg)

	return r
}
