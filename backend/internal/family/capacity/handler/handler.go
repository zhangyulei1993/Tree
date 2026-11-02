package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/response"
	"tree/backend/internal/family/capacity/dto"
	capacityservice "tree/backend/internal/family/capacity/service"
	"tree/backend/internal/family/capacity/wechatpay"
)

type Handler struct {
	service  capacityservice.Service
	provider *wechatpay.Provider
	appID    string
	mchID    string
}

func New(s capacityservice.Service, p *wechatpay.Provider, appID, mchID string) *Handler {
	return &Handler{service: s, provider: p, appID: appID, mchID: mchID}
}
func (h *Handler) Products(c *gin.Context) {
	result, err := h.service.ListProducts(c)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) AdminProducts(c *gin.Context) {
	if !isRefundReviewer(c) {
		response.Abort(c, http.StatusForbidden, apperrors.CodeForbidden)
		return
	}
	result, err := h.service.AdminListProducts(c)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err)
		return
	}
	response.OK(c, result)
}
func (h *Handler) AdminUpdateProduct(c *gin.Context) {
	if !isRefundReviewer(c) {
		response.Abort(c, http.StatusForbidden, apperrors.CodeForbidden)
		return
	}
	adminID, err := middleware.CurrentAdminID(c)
	if err != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	var req dto.UpdateProductRequest
	if c.ShouldBindJSON(&req) != nil {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.AdminUpdateProduct(c, adminID, c.Param("productKey"), req)
	if businessErr != nil {
		status := http.StatusBadRequest
		if businessErr.Code == apperrors.CodeResourceNotFound {
			status = http.StatusNotFound
		}
		response.Error(c, status, businessErr)
		return
	}
	response.OK(c, result)
}
func (h *Handler) AdminOrders(c *gin.Context) {
	if !isRefundReviewer(c) {
		response.Abort(c, http.StatusForbidden, apperrors.CodeForbidden)
		return
	}
	query := dto.AdminOrderQuery{Page: 1, PageSize: 20}
	if value := c.Query("page"); value != "" {
		page, err := strconv.Atoi(value)
		if err != nil || page < 1 || page > 1000000 {
			response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
		query.Page = page
	}
	if value := c.Query("pageSize"); value != "" {
		pageSize, err := strconv.Atoi(value)
		if err != nil || pageSize < 1 || pageSize > 100 {
			response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
		query.PageSize = pageSize
	}
	query.Status = strings.TrimSpace(c.Query("status"))
	if query.Status == "ALL" {
		query.Status = ""
	}
	if query.Status != "" && query.Status != "PENDING" && query.Status != "PAID" && query.Status != "REFUNDING" && query.Status != "REFUNDED" && query.Status != "CANCELLED" {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	if value := c.Query("startDate"); value != "" {
		date, err := time.ParseInLocation("2006-01-02", value, time.Local)
		if err != nil {
			response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
		query.StartDate = &date
	}
	if value := c.Query("endDate"); value != "" {
		date, err := time.ParseInLocation("2006-01-02", value, time.Local)
		if err != nil {
			response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
		endDate := date.AddDate(0, 0, 1)
		query.EndDate = &endDate
	}
	if query.StartDate != nil && query.EndDate != nil && !query.StartDate.Before(*query.EndDate) {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, err := h.service.AdminListOrders(c, query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err)
		return
	}
	response.OK(c, result)
}
func (h *Handler) AdminReconcilePayment(c *gin.Context) {
	if !isRefundReviewer(c) {
		response.Abort(c, http.StatusForbidden, apperrors.CodeForbidden)
		return
	}
	adminID, err := middleware.CurrentAdminID(c)
	if err != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	status, businessErr := h.service.AdminReconcilePayment(c, adminID, c.Param("orderNo"))
	if businessErr != nil {
		response.Error(c, http.StatusBadRequest, businessErr)
		return
	}
	response.OK(c, gin.H{"status": status})
}
func (h *Handler) AdminApproveRefund(c *gin.Context) {
	if !isRefundReviewer(c) {
		response.Abort(c, http.StatusForbidden, apperrors.CodeForbidden)
		return
	}
	adminID, err := middleware.CurrentAdminID(c)
	if err != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	var req dto.AdminRefundRequest
	if c.Request.ContentLength > 0 && c.ShouldBindJSON(&req) != nil {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	if businessErr := h.service.AdminApproveRefund(c, adminID, c.Param("refundNo"), req.Reason); businessErr != nil {
		response.Error(c, http.StatusBadRequest, businessErr)
		return
	}
	response.OK(c, gin.H{"status": "PROCESSING"})
}
func (h *Handler) AdminRejectRefund(c *gin.Context) {
	if !isRefundReviewer(c) {
		response.Abort(c, http.StatusForbidden, apperrors.CodeForbidden)
		return
	}
	adminID, err := middleware.CurrentAdminID(c)
	if err != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	var req dto.AdminRefundRequest
	if c.ShouldBindJSON(&req) != nil || req.Reason == "" {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	if businessErr := h.service.AdminRejectRefund(c, adminID, c.Param("refundNo"), req.Reason); businessErr != nil {
		response.Error(c, http.StatusBadRequest, businessErr)
		return
	}
	response.OK(c, gin.H{"status": "REJECTED"})
}

func (h *Handler) AdminReconcileRefund(c *gin.Context) {
	if !isRefundReviewer(c) {
		response.Abort(c, http.StatusForbidden, apperrors.CodeForbidden)
		return
	}
	adminID, err := middleware.CurrentAdminID(c)
	if err != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	status, businessErr := h.service.AdminReconcileRefund(c, adminID, c.Param("refundNo"))
	if businessErr != nil {
		response.Error(c, http.StatusBadRequest, businessErr)
		return
	}
	response.OK(c, gin.H{"status": status})
}

func (h *Handler) AdminRetryRefund(c *gin.Context) {
	if !isRefundReviewer(c) {
		response.Abort(c, http.StatusForbidden, apperrors.CodeForbidden)
		return
	}
	adminID, err := middleware.CurrentAdminID(c)
	if err != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return
	}
	status, businessErr := h.service.AdminRetryRefund(c, adminID, c.Param("refundNo"))
	if businessErr != nil {
		response.Error(c, http.StatusBadRequest, businessErr)
		return
	}
	response.OK(c, gin.H{"status": status})
}
func (h *Handler) Orders(c *gin.Context) {
	userID, familyID, ok := ids(c)
	if !ok {
		return
	}
	page, pageSize := 1, 20
	var err error
	if value := c.Query("page"); value != "" {
		page, err = strconv.Atoi(value)
		if err != nil || page < 1 {
			response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
	}
	if value := c.Query("pageSize"); value != "" {
		pageSize, err = strconv.Atoi(value)
		if err != nil || pageSize < 1 || pageSize > 50 {
			response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
	}
	status := strings.TrimSpace(c.Query("status"))
	if status == "ALL" {
		status = ""
	}
	validStatuses := map[string]bool{"": true, "PENDING": true, "PAID": true, "REFUNDING": true, "REFUNDED": true, "CANCELLED": true}
	if !validStatuses[status] {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	var startDate, endDate *time.Time
	if value := c.Query("startDate"); value != "" {
		date, parseErr := time.ParseInLocation("2006-01-02", value, time.Local)
		if parseErr != nil {
			response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
		startDate = &date
	}
	if value := c.Query("endDate"); value != "" {
		date, parseErr := time.ParseInLocation("2006-01-02", value, time.Local)
		if parseErr != nil {
			response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
			return
		}
		end := date.AddDate(0, 0, 1)
		endDate = &end
	}
	if startDate != nil && endDate != nil && !(*startDate).Before(*endDate) {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, businessErr := h.service.ListOrdersPage(c, userID, familyID, page, pageSize, status, startDate, endDate)
	if businessErr != nil {
		response.Error(c, http.StatusBadRequest, businessErr)
		return
	}
	response.OK(c, result)
}
func (h *Handler) CreateOrder(c *gin.Context) {
	userID, familyID, ok := ids(c)
	if !ok {
		return
	}
	var req dto.CreateOrderRequest
	if c.ShouldBindJSON(&req) != nil {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result, err := h.service.CreateOrder(c, userID, familyID, req.ProductKey, "")
	if err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}
	response.Created(c, result)
}
func (h *Handler) ResumePayment(c *gin.Context) {
	userID, familyID, ok := ids(c)
	if !ok {
		return
	}
	result, err := h.service.ResumePayment(c, userID, familyID, c.Param("orderNo"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}
	response.OK(c, result)
}
func (h *Handler) ReconcilePayment(c *gin.Context) {
	userID, familyID, ok := ids(c)
	if !ok {
		return
	}
	status, err := h.service.ReconcilePayment(c, userID, familyID, c.Param("orderNo"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}
	response.OK(c, gin.H{"status": status})
}
func (h *Handler) Refund(c *gin.Context) {
	userID, familyID, ok := ids(c)
	if !ok {
		return
	}
	var req dto.AdminRefundRequest
	if c.Request.ContentLength > 0 && c.ShouldBindJSON(&req) != nil {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return
	}
	result := h.service.Refund(c, userID, familyID, c.Param("orderNo"), req.Reason)
	if result != nil {
		response.Error(c, http.StatusBadRequest, result)
		return
	}
	response.OK(c, gin.H{"status": "REFUNDING"})
}
func (h *Handler) PaymentNotify(c *gin.Context) {
	if h.provider == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "PAYMENT_DISABLED"})
		return
	}
	content, err := h.provider.ParsePaymentNotification(c, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_NOTIFY"})
		return
	}
	if content.TradeState != "SUCCESS" || content.Amount.Total <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS"})
		return
	}
	if content.AppID != h.appID || content.MchID != h.mchID || content.Amount.PayerTotal != content.Amount.Total {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_NOTIFY"})
		return
	}
	if err := h.service.MarkPaid(c, content.OutTradeNo, content.TransactionID, content.Amount.Total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}
func (h *Handler) RefundNotify(c *gin.Context) {
	if h.provider == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "PAYMENT_DISABLED"})
		return
	}
	content, err := h.provider.ParseRefundNotification(c, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_NOTIFY"})
		return
	}
	if err := h.service.ConfirmRefund(c, content.OutRefundNo, content.RefundID, content.RefundStatus, content.TransactionID, content.Amount.Total, content.Amount.Refund); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}
func ids(c *gin.Context) (uint64, uint64, bool) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		response.Abort(c, http.StatusUnauthorized, apperrors.CodeUnauthorized)
		return 0, 0, false
	}
	familyID, err := strconv.ParseUint(c.Param("familyId"), 10, 64)
	if err != nil || familyID == 0 {
		response.Abort(c, http.StatusBadRequest, apperrors.CodeInvalidParams)
		return 0, 0, false
	}
	return userID, familyID, true
}

func isRefundReviewer(c *gin.Context) bool {
	role, err := middleware.CurrentAdminRole(c)
	return err == nil && (role == "ROOT_ADMIN" || role == "SUPER_ADMIN")
}
