package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	treeservice "tree/backend/internal/family/tree/service"
)

type TreeHandler struct {
	service treeservice.TreeService
}

func NewTreeHandler(service treeservice.TreeService) *TreeHandler {
	return &TreeHandler{service: service}
}

func (h *TreeHandler) PrivateTree(ctx *gin.Context) {
	actorID, err := middleware.CurrentUserID(ctx)
	if err != nil {
		response.Abort(ctx, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	familyID, ok := familyID(ctx)
	if !ok {
		return
	}
	result, businessErr := h.service.GetPrivateTree(ctx.Request.Context(), actorID, familyID)
	writeResult(ctx, result, businessErr)
}

func (h *TreeHandler) PublicTree(ctx *gin.Context) {
	familyID, ok := familyID(ctx)
	if !ok {
		return
	}
	result, businessErr := h.service.GetPublicTree(ctx.Request.Context(), familyID)
	writeResult(ctx, result, businessErr)
}

func familyID(ctx *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(ctx.Param("familyId"), 10, 64)
	if err != nil || id == 0 {
		response.Abort(ctx, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return 0, false
	}
	return id, true
}

func writeResult(ctx *gin.Context, result any, businessErr *apperrors.BusinessError) {
	if businessErr == nil {
		response.OK(ctx, result)
		return
	}
	status := http.StatusBadRequest
	switch businessErr.Code {
	case treeservice.CodePrivateTreeForbidden, treeservice.CodePublicTreeForbidden:
		status = http.StatusForbidden
	case apperrors.CodeSystemError:
		status = http.StatusInternalServerError
	}
	response.Error(ctx, status, businessErr)
}
