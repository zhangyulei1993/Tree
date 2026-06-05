package model

import "time"

type FamilyPublicApplication struct {
	ID                      uint64     `gorm:"primaryKey;column:id"`
	FamilyID                uint64     `gorm:"column:family_id;not null"`
	ApplicantUserID         *uint64    `gorm:"column:applicant_user_id"`
	ApplicantAdminID        *uint64    `gorm:"column:applicant_admin_id"`
	ApplicationStatus       string     `gorm:"column:application_status;size:40;not null;default:PENDING"`
	ApplicationReason       *string    `gorm:"column:application_reason;size:500"`
	ApplicationSnapshotJSON []byte     `gorm:"column:application_snapshot_json;type:json"`
	ReviewResult            *string    `gorm:"column:review_result;size:40"`
	ReviewedByAdminID       *uint64    `gorm:"column:reviewed_by_admin_id"`
	ReviewedAt              *time.Time `gorm:"column:reviewed_at"`
	ReviewComment           *string    `gorm:"column:review_comment;size:500"`
	CancelledAt             *time.Time `gorm:"column:cancelled_at"`
	CancelReason            *string    `gorm:"column:cancel_reason;size:500"`
	CreatedAt               time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt               time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (FamilyPublicApplication) TableName() string {
	return "family_public_applications"
}

// TODO(service): one family can have only one PENDING public application.
// TODO(service): PLATFORM_ADMIN cannot review its own public application.
