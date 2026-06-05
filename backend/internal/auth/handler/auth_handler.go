package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tree/backend/internal/auth/dto"
	authservice "tree/backend/internal/auth/service"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
)

type AuthHandler struct {
	service authservice.AuthService
}

func NewAuthHandler(service authservice.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) SendCode(ctx *gin.Context) {
	var req dto.SendCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}

	result, businessErr := h.service.SendCode(ctx.Request.Context(), authservice.SendCodeInput{
		Phone:      req.Phone,
		Scene:      req.Scene,
		ClientType: req.ClientType,
		IP:         ctx.ClientIP(),
		UserAgent:  ctx.Request.UserAgent(),
	})
	if businessErr != nil {
		response.Error(ctx, http.StatusBadRequest, businessErr)
		return
	}

	response.OK(ctx, result)
}

func (h *AuthHandler) RegisterPhone(ctx *gin.Context) {
	var req dto.RegisterPhoneRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}

	result, businessErr := h.service.RegisterPhone(ctx.Request.Context(), authservice.RegisterPhoneInput{
		Phone:      req.Phone,
		Code:       req.Code,
		Password:   req.Password,
		Nickname:   req.Nickname,
		ClientType: req.ClientType,
		IP:         ctx.ClientIP(),
		UserAgent:  ctx.Request.UserAgent(),
	})
	if businessErr != nil {
		response.Error(ctx, http.StatusBadRequest, businessErr)
		return
	}

	response.OK(ctx, result)
}

func (h *AuthHandler) LoginPhone(ctx *gin.Context) {
	var req dto.LoginPhoneRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}

	result, businessErr := h.service.LoginPhone(ctx.Request.Context(), authservice.LoginPhoneInput{
		Phone:      req.Phone,
		Password:   req.Password,
		ClientType: req.ClientType,
		IP:         ctx.ClientIP(),
		UserAgent:  ctx.Request.UserAgent(),
	})
	if businessErr != nil {
		response.Error(ctx, http.StatusUnauthorized, businessErr)
		return
	}

	response.OK(ctx, result)
}

func (h *AuthHandler) Logout(ctx *gin.Context) {
	userID, err := middleware.CurrentUserID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	tokenID, _ := middleware.CurrentJWTID(ctx)
	expiresAt, _ := middleware.CurrentJWTExpiresAt(ctx)

	if businessErr := h.service.Logout(ctx.Request.Context(), authservice.LogoutInput{
		UserID:    userID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
		IP:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
	}); businessErr != nil {
		response.Error(ctx, http.StatusInternalServerError, businessErr)
		return
	}

	response.OK(ctx, gin.H{"status": "ok"})
}
