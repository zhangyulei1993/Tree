package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	"tree/backend/internal/notification/service"
)

type Handler struct{ service service.Service }

func New(s service.Service) *Handler { return &Handler{service: s} }

func (h *Handler) List(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	result, businessErr := h.service.List(c, userID, page, pageSize)
	if businessErr != nil {
		response.Error(c, http.StatusBadRequest, businessErr)
		return
	}
	response.OK(c, result)
}
func (h *Handler) MarkRead(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("notificationId"), 10, 64)
	if err != nil || id == 0 {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	if businessErr := h.service.MarkRead(c, userID, id); businessErr != nil {
		response.Error(c, http.StatusBadRequest, businessErr)
		return
	}
	response.OK(c, gin.H{"read": true})
}

func (h *Handler) MarkAllRead(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	if businessErr := h.service.MarkAllRead(c, userID); businessErr != nil {
		response.Error(c, http.StatusBadRequest, businessErr)
		return
	}
	response.OK(c, gin.H{"read": true})
}
