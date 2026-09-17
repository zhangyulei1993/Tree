package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	sharedto "tree/backend/internal/shareconfig/dto"
	shareservice "tree/backend/internal/shareconfig/service"
)

type Handler struct{ service shareservice.Service }

func New(service shareservice.Service) *Handler { return &Handler{service: service} }

func (h *Handler) PublicList(ctx *gin.Context) {
	result, err := h.service.List(ctx.Request.Context())
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) AdminList(ctx *gin.Context) {
	result, err := h.service.List(ctx.Request.Context())
	write(ctx, http.StatusOK, result, err)
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
	var req sharedto.UpdateRequest
	if ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.Update(ctx.Request.Context(), adminID, role, strings.TrimSpace(ctx.Param("configKey")), req)
	write(ctx, http.StatusOK, result, businessErr)
}

func write(ctx *gin.Context, status int, result any, err *apperrors.BusinessError) {
	if err == nil {
		response.Success(ctx, status, result)
		return
	}
	httpStatus := http.StatusBadRequest
	if err.Code == shareservice.CodeNotFound {
		httpStatus = http.StatusNotFound
	}
	if err.Code == apperrors.CodeSystemError {
		httpStatus = http.StatusInternalServerError
	}
	response.Error(ctx, httpStatus, err)
}
