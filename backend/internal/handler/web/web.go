package web

import (
	"strconv"

	"tree/backend/internal/database"
	"tree/backend/internal/model"
	"tree/backend/internal/response"

	"github.com/gin-gonic/gin"
)

func SiteConfig(c *gin.Context) {
	var cfg model.SiteConfig
	database.DB.FirstOrCreate(&cfg, model.SiteConfig{ID: 1})
	response.Success(c, cfg)
}

func BannerList(c *gin.Context) {
	var list []model.Banner
	database.DB.Where("status = ?", 1).Order("sort desc, id desc").Find(&list)
	response.Success(c, list)
}

func ArticleList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var total int64
	var list []model.Article

	db := database.DB.Model(&model.Article{}).Where("status = ?", 1)

	db.Count(&total)
	db.Order("sort desc, id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list)

	response.Page(c, list, total, page, pageSize)
}

func ArticleDetail(c *gin.Context) {
	id := c.Param("id")

	var article model.Article
	if err := database.DB.Where("id = ? AND status = ?", id, 1).First(&article).Error; err != nil {
		response.Error(c, 404, "article not found")
		return
	}

	response.Success(c, article)
}

func ContactSubmit(c *gin.Context) {
	var req model.ContactMessage

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.Name == "" && req.Phone == "" && req.Email == "" {
		response.Error(c, 400, "contact is required")
		return
	}

	if err := database.DB.Create(&req).Error; err != nil {
		response.Error(c, 500, "submit failed")
		return
	}

	response.SuccessMsg(c, "submitted")
}
