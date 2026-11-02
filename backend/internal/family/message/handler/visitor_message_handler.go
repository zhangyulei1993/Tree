package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	"tree/backend/internal/family/message/dto"
	messageservice "tree/backend/internal/family/message/service"
)

type Handler struct{ service messageservice.Service }

func NewHandler(service messageservice.Service) *Handler { return &Handler{service: service} }

func (h *Handler) CreatePublic(ctx *gin.Context) {
	familyID, ok := pathID(ctx, "familyId")
	if !ok {
		return
	}
	var req dto.CreateVisitorMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		bad(ctx)
		return
	}
	result, err := h.service.CreatePublic(ctx.Request.Context(), familyID, req, audit(ctx))
	write(ctx, http.StatusCreated, result, err)
}

func (h *Handler) ListPublic(ctx *gin.Context) {
	familyID, ok := pathID(ctx, "familyId")
	if !ok {
		return
	}
	result, err := h.service.ListPublic(ctx.Request.Context(), familyID, listQuery(ctx))
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
	messageID, ok := pathID(ctx, "messageId")
	if !ok {
		return
	}
	var req dto.ReviewVisitorMessageRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	if approve {
		result, err := h.service.Approve(ctx.Request.Context(), adminID, role, messageID, req, audit(ctx))
		write(ctx, http.StatusOK, result, err)
		return
	}
	result, err := h.service.Reject(ctx.Request.Context(), adminID, role, messageID, req, audit(ctx))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) Delete(ctx *gin.Context) {
	adminID, role, ok := adminAndRole(ctx)
	if !ok {
		return
	}
	messageID, ok := pathID(ctx, "messageId")
	if !ok {
		return
	}
	var req dto.DeleteVisitorMessageRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	err := h.service.Delete(ctx.Request.Context(), adminID, role, messageID, req, audit(ctx))
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

func listQuery(ctx *gin.Context) dto.ListMessagesQuery {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	familyID, _ := strconv.ParseUint(ctx.Query("familyId"), 10, 64)
	return dto.ListMessagesQuery{Status: ctx.Query("status"), FamilyID: familyID, Page: page, PageSize: pageSize}
}

func audit(ctx *gin.Context) messageservice.AuditInput {
	return messageservice.AuditInput{IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent()}
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
	case messageservice.CodeVisitorMessageForbidden:
		httpStatus = http.StatusForbidden
	case messageservice.CodeVisitorMessageNotFound:
		httpStatus = http.StatusNotFound
	case messageservice.CodeVisitorMessageRateLimited:
		httpStatus = http.StatusTooManyRequests
	case messageservice.CodeVisitorMessageFamilyPrivate:
		httpStatus = http.StatusForbidden
	case apperrors.CodeSystemError:
		httpStatus = http.StatusInternalServerError
	}
	response.Error(ctx, httpStatus, err)
}
