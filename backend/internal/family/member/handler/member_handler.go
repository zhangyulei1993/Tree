package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	"tree/backend/internal/family/member/dto"
	memberservice "tree/backend/internal/family/member/service"
)

type MemberHandler struct {
	service memberservice.MemberService
}

func NewMemberHandler(service memberservice.MemberService) *MemberHandler {
	return &MemberHandler{service: service}
}

func (h *MemberHandler) Create(ctx *gin.Context) {
	actorID, familyID, ok := actorAndFamily(ctx)
	if !ok {
		return
	}
	var req dto.CreateMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.Create(ctx.Request.Context(), actorID, familyID, req, audit(ctx))
	writeResult(ctx, http.StatusCreated, result, businessErr)
}

func (h *MemberHandler) List(ctx *gin.Context) {
	actorID, familyID, ok := actorAndFamily(ctx)
	if !ok {
		return
	}
	result, businessErr := h.service.List(ctx.Request.Context(), actorID, familyID)
	writeResult(ctx, http.StatusOK, result, businessErr)
}

func (h *MemberHandler) Detail(ctx *gin.Context) {
	actorID, familyID, memberID, ok := actorFamilyMember(ctx)
	if !ok {
		return
	}
	result, businessErr := h.service.Detail(ctx.Request.Context(), actorID, familyID, memberID)
	writeResult(ctx, http.StatusOK, result, businessErr)
}

func (h *MemberHandler) Update(ctx *gin.Context) {
	actorID, familyID, memberID, ok := actorFamilyMember(ctx)
	if !ok {
		return
	}
	var req dto.UpdateMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.Update(ctx.Request.Context(), actorID, familyID, memberID, req, audit(ctx))
	writeResult(ctx, http.StatusOK, result, businessErr)
}

func (h *MemberHandler) Delete(ctx *gin.Context) {
	actorID, familyID, memberID, ok := actorFamilyMember(ctx)
	if !ok {
		return
	}
	var req dto.DeleteMemberRequest
	if ctx.Request.ContentLength != 0 {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
	}
	if businessErr := h.service.Delete(ctx.Request.Context(), actorID, familyID, memberID, req, audit(ctx)); businessErr != nil {
		writeError(ctx, businessErr)
		return
	}
	response.OK(ctx, gin.H{"status": "ok"})
}

func (h *MemberHandler) BindUser(ctx *gin.Context) {
	actorID, familyID, memberID, ok := actorFamilyMember(ctx)
	if !ok {
		return
	}
	var req dto.BindUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.BindUser(ctx.Request.Context(), actorID, familyID, memberID, req, audit(ctx))
	writeResult(ctx, http.StatusOK, result, businessErr)
}

func (h *MemberHandler) UnbindUser(ctx *gin.Context) {
	actorID, familyID, memberID, ok := actorFamilyMember(ctx)
	if !ok {
		return
	}
	var req dto.UnbindUserRequest
	if ctx.Request.ContentLength != 0 {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
	}
	result, businessErr := h.service.UnbindUser(ctx.Request.Context(), actorID, familyID, memberID, req, audit(ctx))
	writeResult(ctx, http.StatusOK, result, businessErr)
}

func actorAndFamily(ctx *gin.Context) (uint64, uint64, bool) {
	actorID, err := middleware.CurrentUserID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return 0, 0, false
	}
	familyID, err := positiveID(ctx.Param("familyId"))
	if err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return 0, 0, false
	}
	return actorID, familyID, true
}

func actorFamilyMember(ctx *gin.Context) (uint64, uint64, uint64, bool) {
	actorID, familyID, ok := actorAndFamily(ctx)
	if !ok {
		return 0, 0, 0, false
	}
	memberID, err := positiveID(ctx.Param("memberId"))
	if err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return 0, 0, 0, false
	}
	return actorID, familyID, memberID, true
}

func positiveID(value string) (uint64, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}

func audit(ctx *gin.Context) memberservice.AuditInput {
	return memberservice.AuditInput{IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent()}
}

func writeResult[T any](ctx *gin.Context, status int, result T, businessErr *apperrors.BusinessError) {
	if businessErr != nil {
		writeError(ctx, businessErr)
		return
	}
	response.Success(ctx, status, result)
}

func writeError(ctx *gin.Context, businessErr *apperrors.BusinessError) {
	status := apperrors.StatusOr(businessErr.Code, http.StatusBadRequest)
	if status == http.StatusBadRequest {
		switch businessErr.Code {
		case memberservice.CodeMemberCreateForbidden,
			memberservice.CodeMemberEditForbidden,
			memberservice.CodeMemberViewForbidden,
			memberservice.CodeMemberDeleteForbidden,
			memberservice.CodeMemberUnbindProtected:
			status = http.StatusForbidden
		case memberservice.CodeMemberNotFound,
			memberservice.CodeMemberUserUnavailable,
			apperrors.CodeResourceNotFound:
			status = http.StatusNotFound
		case apperrors.CodeSystemError:
			status = http.StatusInternalServerError
		}
	}
	response.Error(ctx, status, businessErr)
}
