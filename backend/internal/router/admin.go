package router

import (
	"tree/backend/internal/config"
	adminHandler "tree/backend/internal/handler/admin"
	"tree/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(r *gin.RouterGroup, cfg *config.Config) {
	h := adminHandler.New(cfg)

	r.POST("/login", h.Login)

	auth := r.Group("")
	auth.Use(middleware.AdminAuth(cfg.JWTSecret))
	{
		auth.GET("/dashboard", h.Dashboard)

		auth.GET("/site/config", h.SiteConfig)
		auth.POST("/site/config/update", h.UpdateSiteConfig)

		auth.GET("/banner/list", h.BannerList)
		auth.POST("/banner/create", h.BannerCreate)
		auth.POST("/banner/update", h.BannerUpdate)
		auth.DELETE("/banner/delete/:id", h.BannerDelete)

		auth.GET("/article/list", h.ArticleList)
		auth.POST("/article/create", h.ArticleCreate)
		auth.POST("/article/update", h.ArticleUpdate)
		auth.DELETE("/article/delete/:id", h.ArticleDelete)

		auth.GET("/contact/list", h.ContactList)
		auth.DELETE("/contact/delete/:id", h.ContactDelete)

		auth.GET("/family/list", h.FamilyList)
		auth.POST("/family/create", h.FamilyCreate)
		auth.POST("/family/update", h.FamilyUpdate)
		auth.DELETE("/family/delete/:id", h.FamilyDelete)

		auth.GET("/member/list", h.MemberList)
		auth.POST("/member/create", h.MemberCreate)
		auth.POST("/member/update", h.MemberUpdate)
		auth.DELETE("/member/delete/:id", h.MemberDelete)

		auth.GET("/relationship/list", h.RelationshipList)
		auth.POST("/relationship/create", h.RelationshipCreate)
		auth.POST("/relationship/update", h.RelationshipUpdate)
		auth.DELETE("/relationship/delete/:id", h.RelationshipDelete)

		auth.GET("/public-tree/config", h.PublicTreeConfig)
		auth.POST("/public-tree/config/update", h.PublicTreeConfigUpdate)
		auth.GET("/public-tree/relationship-type/list", h.RelationshipTypeList)
		auth.POST("/public-tree/relationship-type/create", h.RelationshipTypeCreate)
		auth.POST("/public-tree/relationship-type/update", h.RelationshipTypeUpdate)
		auth.DELETE("/public-tree/relationship-type/delete/:id", h.RelationshipTypeDelete)
		auth.GET("/public-tree/node/list", h.PublicTreeNodeList)
		auth.POST("/public-tree/node/create", h.PublicTreeNodeCreate)
		auth.POST("/public-tree/node/update", h.PublicTreeNodeUpdate)
		auth.DELETE("/public-tree/node/delete/:id", h.PublicTreeNodeDelete)
		auth.GET("/public-tree/relationship/list", h.PublicTreeRelationshipList)
		auth.POST("/public-tree/relationship/create", h.PublicTreeRelationshipCreate)
		auth.POST("/public-tree/relationship/update", h.PublicTreeRelationshipUpdate)
		auth.DELETE("/public-tree/relationship/delete/:id", h.PublicTreeRelationshipDelete)

		auth.POST("/upload", h.Upload)
	}
}
