package router

import (
	webHandler "tree/backend/internal/handler/web"

	"github.com/gin-gonic/gin"
)

func RegisterWebRoutes(r *gin.RouterGroup) {
	r.GET("/site/config", webHandler.SiteConfig)
	r.GET("/banner/list", webHandler.BannerList)
	r.GET("/article/list", webHandler.ArticleList)
	r.GET("/article/detail/:id", webHandler.ArticleDetail)
	r.POST("/contact/submit", webHandler.ContactSubmit)

	r.GET("/family/tree", webHandler.FamilyTree)
	r.GET("/family/member/:id", webHandler.FamilyMemberDetail)
}
