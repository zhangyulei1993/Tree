package model

import "time"

type AccountFeatureOverride struct {
	ID               uint64    `gorm:"primaryKey;column:id"`
	FeatureKey       string    `gorm:"column:feature_key;size:80;not null"`
	PhoneHash        string    `gorm:"column:phone_hash;size:128;not null"`
	PhoneMask        string    `gorm:"column:phone_mask;size:30;not null"`
	UpdatedByAdminID *uint64   `gorm:"column:updated_by_admin_id"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AccountFeatureOverride) TableName() string {
	return "account_feature_overrides"
}
