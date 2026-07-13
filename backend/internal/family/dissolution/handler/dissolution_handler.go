package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	dissolutiondto "tree/backend/internal/family/dissolution/dto"
	dissolutionservice "tree/backend/internal/family/dissolution/service"
)

type Handler struct{ service dissolutionservice.Service }

func NewHandler(service dissolutionservice.Service) *Handler { return &Handler{service: service} }

func (h *Handler) ListAdmin(ctx *gin.Context) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	result, err := h.service.ListAdmin(ctx.Request.Context(), adminID, role, listQuery(ctx))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) Approve(ctx *gin.Context) { h.review(ctx, true) }
func (h *Handler) Reject(ctx *gin.Context)  { h.review(ctx, false) }

func (h *Handler) review(ctx *gin.Context, approve bool) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	requestID, ok := pathID(ctx, "requestId")
	if !ok {
		return
	}
	var req dissolutiondto.ReviewDissolutionRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	if approve {
		result, err := h.service.Approve(ctx.Request.Context(), adminID, role, requestID, req, audit(ctx))
		write(ctx, http.StatusOK, result, err)
		return
	}
	result, err := h.service.Reject(ctx.Request.Context(), adminID, role, requestID, req, audit(ctx))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) Restore(ctx *gin.Context) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	familyID, ok := pathID(ctx, "familyId")
	if !ok {
		return
	}
	var req dissolutiondto.RestoreFamilyRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	result, err := h.service.Restore(ctx.Request.Context(), adminID, role, familyID, req, audit(ctx))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) Finalize(ctx *gin.Context) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	familyID, ok := pathID(ctx, "familyId")
	if !ok {
		return
	}
	result, err := h.service.Finalize(ctx.Request.Context(), adminID, role, familyID, audit(ctx))
	write(ctx, http.StatusOK, result, err)
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

func listQuery(ctx *gin.Context) dissolutiondto.ListDissolutionQuery {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	familyID, _ := strconv.ParseUint(ctx.Query("familyId"), 10, 64)
	return dissolutiondto.ListDissolutionQuery{Status: ctx.Query("status"), FamilyID: familyID, Page: page, PageSize: pageSize}
}

func audit(ctx *gin.Context) dissolutionservice.AuditInput {
	return dissolutionservice.AuditInput{IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent()}
}

func bad(ctx *gin.Context) { response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams) }

func write(ctx *gin.Context, status int, result any, err *apperrors.BusinessError) {
	if err == nil {
		response.Success(ctx, status, result)
		return
	}
	httpStatus := http.StatusBadRequest
	switch err.Code {
	case dissolutionservice.CodeDissolutionAdminDenied, dissolutionservice.CodeFamilyRestoreAdminDenied, dissolutionservice.CodeFamilyFinalizeAdminDenied:
		httpStatus = http.StatusForbidden
	case dissolutionservice.CodeDissolutionNotFound,
		apperrors.CodeResourceNotFound:
		httpStatus = http.StatusNotFound
	case apperrors.CodeSystemError:
		httpStatus = http.StatusInternalServerError
	}
	response.Error(ctx, httpStatus, err)
}
