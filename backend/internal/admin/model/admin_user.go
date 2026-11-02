package model

import "time"

type AdminUser struct {
	ID                uint64     `gorm:"primaryKey;column:id"`
	Username          string     `gorm:"column:username;size:100;not null;uniqueIndex:uk_admin_users_username"`
	PasswordHash      string     `gorm:"column:password_hash;size:255;not null"`
	DisplayName       *string    `gorm:"column:display_name;size:100"`
	Phone             *string    `gorm:"column:phone;size:30"`
	Email             *string    `gorm:"column:email;size:150"`
	Role              string     `gorm:"column:role;size:40;not null"`
	Status            string     `gorm:"column:status;size:40;not null;default:ACTIVE"`
	FailedLoginCount  int        `gorm:"column:failed_login_count;not null;default:0"`
	LockedUntil       *time.Time `gorm:"column:locked_until"`
	LastLoginAt       *time.Time `gorm:"column:last_login_at"`
	LastLoginIP       *string    `gorm:"column:last_login_ip;size:80"`
	PasswordChangedAt *time.Time `gorm:"column:password_changed_at"`
	DisabledAt        *time.Time `gorm:"column:disabled_at"`
	DisabledByAdminID *uint64    `gorm:"column:disabled_by_admin_id"`
	DisabledReason    *string    `gorm:"column:disabled_reason;size:500"`
	DeletedAt         *time.Time `gorm:"column:deleted_at"`
	DeletedByAdminID  *uint64    `gorm:"column:deleted_by_admin_id"`
	UnlockedAt        *time.Time `gorm:"column:unlocked_at"`
	UnlockedByAdminID *uint64    `gorm:"column:unlocked_by_admin_id"`
	Remark            *string    `gorm:"column:remark;size:500"`
	CreatedByAdminID  *uint64    `gorm:"column:created_by_admin_id"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (AdminUser) TableName() string {
	return "admin_users"
}

// TODO(service): ROOT_ADMIN is unique and cannot be managed by lower roles.
