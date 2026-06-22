package model

import "time"

type ContentCategory struct {
	ID          uint64     `gorm:"primaryKey;column:id"`
	Key         string     `gorm:"column:category_key;size:50;not null;uniqueIndex:uk_content_categories_key"`
	Name        string     `gorm:"column:name;size:80;not null"`
	Description *string    `gorm:"column:description;size:500"`
	SortOrder   int        `gorm:"column:sort_order;not null;default:0"`
	IsActive    bool       `gorm:"column:is_active;not null;default:true"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (ContentCategory) TableName() string {
	return "content_categories"
}
