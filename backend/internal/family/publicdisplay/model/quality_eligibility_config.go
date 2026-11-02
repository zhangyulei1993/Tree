package model

import "time"

type QualityEligibilityConfig struct {
	ID                     uint64    `gorm:"primaryKey;column:id"`
	Enabled                bool      `gorm:"column:enabled;not null;default:true"`
	MinMemberCount         int       `gorm:"column:min_member_count;not null;default:3"`
	MinRelationshipCount   int       `gorm:"column:min_relationship_count;not null;default:2"`
	MinBoundMemberCount    int       `gorm:"column:min_bound_member_count;not null;default:1"`
	MinProfileCompleteRate int       `gorm:"column:min_profile_complete_rate;not null;default:0"`
	MinBoundMemberRate     int       `gorm:"column:min_bound_member_rate;not null;default:0"`
	RequireDescription     bool      `gorm:"column:require_description;not null;default:false"`
	RequireLocation        bool      `gorm:"column:require_location;not null;default:false"`
	UpdatedByAdminID       *uint64   `gorm:"column:updated_by_admin_id"`
	CreatedAt              time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt              time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (QualityEligibilityConfig) TableName() string { return "quality_family_eligibility_configs" }
