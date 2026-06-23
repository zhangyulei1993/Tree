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

func (h *AuthHandler) GetMe(ctx *gin.Context) {
	userID, err := middleware.CurrentUserID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}

	result, businessErr := h.service.GetMe(ctx.Request.Context(), userID)
	if businessErr != nil {
		response.Error(ctx, http.StatusBadRequest, businessErr)
		return
	}
	response.OK(ctx, result)
}

func (h *AuthHandler) UpdateProfile(ctx *gin.Context) {
	userID, err := middleware.CurrentUserID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}

	var req dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}

	result, businessErr := h.service.UpdateProfile(ctx.Request.Context(), authservice.UpdateProfileInput{
		UserID:    userID,
		Nickname:  req.Nickname,
		IP:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
	})
	if businessErr != nil {
		response.Error(ctx, http.StatusBadRequest, businessErr)
		return
	}
	response.OK(ctx, result)
}

func (h *AuthHandler) UploadAvatar(ctx *gin.Context) {
	userID, err := middleware.CurrentUserID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}

	result, businessErr := h.service.UploadAvatar(ctx.Request.Context(), authservice.UploadAvatarInput{
		UserID:    userID,
		File:      file,
		IP:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
	})
	if businessErr != nil {
		response.Error(ctx, http.StatusBadRequest, businessErr)
		return
	}
	response.OK(ctx, result)
}
