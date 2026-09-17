package model

import "time"

type ToolConfig struct {
	ID               uint64    `gorm:"primaryKey;column:id"`
	ToolKey          string    `gorm:"column:tool_key;size:80;not null;uniqueIndex:uk_tool_configs_key"`
	DisplayName      string    `gorm:"column:display_name;size:120;not null"`
	Description      string    `gorm:"column:description;size:300;not null"`
	Visible          bool      `gorm:"column:visible;not null;default:false"`
	Enabled          bool      `gorm:"column:enabled;not null;default:true"`
	Pinned           bool      `gorm:"column:pinned;not null;default:false"`
	Highlighted      bool      `gorm:"column:highlighted;not null;default:false"`
	SortOrder        int       `gorm:"column:sort_order;not null;default:0"`
	UpdatedByAdminID *uint64   `gorm:"column:updated_by_admin_id"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ToolConfig) TableName() string { return "tool_configs" }
