package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	"tree/backend/internal/family/invitationreward/dto"
	reward "tree/backend/internal/family/invitationreward/service"
)

type Handler struct{ service reward.Service }

func New(s reward.Service) *Handler { return &Handler{service: s} }
func ids(c *gin.Context) (uint64, uint64, bool) {
	u, e := middleware.CurrentUserID(c)
	if e != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return 0, 0, false
	}
	f, e := strconv.ParseUint(c.Param("familyId"), 10, 64)
	if e != nil || f == 0 {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return 0, 0, false
	}
	return u, f, true
}
func (h *Handler) Summary(c *gin.Context) {
	u, f, ok := ids(c)
	if !ok {
		return
	}
	r, e := h.service.Summary(c, u, f)
	if e != nil {
		response.Error(c, http.StatusBadRequest, e)
		return
	}
	response.OK(c, r)
}
func (h *Handler) Share(c *gin.Context) {
	u, f, ok := ids(c)
	if !ok {
		return
	}
	r, e := h.service.CreateShare(c, u, f)
	if e != nil {
		response.Error(c, http.StatusBadRequest, e)
		return
	}
	response.Created(c, r)
}
func (h *Handler) Qualify(c *gin.Context) {
	u, e := middleware.CurrentUserID(c)
	if e != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	r, b := h.service.Qualify(c, u, c.Param("referralCode"))
	if b != nil {
		response.Error(c, http.StatusBadRequest, b)
		return
	}
	response.OK(c, r)
}
func canManage(c *gin.Context) bool {
	role, err := middleware.CurrentAdminRole(c)
	return err == nil && (role == "ROOT_ADMIN" || role == "SUPER_ADMIN")
}
func (h *Handler) AdminGet(c *gin.Context) {
	if !canManage(c) {
		response.Abort(c, http.StatusForbidden, apperrors.CodeForbidden)
		return
	}
	r, e := h.service.AdminGet(c)
	if e != nil {
		response.Error(c, http.StatusInternalServerError, e)
		return
	}
	response.OK(c, r)
}
func (h *Handler) AdminUpdate(c *gin.Context) {
	if !canManage(c) {
		response.Abort(c, http.StatusForbidden, apperrors.CodeForbidden)
		return
	}
	admin, e := middleware.CurrentAdminID(c)
	if e != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	var req dto.Config
	if c.ShouldBindJSON(&req) != nil {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	r, b := h.service.AdminUpdate(c, admin, req)
	if b != nil {
		response.Error(c, http.StatusBadRequest, b)
		return
	}
	response.OK(c, r)
}
