package model

import "time"

type AccountQuotaConfig struct {
	ID                       uint64    `gorm:"primaryKey;column:id"`
	TrustTier                string    `gorm:"column:trust_tier;size:40;not null"`
	MaxOwnedFamilies         int       `gorm:"column:max_owned_families;not null"`
	MaxMembersPerOwnedFamily int       `gorm:"column:max_members_per_owned_family;not null"`
	MaxJoinedFamilies        int       `gorm:"column:max_joined_families;not null"`
	SupportsGenerationNaming bool      `gorm:"column:supports_generation_naming;not null"`
	UpdatedByAdminID         *uint64   `gorm:"column:updated_by_admin_id"`
	CreatedAt                time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AccountQuotaConfig) TableName() string {
	return "account_quota_configs"
}
