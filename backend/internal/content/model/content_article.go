package model

import "time"

type ContentArticle struct {
	ID               uint64     `gorm:"primaryKey;column:id"`
	CategoryID       uint64     `gorm:"column:category_id;not null"`
	CategoryKey      string     `gorm:"column:category_key;size:50;not null"`
	ContentType      string     `gorm:"column:content_type;size:40;not null;default:INTERNAL"`
	Title            string     `gorm:"column:title;size:160;not null"`
	Slug             string     `gorm:"column:slug;size:160;not null;uniqueIndex:uk_content_articles_slug"`
	Summary          *string    `gorm:"column:summary;size:500"`
	CoverURL         *string    `gorm:"column:cover_url;size:500"`
	Body             string     `gorm:"column:body;type:text;not null"`
	ExternalURL      *string    `gorm:"column:external_url;size:1000"`
	AuthorName       *string    `gorm:"column:author_name;size:80"`
	Source           *string    `gorm:"column:source;size:160"`
	Status           string     `gorm:"column:status;size:40;not null;default:DRAFT"`
	IsFeatured       bool       `gorm:"column:is_featured;not null;default:false"`
	SortOrder        int        `gorm:"column:sort_order;not null;default:0"`
	PublishedAt      *time.Time `gorm:"column:published_at"`
	CreatedByAdminID *uint64    `gorm:"column:created_by_admin_id"`
	UpdatedByAdminID *uint64    `gorm:"column:updated_by_admin_id"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
}

func (ContentArticle) TableName() string {
	return "content_articles"
}
