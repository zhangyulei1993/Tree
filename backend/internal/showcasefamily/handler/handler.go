package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	showcasedto "tree/backend/internal/showcasefamily/dto"
	showcaseservice "tree/backend/internal/showcasefamily/service"
)

type Handler struct{ service showcaseservice.Service }

func New(service showcaseservice.Service) *Handler { return &Handler{service: service} }

func (h *Handler) PublicList(ctx *gin.Context) {
	result, businessErr := h.service.ListPublished(ctx.Request.Context())
	write(ctx, result, businessErr)
}

func (h *Handler) PublicTree(ctx *gin.Context) {
	id, ok := idParam(ctx)
	if !ok {
		return
	}
	result, businessErr := h.service.PublicTree(ctx.Request.Context(), id)
	write(ctx, result, businessErr)
}

func (h *Handler) AdminList(ctx *gin.Context) {
	result, businessErr := h.service.List(ctx.Request.Context())
	write(ctx, result, businessErr)
}

func (h *Handler) AdminGetDisplaySettings(ctx *gin.Context) {
	result, businessErr := h.service.GetDisplaySettings(ctx.Request.Context())
	write(ctx, result, businessErr)
}

func (h *Handler) AdminUpdateDisplaySettings(ctx *gin.Context) {
	adminID, role, ok := currentAdmin(ctx)
	if !ok {
		return
	}
	var req showcasedto.DisplaySettingsRequest
	if ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.UpdateDisplaySettings(ctx.Request.Context(), adminID, role, req)
	write(ctx, result, businessErr)
}

func (h *Handler) AdminTree(ctx *gin.Context) {
	id, ok := idParam(ctx)
	if !ok {
		return
	}
	result, businessErr := h.service.Tree(ctx.Request.Context(), id)
	write(ctx, result, businessErr)
}

func (h *Handler) AdminCreate(ctx *gin.Context) {
	adminID, role, ok := currentAdmin(ctx)
	if !ok {
		return
	}
	var req showcasedto.CreateRequest
	if ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.Create(ctx.Request.Context(), adminID, role, req)
	if businessErr == nil {
		response.Created(ctx, result)
		return
	}
	write(ctx, result, businessErr)
}

func (h *Handler) AdminUpdate(ctx *gin.Context) {
	adminID, role, ok := currentAdmin(ctx)
	if !ok {
		return
	}
	id, ok := idParam(ctx)
	if !ok {
		return
	}
	var req showcasedto.UpdateRequest
	if ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.Update(ctx.Request.Context(), adminID, role, id, req)
	write(ctx, result, businessErr)
}

func (h *Handler) AdminPublish(ctx *gin.Context) {
	h.changePublication(ctx, true)
}

func (h *Handler) AdminUnpublish(ctx *gin.Context) {
	h.changePublication(ctx, false)
}

func (h *Handler) AdminDelete(ctx *gin.Context) {
	adminID, role, ok := currentAdmin(ctx)
	if !ok {
		return
	}
	id, ok := idParam(ctx)
	if !ok {
		return
	}
	businessErr := h.service.Delete(ctx.Request.Context(), adminID, role, id)
	if businessErr == nil {
		ctx.Status(http.StatusNoContent)
		return
	}
	write(ctx, nil, businessErr)
}

func (h *Handler) AdminCreateMember(ctx *gin.Context) {
	adminID, role, ok := currentAdmin(ctx)
	if !ok {
		return
	}
	showcaseFamilyID, ok := idParam(ctx)
	if !ok {
		return
	}
	var req showcasedto.MemberRequest
	if ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.CreateMember(ctx.Request.Context(), adminID, role, showcaseFamilyID, req)
	if businessErr == nil {
		response.Created(ctx, result)
		return
	}
	write(ctx, result, businessErr)
}

func (h *Handler) AdminUpdateMember(ctx *gin.Context) {
	adminID, role, ok := currentAdmin(ctx)
	if !ok {
		return
	}
	showcaseFamilyID, ok := idParam(ctx)
	if !ok {
		return
	}
	memberID, ok := nestedIDParam(ctx, "memberId")
	if !ok {
		return
	}
	var req showcasedto.MemberRequest
	if ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.UpdateMember(ctx.Request.Context(), adminID, role, showcaseFamilyID, memberID, req)
	write(ctx, result, businessErr)
}

func (h *Handler) AdminDeleteMember(ctx *gin.Context) {
	adminID, role, ok := currentAdmin(ctx)
	if !ok {
		return
	}
	showcaseFamilyID, ok := idParam(ctx)
	if !ok {
		return
	}
	memberID, ok := nestedIDParam(ctx, "memberId")
	if !ok {
		return
	}
	businessErr := h.service.DeleteMember(ctx.Request.Context(), adminID, role, showcaseFamilyID, memberID)
	if businessErr == nil {
		ctx.Status(http.StatusNoContent)
		return
	}
	write(ctx, nil, businessErr)
}

func (h *Handler) AdminCreateRelationship(ctx *gin.Context) {
	adminID, role, ok := currentAdmin(ctx)
	if !ok {
		return
	}
	showcaseFamilyID, ok := idParam(ctx)
	if !ok {
		return
	}
	var req showcasedto.RelationshipRequest
	if ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.CreateRelationship(ctx.Request.Context(), adminID, role, showcaseFamilyID, req)
	if businessErr == nil {
		response.Created(ctx, result)
		return
	}
	write(ctx, result, businessErr)
}

func (h *Handler) AdminAddRelative(ctx *gin.Context) {
	adminID, role, ok := currentAdmin(ctx)
	if !ok {
		return
	}
	showcaseFamilyID, ok := idParam(ctx)
	if !ok {
		return
	}
	anchorMemberID, ok := nestedIDParam(ctx, "memberId")
	if !ok {
		return
	}
	var req showcasedto.AddRelativeRequest
	if ctx.ShouldBindJSON(&req) != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.AddRelative(ctx.Request.Context(), adminID, role, showcaseFamilyID, anchorMemberID, req)
	if businessErr == nil {
		response.Created(ctx, result)
		return
	}
	write(ctx, result, businessErr)
}

func (h *Handler) AdminDeleteRelationship(ctx *gin.Context) {
	adminID, role, ok := currentAdmin(ctx)
	if !ok {
		return
	}
	showcaseFamilyID, ok := idParam(ctx)
	if !ok {
		return
	}
	relationshipID, ok := nestedIDParam(ctx, "relationshipId")
	if !ok {
		return
	}
	businessErr := h.service.DeleteRelationship(ctx.Request.Context(), adminID, role, showcaseFamilyID, relationshipID)
	if businessErr == nil {
		ctx.Status(http.StatusNoContent)
		return
	}
	write(ctx, nil, businessErr)
}

func (h *Handler) changePublication(ctx *gin.Context, publish bool) {
	adminID, role, ok := currentAdmin(ctx)
	if !ok {
		return
	}
	id, ok := idParam(ctx)
	if !ok {
		return
	}
	result, businessErr := h.service.Publish(ctx.Request.Context(), adminID, role, id, publish)
	write(ctx, result, businessErr)
}

func currentAdmin(ctx *gin.Context) (uint64, string, bool) {
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

func idParam(ctx *gin.Context) (uint64, bool) {
	return nestedIDParam(ctx, "showcaseFamilyId")
}

func nestedIDParam(ctx *gin.Context, key string) (uint64, bool) {
	id, err := strconv.ParseUint(ctx.Param(key), 10, 64)
	if err != nil || id == 0 {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return 0, false
	}
	return id, true
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
	case showcaseservice.CodeNotFound:
		status = http.StatusNotFound
	case apperrors.CodeSystemError:
		status = http.StatusInternalServerError
	}
	response.Error(ctx, status, businessErr)
}
