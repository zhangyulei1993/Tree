package model

import "time"

type FamilyMemberUserLink struct {
	ID                   uint64     `gorm:"primaryKey;column:id"`
	FamilyID             uint64     `gorm:"column:family_id;not null"`
	MemberID             uint64     `gorm:"column:member_id;not null"`
	UserID               uint64     `gorm:"column:user_id;not null"`
	LinkStatus           string     `gorm:"column:link_status;size:40;not null;default:ACTIVE"`
	LinkSource           string     `gorm:"column:link_source;size:80;not null"`
	FamilyRole           string     `gorm:"column:family_role;size:40;not null;default:MEMBER"`
	InvitationID         *uint64    `gorm:"column:invitation_id"`
	JoinRequestID        *uint64    `gorm:"column:join_request_id"`
	RoleGrantedAt        *time.Time `gorm:"column:role_granted_at"`
	RoleGrantedByUserID  *uint64    `gorm:"column:role_granted_by_user_id"`
	RoleGrantedByAdminID *uint64    `gorm:"column:role_granted_by_admin_id"`
	UnlinkedAt           *time.Time `gorm:"column:unlinked_at"`
	UnlinkedReason       *string    `gorm:"column:unlinked_reason;size:500"`
	UnlinkedByUserID     *uint64    `gorm:"column:unlinked_by_user_id"`
	UnlinkedByAdminID    *uint64    `gorm:"column:unlinked_by_admin_id"`
	CreatedAt            time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (FamilyMemberUserLink) TableName() string {
	return "family_member_user_links"
}

// TODO(service): one ACTIVE user link per user per family.
// TODO(service): one ACTIVE user link per member.
// TODO(service): one FOUNDER per family.
