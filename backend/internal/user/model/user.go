package model

import "time"

type User struct {
	ID                uint64     `gorm:"primaryKey;column:id"`
	Phone             *string    `gorm:"column:phone;size:30"`
	PhoneHash         *string    `gorm:"column:phone_hash;size:128"`
	PhoneVerified     bool       `gorm:"column:phone_verified"`
	PasswordHash      *string    `gorm:"column:password_hash;size:255"`
	Nickname          *string    `gorm:"column:nickname;size:100"`
	RealName          *string    `gorm:"column:real_name;size:100"`
	AvatarURL         *string    `gorm:"column:avatar_url;size:500"`
	AccountOrigin     string     `gorm:"column:account_origin;size:80;not null"`
	RegisterClient    string     `gorm:"column:register_client;size:80;not null"`
	Status            string     `gorm:"column:status;size:40;not null;default:ACTIVE"`
	MergedToUserID    *uint64    `gorm:"column:merged_to_user_id"`
	MergedAt          *time.Time `gorm:"column:merged_at"`
	ClaimedAt         *time.Time `gorm:"column:claimed_at"`
	ClaimedVia        *string    `gorm:"column:claimed_via;size:80"`
	Remark            *string    `gorm:"column:remark;size:500"`
	LastLoginAt       *time.Time `gorm:"column:last_login_at"`
	LastLoginIP       *string    `gorm:"column:last_login_ip;size:80"`
	LastLoginClient   *string    `gorm:"column:last_login_client;size:80"`
	DisabledAt        *time.Time `gorm:"column:disabled_at"`
	DisabledByAdminID *uint64    `gorm:"column:disabled_by_admin_id"`
	DisabledReason    *string    `gorm:"column:disabled_reason;size:500"`
	CancelledAt       *time.Time `gorm:"column:cancelled_at"`
	CancelReason      *string    `gorm:"column:cancel_reason;size:500"`
	CreatedByAdminID  *uint64    `gorm:"column:created_by_admin_id"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt         *time.Time `gorm:"column:deleted_at"`
}

func (User) TableName() string {
	return "users"
}

// TODO(service): enforce active phone uniqueness across ACTIVE/PENDING_CLAIM users.
// TODO(service): users and family_members must remain separate domain models.
