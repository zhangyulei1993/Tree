package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	adminrepo "tree/backend/internal/admin/repository"
	adminservice "tree/backend/internal/admin/service"
	authservice "tree/backend/internal/auth/service"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
)

type ManagementHandler struct {
	service     *adminservice.ManagementService
	authService authservice.AuthService
}

func NewManagementHandler(service *adminservice.ManagementService, authService authservice.AuthService) *ManagementHandler {
	return &ManagementHandler{service: service, authService: authService}
}
func query(ctx *gin.Context) adminrepo.PageQuery {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	familyID, _ := strconv.ParseUint(ctx.Query("familyId"), 10, 64)
	return adminrepo.PageQuery{Keyword: ctx.Query("keyword"), Status: ctx.Query("status"), Role: ctx.Query("role"), Result: ctx.Query("result"), FamilyID: familyID, Page: page, PageSize: size}
}
func id(ctx *gin.Context, name string) (uint64, bool) {
	value, err := strconv.ParseUint(ctx.Param(name), 10, 64)
	if err != nil || value == 0 {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return 0, false
	}
	return value, true
}
func write(ctx *gin.Context, value any, businessErr *apperrors.BusinessError) {
	if businessErr != nil {
		status := http.StatusInternalServerError
		switch businessErr.Code {
		case apperrors.CodeResourceNotFound:
			status = http.StatusNotFound
		case apperrors.CodeAdminUnbindPhoneLoginNoWechat:
			status = http.StatusConflict
		}
		response.Error(ctx, status, businessErr)
		return
	}
	response.OK(ctx, value)
}
func (h *ManagementHandler) Dashboard(ctx *gin.Context) {
	v, e := h.service.Dashboard(ctx.Request.Context())
	write(ctx, v, e)
}
func (h *ManagementHandler) Users(ctx *gin.Context) {
	v, e := h.service.Users(ctx.Request.Context(), query(ctx))
	write(ctx, v, e)
}
func (h *ManagementHandler) User(ctx *gin.Context) {
	value, ok := id(ctx, "userId")
	if !ok {
		return
	}
	v, e := h.service.User(ctx.Request.Context(), value)
	write(ctx, v, e)
}
func (h *ManagementHandler) Families(ctx *gin.Context) {
	v, e := h.service.Families(ctx.Request.Context(), query(ctx))
	write(ctx, v, e)
}
func (h *ManagementHandler) Family(ctx *gin.Context) {
	value, ok := id(ctx, "familyId")
	if !ok {
		return
	}
	v, e := h.service.Family(ctx.Request.Context(), value)
	write(ctx, v, e)
}
func (h *ManagementHandler) Members(ctx *gin.Context) {
	value, ok := id(ctx, "familyId")
	if !ok {
		return
	}
	v, e := h.service.Members(ctx.Request.Context(), value)
	write(ctx, v, e)
}
func (h *ManagementHandler) Admins(ctx *gin.Context) {
	role, _ := middleware.CurrentAdminRole(ctx)
	if role != string(enums.AdminRoleRootAdmin) && role != string(enums.AdminRoleSuperAdmin) {
		response.Abort(ctx, http.StatusForbidden, apperrors.CodeForbidden)
		return
	}
	v, e := h.service.Admins(ctx.Request.Context(), query(ctx))
	write(ctx, v, e)
}
func (h *ManagementHandler) Logs(ctx *gin.Context) {
	v, e := h.service.Logs(ctx.Request.Context(), query(ctx))
	write(ctx, v, e)
}

func (h *ManagementHandler) UnbindPhoneLogin(ctx *gin.Context) {
	role, _ := middleware.CurrentAdminRole(ctx)
	if role != string(enums.AdminRoleRootAdmin) && role != string(enums.AdminRoleSuperAdmin) {
		response.Abort(ctx, http.StatusForbidden, apperrors.CodeForbidden)
		return
	}
	adminID, err := middleware.CurrentAdminID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	userID, ok := id(ctx, "userId")
	if !ok {
		return
	}
	var req struct {
		Confirm bool   `json:"confirm"`
		Reason  string `json:"reason"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	if h.authService == nil {
		response.Abort(ctx, http.StatusInternalServerError, apperrors.CodeSystemError)
		return
	}
	if businessErr := h.authService.AdminUnbindPhoneLogin(ctx.Request.Context(), authservice.AdminUnbindPhoneLoginInput{
		AdminID: adminID, AdminRole: role, TargetUserID: userID, Confirm: req.Confirm, Reason: req.Reason,
		IP: ctx.ClientIP(), UserAgent: ctx.Request.UserAgent(),
	}); businessErr != nil {
		write(ctx, nil, businessErr)
		return
	}
	response.OK(ctx, gin.H{"status": "ok"})
}
