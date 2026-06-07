package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	"tree/backend/internal/family/joinrequest/dto"
	joinservice "tree/backend/internal/family/joinrequest/service"
)

type Handler struct{ service joinservice.Service }

func NewHandler(service joinservice.Service) *Handler { return &Handler{service: service} }

func (h *Handler) Create(ctx *gin.Context) {
	actor, family, ok := actorFamily(ctx)
	if !ok {
		return
	}
	var req dto.CreateJoinRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	result, err := h.service.Create(ctx, actor, family, req, audit(ctx))
	write(ctx, http.StatusCreated, result, err)
}
func (h *Handler) ListMine(ctx *gin.Context) {
	actor, ok := actorID(ctx)
	if !ok {
		return
	}
	result, err := h.service.ListMine(ctx, actor)
	write(ctx, http.StatusOK, result, err)
}
func (h *Handler) ListFamily(ctx *gin.Context) {
	actor, family, ok := actorFamily(ctx)
	if !ok {
		return
	}
	result, err := h.service.ListFamily(ctx, actor, family)
	write(ctx, http.StatusOK, result, err)
}
func (h *Handler) Approve(ctx *gin.Context) {
	actor, family, request, ok := actorFamilyRequest(ctx)
	if !ok {
		return
	}
	var req dto.ApproveJoinRequest
	if ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	result, err := h.service.Approve(ctx, actor, family, request, req, audit(ctx))
	write(ctx, http.StatusOK, result, err)
}
func (h *Handler) Reject(ctx *gin.Context) {
	actor, family, request, ok := actorFamilyRequest(ctx)
	if !ok {
		return
	}
	var req dto.RejectJoinRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	result, err := h.service.Reject(ctx, actor, family, request, req, audit(ctx))
	write(ctx, http.StatusOK, result, err)
}
func (h *Handler) Cancel(ctx *gin.Context) {
	actor, family, request, ok := actorFamilyRequest(ctx)
	if !ok {
		return
	}
	var req dto.CancelJoinRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		bad(ctx)
		return
	}
	result, err := h.service.Cancel(ctx, actor, family, request, req, audit(ctx))
	write(ctx, http.StatusOK, result, err)
}
func actorFamily(ctx *gin.Context) (uint64, uint64, bool) {
	actor, ok := actorID(ctx)
	if !ok {
		return 0, 0, false
	}
	family, ok := pathID(ctx, "familyId")
	return actor, family, ok
}
func actorFamilyRequest(ctx *gin.Context) (uint64, uint64, uint64, bool) {
	actor, family, ok := actorFamily(ctx)
	if !ok {
		return 0, 0, 0, false
	}
	request, ok := pathID(ctx, "requestId")
	return actor, family, request, ok
}
func actorID(ctx *gin.Context) (uint64, bool) {
	id, err := middleware.CurrentUserID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return 0, false
	}
	return id, true
}
func pathID(ctx *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(ctx.Param(name), 10, 64)
	if err != nil || id == 0 {
		bad(ctx)
		return 0, false
	}
	return id, true
}
func bad(ctx *gin.Context) { response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams) }
func audit(ctx *gin.Context) joinservice.AuditInput {
	return joinservice.AuditInput{IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent()}
}
func write(ctx *gin.Context, status int, result any, err *apperrors.BusinessError) {
	if err == nil {
		response.Success(ctx, status, result)
		return
	}
	httpStatus := http.StatusBadRequest
	switch err.Code {
	case joinservice.CodeJoinRequestForbidden:
		httpStatus = http.StatusForbidden
	case joinservice.CodeJoinRequestNotFound:
		httpStatus = http.StatusNotFound
	case apperrors.CodeSystemError:
		httpStatus = http.StatusInternalServerError
	}
	response.Error(ctx, httpStatus, err)
}
