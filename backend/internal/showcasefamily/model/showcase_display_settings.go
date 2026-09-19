package model

import "time"

type ShowcaseDisplaySettings struct {
	ID                     uint8      `gorm:"primaryKey;column:id"`
	RotationDays           int        `gorm:"column:rotation_days;not null;default:7"`
	PinnedShowcaseFamilyID *uint64    `gorm:"column:pinned_showcase_family_id"`
	PinnedUntil            *time.Time `gorm:"column:pinned_until"`
	UpdatedByAdminID       *uint64    `gorm:"column:updated_by_admin_id"`
	CreatedAt              time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (ShowcaseDisplaySettings) TableName() string { return "showcase_display_settings" }
