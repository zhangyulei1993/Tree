package admin

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"tree/backend/internal/config"
	"tree/backend/internal/database"
	"tree/backend/internal/middleware"
	"tree/backend/internal/model"
	"tree/backend/internal/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	var admin model.AdminUser
	if err := database.DB.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		response.Error(c, 400, "wrong username or password")
		return
	}

	if admin.Status != 1 {
		response.Error(c, 403, "account disabled")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		response.Error(c, 400, "wrong username or password")
		return
	}

	token, err := middleware.GenerateAdminToken(admin.ID, admin.Username, admin.Role, h.cfg.JWTSecret)
	if err != nil {
		response.Error(c, 500, "token create failed")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       admin.ID,
			"username": admin.Username,
			"nickname": admin.Nickname,
			"role":     admin.Role,
		},
	})
}

func (h *Handler) Dashboard(c *gin.Context) {
	var bannerCount int64
	var articleCount int64
	var contactCount int64

	database.DB.Model(&model.Banner{}).Count(&bannerCount)
	database.DB.Model(&model.Article{}).Count(&articleCount)
	database.DB.Model(&model.ContactMessage{}).Count(&contactCount)

	response.Success(c, gin.H{
		"banner_count":  bannerCount,
		"article_count": articleCount,
		"contact_count": contactCount,
	})
}

func (h *Handler) SiteConfig(c *gin.Context) {
	var cfg model.SiteConfig
	database.DB.FirstOrCreate(&cfg, model.SiteConfig{ID: 1})
	response.Success(c, cfg)
}

func (h *Handler) UpdateSiteConfig(c *gin.Context) {
	var req model.SiteConfig

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	var cfg model.SiteConfig
	database.DB.FirstOrCreate(&cfg, model.SiteConfig{ID: 1})

	cfg.SiteName = req.SiteName
	cfg.Logo = req.Logo
	cfg.Phone = req.Phone
	cfg.Email = req.Email
	cfg.Address = req.Address
	cfg.Description = req.Description

	database.DB.Save(&cfg)

	response.Success(c, cfg)
}

func (h *Handler) BannerList(c *gin.Context) {
	var list []model.Banner
	database.DB.Order("sort desc, id desc").Find(&list)
	response.Success(c, list)
}

func (h *Handler) BannerCreate(c *gin.Context) {
	var req model.Banner

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.ImageURL == "" {
		response.Error(c, 400, "image is required")
		return
	}

	if err := database.DB.Create(&req).Error; err != nil {
		response.Error(c, 500, "create failed")
		return
	}

	response.Success(c, req)
}

func (h *Handler) BannerUpdate(c *gin.Context) {
	var req model.Banner

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.ID == 0 {
		response.Error(c, 400, "id is required")
		return
	}

	var item model.Banner
	if err := database.DB.First(&item, req.ID).Error; err != nil {
		response.Error(c, 404, "not found")
		return
	}

	item.Title = req.Title
	item.ImageURL = req.ImageURL
	item.LinkURL = req.LinkURL
	item.Sort = req.Sort
	item.Status = req.Status

	database.DB.Save(&item)

	response.Success(c, item)
}

func (h *Handler) BannerDelete(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&model.Banner{}, id)
	response.SuccessMsg(c, "deleted")
}

func (h *Handler) ArticleList(c *gin.Context) {
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

	db := database.DB.Model(&model.Article{})

	db.Count(&total)
	db.Order("sort desc, id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list)

	response.Page(c, list, total, page, pageSize)
}

func (h *Handler) ArticleCreate(c *gin.Context) {
	var req model.Article

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.Title == "" {
		response.Error(c, 400, "title is required")
		return
	}

	if err := database.DB.Create(&req).Error; err != nil {
		response.Error(c, 500, "create failed")
		return
	}

	response.Success(c, req)
}

func (h *Handler) ArticleUpdate(c *gin.Context) {
	var req model.Article

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.ID == 0 {
		response.Error(c, 400, "id is required")
		return
	}

	var item model.Article
	if err := database.DB.First(&item, req.ID).Error; err != nil {
		response.Error(c, 404, "not found")
		return
	}

	item.Title = req.Title
	item.Cover = req.Cover
	item.Summary = req.Summary
	item.Content = req.Content
	item.CategoryID = req.CategoryID
	item.Sort = req.Sort
	item.Status = req.Status

	database.DB.Save(&item)

	response.Success(c, item)
}

func (h *Handler) ArticleDelete(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&model.Article{}, id)
	response.SuccessMsg(c, "deleted")
}

func (h *Handler) ContactList(c *gin.Context) {
	var list []model.ContactMessage
	database.DB.Order("id desc").Find(&list)
	response.Success(c, list)
}

func (h *Handler) ContactDelete(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&model.ContactMessage{}, id)
	response.SuccessMsg(c, "deleted")
}

func (h *Handler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 400, "file is required")
		return
	}

	if err := os.MkdirAll(h.cfg.UploadDir, os.ModePerm); err != nil {
		response.Error(c, 500, "mkdir failed")
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
	savePath := filepath.Join(h.cfg.UploadDir, filename)

	src, err := file.Open()
	if err != nil {
		response.Error(c, 500, "open file failed")
		return
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		response.Error(c, 500, "save file failed")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		response.Error(c, 500, "copy file failed")
		return
	}

	url := h.cfg.StaticURL + "/" + filename

	record := model.UploadFile{
		Name:     file.Filename,
		URL:      url,
		Path:     savePath,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	database.DB.Create(&record)

	response.Success(c, gin.H{
		"url":  url,
		"file": record,
	})
}
