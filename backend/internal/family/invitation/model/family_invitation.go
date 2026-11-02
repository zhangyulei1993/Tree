package model

import "time"

type FamilyInvitation struct {
	ID                    uint64     `gorm:"primaryKey;column:id"`
	FamilyID              uint64     `gorm:"column:family_id;not null"`
	TargetMemberID        uint64     `gorm:"column:target_member_id;not null"`
	InviterUserID         *uint64    `gorm:"column:inviter_user_id"`
	InviterAdminID        *uint64    `gorm:"column:inviter_admin_id"`
	TargetUserID          *uint64    `gorm:"column:target_user_id"`
	TargetPhone           *string    `gorm:"column:target_phone;size:30"`
	TargetPhoneHash       *string    `gorm:"column:target_phone_hash;size:128"`
	InviteType            string     `gorm:"column:invite_type;size:80;not null"`
	InviteChannel         string     `gorm:"column:invite_channel;size:40;not null"`
	InviteActorType       string     `gorm:"column:invite_actor_type;size:80;not null"`
	FamilyRoleAfterAccept string     `gorm:"column:family_role_after_accept;size:40;not null;default:MEMBER"`
	InviteToken           *string    `gorm:"column:invite_token;size:180"`
	InviteMessage         *string    `gorm:"column:invite_message;size:500"`
	PendingMemberLabel    *string    `gorm:"column:pending_member_label;size:120"`
	Status                string     `gorm:"column:status;size:40;not null;default:PENDING"`
	AcceptedByUserID      *uint64    `gorm:"column:accepted_by_user_id"`
	AcceptedAt            *time.Time `gorm:"column:accepted_at"`
	RejectedByUserID      *uint64    `gorm:"column:rejected_by_user_id"`
	RejectedAt            *time.Time `gorm:"column:rejected_at"`
	RejectReason          *string    `gorm:"column:reject_reason;size:500"`
	CancelledByUserID     *uint64    `gorm:"column:cancelled_by_user_id"`
	CancelledByAdminID    *uint64    `gorm:"column:cancelled_by_admin_id"`
	CancelledAt           *time.Time `gorm:"column:cancelled_at"`
	CancelReason          *string    `gorm:"column:cancel_reason;size:500"`
	ExpiredAt             time.Time  `gorm:"column:expired_at;not null"`
	CreatedAt             time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt             time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (FamilyInvitation) TableName() string {
	return "family_invitations"
}

// TODO(service): prevent duplicate PENDING invitation for the same member.
