package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
	"tree/backend/internal/common/enums"
	"tree/backend/internal/family/capacity/dto"
	"tree/backend/internal/family/capacity/model"
	capacityrepo "tree/backend/internal/family/capacity/repository"
	familymodel "tree/backend/internal/family/core/model"
	operationlog "tree/backend/internal/operationlog/service"
	"tree/backend/internal/testsupport/freshquota"
)

type fakeRefundProvider struct {
	query        dto.RefundQueryResult
	paymentQuery dto.PaymentQueryResult
	refundCalls  int
}

func (p *fakeRefundProvider) CreateJSAPI(context.Context, *model.Order, string) (dto.PaymentParams, error) {
	return dto.PaymentParams{}, nil
}
func (p *fakeRefundProvider) QueryPayment(context.Context, string) (dto.PaymentQueryResult, error) {
	return p.paymentQuery, nil
}
func (p *fakeRefundProvider) Refund(context.Context, *model.Order, *model.Refund) error {
	p.refundCalls++
	return nil
}
func (p *fakeRefundProvider) QueryRefund(context.Context, string) (dto.RefundQueryResult, error) {
	return p.query, nil
}

func TestAdminReconcileRefundUsesWechatResult(t *testing.T) {
	for _, merchantStatus := range []string{"SUCCESS", "CLOSED", "NOT_FOUND", "MISMATCH"} {
		t.Run(merchantStatus, func(t *testing.T) {
			tx, runID := freshquota.TestDB(t)
			if !tx.Migrator().HasTable(&model.Product{}) {
				t.Skip("local MySQL has not applied capacity migrations")
			}
			order, refund := seedRefundForReconciliation(t, tx, runID)
			provider := &fakeRefundProvider{query: dto.RefundQueryResult{
				RefundNo: refund.RefundNo, RefundID: "wx-refund-1", OrderNo: order.OrderNo,
				TransactionID: *order.WechatTransactionID, Status: merchantStatus,
				TotalAmount: 1, RefundAmount: 1,
			}}
			if merchantStatus == "MISMATCH" {
				provider.query.Status = "SUCCESS"
				provider.query.TotalAmount = 2
			}
			svc := New(tx, capacityrepo.NewRepository(tx), nil, provider, operationlog.NewGormService(tx), nil, nil, "", nil, 30)
			status, businessErr := svc.AdminReconcileRefund(context.Background(), 1, refund.RefundNo)
			switch merchantStatus {
			case "SUCCESS":
				if businessErr != nil || status != model.RefundStatusSuccess {
					t.Fatalf("reconcile success: status=%s error=%v", status, businessErr)
				}
				assertCapacityStates(t, tx, order.ID, refund.ID, model.OrderStatusRefunded, model.RefundStatusSuccess, model.EntitlementRevoked)
			case "CLOSED":
				if businessErr != nil || status != model.RefundStatusFailed {
					t.Fatalf("reconcile closed: status=%s error=%v", status, businessErr)
				}
				assertCapacityStates(t, tx, order.ID, refund.ID, model.OrderStatusPaid, model.RefundStatusFailed, model.EntitlementActive)
			case "NOT_FOUND":
				if businessErr != nil || status != "NOT_FOUND" {
					t.Fatalf("reconcile missing: status=%s error=%v", status, businessErr)
				}
				assertCapacityStates(t, tx, order.ID, refund.ID, model.OrderStatusRefunding, model.RefundStatusReview, model.EntitlementRefunding)
				status, businessErr = svc.AdminRetryRefund(context.Background(), 1, refund.RefundNo)
				if businessErr != nil || status != model.RefundStatusProcessing || provider.refundCalls != 1 {
					t.Fatalf("retry: status=%s error=%v calls=%d", status, businessErr, provider.refundCalls)
				}
				assertCapacityStates(t, tx, order.ID, refund.ID, model.OrderStatusRefunding, model.RefundStatusProcessing, model.EntitlementRefunding)
			case "MISMATCH":
				if businessErr == nil {
					t.Fatal("mismatched merchant result was accepted")
				}
				assertCapacityStates(t, tx, order.ID, refund.ID, model.OrderStatusRefunding, model.RefundStatusReview, model.EntitlementRefunding)
			}
		})
	}
}

func TestAdminReconcilePaymentUsesVerifiedWechatResult(t *testing.T) {
	tx, runID := freshquota.TestDB(t)
	if !tx.Migrator().HasTable(&model.Entitlement{}) {
		t.Skip("local MySQL has not applied capacity migrations")
	}
	user := freshquota.CreateActiveUser(t, tx, "payment-reconcile", false)
	family := &familymodel.Family{FamilyName: "支付对账家庭", FamilySurname: "张", Status: string(enums.StatusNormal), PublicDisplayStatus: string(enums.StatusPrivate), GraphVersion: 1}
	if err := tx.Create(family).Error; err != nil {
		t.Fatal(err)
	}
	product := &model.Product{ProductKey: runID, DisplayName: "+1", AddedMembers: 1, AmountFen: 1, RefundWindowMinutes: 30, Status: model.ProductStatusActive}
	if err := tx.Create(product).Error; err != nil {
		t.Fatal(err)
	}
	order := &model.Order{OrderNo: "payment-" + runID, FamilyID: family.ID, BuyerUserID: user.ID, ProductID: product.ID, ProductName: product.DisplayName, AddedMembers: 1, AmountFen: 1, RefundWindowMinutes: 30, Status: model.OrderStatusCancelled}
	if err := tx.Create(order).Error; err != nil {
		t.Fatal(err)
	}
	provider := &fakeRefundProvider{paymentQuery: dto.PaymentQueryResult{OrderNo: order.OrderNo, TransactionID: "wx-payment-" + runID, Status: "SUCCESS", TotalAmount: 1, PayerTotal: 1}}
	svc := New(tx, capacityrepo.NewRepository(tx), nil, provider, operationlog.NewGormService(tx), nil, nil, "", nil, 30)
	status, businessErr := svc.AdminReconcilePayment(context.Background(), 1, order.OrderNo)
	if businessErr != nil || status != model.OrderStatusPaid {
		t.Fatalf("payment reconciliation: status=%s error=%v", status, businessErr)
	}
	var current model.Order
	if err := tx.First(&current, order.ID).Error; err != nil {
		t.Fatal(err)
	}
	if current.Status != model.OrderStatusPaid || current.WechatTransactionID == nil || *current.WechatTransactionID != provider.paymentQuery.TransactionID {
		t.Fatalf("payment result not persisted: %#v", current)
	}
	var entitlement model.Entitlement
	if err := tx.Where("order_id = ?", order.ID).First(&entitlement).Error; err != nil {
		t.Fatal(err)
	}
	if entitlement.Status != model.EntitlementActive || entitlement.AddedMembers != 1 {
		t.Fatalf("unexpected entitlement: %#v", entitlement)
	}
}

func TestLateRefundSuccessCorrectsOldManualFailure(t *testing.T) {
	tx, runID := freshquota.TestDB(t)
	if !tx.Migrator().HasTable(&model.Product{}) {
		t.Skip("local MySQL has not applied capacity migrations")
	}
	order, refund := seedRefundForReconciliation(t, tx, runID)
	if err := tx.Model(order).Update("status", model.OrderStatusPaid).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Model(refund).Update("status", model.RefundStatusFailed).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Model(&model.Entitlement{}).Where("order_id = ?", order.ID).Update("status", model.EntitlementActive).Error; err != nil {
		t.Fatal(err)
	}
	svc := New(tx, capacityrepo.NewRepository(tx), nil, nil, operationlog.NewGormService(tx), nil, nil, "", nil, 30)
	if err := svc.ConfirmRefund(context.Background(), refund.RefundNo, "wx-refund-1", "SUCCESS", *order.WechatTransactionID, 1, 1); err != nil {
		t.Fatal(err)
	}
	assertCapacityStates(t, tx, order.ID, refund.ID, model.OrderStatusRefunded, model.RefundStatusSuccess, model.EntitlementRevoked)
}

func TestLateRefundProcessingRefreezesOldManualFailure(t *testing.T) {
	tx, runID := freshquota.TestDB(t)
	if !tx.Migrator().HasTable(&model.Product{}) {
		t.Skip("local MySQL has not applied capacity migrations")
	}
	order, refund := seedRefundForReconciliation(t, tx, runID)
	if err := tx.Model(order).Update("status", model.OrderStatusPaid).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Model(refund).Update("status", model.RefundStatusFailed).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Model(&model.Entitlement{}).Where("order_id = ?", order.ID).Update("status", model.EntitlementActive).Error; err != nil {
		t.Fatal(err)
	}
	svc := New(tx, capacityrepo.NewRepository(tx), nil, nil, operationlog.NewGormService(tx), nil, nil, "", nil, 30)
	if err := svc.ConfirmRefund(context.Background(), refund.RefundNo, "wx-refund-1", "PROCESSING", *order.WechatTransactionID, 1, 1); err != nil {
		t.Fatal(err)
	}
	assertCapacityStates(t, tx, order.ID, refund.ID, model.OrderStatusRefunding, model.RefundStatusReview, model.EntitlementRefunding)
	if err := svc.ConfirmRefund(context.Background(), refund.RefundNo, "wx-refund-1", "SUCCESS", *order.WechatTransactionID, 1, 1); err != nil {
		t.Fatal(err)
	}
	assertCapacityStates(t, tx, order.ID, refund.ID, model.OrderStatusRefunded, model.RefundStatusSuccess, model.EntitlementRevoked)
}

func TestLateRefundSuccessDoesNotOverwriteNewAttempt(t *testing.T) {
	tx, runID := freshquota.TestDB(t)
	if !tx.Migrator().HasTable(&model.Product{}) {
		t.Skip("local MySQL has not applied capacity migrations")
	}
	order, refund := seedRefundForReconciliation(t, tx, runID)
	if err := tx.Model(refund).Update("status", model.RefundStatusFailed).Error; err != nil {
		t.Fatal(err)
	}
	next := &model.Refund{RefundNo: "next-" + runID, OrderID: order.ID, AttemptNo: 2, FamilyID: refund.FamilyID, OperatorUserID: refund.OperatorUserID, AmountFen: 1, Status: model.RefundStatusPending, RequestedAt: time.Now()}
	if err := tx.Create(next).Error; err != nil {
		t.Fatal(err)
	}
	svc := New(tx, capacityrepo.NewRepository(tx), nil, nil, operationlog.NewGormService(tx), nil, nil, "", nil, 30)
	if err := svc.ConfirmRefund(context.Background(), refund.RefundNo, "wx-refund-1", "SUCCESS", *order.WechatTransactionID, 1, 1); err == nil {
		t.Fatal("older refund must not silently overwrite an active retry")
	}
	assertCapacityStates(t, tx, order.ID, refund.ID, model.OrderStatusRefunding, model.RefundStatusFailed, model.EntitlementRefunding)
}

func seedRefundForReconciliation(t *testing.T, tx *gorm.DB, runID string) (*model.Order, *model.Refund) {
	t.Helper()
	user := freshquota.CreateActiveUser(t, tx, "refund-reconcile", false)
	family := &familymodel.Family{FamilyName: "退款核对家庭", FamilySurname: "张", Status: string(enums.StatusNormal), PublicDisplayStatus: string(enums.StatusPrivate), GraphVersion: 1}
	if err := tx.Create(family).Error; err != nil {
		t.Fatal(err)
	}
	product := &model.Product{ProductKey: runID, DisplayName: "+1", AddedMembers: 1, AmountFen: 1, RefundWindowMinutes: 30, Status: model.ProductStatusActive}
	if err := tx.Create(product).Error; err != nil {
		t.Fatal(err)
	}
	transactionID := "wx-payment-" + runID
	order := &model.Order{OrderNo: "order-" + runID, FamilyID: family.ID, BuyerUserID: user.ID, ProductID: product.ID, ProductName: product.DisplayName, AddedMembers: 1, AmountFen: 1, RefundWindowMinutes: 30, Status: model.OrderStatusRefunding, WechatTransactionID: &transactionID}
	if err := tx.Create(order).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Create(&model.Entitlement{FamilyID: family.ID, OrderID: order.ID, AddedMembers: 1, Status: model.EntitlementRefunding}).Error; err != nil {
		t.Fatal(err)
	}
	refund := &model.Refund{RefundNo: "refund-" + runID, OrderID: order.ID, AttemptNo: 1, FamilyID: family.ID, OperatorUserID: user.ID, AmountFen: 1, Status: model.RefundStatusReview, RequestedAt: time.Now()}
	if err := tx.Create(refund).Error; err != nil {
		t.Fatal(err)
	}
	return order, refund
}

func TestPaidOrderEntitlementLookupIgnoresNewPendingOrder(t *testing.T) {
	tx, runID := freshquota.TestDB(t)
	if !tx.Migrator().HasTable(&model.Entitlement{}) {
		t.Skip("local MySQL has not applied capacity migrations")
	}
	paidOrder, _ := seedRefundForReconciliation(t, tx, runID)
	pending := &model.Order{
		OrderNo: "pending-" + runID, FamilyID: paidOrder.FamilyID,
		BuyerUserID: paidOrder.BuyerUserID, ProductID: paidOrder.ProductID,
		ProductName: paidOrder.ProductName, AddedMembers: 1, AmountFen: 1,
		RefundWindowMinutes: 30, Status: model.OrderStatusPending,
	}
	if err := tx.Create(pending).Error; err != nil {
		t.Fatal(err)
	}
	entitlement, err := capacityrepo.NewRepository(tx).FindEntitlement(context.Background(), paidOrder.ID, true)
	if err != nil || entitlement.OrderID != paidOrder.ID || entitlement.Status != model.EntitlementRefunding {
		t.Fatalf("paid order entitlement lookup failed: entitlement=%+v err=%v", entitlement, err)
	}
}

func TestRefundNotificationsPreserveTerminalCapacity(t *testing.T) {
	tx, runID := freshquota.TestDB(t)
	if !tx.Migrator().HasTable(&model.Product{}) {
		t.Skip("local MySQL has not applied capacity migrations")
	}
	ctx := context.Background()
	user := freshquota.CreateActiveUser(t, tx, "refund-test", false)
	family := &familymodel.Family{FamilyName: "退款测试家庭", FamilySurname: "张", Status: string(enums.StatusNormal), PublicDisplayStatus: string(enums.StatusPrivate), GraphVersion: 1}
	if err := tx.Create(family).Error; err != nil {
		t.Fatal(err)
	}
	product := &model.Product{ProductKey: runID, DisplayName: "+1", AddedMembers: 1, AmountFen: 1, RefundWindowMinutes: 30, Status: model.ProductStatusActive}
	if err := tx.Create(product).Error; err != nil {
		t.Fatal(err)
	}
	transactionID := "wx-" + runID
	order := &model.Order{OrderNo: "order-" + runID, FamilyID: family.ID, BuyerUserID: user.ID, ProductID: product.ID, ProductName: product.DisplayName, AddedMembers: 1, AmountFen: 1, RefundWindowMinutes: 30, Status: model.OrderStatusRefunding, WechatTransactionID: &transactionID}
	if err := tx.Create(order).Error; err != nil {
		t.Fatal(err)
	}
	entitlement := &model.Entitlement{FamilyID: family.ID, OrderID: order.ID, AddedMembers: 1, Status: model.EntitlementRefunding}
	if err := tx.Create(entitlement).Error; err != nil {
		t.Fatal(err)
	}
	refund := &model.Refund{RefundNo: "refund-" + runID, OrderID: order.ID, AttemptNo: 1, FamilyID: family.ID, OperatorUserID: user.ID, AmountFen: 1, Status: model.RefundStatusProcessing, RequestedAt: time.Now()}
	if err := tx.Create(refund).Error; err != nil {
		t.Fatal(err)
	}
	svc := New(tx, capacityrepo.NewRepository(tx), nil, nil, operationlog.NewGormService(tx), nil, nil, "", nil, 30)

	if err := svc.ConfirmRefund(ctx, refund.RefundNo, "wx-refund-1", "CLOSED", transactionID, 2, 1); err == nil {
		t.Fatal("wrong total must not restore entitlement")
	}
	assertCapacityStates(t, tx, order.ID, refund.ID, model.OrderStatusRefunding, model.RefundStatusProcessing, model.EntitlementRefunding)

	if err := svc.ConfirmRefund(ctx, refund.RefundNo, "wx-refund-1", "ABNORMAL", transactionID, 1, 1); err != nil {
		t.Fatal(err)
	}
	if err := svc.ConfirmRefund(ctx, refund.RefundNo, "wx-refund-1", "PROCESSING", transactionID, 1, 1); err != nil {
		t.Fatal(err)
	}
	assertCapacityStates(t, tx, order.ID, refund.ID, model.OrderStatusRefunding, model.RefundStatusReview, model.EntitlementRefunding)

	if err := svc.ConfirmRefund(ctx, refund.RefundNo, "wx-refund-1", "SUCCESS", transactionID, 1, 1); err != nil {
		t.Fatal(err)
	}
	if err := svc.ConfirmRefund(ctx, refund.RefundNo, "wx-refund-1", "CLOSED", transactionID, 1, 1); err != nil {
		t.Fatal(err)
	}
	assertCapacityStates(t, tx, order.ID, refund.ID, model.OrderStatusRefunded, model.RefundStatusSuccess, model.EntitlementRevoked)
}

func TestRefundNotificationMatchesOriginalPayment(t *testing.T) {
	transactionID := "wx-payment-1"
	order := &model.Order{AmountFen: 100, WechatTransactionID: &transactionID}
	refund := &model.Refund{AmountFen: 100}
	if !refundNotificationMatches(refund, order, "wx-refund-1", transactionID, 100, 100) {
		t.Fatal("matching notification was rejected")
	}
	for _, tc := range []struct {
		name        string
		refundID    string
		transaction string
		total       int64
		amount      int64
	}{
		{"missing refund id", "", transactionID, 100, 100},
		{"wrong payment", "wx-refund-1", "other", 100, 100},
		{"wrong total", "wx-refund-1", transactionID, 99, 100},
		{"wrong refund", "wx-refund-1", transactionID, 100, 99},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if refundNotificationMatches(refund, order, tc.refundID, tc.transaction, tc.total, tc.amount) {
				t.Fatal("mismatched notification was accepted")
			}
		})
	}
}

func assertCapacityStates(t *testing.T, tx *gorm.DB, orderID, refundID uint64, wantOrder, wantRefund, wantEntitlement string) {
	t.Helper()
	var order model.Order
	var refund model.Refund
	var entitlement model.Entitlement
	if err := tx.First(&order, orderID).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.First(&refund, refundID).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Where("order_id = ?", orderID).First(&entitlement).Error; err != nil {
		t.Fatal(err)
	}
	if order.Status != wantOrder || refund.Status != wantRefund || entitlement.Status != wantEntitlement {
		t.Fatalf("states: order=%s refund=%s entitlement=%s; want %s/%s/%s", order.Status, refund.Status, entitlement.Status, wantOrder, wantRefund, wantEntitlement)
	}
}
