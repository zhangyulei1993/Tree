package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	tooldto "tree/backend/internal/toolconfig/dto"
	toolservice "tree/backend/internal/toolconfig/service"
)

type Handler struct{ service toolservice.Service }

func New(service toolservice.Service) *Handler { return &Handler{service: service} }

func (h *Handler) PublicList(ctx *gin.Context) {
	result, err := h.service.List(ctx.Request.Context())
	write(ctx, result, err)
}

func (h *Handler) AdminList(ctx *gin.Context) {
	result, err := h.service.List(ctx.Request.Context())
	write(ctx, result, err)
}

func (h *Handler) AdminCreate(ctx *gin.Context) {
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
	var req tooldto.CreateRequest
	if ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.Create(ctx.Request.Context(), adminID, role, req.ToolKey, req.DisplayName, req.Description)
	if businessErr == nil {
		response.Created(ctx, result)
		return
	}
	write(ctx, result, businessErr)
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
	var req tooldto.UpdateRequest
	if ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.Update(ctx.Request.Context(), adminID, role, strings.TrimSpace(ctx.Param("toolKey")), req)
	write(ctx, result, businessErr)
}

func write(ctx *gin.Context, result any, businessErr *apperrors.BusinessError) {
	if businessErr == nil {
		response.OK(ctx, result)
		return
	}
	status := http.StatusBadRequest
	switch businessErr.Code {
	case apperrors.CodeForbidden:
		status = http.StatusForbidden
	case toolservice.CodeNotFound:
		status = http.StatusNotFound
	case toolservice.CodeDuplicate:
		status = http.StatusConflict
	case apperrors.CodeSystemError:
		status = http.StatusInternalServerError
	}
	response.Error(ctx, status, businessErr)
}
