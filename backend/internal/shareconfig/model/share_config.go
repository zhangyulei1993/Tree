package model

import "time"

type ShareConfig struct {
	ID               uint64    `gorm:"primaryKey;column:id"`
	ConfigKey        string    `gorm:"column:config_key;size:64;not null;uniqueIndex:uk_share_configs_key"`
	TitleTemplate    string    `gorm:"column:title_template;size:300;not null"`
	ImageURL         string    `gorm:"column:image_url;size:500;not null"`
	ImageURLs        *string   `gorm:"column:image_urls;type:text"`
	Description      *string   `gorm:"column:description;size:500"`
	UpdatedByAdminID *uint64   `gorm:"column:updated_by_admin_id"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ShareConfig) TableName() string { return "share_configs" }
