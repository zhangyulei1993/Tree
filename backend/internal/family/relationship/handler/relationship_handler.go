package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	"tree/backend/internal/family/relationship/dto"
	relationshipservice "tree/backend/internal/family/relationship/service"
)

type RelationshipHandler struct {
	service relationshipservice.RelationshipService
}

func NewRelationshipHandler(service relationshipservice.RelationshipService) *RelationshipHandler {
	return &RelationshipHandler{service: service}
}

func (h *RelationshipHandler) Create(ctx *gin.Context) {
	actorID, familyID, ok := actorAndFamily(ctx)
	if !ok {
		return
	}
	var req dto.CreateRelationshipRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.Create(ctx.Request.Context(), actorID, familyID, req, audit(ctx))
	writeResult(ctx, http.StatusCreated, result, businessErr)
}

func (h *RelationshipHandler) PlaceExisting(ctx *gin.Context) {
	actorID, familyID, ok := actorAndFamily(ctx)
	if !ok {
		return
	}
	var req dto.PlaceExistingMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.PlaceExisting(ctx.Request.Context(), actorID, familyID, req, audit(ctx))
	writeResult(ctx, http.StatusCreated, result, businessErr)
}

func (h *RelationshipHandler) Update(ctx *gin.Context) {
	actorID, familyID, relationshipID, ok := actorFamilyRelationship(ctx)
	if !ok {
		return
	}
	var req dto.UpdateRelationshipRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.Update(ctx.Request.Context(), actorID, familyID, relationshipID, req, audit(ctx))
	writeResult(ctx, http.StatusOK, result, businessErr)
}

func (h *RelationshipHandler) Delete(ctx *gin.Context) {
	actorID, familyID, relationshipID, ok := actorFamilyRelationship(ctx)
	if !ok {
		return
	}
	var req dto.DeleteRelationshipRequest
	if ctx.Request.ContentLength != 0 {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
	}
	result, businessErr := h.service.Delete(ctx.Request.Context(), actorID, familyID, relationshipID, req, audit(ctx))
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

func actorFamilyRelationship(ctx *gin.Context) (uint64, uint64, uint64, bool) {
	actorID, familyID, ok := actorAndFamily(ctx)
	if !ok {
		return 0, 0, 0, false
	}
	relationshipID, err := positiveID(ctx.Param("relationshipId"))
	if err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return 0, 0, 0, false
	}
	return actorID, familyID, relationshipID, true
}

func positiveID(value string) (uint64, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}

func audit(ctx *gin.Context) relationshipservice.AuditInput {
	return relationshipservice.AuditInput{IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent()}
}

func writeResult[T any](ctx *gin.Context, status int, result T, businessErr *apperrors.BusinessError) {
	if businessErr != nil {
		writeError(ctx, businessErr)
		return
	}
	response.Success(ctx, status, result)
}

func writeError(ctx *gin.Context, businessErr *apperrors.BusinessError) {
	status := http.StatusBadRequest
	switch businessErr.Code {
	case relationshipservice.CodeRelationshipForbidden:
		status = http.StatusForbidden
	case relationshipservice.CodeRelationshipMember,
		relationshipservice.CodeRelationshipNotFound,
		apperrors.CodeResourceNotFound:
		status = http.StatusNotFound
	case apperrors.CodeSystemError:
		status = http.StatusInternalServerError
	}
	response.Error(ctx, status, businessErr)
}
