package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	quotadto "tree/backend/internal/accountquota/dto"
	quotaservice "tree/backend/internal/accountquota/service"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
)

type Handler struct {
	service quotaservice.Service
}

func NewHandler(service quotaservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Capabilities(ctx *gin.Context) {
	userID, err := middleware.CurrentUserID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	result, businessErr := h.service.GetCapabilities(ctx.Request.Context(), userID)
	write(ctx, result, businessErr)
}

func (h *Handler) ListConfigs(ctx *gin.Context) {
	role, err := middleware.CurrentAdminRole(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	result, businessErr := h.service.ListConfigs(ctx.Request.Context(), role)
	write(ctx, result, businessErr)
}

func (h *Handler) UpdateConfig(ctx *gin.Context) {
	adminID, err := middleware.CurrentAdminID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	role, err := middleware.CurrentAdminRole(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	tier := strings.ToUpper(strings.TrimSpace(ctx.Param("tier")))
	if tier == "" {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	var req quotadto.UpdateConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	values, err := req.Parse()
	if err != nil {
		if errors.Is(err, quotadto.ErrMissingConfigField) {
			response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.UpdateConfig(ctx.Request.Context(), adminID, role, tier, values, audit(ctx))
	write(ctx, result, businessErr)
}

func (h *Handler) PreviewImpact(ctx *gin.Context) {
	role, err := middleware.CurrentAdminRole(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	tier := strings.ToUpper(strings.TrimSpace(ctx.Param("tier")))
	if tier == "" {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	var req quotadto.ImpactPreviewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	values, err := req.Parse()
	if err != nil {
		if errors.Is(err, quotadto.ErrMissingConfigField) {
			response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.PreviewImpact(ctx.Request.Context(), role, tier, values)
	write(ctx, result, businessErr)
}

func (h *Handler) ListFeatureOverrides(ctx *gin.Context) {
	role, err := middleware.CurrentAdminRole(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	featureKey := strings.ToUpper(strings.TrimSpace(ctx.Param("featureKey")))
	if featureKey == "" {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.ListFeatureOverrides(ctx.Request.Context(), role, featureKey)
	write(ctx, result, businessErr)
}

func (h *Handler) UpdateFeatureOverrides(ctx *gin.Context) {
	adminID, err := middleware.CurrentAdminID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	role, err := middleware.CurrentAdminRole(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	featureKey := strings.ToUpper(strings.TrimSpace(ctx.Param("featureKey")))
	if featureKey == "" {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	var req quotadto.UpdateFeatureOverridesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.UpdateFeatureOverrides(ctx.Request.Context(), adminID, role, featureKey, req.Phones, audit(ctx))
	write(ctx, result, businessErr)
}

func (h *Handler) DeleteFeatureOverride(ctx *gin.Context) {
	adminID, err := middleware.CurrentAdminID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	role, err := middleware.CurrentAdminRole(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	featureKey := strings.ToUpper(strings.TrimSpace(ctx.Param("featureKey")))
	if featureKey == "" {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	overrideID, parseErr := strconv.ParseUint(strings.TrimSpace(ctx.Param("overrideId")), 10, 64)
	if parseErr != nil || overrideID == 0 {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.DeleteFeatureOverride(ctx.Request.Context(), adminID, role, featureKey, overrideID, audit(ctx))
	write(ctx, result, businessErr)
}

func write(ctx *gin.Context, value any, businessErr *apperrors.BusinessError) {
	if businessErr != nil {
		status := http.StatusBadRequest
		switch businessErr.Code {
		case apperrors.CodeForbidden, apperrors.CodeQuotaConfigForbidden:
			status = http.StatusForbidden
		case apperrors.CodeResourceNotFound:
			status = http.StatusNotFound
		case apperrors.CodeSystemError:
			status = http.StatusInternalServerError
		}
		response.Error(ctx, status, businessErr)
		return
	}
	response.OK(ctx, value)
}

func audit(ctx *gin.Context) quotaservice.AuditInput {
	return quotaservice.AuditInput{
		IP:        ctx.ClientIP(),
		UserAgent: ctx.GetHeader("User-Agent"),
	}
}
