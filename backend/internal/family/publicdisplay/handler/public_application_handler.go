package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	"tree/backend/internal/family/publicdisplay/dto"
	publicservice "tree/backend/internal/family/publicdisplay/service"
)

type Handler struct{ service publicservice.Service }

func NewHandler(service publicservice.Service) *Handler { return &Handler{service: service} }

func (h *Handler) Submit(ctx *gin.Context) {
	userID, familyID, ok := userAndFamily(ctx)
	if !ok {
		return
	}
	var req dto.CreatePublicApplicationRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	result, err := h.service.Submit(ctx.Request.Context(), userID, familyID, req, audit(ctx))
	write(ctx, http.StatusCreated, result, err)
}

func (h *Handler) ListFamily(ctx *gin.Context) {
	userID, familyID, ok := userAndFamily(ctx)
	if !ok {
		return
	}
	result, err := h.service.ListFamily(ctx.Request.Context(), userID, familyID, listQuery(ctx))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) Cancel(ctx *gin.Context) {
	userID, familyID, ok := userAndFamily(ctx)
	if !ok {
		return
	}
	applicationID, ok := pathID(ctx, "applicationId")
	if !ok {
		return
	}
	var req dto.CancelPublicApplicationRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	result, err := h.service.Cancel(ctx.Request.Context(), userID, familyID, applicationID, req, audit(ctx))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) ListAdmin(ctx *gin.Context) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	result, err := h.service.ListAdmin(ctx.Request.Context(), adminID, role, listQuery(ctx))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) Approve(ctx *gin.Context) {
	h.review(ctx, true)
}

func (h *Handler) Reject(ctx *gin.Context) {
	h.review(ctx, false)
}

func (h *Handler) review(ctx *gin.Context, approve bool) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	applicationID, ok := pathID(ctx, "applicationId")
	if !ok {
		return
	}
	var req dto.ReviewPublicApplicationRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	if approve {
		result, err := h.service.Approve(ctx.Request.Context(), adminID, role, applicationID, req, audit(ctx))
		write(ctx, http.StatusOK, result, err)
		return
	}
	result, err := h.service.Reject(ctx.Request.Context(), adminID, role, applicationID, req, audit(ctx))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) TakeDown(ctx *gin.Context) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	familyID, ok := pathID(ctx, "familyId")
	if !ok {
		return
	}
	var req dto.TakeDownPublicFamilyRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	result, err := h.service.TakeDown(ctx.Request.Context(), adminID, role, familyID, req, audit(ctx))
	write(ctx, http.StatusOK, result, err)
}

func userAndFamily(ctx *gin.Context) (uint64, uint64, bool) {
	userID, err := middleware.CurrentUserID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return 0, 0, false
	}
	familyID, ok := pathID(ctx, "familyId")
	return userID, familyID, ok
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

func listQuery(ctx *gin.Context) dto.ListApplicationsQuery {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	return dto.ListApplicationsQuery{Status: ctx.Query("status"), Page: page, PageSize: pageSize}
}

func audit(ctx *gin.Context) publicservice.AuditInput {
	return publicservice.AuditInput{IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent()}
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
	case publicservice.CodePublicApplicationForbidden:
		httpStatus = http.StatusForbidden
	case publicservice.CodePublicApplicationNotFound:
		httpStatus = http.StatusNotFound
	case apperrors.CodeSystemError:
		httpStatus = http.StatusInternalServerError
	}
	response.Error(ctx, httpStatus, err)
}
