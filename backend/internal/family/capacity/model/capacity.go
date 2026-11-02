package model

import "time"

const (
	ProductStatusActive    = "ACTIVE"
	ProductStatusInactive  = "INACTIVE"
	OrderStatusPending     = "PENDING"
	OrderStatusCancelled   = "CANCELLED"
	OrderStatusPaid        = "PAID"
	OrderStatusRefunding   = "REFUNDING"
	OrderStatusRefunded    = "REFUNDED"
	RefundMaxAttempts      = 3
	EntitlementActive      = "ACTIVE"
	EntitlementRefunding   = "REFUNDING"
	EntitlementRevoked     = "REVOKED"
	RefundStatusPending    = "PENDING"
	RefundStatusProcessing = "PROCESSING"
	RefundStatusSuccess    = "SUCCESS"
	RefundStatusFailed     = "FAILED"
	RefundStatusRejected   = "REJECTED"
	RefundStatusReview     = "REVIEW_REQUIRED"
)

type Product struct {
	ID                  uint64    `gorm:"primaryKey"`
	ProductKey          string    `gorm:"column:product_key"`
	DisplayName         string    `gorm:"column:display_name"`
	AddedMembers        int       `gorm:"column:added_members"`
	AmountFen           int       `gorm:"column:amount_fen"`
	RefundWindowMinutes int       `gorm:"column:refund_window_minutes"`
	Status              string    `gorm:"column:status"`
	SortOrder           int       `gorm:"column:sort_order"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

func (Product) TableName() string { return "family_capacity_products" }

type Order struct {
	ID                  uint64     `gorm:"primaryKey"`
	OrderNo             string     `gorm:"column:order_no"`
	FamilyID            uint64     `gorm:"column:family_id"`
	BuyerUserID         uint64     `gorm:"column:buyer_user_id"`
	ProductID           uint64     `gorm:"column:product_id"`
	ProductName         string     `gorm:"column:product_name"`
	AddedMembers        int        `gorm:"column:added_members"`
	AmountFen           int        `gorm:"column:amount_fen"`
	RefundWindowMinutes int        `gorm:"column:refund_window_minutes"`
	Status              string     `gorm:"column:status"`
	WechatTransactionID *string    `gorm:"column:wechat_transaction_id"`
	PaidAt              *time.Time `gorm:"column:paid_at"`
	ClosedAt            *time.Time `gorm:"column:closed_at"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at"`
}

func (Order) TableName() string { return "family_capacity_orders" }

type Entitlement struct {
	ID           uint64     `gorm:"primaryKey"`
	FamilyID     uint64     `gorm:"column:family_id"`
	OrderID      uint64     `gorm:"column:order_id"`
	AddedMembers int        `gorm:"column:added_members"`
	Status       string     `gorm:"column:status"`
	RevokedAt    *time.Time `gorm:"column:revoked_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (Entitlement) TableName() string { return "family_capacity_entitlements" }

type Refund struct {
	ID                uint64     `gorm:"primaryKey"`
	RefundNo          string     `gorm:"column:refund_no"`
	OrderID           uint64     `gorm:"column:order_id"`
	AttemptNo         int        `gorm:"column:attempt_no"`
	FamilyID          uint64     `gorm:"column:family_id"`
	OperatorUserID    uint64     `gorm:"column:operator_user_id"`
	ReviewedByAdminID *uint64    `gorm:"column:reviewed_by_admin_id"`
	AmountFen         int        `gorm:"column:amount_fen"`
	Status            string     `gorm:"column:status"`
	WechatRefundID    *string    `gorm:"column:wechat_refund_id"`
	RequestedAt       time.Time  `gorm:"column:requested_at"`
	RequestReason     *string    `gorm:"column:request_reason;size:500"`
	SucceededAt       *time.Time `gorm:"column:succeeded_at"`
	FailedAt          *time.Time `gorm:"column:failed_at"`
	ReviewedAt        *time.Time `gorm:"column:reviewed_at"`
	ReviewReason      *string    `gorm:"column:review_reason;size:500"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}

func (Refund) TableName() string { return "family_capacity_refunds" }
