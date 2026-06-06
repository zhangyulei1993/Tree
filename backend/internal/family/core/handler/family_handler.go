package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	"tree/backend/internal/family/core/dto"
	familyservice "tree/backend/internal/family/core/service"
)

type FamilyHandler struct {
	service familyservice.FamilyService
}

func NewFamilyHandler(service familyservice.FamilyService) *FamilyHandler {
	return &FamilyHandler{service: service}
}

func (h *FamilyHandler) Create(ctx *gin.Context) {
	userID, ok := currentUser(ctx)
	if !ok {
		return
	}
	var req dto.CreateFamilyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.Create(ctx.Request.Context(), userID, req, audit(ctx))
	if businessErr != nil {
		response.Error(ctx, http.StatusBadRequest, businessErr)
		return
	}
	response.Created(ctx, result)
}

func (h *FamilyHandler) List(ctx *gin.Context) {
	userID, ok := currentUser(ctx)
	if !ok {
		return
	}
	result, businessErr := h.service.List(ctx.Request.Context(), userID)
	if businessErr != nil {
		response.Error(ctx, http.StatusInternalServerError, businessErr)
		return
	}
	response.OK(ctx, result)
}

func (h *FamilyHandler) Detail(ctx *gin.Context) {
	h.withFamilyID(ctx, func(familyID uint64) {
		userID, ok := currentUser(ctx)
		if !ok {
			return
		}
		result, businessErr := h.service.Detail(ctx.Request.Context(), userID, familyID)
		writeResult(ctx, result, businessErr)
	})
}

func (h *FamilyHandler) PublicDetail(ctx *gin.Context) {
	h.withFamilyID(ctx, func(familyID uint64) {
		result, businessErr := h.service.PublicDetail(ctx.Request.Context(), familyID)
		writeResult(ctx, result, businessErr)
	})
}

func (h *FamilyHandler) Update(ctx *gin.Context) {
	h.withFamilyID(ctx, func(familyID uint64) {
		userID, ok := currentUser(ctx)
		if !ok {
			return
		}
		var req dto.UpdateFamilyRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
		result, businessErr := h.service.Update(ctx.Request.Context(), userID, familyID, req, audit(ctx))
		writeResult(ctx, result, businessErr)
	})
}

func (h *FamilyHandler) CreateDissolutionRequest(ctx *gin.Context) {
	h.withFamilyID(ctx, func(familyID uint64) {
		userID, ok := currentUser(ctx)
		if !ok {
			return
		}
		var req dto.CreateDissolutionRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
		result, businessErr := h.service.CreateDissolutionRequest(ctx.Request.Context(), userID, familyID, req, audit(ctx))
		writeResult(ctx, result, businessErr)
	})
}

func (h *FamilyHandler) CurrentDissolutionRequest(ctx *gin.Context) {
	h.withFamilyID(ctx, func(familyID uint64) {
		userID, ok := currentUser(ctx)
		if !ok {
			return
		}
		result, businessErr := h.service.CurrentDissolutionRequest(ctx.Request.Context(), userID, familyID)
		writeResult(ctx, result, businessErr)
	})
}

func (h *FamilyHandler) CancelDissolutionRequest(ctx *gin.Context) {
	h.withFamilyID(ctx, func(familyID uint64) {
		requestID, err := strconv.ParseUint(ctx.Param("requestId"), 10, 64)
		if err != nil || requestID == 0 {
			response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
		userID, ok := currentUser(ctx)
		if !ok {
			return
		}
		var req dto.CancelDissolutionRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
		result, businessErr := h.service.CancelDissolutionRequest(ctx.Request.Context(), userID, familyID, requestID, req, audit(ctx))
		writeResult(ctx, result, businessErr)
	})
}

func (h *FamilyHandler) withFamilyID(ctx *gin.Context, fn func(uint64)) {
	familyID, err := strconv.ParseUint(ctx.Param("familyId"), 10, 64)
	if err != nil || familyID == 0 {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	fn(familyID)
}

func currentUser(ctx *gin.Context) (uint64, bool) {
	userID, err := middleware.CurrentUserID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return 0, false
	}
	return userID, true
}

func audit(ctx *gin.Context) familyservice.AuditInput {
	return familyservice.AuditInput{IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent()}
}

func writeResult[T any](ctx *gin.Context, result T, businessErr *apperrors.BusinessError) {
	if businessErr != nil {
		status := http.StatusBadRequest
		if businessErr.Code == familyservice.CodeFamilyDetailForbidden ||
			businessErr.Code == familyservice.CodeFamilyUpdateForbidden ||
			businessErr.Code == familyservice.CodeFamilyDissolutionForbidden {
			status = http.StatusForbidden
		}
		if businessErr.Code == familyservice.CodeFamilyNotFound || businessErr.Code == apperrors.CodeResourceNotFound {
			status = http.StatusNotFound
		}
		response.Error(ctx, status, businessErr)
		return
	}
	response.OK(ctx, result)
}
