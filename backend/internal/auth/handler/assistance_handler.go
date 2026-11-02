package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"tree/backend/internal/auth/dto"
	service "tree/backend/internal/auth/service"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
)

type AssistanceHandler struct{ service *service.AssistanceService }

func NewAssistanceHandler(s *service.AssistanceService) *AssistanceHandler {
	return &AssistanceHandler{service: s}
}
func (h *AssistanceHandler) Generate(c *gin.Context) {
	id, e := middleware.CurrentUserID(c)
	if e != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	var req dto.GenerateAssistanceCodeRequest
	if c.Request.ContentLength > 0 && c.ShouldBindJSON(&req) != nil {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	v, b := h.service.Generate(c.Request.Context(), service.GenerateAssistanceInput{UserID: id, ValidDays: req.ValidDays, IP: c.ClientIP(), UserAgent: c.Request.UserAgent()})
	if b != nil {
		response.Error(c, http.StatusBadRequest, b)
		return
	}
	response.OK(c, v)
}
func (h *AssistanceHandler) List(c *gin.Context) {
	id, e := middleware.CurrentUserID(c)
	if e != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	v, b := h.service.List(c.Request.Context(), id)
	if b != nil {
		response.Error(c, http.StatusBadRequest, b)
		return
	}
	response.OK(c, v)
}
func (h *AssistanceHandler) Redeem(c *gin.Context) {
	var req dto.RedeemAssistanceCodeRequest
	if c.ShouldBindJSON(&req) != nil {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	v, b := h.service.Redeem(c.Request.Context(), service.RedeemAssistanceInput{Phone: req.Phone, Code: req.Code, IP: c.ClientIP(), UserAgent: c.Request.UserAgent()})
	if b != nil {
		response.Error(c, http.StatusUnauthorized, b)
		return
	}
	response.OK(c, v)
}
func (h *AssistanceHandler) Revoke(c *gin.Context) {
	id, e := middleware.CurrentUserID(c)
	codeID, e2 := strconv.ParseUint(c.Param("codeId"), 10, 64)
	if e != nil || e2 != nil || codeID == 0 {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	if b := h.service.Revoke(c.Request.Context(), service.RevokeAssistanceInput{UserID: id, CodeID: codeID, IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}); b != nil {
		response.Error(c, http.StatusBadRequest, b)
		return
	}
	response.OK(c, gin.H{"status": "ok"})
}
