package model

import "time"

type FamilyFounderTransferRequest struct {
	ID                uint64     `gorm:"primaryKey;column:id"`
	FamilyID          uint64     `gorm:"column:family_id;not null"`
	FromMemberID      uint64     `gorm:"column:from_member_id;not null"`
	FromUserID        uint64     `gorm:"column:from_user_id;not null"`
	ToMemberID        uint64     `gorm:"column:to_member_id;not null"`
	ToUserID          uint64     `gorm:"column:to_user_id;not null"`
	RequestStatus     string     `gorm:"column:request_status;size:40;not null;default:PENDING"`
	RequestReason     *string    `gorm:"column:request_reason;size:500"`
	ReviewResult      *string    `gorm:"column:review_result;size:40"`
	ReviewedByAdminID *uint64    `gorm:"column:reviewed_by_admin_id"`
	ReviewedAt        *time.Time `gorm:"column:reviewed_at"`
	ReviewComment     *string    `gorm:"column:review_comment;size:500"`
	CompletedAt       *time.Time `gorm:"column:completed_at"`
	CancelledAt       *time.Time `gorm:"column:cancelled_at"`
	CancelReason      *string    `gorm:"column:cancel_reason;size:500"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (FamilyFounderTransferRequest) TableName() string {
	return "family_founder_transfer_requests"
}

// TODO(service): PLATFORM_ADMIN cannot approve founder transfer.
// TODO(service): one family can have only one PENDING founder transfer request.
