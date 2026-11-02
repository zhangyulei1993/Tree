package model

import "time"

type Notification struct {
	ID          uint64     `gorm:"primaryKey;column:id"`
	UserID      uint64     `gorm:"column:user_id;not null"`
	Category    string     `gorm:"column:category;size:40;not null"`
	Title       string     `gorm:"column:title;size:120;not null"`
	Content     string     `gorm:"column:content;size:500;not null"`
	Route       *string    `gorm:"column:route;size:500"`
	RelatedType *string    `gorm:"column:related_type;size:60"`
	RelatedID   *uint64    `gorm:"column:related_id"`
	ReadAt      *time.Time `gorm:"column:read_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (Notification) TableName() string { return "user_notifications" }
