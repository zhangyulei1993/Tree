package model

import "time"

const (
	ReferralStatusPending   = "PENDING"
	ReferralStatusQualified = "QUALIFIED"
	GrantStatusActive       = "ACTIVE"
)

type Config struct {
	ID                       uint64    `gorm:"primaryKey"`
	Enabled                  bool      `gorm:"column:enabled"`
	ActivationRemainingSlots int       `gorm:"column:activation_remaining_slots"`
	MaxRewardMembers         int       `gorm:"column:max_reward_members"`
	RequirePhoneBound        bool      `gorm:"column:require_phone_bound"`
	UpdatedByAdminID         *uint64   `gorm:"column:updated_by_admin_id"`
	CreatedAt                time.Time `gorm:"column:created_at"`
	UpdatedAt                time.Time `gorm:"column:updated_at"`
}

func (Config) TableName() string { return "family_invitation_reward_configs" }

type Tier struct {
	ID                      uint64    `gorm:"primaryKey"`
	ConfigID                uint64    `gorm:"column:config_id"`
	RequiredInvites         int       `gorm:"column:required_invites"`
	CumulativeRewardMembers int       `gorm:"column:cumulative_reward_members"`
	SortOrder               int       `gorm:"column:sort_order"`
	Retired                 bool      `gorm:"column:retired"`
	CreatedAt               time.Time `gorm:"column:created_at"`
	UpdatedAt               time.Time `gorm:"column:updated_at"`
}

func (Tier) TableName() string { return "family_invitation_reward_tiers" }

type Referral struct {
	ID            uint64     `gorm:"primaryKey"`
	FamilyID      uint64     `gorm:"column:family_id"`
	InviterUserID uint64     `gorm:"column:inviter_user_id"`
	ReferralCode  string     `gorm:"column:referral_code"`
	InviteeUserID *uint64    `gorm:"column:invitee_user_id"`
	Status        string     `gorm:"column:status"`
	QualifiedAt   *time.Time `gorm:"column:qualified_at"`
	ExpiresAt     *time.Time `gorm:"column:expires_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (Referral) TableName() string { return "family_invitation_reward_referrals" }

type Grant struct {
	ID           uint64    `gorm:"primaryKey"`
	FamilyID     uint64    `gorm:"column:family_id"`
	TierID       uint64    `gorm:"column:tier_id"`
	AddedMembers int       `gorm:"column:added_members"`
	Status       string    `gorm:"column:status"`
	GrantedAt    time.Time `gorm:"column:granted_at"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (Grant) TableName() string { return "family_invitation_reward_grants" }
