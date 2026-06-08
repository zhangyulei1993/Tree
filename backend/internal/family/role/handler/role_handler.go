package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	roledto "tree/backend/internal/family/role/dto"
	roleservice "tree/backend/internal/family/role/service"
)

type Handler struct{ service roleservice.Service }

func NewHandler(service roleservice.Service) *Handler { return &Handler{service: service} }

func (h *Handler) SetAdmin(ctx *gin.Context) {
	h.change(ctx, true)
}

func (h *Handler) UnsetAdmin(ctx *gin.Context) {
	h.change(ctx, false)
}

func (h *Handler) change(ctx *gin.Context, set bool) {
	userID, err := middleware.CurrentUserID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	familyID, ok := pathID(ctx, "familyId")
	if !ok {
		return
	}
	memberID, ok := pathID(ctx, "memberId")
	if !ok {
		return
	}
	var req roledto.RoleChangeRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	if set {
		result, businessErr := h.service.SetAdmin(ctx.Request.Context(), userID, familyID, memberID, req, audit(ctx))
		write(ctx, result, businessErr)
		return
	}
	result, businessErr := h.service.UnsetAdmin(ctx.Request.Context(), userID, familyID, memberID, req, audit(ctx))
	write(ctx, result, businessErr)
}

func pathID(ctx *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(ctx.Param(name), 10, 64)
	if err != nil || id == 0 {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return 0, false
	}
	return id, true
}

func audit(ctx *gin.Context) roleservice.AuditInput {
	return roleservice.AuditInput{IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent()}
}

func write(ctx *gin.Context, result any, err *apperrors.BusinessError) {
	if err == nil {
		response.Success(ctx, http.StatusOK, result)
		return
	}
	status := http.StatusBadRequest
	if err.Code == roleservice.CodeRoleForbidden {
		status = http.StatusForbidden
	}
	if err.Code == apperrors.CodeSystemError {
		status = http.StatusInternalServerError
	}
	response.Error(ctx, status, err)
}
