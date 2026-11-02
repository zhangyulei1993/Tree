package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tree/backend/internal/admin/dto"
	adminservice "tree/backend/internal/admin/service"
	"tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
)

type AuthHandler struct {
	service adminservice.AuthService
}

func NewAuthHandler(service adminservice.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, errors.CodeInvalidParams)
		return
	}

	result, businessErr := h.service.Login(ctx.Request.Context(), adminservice.LoginInput{
		Username:  req.Username,
		Password:  req.Password,
		IP:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
	})
	if businessErr != nil {
		response.Error(ctx, http.StatusUnauthorized, businessErr)
		return
	}

	response.OK(ctx, result)
}

func (h *AuthHandler) Logout(ctx *gin.Context) {
	adminID, err := middleware.CurrentAdminID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, errors.CodeUnauthorized)
		return
	}
	role, _ := middleware.CurrentAdminRole(ctx)
	tokenID, _ := middleware.CurrentJWTID(ctx)
	expiresAt, _ := middleware.CurrentJWTExpiresAt(ctx)

	if businessErr := h.service.Logout(ctx.Request.Context(), adminservice.LogoutInput{
		AdminID:   adminID,
		Role:      role,
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

func (h *AuthHandler) Me(ctx *gin.Context) {
	adminID, err := middleware.CurrentAdminID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, errors.CodeUnauthorized)
		return
	}

	info, businessErr := h.service.Me(ctx.Request.Context(), adminID)
	if businessErr != nil {
		response.Error(ctx, http.StatusForbidden, businessErr)
		return
	}

	response.OK(ctx, info)
}

func (h *AuthHandler) ChangePassword(ctx *gin.Context) {
	adminID, err := middleware.CurrentAdminID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, errors.CodeUnauthorized)
		return
	}
	role, _ := middleware.CurrentAdminRole(ctx)
	tokenID, _ := middleware.CurrentJWTID(ctx)
	expiresAt, _ := middleware.CurrentJWTExpiresAt(ctx)

	var req dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, errors.CodeInvalidParams)
		return
	}

	if businessErr := h.service.ChangePassword(ctx.Request.Context(), adminservice.ChangePasswordInput{
		AdminID:     adminID,
		Role:        role,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
		TokenID:     tokenID,
		ExpiresAt:   expiresAt,
		IP:          ctx.ClientIP(),
		UserAgent:   ctx.Request.UserAgent(),
	}); businessErr != nil {
		response.Error(ctx, http.StatusBadRequest, businessErr)
		return
	}

	response.OK(ctx, gin.H{"status": "ok"})
}
