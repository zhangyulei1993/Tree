package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	choicedto "tree/backend/internal/choicescenario/dto"
	choiceservice "tree/backend/internal/choicescenario/service"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
)

type Handler struct {
	service choiceservice.Service
}

func New(service choiceservice.Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(ctx *gin.Context) {
	result, businessErr := h.service.List(ctx.Request.Context(), choicedto.ListQuery{
		Period: ctx.DefaultQuery("period", "today"),
	})
	write(ctx, result, businessErr)
}

func (h *Handler) Create(ctx *gin.Context) {
	userID, err := middleware.CurrentUserID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	var req choicedto.CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, apperrors.New(apperrors.CodeChoiceScenarioInvalidInput))
		return
	}
	result, businessErr := h.service.Create(ctx.Request.Context(), userID, req, choiceservice.AuditInput{
		IP: ctx.ClientIP(), UserAgent: ctx.GetHeader("User-Agent"),
	})
	if businessErr != nil {
		write(ctx, result, businessErr)
		return
	}
	response.Created(ctx, result)
}

func (h *Handler) Use(ctx *gin.Context) {
	userID, err := middleware.CurrentUserID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	scenarioID, err := strconv.ParseUint(ctx.Param("scenarioId"), 10, 64)
	if err != nil || scenarioID == 0 {
		response.Error(ctx, http.StatusBadRequest, apperrors.New(apperrors.CodeChoiceScenarioInvalidInput))
		return
	}
	businessErr := h.service.Use(ctx.Request.Context(), userID, scenarioID)
	write(ctx, map[string]bool{"used": businessErr == nil}, businessErr)
}

func write(ctx *gin.Context, result any, businessErr *apperrors.BusinessError) {
	if businessErr == nil {
		response.OK(ctx, result)
		return
	}
	status := http.StatusBadRequest
	switch businessErr.Code {
	case apperrors.CodeChoiceScenarioNotFound:
		status = http.StatusNotFound
	case apperrors.CodeChoiceScenarioRateLimited:
		status = http.StatusTooManyRequests
	case apperrors.CodeContentSafetyUnavailable:
		status = http.StatusServiceUnavailable
	case apperrors.CodeContentSafetyWechatRequired, apperrors.CodeForbidden:
		status = http.StatusForbidden
	case apperrors.CodeSystemError:
		status = http.StatusInternalServerError
	}
	response.Error(ctx, status, businessErr)
}
