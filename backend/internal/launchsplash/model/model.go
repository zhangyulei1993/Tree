package model

import "time"

type Config struct {
	ID               uint64    `gorm:"primaryKey;column:id"`
	ConfigKey        string    `gorm:"column:config_key;size:64;not null;uniqueIndex:uk_launch_splash_configs_key"`
	Enabled          bool      `gorm:"column:enabled;not null"`
	Title            string    `gorm:"column:title;size:120;not null"`
	Tagline          string    `gorm:"column:tagline;size:240;not null"`
	ImageURL         string    `gorm:"column:image_url;size:500;not null"`
	DayImageURL      string    `gorm:"column:day_image_url;size:500;not null"`
	NightImageURL    string    `gorm:"column:night_image_url;size:500;not null"`
	DurationMs       int       `gorm:"column:duration_ms;not null"`
	SkipEnabled      bool      `gorm:"column:skip_enabled;not null"`
	UpdatedByAdminID *uint64   `gorm:"column:updated_by_admin_id"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Config) TableName() string { return "launch_splash_configs" }
