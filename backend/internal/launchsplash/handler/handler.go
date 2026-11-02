package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	launchdto "tree/backend/internal/launchsplash/dto"
	launchservice "tree/backend/internal/launchsplash/service"
)

type Handler struct{ service launchservice.Service }

func New(service launchservice.Service) *Handler { return &Handler{service: service} }
func (h *Handler) Public(ctx *gin.Context) {
	result, err := h.service.Public(ctx.Request.Context())
	write(ctx, result, err)
}
func (h *Handler) AdminGet(ctx *gin.Context) {
	if !canManage(ctx) {
		return
	}
	result, err := h.service.AdminGet(ctx.Request.Context())
	write(ctx, result, err)
}

func canManage(ctx *gin.Context) bool {
	role, err := middleware.CurrentAdminRole(ctx)
	if err != nil || (role != "ROOT_ADMIN" && role != "SUPER_ADMIN") {
		response.Abort(ctx, http.StatusForbidden, apperrors.CodeForbidden)
		return false
	}
	return true
}
func (h *Handler) AdminUpdate(ctx *gin.Context) {
	adminID, err := middleware.CurrentAdminID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	role, err := middleware.CurrentAdminRole(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	var req launchdto.UpdateRequest
	if ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.AdminUpdate(ctx.Request.Context(), adminID, role, req)
	write(ctx, result, businessErr)
}
func write(ctx *gin.Context, result any, err *apperrors.BusinessError) {
	if err == nil {
		response.OK(ctx, result)
		return
	}
	status := http.StatusBadRequest
	if err.Code == launchservice.CodeNotFound {
		status = http.StatusNotFound
	}
	if err.Code == apperrors.CodeForbidden {
		status = http.StatusForbidden
	}
	if err.Code == apperrors.CodeSystemError {
		status = http.StatusInternalServerError
	}
	response.Error(ctx, status, err)
}
