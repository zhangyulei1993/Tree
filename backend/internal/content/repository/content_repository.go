package repository

import (
	"context"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	contentmodel "tree/backend/internal/content/model"
	operationlog "tree/backend/internal/operationlog/service"
)

type ArticleRow struct {
	contentmodel.ContentArticle
	CategoryName   string `gorm:"column:category_name"`
	CategoryActive bool   `gorm:"column:category_active"`
}

type ListArticlesQuery struct {
	CategoryKey string
	Status      string
	Keyword     string
	Featured    *bool
	PublicOnly  bool
	Page        int
	PageSize    int
}

type Repository interface {
	WithTx(*gorm.DB) Repository
	CreateCategory(context.Context, *contentmodel.ContentCategory) error
	UpdateCategory(context.Context, uint64, map[string]any) error
	DeleteCategory(context.Context, uint64, time.Time) error
	FindCategoryByID(context.Context, uint64, bool) (*contentmodel.ContentCategory, error)
	FindCategoryByKey(context.Context, string, bool) (*contentmodel.ContentCategory, error)
	ListCategories(context.Context, bool) ([]contentmodel.ContentCategory, error)
	CreateArticle(context.Context, *contentmodel.ContentArticle) error
	UpdateArticle(context.Context, uint64, map[string]any) error
	DeleteArticle(context.Context, uint64, time.Time) error
	FindArticleByID(context.Context, uint64, bool) (*ArticleRow, error)
	FindArticleByIDOrSlug(context.Context, string, bool) (*ArticleRow, error)
	ListArticles(context.Context, ListArticlesQuery) ([]ArticleRow, int64, error)
	WriteLog(context.Context, operationlog.WriteInput) error
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository         { return &GormRepository{db: db} }
func (r *GormRepository) WithTx(tx *gorm.DB) Repository { return &GormRepository{db: tx} }

func (r *GormRepository) CreateCategory(ctx context.Context, category *contentmodel.ContentCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *GormRepository) UpdateCategory(ctx context.Context, id uint64, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&contentmodel.ContentCategory{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(values)
	return rowsError(result)
}

func (r *GormRepository) DeleteCategory(ctx context.Context, id uint64, now time.Time) error {
	result := r.db.WithContext(ctx).Model(&contentmodel.ContentCategory{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"deleted_at": now, "is_active": false})
	return rowsError(result)
}

func (r *GormRepository) FindCategoryByID(ctx context.Context, id uint64, lock bool) (*contentmodel.ContentCategory, error) {
	var category contentmodel.ContentCategory
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.Where("id = ? AND deleted_at IS NULL", id).First(&category).Error
	return &category, err
}

func (r *GormRepository) FindCategoryByKey(ctx context.Context, key string, lock bool) (*contentmodel.ContentCategory, error) {
	var category contentmodel.ContentCategory
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.Where("category_key = ? AND deleted_at IS NULL", key).First(&category).Error
	return &category, err
}

func (r *GormRepository) ListCategories(ctx context.Context, includeInactive bool) ([]contentmodel.ContentCategory, error) {
	var rows []contentmodel.ContentCategory
	query := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	err := query.Order("sort_order ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *GormRepository) CreateArticle(ctx context.Context, article *contentmodel.ContentArticle) error {
	return r.db.WithContext(ctx).Create(article).Error
}

func (r *GormRepository) UpdateArticle(ctx context.Context, id uint64, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&contentmodel.ContentArticle{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(values)
	return rowsError(result)
}

func (r *GormRepository) DeleteArticle(ctx context.Context, id uint64, now time.Time) error {
	result := r.db.WithContext(ctx).Model(&contentmodel.ContentArticle{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"deleted_at": now})
	return rowsError(result)
}

func (r *GormRepository) FindArticleByID(ctx context.Context, id uint64, lock bool) (*ArticleRow, error) {
	return r.findArticle(ctx, "a.id = ?", id, lock)
}

func (r *GormRepository) FindArticleByIDOrSlug(ctx context.Context, idOrSlug string, publicOnly bool) (*ArticleRow, error) {
	if id, err := strconv.ParseUint(idOrSlug, 10, 64); err == nil && id > 0 {
		return r.findArticleWithPublic(ctx, "a.id = ?", id, false, publicOnly)
	}
	return r.findArticleWithPublic(ctx, "a.slug = ?", idOrSlug, false, publicOnly)
}

func (r *GormRepository) ListArticles(ctx context.Context, query ListArticlesQuery) ([]ArticleRow, int64, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Table("content_articles AS a").
		Joins("JOIN content_categories AS c ON c.id = a.category_id").
		Where("a.deleted_at IS NULL AND c.deleted_at IS NULL")
	if query.PublicOnly {
		db = db.Where("a.status = ? AND c.is_active = ? AND a.published_at IS NOT NULL AND a.published_at <= ?", "PUBLISHED", true, time.Now())
	} else if query.Status != "" {
		db = db.Where("a.status = ?", query.Status)
	}
	if query.CategoryKey != "" {
		db = db.Where("a.category_key = ?", query.CategoryKey)
	}
	if query.Featured != nil {
		db = db.Where("a.is_featured = ?", *query.Featured)
	}
	keyword := strings.TrimSpace(query.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("a.title LIKE ? OR a.summary LIKE ? OR a.body LIKE ?", like, like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []ArticleRow
	err := db.Select("a.*, c.name AS category_name, c.is_active AS category_active").
		Order("a.is_featured DESC, a.sort_order ASC, COALESCE(a.published_at, a.created_at) DESC, a.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error
	return rows, total, err
}

func (r *GormRepository) WriteLog(ctx context.Context, input operationlog.WriteInput) error {
	return operationlog.NewGormService(r.db).WriteSuccess(ctx, input)
}

func (r *GormRepository) findArticle(ctx context.Context, condition string, value any, lock bool) (*ArticleRow, error) {
	return r.findArticleWithPublic(ctx, condition, value, lock, false)
}

func (r *GormRepository) findArticleWithPublic(ctx context.Context, condition string, value any, lock bool, publicOnly bool) (*ArticleRow, error) {
	var row ArticleRow
	query := r.db.WithContext(ctx).Table("content_articles AS a").
		Joins("JOIN content_categories AS c ON c.id = a.category_id").
		Where("a.deleted_at IS NULL AND c.deleted_at IS NULL").
		Where(condition, value)
	if publicOnly {
		query = query.Where("a.status = ? AND c.is_active = ? AND a.published_at IS NOT NULL AND a.published_at <= ?", "PUBLISHED", true, time.Now())
	}
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.Select("a.*, c.name AS category_name, c.is_active AS category_active").First(&row).Error
	return &row, err
}

func rowsError(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

type UnitOfWork interface {
	WithinTransaction(context.Context, func(Repository) error) error
}

type GormUnitOfWork struct {
	db   *gorm.DB
	repo Repository
}

func NewUnitOfWork(db *gorm.DB, repo Repository) *GormUnitOfWork {
	return &GormUnitOfWork{db: db, repo: repo}
}

func (u *GormUnitOfWork) WithinTransaction(ctx context.Context, fn func(Repository) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return fn(u.repo.WithTx(tx)) })
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
