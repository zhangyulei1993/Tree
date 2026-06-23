package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	contentdto "tree/backend/internal/content/dto"
	contentservice "tree/backend/internal/content/service"
)

type Handler struct{ service contentservice.Service }

func NewHandler(service contentservice.Service) *Handler { return &Handler{service: service} }

func (h *Handler) PublicCategories(ctx *gin.Context) {
	result, err := h.service.ListPublicCategories(ctx.Request.Context())
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) PublicArticles(ctx *gin.Context) {
	result, err := h.service.ListPublicArticles(ctx.Request.Context(), listArticlesQuery(ctx, false))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) PublicArticleDetail(ctx *gin.Context) {
	result, err := h.service.PublicArticleDetail(ctx.Request.Context(), ctx.Param("articleId"))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) AdminCategories(ctx *gin.Context) {
	result, err := h.service.ListAdminCategories(ctx.Request.Context(), ctx.Query("includeInactive") == "true")
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) CreateCategory(ctx *gin.Context) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	var req contentdto.CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		bad(ctx)
		return
	}
	result, err := h.service.CreateCategory(ctx.Request.Context(), adminID, role, req)
	write(ctx, http.StatusCreated, result, err)
}

func (h *Handler) UpdateCategory(ctx *gin.Context) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	categoryID, ok := pathID(ctx, "categoryId")
	if !ok {
		return
	}
	var req contentdto.UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		bad(ctx)
		return
	}
	result, err := h.service.UpdateCategory(ctx.Request.Context(), adminID, role, categoryID, req)
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) DeleteCategory(ctx *gin.Context) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	categoryID, ok := pathID(ctx, "categoryId")
	if !ok {
		return
	}
	err := h.service.DeleteCategory(ctx.Request.Context(), adminID, role, categoryID)
	write(ctx, http.StatusOK, map[string]bool{"deleted": err == nil}, err)
}

func (h *Handler) AdminArticles(ctx *gin.Context) {
	result, err := h.service.ListAdminArticles(ctx.Request.Context(), listArticlesQuery(ctx, true))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) AdminArticleDetail(ctx *gin.Context) {
	articleID, ok := pathID(ctx, "articleId")
	if !ok {
		return
	}
	result, err := h.service.AdminArticleDetail(ctx.Request.Context(), articleID)
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) CreateArticle(ctx *gin.Context) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	var req contentdto.CreateArticleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		bad(ctx)
		return
	}
	result, err := h.service.CreateArticle(ctx.Request.Context(), adminID, role, req)
	write(ctx, http.StatusCreated, result, err)
}

func (h *Handler) UpdateArticle(ctx *gin.Context) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	articleID, ok := pathID(ctx, "articleId")
	if !ok {
		return
	}
	var req contentdto.UpdateArticleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		bad(ctx)
		return
	}
	result, err := h.service.UpdateArticle(ctx.Request.Context(), adminID, role, articleID, req)
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) PublishArticle(ctx *gin.Context) {
	h.articleStatus(ctx, true)
}

func (h *Handler) UnpublishArticle(ctx *gin.Context) {
	h.articleStatus(ctx, false)
}

func (h *Handler) articleStatus(ctx *gin.Context, publish bool) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	articleID, ok := pathID(ctx, "articleId")
	if !ok {
		return
	}
	var result any
	var err *apperrors.BusinessError
	if publish {
		result, err = h.service.PublishArticle(ctx.Request.Context(), adminID, role, articleID)
	} else {
		result, err = h.service.UnpublishArticle(ctx.Request.Context(), adminID, role, articleID)
	}
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) DeleteArticle(ctx *gin.Context) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	articleID, ok := pathID(ctx, "articleId")
	if !ok {
		return
	}
	err := h.service.DeleteArticle(ctx.Request.Context(), adminID, role, articleID)
	write(ctx, http.StatusOK, map[string]bool{"deleted": err == nil}, err)
}

func adminAndRole(ctx *gin.Context) (uint64, string, bool) {
	adminID, err := middleware.CurrentAdminID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return 0, "", false
	}
	role, err := middleware.CurrentAdminRole(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return 0, "", false
	}
	return adminID, role, true
}

func pathID(ctx *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(ctx.Param(name), 10, 64)
	if err != nil || id == 0 {
		bad(ctx)
		return 0, false
	}
	return id, true
}

func listArticlesQuery(ctx *gin.Context, admin bool) contentdto.ListArticlesQuery {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	var featured *bool
	if value := ctx.Query("featured"); value != "" {
		parsed := value == "true" || value == "1"
		featured = &parsed
	}
	query := contentdto.ListArticlesQuery{
		CategoryKey: ctx.Query("categoryKey"),
		Keyword:     ctx.Query("keyword"),
		Featured:    featured,
		Page:        page,
		PageSize:    pageSize,
	}
	if admin {
		query.Status = ctx.Query("status")
	}
	return query
}

func bad(ctx *gin.Context) {
	response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
}

func write(ctx *gin.Context, status int, result any, err *apperrors.BusinessError) {
	if err == nil {
		response.Success(ctx, status, result)
		return
	}
	httpStatus := http.StatusBadRequest
	switch err.Code {
	case contentservice.CodeContentArticleNotFound, contentservice.CodeContentCategoryNotFound:
		httpStatus = http.StatusNotFound
	case contentservice.CodeContentForbidden:
		httpStatus = http.StatusForbidden
	case apperrors.CodeSystemError:
		httpStatus = http.StatusInternalServerError
	}
	response.Error(ctx, httpStatus, err)
}
