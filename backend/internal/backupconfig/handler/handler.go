package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	backupdto "tree/backend/internal/backupconfig/dto"
	backupservice "tree/backend/internal/backupconfig/service"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
)

type Handler struct{ service backupservice.Service }

func New(service backupservice.Service) *Handler { return &Handler{service: service} }
func (h *Handler) Overview(ctx *gin.Context) {
	role, err := middleware.CurrentAdminRole(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	value, businessErr := h.service.Overview(ctx.Request.Context(), role)
	write(ctx, value, businessErr)
}
func (h *Handler) Update(ctx *gin.Context) {
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
	var req backupdto.UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	value, businessErr := h.service.Update(ctx.Request.Context(), adminID, role, req)
	write(ctx, value, businessErr)
}
func write(ctx *gin.Context, value any, err *apperrors.BusinessError) {
	if err != nil {
		status := http.StatusBadRequest
		if err.Code == apperrors.CodeForbidden {
			status = http.StatusForbidden
		}
		if err.Code == apperrors.CodeSystemError {
			status = http.StatusInternalServerError
		}
		response.Error(ctx, status, err)
		return
	}
	response.OK(ctx, value)
}
