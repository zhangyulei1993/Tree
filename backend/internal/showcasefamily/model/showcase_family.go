package model

import (
	"time"

	"gorm.io/gorm"
)

type ShowcaseFamily struct {
	ID               uint64         `gorm:"primaryKey;column:id"`
	DisplayName      string         `gorm:"column:display_name;size:120;not null"`
	RegionText       string         `gorm:"column:region_text;size:120"`
	Summary          string         `gorm:"column:summary;size:1000;not null"`
	CoverImageURL    string         `gorm:"column:cover_image_url;size:500"`
	TreeImageURL     string         `gorm:"column:tree_image_url;size:500"`
	Status           string         `gorm:"column:status;size:20;not null;default:DRAFT"`
	SortOrder        int            `gorm:"column:sort_order;not null;default:0"`
	PublishedAt      *time.Time     `gorm:"column:published_at"`
	UpdatedByAdminID *uint64        `gorm:"column:updated_by_admin_id"`
	CreatedAt        time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (ShowcaseFamily) TableName() string { return "showcase_family_configs" }
