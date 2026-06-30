package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	"tree/backend/internal/family/invitation/dto"
	inviteservice "tree/backend/internal/family/invitation/service"
)

type Handler struct{ service inviteservice.Service }

func NewHandler(service inviteservice.Service) *Handler { return &Handler{service: service} }

func (h *Handler) Create(ctx *gin.Context) {
	actor, ok := userID(ctx)
	if !ok {
		return
	}
	family, ok := pathID(ctx, "familyId")
	if !ok {
		return
	}
	member, ok := pathID(ctx, "memberId")
	if !ok {
		return
	}
	var req dto.CreateInvitationRequest
	if ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, err := h.service.Create(ctx, actor, family, member, req, audit(ctx))
	write(ctx, http.StatusCreated, result, err)
}

func (h *Handler) Detail(ctx *gin.Context) {
	result, err := h.service.Detail(ctx, ctx.Param("inviteToken"))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) Accept(ctx *gin.Context) {
	actor, invitation, ok := actorAndInvitation(ctx)
	if !ok {
		return
	}
	result, err := h.service.Accept(ctx, actor, invitation, audit(ctx))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) Reject(ctx *gin.Context) {
	actor, invitation, ok := actorAndInvitation(ctx)
	if !ok {
		return
	}
	var req dto.RejectInvitationRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, err := h.service.Reject(ctx, actor, invitation, req, audit(ctx))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) Cancel(ctx *gin.Context) {
	actor, invitation, ok := actorAndInvitation(ctx)
	if !ok {
		return
	}
	var req dto.CancelInvitationRequest
	if ctx.Request.ContentLength != 0 && ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, err := h.service.Cancel(ctx, actor, invitation, req, audit(ctx))
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) ListMine(ctx *gin.Context) {
	actor, ok := userID(ctx)
	if !ok {
		return
	}
	result, err := h.service.ListMine(ctx, actor)
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) ListFamily(ctx *gin.Context) {
	actor, ok := userID(ctx)
	if !ok {
		return
	}
	family, ok := pathID(ctx, "familyId")
	if !ok {
		return
	}
	result, err := h.service.ListFamily(ctx, actor, family)
	write(ctx, http.StatusOK, result, err)
}

func (h *Handler) Regenerate(ctx *gin.Context) {
	actor, invitation, ok := actorAndInvitation(ctx)
	if !ok {
		return
	}
	result, err := h.service.Regenerate(ctx, actor, invitation, audit(ctx))
	write(ctx, http.StatusCreated, result, err)
}

func actorAndInvitation(ctx *gin.Context) (uint64, uint64, bool) {
	actor, ok := userID(ctx)
	if !ok {
		return 0, 0, false
	}
	id, ok := pathID(ctx, "invitationId")
	return actor, id, ok
}

func userID(ctx *gin.Context) (uint64, bool) {
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
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return 0, false
	}
	return id, true
}

func audit(ctx *gin.Context) inviteservice.AuditInput {
	return inviteservice.AuditInput{IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent()}
}

func write(ctx *gin.Context, status int, result any, err *apperrors.BusinessError) {
	if err == nil {
		response.Success(ctx, status, result)
		return
	}
	httpStatus := http.StatusBadRequest
	switch err.Code {
	case inviteservice.CodeInvitationForbidden, inviteservice.CodeInvitationTargetMismatch:
		httpStatus = http.StatusForbidden
	case inviteservice.CodeInvitationNotFound:
		httpStatus = http.StatusNotFound
	case apperrors.CodeSystemError:
		httpStatus = http.StatusInternalServerError
	}
	response.Error(ctx, httpStatus, err)
}
