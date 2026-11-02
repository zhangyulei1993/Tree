package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"tree/backend/internal/family/capacity/model"
)

type Repository interface {
	WithDB(*gorm.DB) Repository
	ListActiveProducts(context.Context) ([]model.Product, error)
	ListAllProducts(context.Context) ([]model.Product, error)
	FindProductByKey(context.Context, string, bool) (*model.Product, error)
	UpdateProduct(context.Context, string, map[string]any) error
	ListOrders(context.Context, uint64, uint64) ([]model.Order, error)
	ListOrdersPage(context.Context, uint64, uint64, int, int, string, *time.Time, *time.Time) ([]model.Order, int64, error)
	ListAllOrders(context.Context, int, int, string, *time.Time, *time.Time) ([]model.Order, int64, error)
	ListRefundsByOrderIDs(context.Context, []uint64) ([]model.Refund, error)
	ExpirePendingOrders(context.Context, time.Time, time.Time) error
	LockFamily(context.Context, uint64) error
	FindActiveProduct(context.Context, string) (*model.Product, error)
	CreateOrder(context.Context, *model.Order) error
	FindOrder(context.Context, string, bool) (*model.Order, error)
	FindOrderByID(context.Context, uint64, bool) (*model.Order, error)
	FindEntitlement(context.Context, uint64, bool) (*model.Entitlement, error)
	MarkPaid(context.Context, uint64, string, time.Time) error
	MarkRefunding(context.Context, uint64) error
	MarkOrderPaid(context.Context, uint64) error
	CreateEntitlement(context.Context, *model.Entitlement) error
	CreateRefund(context.Context, *model.Refund) error
	FindRefund(context.Context, string, bool) (*model.Refund, error)
	FindRefundByOrderID(context.Context, uint64, bool) (*model.Refund, error)
	MarkRefunded(context.Context, uint64, string, time.Time) error
	MarkRefundFailed(context.Context, uint64, time.Time) error
	MarkRefundReviewRequired(context.Context, uint64) error
	ReopenFailedRefundReview(context.Context, uint64) error
	MarkRefundNotificationProcessing(context.Context, uint64, time.Time) error
	MarkReviewRefundProcessing(context.Context, uint64) error
	MarkRefundProcessing(context.Context, uint64, uint64, string, time.Time) error
	MarkRefundRejected(context.Context, uint64, uint64, string, time.Time) error
	MarkOrderRefunded(context.Context, uint64, time.Time) error
	RevokeEntitlement(context.Context, uint64, time.Time) error
	MarkEntitlementRefunding(context.Context, uint64) error
	RestoreEntitlement(context.Context, uint64) error
	CountActiveMembers(context.Context, uint64) (int, error)
	SumActiveEntitlements(context.Context, uint64) (int, error)
}

type gormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository              { return &gormRepository{db: db} }
func (r *gormRepository) WithDB(db *gorm.DB) Repository { return &gormRepository{db: db} }

func (r *gormRepository) ListActiveProducts(ctx context.Context) ([]model.Product, error) {
	var rows []model.Product
	err := r.db.WithContext(ctx).Where("status = ?", model.ProductStatusActive).Order("sort_order, id").Find(&rows).Error
	return rows, err
}
func (r *gormRepository) ListAllProducts(ctx context.Context) ([]model.Product, error) {
	var rows []model.Product
	err := r.db.WithContext(ctx).Order("sort_order, id").Find(&rows).Error
	return rows, err
}
func (r *gormRepository) FindProductByKey(ctx context.Context, key string, lock bool) (*model.Product, error) {
	var row model.Product
	q := r.db.WithContext(ctx).Where("product_key = ?", key)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.First(&row).Error
	return &row, err
}
func (r *gormRepository) UpdateProduct(ctx context.Context, key string, values map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.Product{}).Where("product_key = ?", key).Updates(values).Error
}
func (r *gormRepository) ListOrders(ctx context.Context, familyID, userID uint64) ([]model.Order, error) {
	var rows []model.Order
	err := r.db.WithContext(ctx).
		Where("family_id = ? AND buyer_user_id = ?", familyID, userID).
		Order("id DESC").Find(&rows).Error
	return rows, err
}

func (r *gormRepository) ListOrdersPage(ctx context.Context, familyID, userID uint64, page, pageSize int, status string, startDate, endDate *time.Time) ([]model.Order, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Order{}).
		Where("family_id = ? AND buyer_user_id = ?", familyID, userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at < ?", *endDate)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Order
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}
func (r *gormRepository) ListAllOrders(ctx context.Context, page, pageSize int, status string, startDate, endDate *time.Time) ([]model.Order, int64, error) {
	buildQuery := func() *gorm.DB {
		query := r.db.WithContext(ctx).Model(&model.Order{})
		if status != "" {
			query = query.Where("status = ?", status)
		}
		if startDate != nil {
			query = query.Where("created_at >= ?", *startDate)
		}
		if endDate != nil {
			query = query.Where("created_at < ?", *endDate)
		}
		return query
	}
	var total int64
	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Order
	err := buildQuery().Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}
func (r *gormRepository) ListRefundsByOrderIDs(ctx context.Context, orderIDs []uint64) ([]model.Refund, error) {
	if len(orderIDs) == 0 {
		return []model.Refund{}, nil
	}
	var rows []model.Refund
	err := r.db.WithContext(ctx).Where("order_id IN ?", orderIDs).Order("attempt_no DESC, id DESC").Find(&rows).Error
	return rows, err
}
func (r *gormRepository) ExpirePendingOrders(ctx context.Context, now, before time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Order{}).
		Where("status = ? AND created_at <= ?", model.OrderStatusPending, before).
		Updates(map[string]any{"status": model.OrderStatusCancelled, "closed_at": now}).Error
}
func (r *gormRepository) LockFamily(ctx context.Context, familyID uint64) error {
	var row struct{ ID uint64 }
	return r.db.WithContext(ctx).Table("families").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", familyID).First(&row).Error
}
func (r *gormRepository) FindActiveProduct(ctx context.Context, key string) (*model.Product, error) {
	var row model.Product
	err := r.db.WithContext(ctx).Where("product_key = ? AND status = ?", key, model.ProductStatusActive).First(&row).Error
	return &row, err
}
func (r *gormRepository) CreateOrder(ctx context.Context, row *model.Order) error {
	return r.db.WithContext(ctx).Create(row).Error
}
func (r *gormRepository) FindOrder(ctx context.Context, orderNo string, lock bool) (*model.Order, error) {
	var row model.Order
	q := r.db.WithContext(ctx).Where("order_no = ?", orderNo)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.First(&row).Error
	return &row, err
}
func (r *gormRepository) FindOrderByID(ctx context.Context, id uint64, lock bool) (*model.Order, error) {
	var row model.Order
	q := r.db.WithContext(ctx).Where("id = ?", id)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.First(&row).Error
	return &row, err
}
func (r *gormRepository) FindEntitlement(ctx context.Context, orderID uint64, lock bool) (*model.Entitlement, error) {
	var row model.Entitlement
	q := r.db.WithContext(ctx).Where("order_id = ?", orderID)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.First(&row).Error
	return &row, err
}
func (r *gormRepository) MarkPaid(ctx context.Context, id uint64, transactionID string, now time.Time) error {
	// A delayed WeChat callback may arrive after the expiry worker has closed the
	// local pending order. The signed callback is the source of truth for payment
	// and must be able to reconcile that order instead of losing the entitlement.
	return requireOneRow(r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ? AND status IN ?", id, []string{model.OrderStatusPending, model.OrderStatusCancelled}).Updates(map[string]any{"status": model.OrderStatusPaid, "wechat_transaction_id": transactionID, "paid_at": now}))
}
func (r *gormRepository) MarkRefunding(ctx context.Context, id uint64) error {
	return requireOneRow(r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ? AND status = ?", id, model.OrderStatusPaid).Update("status", model.OrderStatusRefunding))
}
func (r *gormRepository) MarkOrderPaid(ctx context.Context, id uint64) error {
	return requireOneRow(r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ? AND status = ?", id, model.OrderStatusRefunding).Update("status", model.OrderStatusPaid))
}
func (r *gormRepository) CreateEntitlement(ctx context.Context, row *model.Entitlement) error {
	return r.db.WithContext(ctx).Create(row).Error
}
func (r *gormRepository) CreateRefund(ctx context.Context, row *model.Refund) error {
	return r.db.WithContext(ctx).Create(row).Error
}
func (r *gormRepository) FindRefund(ctx context.Context, refundNo string, lock bool) (*model.Refund, error) {
	var row model.Refund
	q := r.db.WithContext(ctx).Where("refund_no = ?", refundNo)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.First(&row).Error
	return &row, err
}
func (r *gormRepository) FindRefundByOrderID(ctx context.Context, orderID uint64, lock bool) (*model.Refund, error) {
	var row model.Refund
	q := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("attempt_no DESC, id DESC")
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.First(&row).Error
	return &row, err
}
func (r *gormRepository) MarkRefunded(ctx context.Context, id uint64, refundID string, now time.Time) error {
	return requireOneRow(r.db.WithContext(ctx).Model(&model.Refund{}).Where("id = ? AND status IN ?", id, []string{model.RefundStatusPending, model.RefundStatusProcessing, model.RefundStatusReview, model.RefundStatusFailed}).Updates(map[string]any{"status": model.RefundStatusSuccess, "wechat_refund_id": refundID, "succeeded_at": now}))
}
func (r *gormRepository) MarkRefundFailed(ctx context.Context, id uint64, now time.Time) error {
	return requireOneRow(r.db.WithContext(ctx).Model(&model.Refund{}).Where("id = ? AND status IN ?", id, []string{model.RefundStatusPending, model.RefundStatusProcessing, model.RefundStatusReview}).Updates(map[string]any{"status": model.RefundStatusFailed, "failed_at": now}))
}
func (r *gormRepository) MarkRefundReviewRequired(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.Refund{}).Where("id = ? AND status IN ?", id, []string{model.RefundStatusPending, model.RefundStatusProcessing, model.RefundStatusReview}).Update("status", model.RefundStatusReview).Error
}
func (r *gormRepository) ReopenFailedRefundReview(ctx context.Context, id uint64) error {
	return requireOneRow(r.db.WithContext(ctx).Model(&model.Refund{}).Where("id = ? AND status = ?", id, model.RefundStatusFailed).Update("status", model.RefundStatusReview))
}
func (r *gormRepository) MarkRefundNotificationProcessing(ctx context.Context, id uint64, now time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Refund{}).Where("id = ? AND status IN ?", id, []string{model.RefundStatusPending, model.RefundStatusProcessing}).Updates(map[string]any{"status": model.RefundStatusProcessing}).Error
}
func (r *gormRepository) MarkReviewRefundProcessing(ctx context.Context, id uint64) error {
	return requireOneRow(r.db.WithContext(ctx).Model(&model.Refund{}).Where("id = ? AND status = ?", id, model.RefundStatusReview).Update("status", model.RefundStatusProcessing))
}
func (r *gormRepository) MarkRefundProcessing(ctx context.Context, id, adminID uint64, reason string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Refund{}).Where("id = ? AND status = ?", id, model.RefundStatusPending).Updates(map[string]any{"status": model.RefundStatusProcessing, "reviewed_by_admin_id": adminID, "reviewed_at": now, "review_reason": reason}).Error
}
func (r *gormRepository) MarkRefundRejected(ctx context.Context, id, adminID uint64, reason string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Refund{}).Where("id = ? AND status = ?", id, model.RefundStatusPending).Updates(map[string]any{"status": model.RefundStatusRejected, "reviewed_by_admin_id": adminID, "reviewed_at": now, "review_reason": reason}).Error
}
func (r *gormRepository) MarkOrderRefunded(ctx context.Context, id uint64, now time.Time) error {
	return requireOneRow(r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ? AND status IN ?", id, []string{model.OrderStatusRefunding, model.OrderStatusPaid}).Updates(map[string]any{"status": model.OrderStatusRefunded, "closed_at": now}))
}
func (r *gormRepository) RevokeEntitlement(ctx context.Context, orderID uint64, now time.Time) error {
	return requireOneRow(r.db.WithContext(ctx).Model(&model.Entitlement{}).Where("order_id = ? AND status IN ?", orderID, []string{model.EntitlementActive, model.EntitlementRefunding}).Updates(map[string]any{"status": model.EntitlementRevoked, "revoked_at": now}))
}
func (r *gormRepository) MarkEntitlementRefunding(ctx context.Context, orderID uint64) error {
	return requireOneRow(r.db.WithContext(ctx).Model(&model.Entitlement{}).Where("order_id = ? AND status = ?", orderID, model.EntitlementActive).Update("status", model.EntitlementRefunding))
}
func (r *gormRepository) RestoreEntitlement(ctx context.Context, orderID uint64) error {
	return requireOneRow(r.db.WithContext(ctx).Model(&model.Entitlement{}).Where("order_id = ? AND status = ?", orderID, model.EntitlementRefunding).Update("status", model.EntitlementActive))
}

func requireOneRow(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("capacity state changed concurrently")
	}
	return nil
}
func (r *gormRepository) CountActiveMembers(ctx context.Context, familyID uint64) (int, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("family_members").Where("family_id = ? AND status = 'ACTIVE' AND deleted_at IS NULL", familyID).Count(&n).Error
	return int(n), err
}
func (r *gormRepository) SumActiveEntitlements(ctx context.Context, familyID uint64) (int, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Entitlement{}).Where("family_id = ? AND status = ?", familyID, model.EntitlementActive).Select("COALESCE(SUM(added_members), 0)").Scan(&n).Error
	return int(n), err
}
