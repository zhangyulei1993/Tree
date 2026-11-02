package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	transferdto "tree/backend/internal/family/transfer/dto"
	transferservice "tree/backend/internal/family/transfer/service"
)

type Handler struct{ service transferservice.Service }

func NewHandler(service transferservice.Service) *Handler { return &Handler{service: service} }

func (h *Handler) Create(ctx *gin.Context) {
	userID, familyID, ok := userAndFamily(ctx)
	if !ok {
		return
	}
	var req transferdto.CreateTransferRequest
	if ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	result, err := h.service.Create(ctx.Request.Context(), userID, familyID, req, audit(ctx))
	write(ctx, http.StatusCreated, result, err)
}

func (h *Handler) Current(ctx *gin.Context) {
	userID, familyID, ok := userAndFamily(ctx)
	if !ok {
		return
	}
	result, err := h.service.Current(ctx.Request.Context(), userID, familyID)
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) Cancel(ctx *gin.Context) {
	userID, familyID, ok := userAndFamily(ctx)
	if !ok {
		return
	}
	requestID, ok := pathID(ctx, "requestId")
	if !ok {
		return
	}
	var req transferdto.CancelTransferRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	result, err := h.service.Cancel(ctx.Request.Context(), userID, familyID, requestID, req, audit(ctx))
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
	var req transferdto.ReviewTransferRequest
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

func listQuery(ctx *gin.Context) transferdto.ListTransferQuery {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	familyID, _ := strconv.ParseUint(ctx.Query("familyId"), 10, 64)
	return transferdto.ListTransferQuery{Status: ctx.Query("status"), FamilyID: familyID, Page: page, PageSize: pageSize}
}

func audit(ctx *gin.Context) transferservice.AuditInput {
	return transferservice.AuditInput{IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent()}
}

func bad(ctx *gin.Context) { response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams) }

func write(ctx *gin.Context, status int, result any, err *apperrors.BusinessError) {
	if err == nil {
		response.Success(ctx, status, result)
		return
	}
	httpStatus := apperrors.StatusOr(err.Code, http.StatusBadRequest)
	if httpStatus == http.StatusBadRequest {
		switch err.Code {
		case transferservice.CodeTransferForbidden, transferservice.CodeTransferAdminDenied:
			httpStatus = http.StatusForbidden
		case transferservice.CodeTransferNotFound:
			httpStatus = http.StatusNotFound
		case apperrors.CodeSystemError:
			httpStatus = http.StatusInternalServerError
		}
	}
	response.Error(ctx, httpStatus, err)
}
