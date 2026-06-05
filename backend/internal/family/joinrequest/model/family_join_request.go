package model

import "time"

type FamilyJoinRequest struct {
	ID                 uint64     `gorm:"primaryKey;column:id"`
	FamilyID           uint64     `gorm:"column:family_id;not null"`
	ApplicantUserID    uint64     `gorm:"column:applicant_user_id;not null"`
	ApplicantRealName  *string    `gorm:"column:applicant_real_name;size:100"`
	ApplicantPhone     *string    `gorm:"column:applicant_phone;size:30"`
	ApplicantPhoneHash *string    `gorm:"column:applicant_phone_hash;size:128"`
	ApplicantMessage   *string    `gorm:"column:applicant_message;size:500"`
	RequestStatus      string     `gorm:"column:request_status;size:40;not null;default:PENDING"`
	ApproveMode        *string    `gorm:"column:approve_mode;size:80"`
	BoundMemberID      *uint64    `gorm:"column:bound_member_id"`
	CreatedMemberID    *uint64    `gorm:"column:created_member_id"`
	HandleResult       *string    `gorm:"column:handle_result;size:80"`
	HandledByUserID    *uint64    `gorm:"column:handled_by_user_id"`
	HandledByAdminID   *uint64    `gorm:"column:handled_by_admin_id"`
	HandledAt          *time.Time `gorm:"column:handled_at"`
	HandleComment      *string    `gorm:"column:handle_comment;size:500"`
	CancelledAt        *time.Time `gorm:"column:cancelled_at"`
	CancelReason       *string    `gorm:"column:cancel_reason;size:500"`
	CreatedAt          time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (FamilyJoinRequest) TableName() string {
	return "family_join_requests"
}

// TODO(service): approving a join request that creates a link must be transactional.
