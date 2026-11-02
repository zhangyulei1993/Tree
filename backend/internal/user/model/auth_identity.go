package model

import "time"

type UserAuthIdentity struct {
	ID             uint64     `gorm:"primaryKey;column:id"`
	UserID         uint64     `gorm:"column:user_id;not null"`
	Provider       string     `gorm:"column:provider;size:80;not null"`
	ProviderAppID  *string    `gorm:"column:provider_app_id;size:120"`
	OpenID         *string    `gorm:"column:openid;size:180"`
	OpenIDHash     *string    `gorm:"column:openid_hash;size:128"`
	UnionID        *string    `gorm:"column:unionid;size:180"`
	UnionIDHash    *string    `gorm:"column:unionid_hash;size:128"`
	IdentityStatus string     `gorm:"column:identity_status;size:40;not null;default:ACTIVE"`
	BoundAt        *time.Time `gorm:"column:bound_at"`
	UnboundAt      *time.Time `gorm:"column:unbound_at"`
	LastLoginAt    *time.Time `gorm:"column:last_login_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (UserAuthIdentity) TableName() string {
	return "user_auth_identities"
}

// TODO(service): never log full openid, full unionid, session_key, or AppSecret.
