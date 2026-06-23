package vo

import "time"

type Category struct {
	ID          uint64    `json:"id"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	SortOrder   int       `json:"sortOrder"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ArticleSummary struct {
	ID           uint64     `json:"id"`
	CategoryID   uint64     `json:"categoryId"`
	CategoryKey  string     `json:"categoryKey"`
	CategoryName string     `json:"categoryName"`
	Title        string     `json:"title"`
	Slug         string     `json:"slug"`
	Summary      *string    `json:"summary,omitempty"`
	CoverURL     *string    `json:"coverUrl,omitempty"`
	AuthorName   *string    `json:"authorName,omitempty"`
	Source       *string    `json:"source,omitempty"`
	Status       string     `json:"status"`
	IsFeatured   bool       `json:"isFeatured"`
	SortOrder    int        `json:"sortOrder"`
	PublishedAt  *time.Time `json:"publishedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type ArticleDetail struct {
	ArticleSummary
	Body string `json:"body"`
}

type ListArticlesResult struct {
	Items    []ArticleSummary `json:"items"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
	Total    int64            `json:"total"`
}
