package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	accountquotaservice "tree/backend/internal/accountquota/service"
	authrepo "tree/backend/internal/auth/repository"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/permission"
	"tree/backend/internal/family/capacity/dto"
	"tree/backend/internal/family/capacity/model"
	capacityrepo "tree/backend/internal/family/capacity/repository"
	notificationservice "tree/backend/internal/notification/service"
	operationlog "tree/backend/internal/operationlog/service"
)

type PaymentProvider interface {
	CreateJSAPI(ctx context.Context, order *model.Order, openID string) (dto.PaymentParams, error)
	QueryPayment(ctx context.Context, orderNo string) (dto.PaymentQueryResult, error)
	Refund(ctx context.Context, order *model.Order, refund *model.Refund) error
	QueryRefund(ctx context.Context, refundNo string) (dto.RefundQueryResult, error)
}

type Service interface {
	ListProducts(context.Context) ([]dto.ProductVO, *apperrors.BusinessError)
	ListOrders(context.Context, uint64, uint64) ([]*dto.OrderVO, *apperrors.BusinessError)
	ListOrdersPage(context.Context, uint64, uint64, int, int, string, *time.Time, *time.Time) (*dto.OrderPage, *apperrors.BusinessError)
	AdminListProducts(context.Context) ([]dto.ProductVO, *apperrors.BusinessError)
	AdminUpdateProduct(context.Context, uint64, string, dto.UpdateProductRequest) (*dto.ProductVO, *apperrors.BusinessError)
	AdminListOrders(context.Context, dto.AdminOrderQuery) (*dto.AdminOrderPage, *apperrors.BusinessError)
	CreateOrder(context.Context, uint64, uint64, string, string) (*dto.OrderVO, *apperrors.BusinessError)
	ResumePayment(context.Context, uint64, uint64, string) (*dto.OrderVO, *apperrors.BusinessError)
	MarkPaid(context.Context, string, string, int64) *apperrors.BusinessError
	ReconcilePayment(context.Context, uint64, uint64, string) (string, *apperrors.BusinessError)
	AdminReconcilePayment(context.Context, uint64, string) (string, *apperrors.BusinessError)
	ConfirmRefund(context.Context, string, string, string, string, int64, int64) *apperrors.BusinessError
	Refund(context.Context, uint64, uint64, string, string) *apperrors.BusinessError
	AdminApproveRefund(context.Context, uint64, string, string) *apperrors.BusinessError
	AdminRejectRefund(context.Context, uint64, string, string) *apperrors.BusinessError
	AdminReconcileRefund(context.Context, uint64, string) (string, *apperrors.BusinessError)
	AdminRetryRefund(context.Context, uint64, string) (string, *apperrors.BusinessError)
	StartOrderExpiryWorker(context.Context)
}

func (s *service) StartOrderExpiryWorker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			if err := s.expirePendingOrders(ctx); err != nil && s.logger != nil {
				s.logger.Warn("expire capacity orders failed", zap.Error(err))
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

type service struct {
	db            *gorm.DB
	repo          capacityrepo.Repository
	permission    permission.FamilyPermissionService
	provider      PaymentProvider
	logs          operationlog.Service
	users         authrepo.UserRepository
	quota         accountquotaservice.Service
	appID         string
	logger        *zap.Logger
	orderTimeout  time.Duration
	notifications notificationservice.Service
}

func New(db *gorm.DB, repo capacityrepo.Repository, perms permission.FamilyPermissionService, provider PaymentProvider, logs operationlog.Service, users authrepo.UserRepository, quota accountquotaservice.Service, appID string, logger *zap.Logger, orderTimeoutMinutes int) Service {
	if orderTimeoutMinutes <= 0 {
		orderTimeoutMinutes = 30
	}
	return &service{db: db, repo: repo, permission: perms, provider: provider, logs: logs, users: users, quota: quota, appID: appID, logger: logger, orderTimeout: time.Duration(orderTimeoutMinutes) * time.Minute}
}

func NewWithNotifications(db *gorm.DB, repo capacityrepo.Repository, perms permission.FamilyPermissionService, provider PaymentProvider, logs operationlog.Service, users authrepo.UserRepository, quota accountquotaservice.Service, appID string, logger *zap.Logger, orderTimeoutMinutes int, notifications notificationservice.Service) Service {
	result := New(db, repo, perms, provider, logs, users, quota, appID, logger, orderTimeoutMinutes)
	result.(*service).notifications = notifications
	return result
}

func (s *service) ListProducts(ctx context.Context) ([]dto.ProductVO, *apperrors.BusinessError) {
	rows, err := s.repo.ListActiveProducts(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := make([]dto.ProductVO, 0, len(rows))
	for _, row := range rows {
		result = append(result, dto.ProductVO{ProductKey: row.ProductKey, DisplayName: row.DisplayName, AddedMembers: row.AddedMembers, AmountFen: row.AmountFen, RefundWindowMinutes: row.RefundWindowMinutes, Status: row.Status})
	}
	return result, nil
}

func (s *service) AdminListProducts(ctx context.Context) ([]dto.ProductVO, *apperrors.BusinessError) {
	rows, err := s.repo.ListAllProducts(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := make([]dto.ProductVO, 0, len(rows))
	for _, row := range rows {
		result = append(result, dto.ProductVO{ProductKey: row.ProductKey, DisplayName: row.DisplayName, AddedMembers: row.AddedMembers, AmountFen: row.AmountFen, RefundWindowMinutes: row.RefundWindowMinutes, Status: row.Status})
	}
	return result, nil
}

func (s *service) AdminUpdateProduct(ctx context.Context, adminID uint64, productKey string, req dto.UpdateProductRequest) (*dto.ProductVO, *apperrors.BusinessError) {
	if adminID == 0 || strings.TrimSpace(productKey) == "" {
		return nil, apperrors.New(apperrors.CodeInvalidParams)
	}
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.DisplayName == "" || len([]rune(req.DisplayName)) > 80 || req.AddedMembers < 1 || req.AddedMembers > 10000 || req.AmountFen < 1 || req.AmountFen > 99999999 || req.RefundWindowMinutes < 1 || req.RefundWindowMinutes > 43200 || (req.Status != model.ProductStatusActive && req.Status != model.ProductStatusInactive) {
		return nil, apperrors.New(apperrors.CodeInvalidParams)
	}
	var product *model.Product
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithDB(tx)
		row, err := repo.FindProductByKey(ctx, productKey, true)
		if err != nil {
			return err
		}
		if err := repo.UpdateProduct(ctx, productKey, map[string]any{
			"display_name":          req.DisplayName,
			"added_members":         req.AddedMembers,
			"amount_fen":            req.AmountFen,
			"refund_window_minutes": req.RefundWindowMinutes,
			"status":                req.Status,
		}); err != nil {
			return err
		}
		row.DisplayName = req.DisplayName
		row.AddedMembers = req.AddedMembers
		row.AmountFen = req.AmountFen
		row.RefundWindowMinutes = req.RefundWindowMinutes
		row.Status = req.Status
		product = row
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeResourceNotFound)
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if s.logs != nil {
		productID := product.ID
		_ = s.logs.WriteSuccess(ctx, operationlog.WriteInput{
			OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID,
			Module: "FAMILY_CAPACITY", Action: "UPDATE_CAPACITY_PRODUCT", TargetType: strptr("FAMILY_CAPACITY_PRODUCT"), TargetID: &productID,
			DetailJSON: []byte(fmt.Sprintf(`{"productKey":%q,"addedMembers":%d,"amountFen":%d,"refundWindowMinutes":%d}`, product.ProductKey, product.AddedMembers, product.AmountFen, product.RefundWindowMinutes)),
		})
	}
	return &dto.ProductVO{ProductKey: product.ProductKey, DisplayName: product.DisplayName, AddedMembers: product.AddedMembers, AmountFen: product.AmountFen, RefundWindowMinutes: product.RefundWindowMinutes, Status: product.Status}, nil
}

func (s *service) AdminListOrders(ctx context.Context, query dto.AdminOrderQuery) (*dto.AdminOrderPage, *apperrors.BusinessError) {
	_ = s.expirePendingOrders(ctx)
	rows, total, err := s.repo.ListAllOrders(ctx, query.Page, query.PageSize, query.Status, query.StartDate, query.EndDate)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	orderIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		orderIDs = append(orderIDs, row.ID)
	}
	refundRows, err := s.repo.ListRefundsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	latestRefunds := make(map[uint64]model.Refund, len(orderIDs))
	for _, refund := range refundRows {
		if _, exists := latestRefunds[refund.OrderID]; !exists {
			latestRefunds[refund.OrderID] = refund
		}
	}
	result := make([]dto.AdminOrderVO, 0, len(rows))
	for _, row := range rows {
		item := dto.AdminOrderVO{OrderNo: row.OrderNo, FamilyID: row.FamilyID, BuyerUserID: row.BuyerUserID, ProductName: row.ProductName, AddedMembers: row.AddedMembers, AmountFen: row.AmountFen, Status: row.Status, CreatedAt: row.CreatedAt.Format(time.RFC3339)}
		item.ExpiresAt = row.CreatedAt.Add(s.orderTimeout).Format(time.RFC3339)
		if row.WechatTransactionID != nil {
			item.TransactionID = *row.WechatTransactionID
		}
		if row.PaidAt != nil {
			item.PaidAt = row.PaidAt.Format(time.RFC3339)
		}
		if row.ClosedAt != nil {
			item.ClosedAt = row.ClosedAt.Format(time.RFC3339)
		}
		if refund, exists := latestRefunds[row.ID]; exists {
			item.RefundNo = refund.RefundNo
			item.RefundStatus = refund.Status
			item.RefundAttemptNo = refund.AttemptNo
			item.RefundAttemptsRemaining = maxRefundAttemptsRemaining(refund.AttemptNo, refund.Status)
			item.RefundRequestedAt = refund.RequestedAt.Format(time.RFC3339)
			if refund.ReviewReason != nil {
				item.RefundReason = *refund.ReviewReason
			}
		}
		result = append(result, item)
	}
	return &dto.AdminOrderPage{Items: result, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func (s *service) ListOrders(ctx context.Context, userID, familyID uint64) ([]*dto.OrderVO, *apperrors.BusinessError) {
	link, err := s.permission.GetActiveLink(ctx, userID, familyID)
	if err != nil || link.FamilyRole != string(enums.FamilyRoleFounder) {
		return nil, apperrors.New(apperrors.CodeCapacityPurchaseForbidden)
	}
	_ = s.expirePendingOrders(ctx)
	rows, err := s.repo.ListOrders(ctx, familyID, userID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	orderIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		orderIDs = append(orderIDs, row.ID)
	}
	refundRows, err := s.repo.ListRefundsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	latestRefunds := make(map[uint64]model.Refund, len(orderIDs))
	for _, refund := range refundRows {
		if _, exists := latestRefunds[refund.OrderID]; !exists {
			latestRefunds[refund.OrderID] = refund
		}
	}
	return s.toOrderVOs(rows, latestRefunds), nil
}

func (s *service) ListOrdersPage(ctx context.Context, userID, familyID uint64, page, pageSize int, status string, startDate, endDate *time.Time) (*dto.OrderPage, *apperrors.BusinessError) {
	link, err := s.permission.GetActiveLink(ctx, userID, familyID)
	if err != nil || link.FamilyRole != string(enums.FamilyRoleFounder) {
		return nil, apperrors.New(apperrors.CodeCapacityPurchaseForbidden)
	}
	_ = s.expirePendingOrders(ctx)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	rows, total, err := s.repo.ListOrdersPage(ctx, familyID, userID, page, pageSize, status, startDate, endDate)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	orderIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		orderIDs = append(orderIDs, row.ID)
	}
	refundRows, err := s.repo.ListRefundsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	latestRefunds := make(map[uint64]model.Refund, len(orderIDs))
	for _, refund := range refundRows {
		if _, exists := latestRefunds[refund.OrderID]; !exists {
			latestRefunds[refund.OrderID] = refund
		}
	}
	return &dto.OrderPage{Items: s.toOrderVOs(rows, latestRefunds), Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *service) toOrderVOs(rows []model.Order, latestRefunds map[uint64]model.Refund) []*dto.OrderVO {
	result := make([]*dto.OrderVO, 0, len(rows))
	for _, row := range rows {
		item := orderVO(&row, nil, s.orderTimeout)
		if refund, exists := latestRefunds[row.ID]; exists {
			item.RefundAttemptCount = refund.AttemptNo
			item.RefundAttemptsRemaining = maxRefundAttemptsRemaining(refund.AttemptNo, refund.Status)
			item.LastRefundStatus = refund.Status
			if refund.ReviewReason != nil {
				item.LastRefundReason = *refund.ReviewReason
			}
		}
		result = append(result, item)
	}
	return result
}

func (s *service) CreateOrder(ctx context.Context, userID, familyID uint64, productKey, openID string) (*dto.OrderVO, *apperrors.BusinessError) {
	_ = s.expirePendingOrders(ctx)
	if openID == "" && s.users != nil {
		var err error
		openID, err = s.users.FindActiveWechatMiniOpenID(ctx, userID, s.appID)
		if err != nil {
			return nil, apperrors.New(apperrors.CodeWechatPayConfigError)
		}
	}
	link, err := s.permission.GetActiveLink(ctx, userID, familyID)
	if err != nil || link.FamilyRole != string(enums.FamilyRoleFounder) {
		return nil, apperrors.New(apperrors.CodeCapacityPurchaseForbidden)
	}
	product, err := s.repo.FindActiveProduct(ctx, productKey)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeCapacityProductNotFound)
	}
	if s.provider == nil {
		return nil, apperrors.New(apperrors.CodeWechatPayConfigError)
	}
	order := &model.Order{OrderNo: newNumber("T"), FamilyID: familyID, BuyerUserID: userID, ProductID: product.ID, ProductName: product.DisplayName, AddedMembers: product.AddedMembers, AmountFen: product.AmountFen, RefundWindowMinutes: product.RefundWindowMinutes, Status: model.OrderStatusPending}
	var params dto.PaymentParams
	err = s.repo.CreateOrder(ctx, order)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	// Do not hold a database transaction while waiting on WeChat. A failed
	// prepay call leaves a harmless PENDING order that can be closed later.
	params, err = s.provider.CreateJSAPI(ctx, order, openID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("wechat jsapi prepay failed", zap.Error(err), zap.String("orderNo", order.OrderNo), zap.Uint64("userId", userID), zap.Uint64("familyId", familyID), zap.String("productKey", productKey))
		}
		return nil, apperrors.New(apperrors.CodeWechatPayConfigError)
	}
	result := orderVO(order, nil, s.orderTimeout)
	result.Payment = &params
	return result, nil
}

// ResumePayment returns fresh JSAPI parameters for the existing merchant order.
// It never creates another order, so a buyer can safely retry after leaving WeChat.
func (s *service) ResumePayment(ctx context.Context, userID, familyID uint64, orderNo string) (*dto.OrderVO, *apperrors.BusinessError) {
	_ = s.expirePendingOrders(ctx)
	link, err := s.permission.GetActiveLink(ctx, userID, familyID)
	if err != nil || link.FamilyRole != string(enums.FamilyRoleFounder) {
		return nil, apperrors.New(apperrors.CodeCapacityPurchaseForbidden)
	}
	order, err := s.repo.FindOrder(ctx, strings.TrimSpace(orderNo), false)
	if errors.Is(err, gorm.ErrRecordNotFound) || order.FamilyID != familyID || order.BuyerUserID != userID {
		return nil, apperrors.New(apperrors.CodeCapacityOrderNotFound)
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if order.Status != model.OrderStatusPending || (s.orderTimeout > 0 && !order.CreatedAt.Add(s.orderTimeout).After(time.Now())) {
		return nil, apperrors.New(apperrors.CodeCapacityOrderNotPayable)
	}
	if s.provider == nil || s.users == nil {
		return nil, apperrors.New(apperrors.CodeWechatPayConfigError)
	}
	openID, err := s.users.FindActiveWechatMiniOpenID(ctx, userID, s.appID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeWechatPayConfigError)
	}
	params, err := s.provider.CreateJSAPI(ctx, order, openID)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("wechat jsapi resume failed", zap.Error(err), zap.String("orderNo", order.OrderNo), zap.Uint64("userId", userID))
		}
		return nil, apperrors.New(apperrors.CodeWechatPayConfigError)
	}
	return orderVO(order, &params, s.orderTimeout), nil
}

func (s *service) MarkPaid(ctx context.Context, orderNo, transactionID string, amountFen int64) *apperrors.BusinessError {
	if orderNo == "" || transactionID == "" {
		return apperrors.New(apperrors.CodeInvalidParams)
	}
	_ = s.expirePendingOrders(ctx)
	var paidOrder *model.Order
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithDB(tx)
		order, err := repo.FindOrder(ctx, orderNo, true)
		if err != nil {
			return err
		}
		if order.Status == model.OrderStatusPaid {
			if amountFen != int64(order.AmountFen) || order.WechatTransactionID == nil || *order.WechatTransactionID != transactionID {
				return errors.New("capacity payment notification mismatch")
			}
			return nil
		}
		if order.Status != model.OrderStatusPending && order.Status != model.OrderStatusCancelled {
			return errors.New("capacity order is not payable")
		}
		// A signed WeChat callback or a signed WeChat order-query result is the
		// payment source of truth. A local timeout must never discard a real
		// payment that completed just before or after the expiry worker ran.
		if amountFen != int64(order.AmountFen) {
			return errors.New("capacity payment amount mismatch")
		}
		if err := repo.MarkPaid(ctx, order.ID, transactionID, time.Now()); err != nil {
			return err
		}
		if err := repo.CreateEntitlement(ctx, &model.Entitlement{FamilyID: order.FamilyID, OrderID: order.ID, AddedMembers: order.AddedMembers, Status: model.EntitlementActive}); err != nil {
			return err
		}
		paidOrder = order
		return nil
	})
	if err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	if paidOrder != nil && s.notifications != nil {
		paidAt := paidOrder.CreatedAt
		if paidOrder.PaidAt != nil {
			paidAt = *paidOrder.PaidAt
		}
		s.notifications.Create(ctx, notificationservice.Input{UserID: paidOrder.BuyerUserID, Category: "PAYMENT", Title: "扩容支付已到账", Content: fmt.Sprintf("%s 支付成功，已增加 %d 位家庭成员容量。", paidOrder.ProductName, paidOrder.AddedMembers), Route: notificationservice.CapacityRoute(paidOrder.FamilyID, paidOrder.OrderNo), RelatedType: "FAMILY_CAPACITY_ORDER", RelatedID: paidOrder.ID, Subscribe: &notificationservice.SubscribeInput{Kind: "PAYMENT", Page: notificationservice.CapacityPageRoute(paidOrder.FamilyID, paidOrder.OrderNo), Values: []string{paidOrder.ProductName, paidOrder.OrderNo, paidOrder.OrderNo, fmt.Sprintf("%.2f元", float64(paidOrder.AmountFen)/100), "家庭创建人", paidAt.Format("2006-01-02 15:04")}}})
	}
	return nil
}

func (s *service) ReconcilePayment(ctx context.Context, userID, familyID uint64, orderNo string) (string, *apperrors.BusinessError) {
	link, err := s.permission.GetActiveLink(ctx, userID, familyID)
	if err != nil || link.FamilyRole != string(enums.FamilyRoleFounder) {
		return "", apperrors.New(apperrors.CodeCapacityPurchaseForbidden)
	}
	return s.reconcilePayment(ctx, userID, 0, familyID, orderNo)
}

func (s *service) AdminReconcilePayment(ctx context.Context, adminID uint64, orderNo string) (string, *apperrors.BusinessError) {
	if adminID == 0 {
		return "", apperrors.New(apperrors.CodeInvalidParams)
	}
	return s.reconcilePayment(ctx, 0, adminID, 0, orderNo)
}

func (s *service) reconcilePayment(ctx context.Context, userID, adminID, expectedFamilyID uint64, orderNo string) (string, *apperrors.BusinessError) {
	if strings.TrimSpace(orderNo) == "" || s.provider == nil {
		return "", apperrors.New(apperrors.CodeInvalidParams)
	}
	order, err := s.repo.FindOrder(ctx, orderNo, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", apperrors.New(apperrors.CodeCapacityOrderNotFound)
		}
		return "", apperrors.New(apperrors.CodeSystemError)
	}
	if expectedFamilyID != 0 && (order.FamilyID != expectedFamilyID || order.BuyerUserID != userID) {
		return "", apperrors.New(apperrors.CodeCapacityPurchaseForbidden)
	}
	if order.Status == model.OrderStatusPaid || order.Status == model.OrderStatusRefunding || order.Status == model.OrderStatusRefunded {
		return order.Status, nil
	}
	if order.Status != model.OrderStatusPending && order.Status != model.OrderStatusCancelled {
		return "", apperrors.New(apperrors.CodeCapacityOrderNotFound)
	}
	result, err := s.provider.QueryPayment(ctx, orderNo)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("wechat payment query failed", zap.Error(err), zap.String("orderNo", orderNo))
		}
		return "", apperrors.New(apperrors.CodeWechatPayConfigError)
	}
	if result.Status != "SUCCESS" {
		return result.Status, nil
	}
	if result.OrderNo != order.OrderNo || result.TransactionID == "" || result.AppID != s.appID || result.TotalAmount != int64(order.AmountFen) || result.PayerTotal != result.TotalAmount {
		if s.logger != nil {
			s.logger.Warn("wechat payment query mismatch", zap.String("orderNo", orderNo), zap.String("wechatOrderNo", result.OrderNo), zap.String("wechatStatus", result.Status))
		}
		return "", apperrors.New(apperrors.CodeSystemError)
	}
	if businessErr := s.MarkPaid(ctx, orderNo, result.TransactionID, result.TotalAmount); businessErr != nil {
		return "", businessErr
	}
	if s.logs != nil {
		orderID := order.ID
		input := operationlog.WriteInput{Module: "FAMILY_CAPACITY", Action: "RECONCILE_CAPACITY_PAYMENT", TargetType: strptr("FAMILY_CAPACITY_ORDER"), TargetID: &orderID, FamilyID: &order.FamilyID, DetailJSON: []byte(fmt.Sprintf(`{"orderNo":%q,"wechatStatus":%q}`, orderNo, result.Status))}
		if adminID != 0 {
			input.OperatorType = string(enums.OperatorTypeAdmin)
			input.OperatorAdminID = &adminID
		} else {
			input.OperatorType = string(enums.OperatorTypeUser)
			input.OperatorUserID = &userID
		}
		if err := s.logs.WriteSuccess(ctx, input); err != nil && s.logger != nil {
			s.logger.Warn("capacity payment reconciliation audit failed", zap.Error(err), zap.String("orderNo", orderNo))
		}
	}
	return model.OrderStatusPaid, nil
}

func (s *service) Refund(ctx context.Context, userID, familyID uint64, orderNo, reason string) *apperrors.BusinessError {
	link, err := s.permission.GetActiveLink(ctx, userID, familyID)
	if err != nil || link.FamilyRole != string(enums.FamilyRoleFounder) {
		return apperrors.New(apperrors.CodeCapacityPurchaseForbidden)
	}
	if s.quota == nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	var order *model.Order
	var refund *model.Refund
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithDB(tx)
		if err := repo.LockFamily(ctx, familyID); err != nil {
			return err
		}
		order, err = repo.FindOrder(ctx, orderNo, true)
		if err != nil {
			return err
		}
		if order.FamilyID != familyID || order.BuyerUserID != userID || order.Status != model.OrderStatusPaid || order.PaidAt == nil {
			return errors.New("refund unavailable")
		}
		attemptNo := 1
		latestRefund, latestErr := repo.FindRefundByOrderID(ctx, order.ID, true)
		if latestErr == nil {
			if latestRefund.Status == model.RefundStatusPending || latestRefund.Status == model.RefundStatusProcessing || latestRefund.Status == model.RefundStatusSuccess {
				return errors.New("refund unavailable")
			}
			if latestRefund.AttemptNo >= model.RefundMaxAttempts {
				return errors.New("refund attempt limit reached")
			}
			attemptNo = latestRefund.AttemptNo + 1
		} else if !errors.Is(latestErr, gorm.ErrRecordNotFound) {
			return latestErr
		}
		reason = strings.TrimSpace(reason)
		if attemptNo > 1 && reason == "" {
			return errors.New("refund reason required")
		}
		if len([]rune(reason)) > 500 {
			return errors.New("refund reason too long")
		}
		if time.Now().After(order.PaidAt.Add(time.Duration(order.RefundWindowMinutes) * time.Minute)) {
			return errors.New("refund window expired")
		}
		entitlement, e := repo.FindEntitlement(ctx, order.ID, true)
		if e != nil || entitlement.Status != model.EntitlementActive || entitlement.AddedMembers != order.AddedMembers {
			return errors.New("capacity entitlement unavailable")
		}
		quotaReader, ok := s.quota.(accountquotaservice.FamilyCapacityReader)
		if !ok {
			return errors.New("family capacity reader unavailable")
		}
		capacity, e := quotaReader.GetFamilyCapacity(ctx, tx, familyID)
		if e != nil {
			return e
		}
		if capacity.UsedMembers > capacity.EffectiveLimit-order.AddedMembers {
			return errors.New("refund would exceed capacity")
		}
		if err := repo.MarkRefunding(ctx, order.ID); err != nil {
			return err
		}
		if err := repo.MarkEntitlementRefunding(ctx, order.ID); err != nil {
			return err
		}
		refund = &model.Refund{RefundNo: newNumber("R"), OrderID: order.ID, AttemptNo: attemptNo, FamilyID: familyID, OperatorUserID: userID, AmountFen: order.AmountFen, Status: model.RefundStatusPending, RequestedAt: time.Now(), RequestReason: optionalString(reason)}
		return repo.CreateRefund(ctx, refund)
	})
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("capacity refund request rejected",
				zap.Error(err),
				zap.Uint64("userId", userID),
				zap.Uint64("familyId", familyID),
				zap.String("orderNo", orderNo),
			)
		}
		return mapRefundError(err)
	}
	if s.logs != nil {
		refundID := refund.ID
		_ = s.logs.WriteSuccess(ctx, operationlog.WriteInput{OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &userID, Module: "FAMILY_CAPACITY", Action: "REQUEST_CAPACITY_REFUND", TargetType: strptr("FAMILY_CAPACITY_REFUND"), TargetID: &refundID, FamilyID: &familyID, DetailJSON: []byte(fmt.Sprintf(`{"refundNo":%q}`, refund.RefundNo))})
	}
	return nil
}

func (s *service) AdminApproveRefund(ctx context.Context, adminID uint64, refundNo, reason string) *apperrors.BusinessError {
	if s.provider == nil || adminID == 0 {
		return apperrors.New(apperrors.CodeWechatPayConfigError)
	}
	reason = strings.TrimSpace(reason)
	var order *model.Order
	var refund *model.Refund
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithDB(tx)
		refundRow, err := repo.FindRefund(ctx, refundNo, true)
		if err != nil {
			return err
		}
		if refundRow.Status != model.RefundStatusPending {
			return errors.New("refund is not awaiting review")
		}
		orderRow, err := repo.FindOrderByID(ctx, refundRow.OrderID, true)
		if err != nil || orderRow.Status != model.OrderStatusRefunding {
			return errors.New("refund order is not awaiting review")
		}
		if err := repo.MarkRefundProcessing(ctx, refundRow.ID, adminID, reason, time.Now()); err != nil {
			return err
		}
		order, refund = orderRow, refundRow
		return nil
	})
	if err != nil {
		return apperrors.New(apperrors.CodeCapacityRefundUnavailable)
	}
	if err := s.provider.Refund(ctx, order, refund); err != nil {
		if s.logger != nil {
			s.logger.Warn("wechat refund submission needs manual verification", zap.Error(err), zap.String("refundNo", refund.RefundNo))
		}
		if markErr := s.markRefundReviewRequired(ctx, refund.RefundNo); markErr != nil {
			return apperrors.New(apperrors.CodeSystemError)
		}
		return apperrors.New(apperrors.CodeSystemError)
	}
	if s.logs != nil {
		refundID := refund.ID
		_ = s.logs.WriteSuccess(ctx, operationlog.WriteInput{OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, Module: "FAMILY_CAPACITY", Action: "APPROVE_CAPACITY_REFUND", TargetType: strptr("FAMILY_CAPACITY_REFUND"), TargetID: &refundID, FamilyID: &refund.FamilyID, DetailJSON: []byte(fmt.Sprintf(`{"refundNo":%q,"reason":%q}`, refund.RefundNo, reason))})
	}
	if s.notifications != nil {
		s.notifications.Create(ctx, notificationservice.Input{UserID: order.BuyerUserID, Category: "REFUND", Title: "退款审核已通过", Content: "退款已提交至微信支付处理，扩容名额将在退款结果确认前保持冻结。", Route: notificationservice.CapacityRoute(order.FamilyID, order.OrderNo), RelatedType: "FAMILY_CAPACITY_REFUND", RelatedID: refund.ID, Subscribe: &notificationservice.SubscribeInput{Kind: "REFUND", Page: notificationservice.CapacityPageRoute(order.FamilyID, order.OrderNo), Values: []string{refund.RefundNo, fmt.Sprintf("%.2f元", float64(refund.AmountFen)/100), refund.RequestedAt.Format("2006-01-02 15:04"), "审核通过", "家庭扩容退款"}}})
	}
	return nil
}

func (s *service) AdminRejectRefund(ctx context.Context, adminID uint64, refundNo, reason string) *apperrors.BusinessError {
	if adminID == 0 {
		return apperrors.New(apperrors.CodeInvalidParams)
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "后台审核未通过"
	}
	var refund *model.Refund
	var order *model.Order
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithDB(tx)
		refundRow, err := repo.FindRefund(ctx, refundNo, true)
		if err != nil {
			return err
		}
		if refundRow.Status != model.RefundStatusPending {
			return errors.New("refund is not awaiting review")
		}
		orderRow, err := repo.FindOrderByID(ctx, refundRow.OrderID, true)
		if err != nil || orderRow.Status != model.OrderStatusRefunding {
			return errors.New("refund order is not awaiting review")
		}
		if err := repo.MarkRefundRejected(ctx, refundRow.ID, adminID, reason, time.Now()); err != nil {
			return err
		}
		if err := repo.MarkOrderPaid(ctx, orderRow.ID); err != nil {
			return err
		}
		if err := repo.RestoreEntitlement(ctx, orderRow.ID); err != nil {
			return err
		}
		refund, order = refundRow, orderRow
		return nil
	})
	if err != nil {
		return apperrors.New(apperrors.CodeCapacityRefundUnavailable)
	}
	if s.logs != nil {
		refundID := refund.ID
		_ = s.logs.WriteSuccess(ctx, operationlog.WriteInput{OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, Module: "FAMILY_CAPACITY", Action: "REJECT_CAPACITY_REFUND", TargetType: strptr("FAMILY_CAPACITY_REFUND"), TargetID: &refundID, FamilyID: &refund.FamilyID, DetailJSON: []byte(fmt.Sprintf(`{"refundNo":%q,"reason":%q}`, refund.RefundNo, reason))})
	}
	if s.notifications != nil {
		s.notifications.Create(ctx, notificationservice.Input{UserID: order.BuyerUserID, Category: "REFUND", Title: "退款申请未通过", Content: fmt.Sprintf("退款申请未通过：%s。扩容名额已恢复使用。", reason), Route: notificationservice.CapacityRoute(order.FamilyID, order.OrderNo), RelatedType: "FAMILY_CAPACITY_REFUND", RelatedID: refund.ID, Subscribe: &notificationservice.SubscribeInput{Kind: "REFUND", Page: notificationservice.CapacityPageRoute(order.FamilyID, order.OrderNo), Values: []string{refund.RefundNo, fmt.Sprintf("%.2f元", float64(refund.AmountFen)/100), refund.RequestedAt.Format("2006-01-02 15:04"), "审核拒绝", "家庭扩容退款"}}})
	}
	return nil
}

func (s *service) AdminReconcileRefund(ctx context.Context, adminID uint64, refundNo string) (string, *apperrors.BusinessError) {
	if adminID == 0 || refundNo == "" || s.provider == nil {
		return "", apperrors.New(apperrors.CodeInvalidParams)
	}
	refund, err := s.repo.FindRefund(ctx, refundNo, false)
	if err != nil || (refund.Status != model.RefundStatusReview && refund.Status != model.RefundStatusProcessing && refund.Status != model.RefundStatusFailed) {
		return "", apperrors.New(apperrors.CodeCapacityRefundUnavailable)
	}
	order, err := s.repo.FindOrderByID(ctx, refund.OrderID, false)
	if err != nil {
		return "", apperrors.New(apperrors.CodeSystemError)
	}
	result, err := s.provider.QueryRefund(ctx, refundNo)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("wechat refund query failed", zap.Error(err), zap.String("refundNo", refundNo))
		}
		return "", apperrors.New(apperrors.CodeCapacityRefundQueryFailed)
	}
	if result.Status == "NOT_FOUND" {
		return "NOT_FOUND", nil
	}
	if result.RefundNo != refundNo || result.OrderNo != order.OrderNo || !refundNotificationMatches(refund, order, result.RefundID, result.TransactionID, result.TotalAmount, result.RefundAmount) {
		return "", apperrors.New(apperrors.CodeCapacityRefundMismatch)
	}
	if businessErr := s.ConfirmRefund(ctx, refundNo, result.RefundID, result.Status, result.TransactionID, result.TotalAmount, result.RefundAmount); businessErr != nil {
		return "", apperrors.New(apperrors.CodeCapacityRefundStateConflict)
	}
	current, err := s.repo.FindRefund(ctx, refundNo, false)
	if err != nil {
		return "", apperrors.New(apperrors.CodeSystemError)
	}
	if s.logs != nil {
		refundID := refund.ID
		if err := s.logs.WriteSuccess(ctx, operationlog.WriteInput{OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, Module: "FAMILY_CAPACITY", Action: "RECONCILE_CAPACITY_REFUND", TargetType: strptr("FAMILY_CAPACITY_REFUND"), TargetID: &refundID, FamilyID: &refund.FamilyID, DetailJSON: []byte(fmt.Sprintf(`{"refundNo":%q,"wechatStatus":%q}`, refundNo, result.Status))}); err != nil && s.logger != nil {
			s.logger.Warn("capacity refund reconciliation audit failed", zap.Error(err), zap.String("refundNo", refundNo))
		}
	}
	return current.Status, nil
}

func (s *service) AdminRetryRefund(ctx context.Context, adminID uint64, refundNo string) (string, *apperrors.BusinessError) {
	if adminID == 0 || refundNo == "" || s.provider == nil {
		return "", apperrors.New(apperrors.CodeInvalidParams)
	}
	refund, err := s.repo.FindRefund(ctx, refundNo, false)
	if err != nil || refund.Status != model.RefundStatusReview {
		return "", apperrors.New(apperrors.CodeCapacityRefundUnavailable)
	}
	order, err := s.repo.FindOrderByID(ctx, refund.OrderID, false)
	if err != nil || order.Status != model.OrderStatusRefunding {
		return "", apperrors.New(apperrors.CodeCapacityRefundUnavailable)
	}
	result, err := s.provider.QueryRefund(ctx, refundNo)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("wechat refund query before retry failed", zap.Error(err), zap.String("refundNo", refundNo))
		}
		return "", apperrors.New(apperrors.CodeCapacityRefundQueryFailed)
	}
	if result.Status != "NOT_FOUND" || result.RefundNo != refundNo {
		return "", apperrors.New(apperrors.CodeCapacityRefundUnavailable)
	}
	if err := s.provider.Refund(ctx, order, refund); err != nil {
		if s.logger != nil {
			s.logger.Warn("wechat refund resubmission needs verification", zap.Error(err), zap.String("refundNo", refundNo))
		}
		return "", apperrors.New(apperrors.CodeCapacityRefundRetryUnknown)
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithDB(tx)
		current, err := repo.FindRefund(ctx, refundNo, true)
		if err != nil {
			return err
		}
		if current.Status == model.RefundStatusSuccess || current.Status == model.RefundStatusProcessing || current.Status == model.RefundStatusFailed {
			return nil
		}
		if current.Status != model.RefundStatusReview {
			return errors.New("refund state changed during retry")
		}
		return repo.MarkReviewRefundProcessing(ctx, current.ID)
	})
	if err != nil {
		return "", apperrors.New(apperrors.CodeCapacityRefundRetryUnknown)
	}
	current, err := s.repo.FindRefund(ctx, refundNo, false)
	if err != nil {
		return "", apperrors.New(apperrors.CodeCapacityRefundRetryUnknown)
	}
	if s.logs != nil {
		refundID := refund.ID
		if err := s.logs.WriteSuccess(ctx, operationlog.WriteInput{OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, Module: "FAMILY_CAPACITY", Action: "RETRY_CAPACITY_REFUND", TargetType: strptr("FAMILY_CAPACITY_REFUND"), TargetID: &refundID, FamilyID: &refund.FamilyID, DetailJSON: []byte(fmt.Sprintf(`{"refundNo":%q}`, refundNo))}); err != nil && s.logger != nil {
			s.logger.Warn("capacity refund retry audit failed", zap.Error(err), zap.String("refundNo", refundNo))
		}
	}
	return current.Status, nil
}

func (s *service) ConfirmRefund(ctx context.Context, refundNo, refundID, status, transactionID string, totalAmount, refundAmount int64) *apperrors.BusinessError {
	if status != "SUCCESS" && status != "CLOSED" && status != "ABNORMAL" && status != "PROCESSING" {
		return apperrors.New(apperrors.CodeInvalidParams)
	}
	var notifyUserID, notifyFamilyID, notifyRefundID uint64
	var notifyOrderNo string
	var notifyTitle, notifyContent, notifyStatus string
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithDB(tx)
		refund, err := repo.FindRefund(ctx, refundNo, true)
		if err != nil {
			return err
		}
		order, err := repo.FindOrderByID(ctx, refund.OrderID, false)
		if err != nil || !refundNotificationMatches(refund, order, refundID, transactionID, totalAmount, refundAmount) {
			return errors.New("refund notification does not match order")
		}
		notifyUserID, notifyFamilyID, notifyRefundID, notifyOrderNo = order.BuyerUserID, order.FamilyID, refund.ID, order.OrderNo
		if refund.Status == model.RefundStatusSuccess {
			if status == "SUCCESS" && (refund.WechatRefundID == nil || *refund.WechatRefundID != refundID) {
				return errors.New("conflicting refund notification after success")
			}
			return nil
		}
		if refund.Status == model.RefundStatusRejected {
			if status == "SUCCESS" {
				return errors.New("unsubmitted refund received a success notification")
			}
			return nil
		}
		if refund.Status == model.RefundStatusFailed {
			if status == "CLOSED" {
				return nil
			}
			latest, err := repo.FindRefundByOrderID(ctx, order.ID, true)
			if err != nil || latest.ID != refund.ID || order.Status != model.OrderStatusPaid {
				return errors.New("late refund notification conflicts with a newer attempt")
			}
			if status == "PROCESSING" || status == "ABNORMAL" {
				if err := repo.MarkRefunding(ctx, order.ID); err != nil {
					return err
				}
				if err := repo.MarkEntitlementRefunding(ctx, order.ID); err != nil {
					return err
				}
				return repo.ReopenFailedRefundReview(ctx, refund.ID)
			}
		} else if order.Status != model.OrderStatusRefunding {
			return errors.New("refund order is not refunding")
		}
		switch status {
		case "CLOSED":
			if err := repo.MarkRefundFailed(ctx, refund.ID, time.Now()); err != nil {
				return err
			}
			if err := repo.MarkOrderPaid(ctx, order.ID); err != nil {
				return err
			}
			notifyTitle, notifyContent, notifyStatus = "退款未完成", "微信支付未完成本次退款，扩容名额已恢复使用。", "退款失败"
			return repo.RestoreEntitlement(ctx, order.ID)
		case "ABNORMAL":
			notifyTitle, notifyContent, notifyStatus = "退款需要人工处理", "微信支付返回异常状态，平台将继续核查本次退款。", "退款异常"
			return repo.MarkRefundReviewRequired(ctx, refund.ID)
		case "PROCESSING":
			if refund.Status == model.RefundStatusReview {
				return nil
			}
			notifyTitle, notifyContent, notifyStatus = "退款处理中", "退款已提交微信支付，结果确认后会再次更新。", "退款处理中"
			return repo.MarkRefundNotificationProcessing(ctx, refund.ID, time.Now())
		}
		if err := repo.MarkRefunded(ctx, refund.ID, refundID, time.Now()); err != nil {
			return err
		}
		if err := repo.MarkOrderRefunded(ctx, refund.OrderID, time.Now()); err != nil {
			return err
		}
		if err := repo.RevokeEntitlement(ctx, refund.OrderID, time.Now()); err != nil {
			return err
		}
		notifyTitle, notifyContent, notifyStatus = "退款成功", "退款已成功，扩容名额已扣除。", "退款成功"
		target := refund.OrderID
		if s.logs == nil {
			return nil
		}
		return s.logs.WriteSuccess(ctx, operationlog.WriteInput{OperatorType: string(enums.OperatorTypeSystem), Module: "FAMILY_CAPACITY", Action: "REFUND_CAPACITY_ORDER", TargetType: strptr("FAMILY_CAPACITY_ORDER"), TargetID: &target, FamilyID: &refund.FamilyID, DetailJSON: []byte(fmt.Sprintf(`{"refundId":%q}`, refundID))})
	})
	if err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	if notifyUserID != 0 && notifyTitle != "" && s.notifications != nil {
		s.notifications.Create(ctx, notificationservice.Input{UserID: notifyUserID, Category: "REFUND", Title: notifyTitle, Content: notifyContent, Route: notificationservice.CapacityRoute(notifyFamilyID, notifyOrderNo), RelatedType: "FAMILY_CAPACITY_REFUND", RelatedID: notifyRefundID, Subscribe: &notificationservice.SubscribeInput{Kind: "REFUND", Page: notificationservice.CapacityPageRoute(notifyFamilyID, notifyOrderNo), Values: []string{refundNo, fmt.Sprintf("%.2f元", float64(refundAmount)/100), time.Now().Format("2006-01-02 15:04"), notifyStatus, "家庭扩容退款"}}})
	}
	return nil
}

func refundNotificationMatches(refund *model.Refund, order *model.Order, refundID, transactionID string, totalAmount, refundAmount int64) bool {
	return refund != nil && order != nil && refundID != "" && transactionID != "" &&
		order.WechatTransactionID != nil && transactionID == *order.WechatTransactionID &&
		totalAmount == int64(order.AmountFen) && refundAmount == int64(refund.AmountFen) &&
		refund.AmountFen == order.AmountFen
}

func (s *service) markRefundReviewRequired(ctx context.Context, refundNo string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithDB(tx)
		refund, err := repo.FindRefund(ctx, refundNo, true)
		if err != nil {
			return err
		}
		if refund.Status == model.RefundStatusSuccess || refund.Status == model.RefundStatusFailed {
			return nil
		}
		return repo.MarkRefundReviewRequired(ctx, refund.ID)
	})
}

func orderVO(order *model.Order, params *dto.PaymentParams, orderTimeout time.Duration) *dto.OrderVO {
	result := &dto.OrderVO{OrderNo: order.OrderNo, FamilyID: order.FamilyID, ProductName: order.ProductName, AddedMembers: order.AddedMembers, AmountFen: order.AmountFen, Status: order.Status, RefundWindowMinutes: order.RefundWindowMinutes, RefundAttemptsRemaining: model.RefundMaxAttempts, CreatedAt: order.CreatedAt.Format(time.RFC3339)}
	if orderTimeout > 0 {
		result.ExpiresAt = order.CreatedAt.Add(orderTimeout).Format(time.RFC3339)
	}
	if order.PaidAt != nil {
		result.RefundableUntil = order.PaidAt.Add(time.Duration(order.RefundWindowMinutes) * time.Minute).Format(time.RFC3339)
	}
	if order.ClosedAt != nil {
		result.ClosedAt = order.ClosedAt.Format(time.RFC3339)
	}
	result.Payment = params
	return result
}

func (s *service) expirePendingOrders(ctx context.Context) error {
	now := time.Now()
	return s.repo.ExpirePendingOrders(ctx, now, now.Add(-s.orderTimeout))
}
func newNumber(prefix string) string {
	// WeChat Pay out_trade_no accepts at most 32 characters.
	var b [8]byte
	_, _ = rand.Read(b[:])
	return prefix + time.Now().Format("20060102150405") + hex.EncodeToString(b[:])
}
func strptr(v string) *string { return &v }
func optionalString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
func maxRefundAttemptsRemaining(attemptNo int, status string) int {
	if status == model.RefundStatusSuccess || attemptNo >= model.RefundMaxAttempts {
		return 0
	}
	return model.RefundMaxAttempts - attemptNo
}
func mapRefundError(err error) *apperrors.BusinessError {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.New(apperrors.CodeCapacityOrderNotFound)
	}
	if err.Error() == "refund would exceed capacity" {
		return apperrors.New(apperrors.CodeCapacityRefundBlocked)
	}
	if err.Error() == "refund reason required" {
		return apperrors.New(apperrors.CodeCapacityRefundReasonRequired)
	}
	return apperrors.New(apperrors.CodeCapacityRefundUnavailable)
}
