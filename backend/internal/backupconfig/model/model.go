package model

import "time"

type Config struct {
	ID                  uint8     `gorm:"primaryKey;column:id"`
	Enabled             bool      `gorm:"column:enabled"`
	RetentionDays       int       `gorm:"column:retention_days"`
	MaxStorageMB        int       `gorm:"column:max_storage_mb"`
	MaxDiskUsagePercent int       `gorm:"column:max_disk_usage_percent"`
	UpdatedByAdminID    *uint64   `gorm:"column:updated_by_admin_id"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

func (Config) TableName() string { return "backup_operation_configs" }
