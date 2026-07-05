package dto

type ListCategoriesQuery struct {
	IncludeInactive bool
}

type CreateCategoryRequest struct {
	Key         string  `json:"key" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	SortOrder   int     `json:"sortOrder"`
	IsActive    *bool   `json:"isActive"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sortOrder"`
	IsActive    *bool   `json:"isActive"`
}

type ListArticlesQuery struct {
	CategoryKey string
	Status      string
	Keyword     string
	Featured    *bool
	Page        int
	PageSize    int
}

type CreateArticleRequest struct {
	CategoryKey string  `json:"categoryKey" binding:"required"`
	ContentType string  `json:"contentType"`
	Title       string  `json:"title" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
	Summary     *string `json:"summary"`
	CoverURL    *string `json:"coverUrl"`
	Body        string  `json:"body"`
	ExternalURL *string `json:"externalUrl"`
	AuthorName  *string `json:"authorName"`
	Source      *string `json:"source"`
	Status      string  `json:"status"`
	IsFeatured  bool    `json:"isFeatured"`
	SortOrder   int     `json:"sortOrder"`
}

type UpdateArticleRequest struct {
	CategoryKey *string `json:"categoryKey"`
	ContentType *string `json:"contentType"`
	Title       *string `json:"title"`
	Slug        *string `json:"slug"`
	Summary     *string `json:"summary"`
	CoverURL    *string `json:"coverUrl"`
	Body        *string `json:"body"`
	ExternalURL *string `json:"externalUrl"`
	AuthorName  *string `json:"authorName"`
	Source      *string `json:"source"`
	Status      *string `json:"status"`
	IsFeatured  *bool   `json:"isFeatured"`
	SortOrder   *int    `json:"sortOrder"`
}

type PublishArticleRequest struct {
	ReviewComment *string `json:"reviewComment"`
}
