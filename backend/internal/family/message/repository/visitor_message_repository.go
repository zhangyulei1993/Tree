package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	familymodel "tree/backend/internal/family/core/model"
	messagemodel "tree/backend/internal/family/message/model"
	operationlog "tree/backend/internal/operationlog/service"
)

type MessageRow struct {
	messagemodel.VisitorMessage
	FamilyName string `gorm:"column:family_name"`
}

type ListQuery struct {
	Status     string
	FamilyID   uint64
	PublicOnly bool
	Page       int
	PageSize   int
}

type Repository interface {
	WithTx(*gorm.DB) Repository
	FindFamily(context.Context, uint64, bool) (*familymodel.Family, error)
	CountByIP(context.Context, uint64, string, time.Time) (int64, error)
	Create(context.Context, *messagemodel.VisitorMessage) error
	FindByID(context.Context, uint64, bool) (*messagemodel.VisitorMessage, error)
	List(context.Context, ListQuery) ([]MessageRow, int64, error)
	UpdateStatus(context.Context, uint64, map[string]any) error
	WriteLog(context.Context, operationlog.WriteInput) error
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository         { return &GormRepository{db: db} }
func (r *GormRepository) WithTx(tx *gorm.DB) Repository { return &GormRepository{db: tx} }

func (r *GormRepository) FindFamily(ctx context.Context, familyID uint64, lock bool) (*familymodel.Family, error) {
	var family familymodel.Family
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.Where("id = ? AND deleted_at IS NULL", familyID).First(&family).Error
	return &family, err
}

func (r *GormRepository) CountByIP(ctx context.Context, familyID uint64, ip string, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&messagemodel.VisitorMessage{}).
		Where("family_id = ? AND ip = ? AND created_at >= ?", familyID, ip, since).
		Count(&count).Error
	return count, err
}

func (r *GormRepository) Create(ctx context.Context, message *messagemodel.VisitorMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *GormRepository) FindByID(ctx context.Context, id uint64, lock bool) (*messagemodel.VisitorMessage, error) {
	var message messagemodel.VisitorMessage
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.First(&message, id).Error
	return &message, err
}

func (r *GormRepository) List(ctx context.Context, query ListQuery) ([]MessageRow, int64, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Table("visitor_messages AS m").
		Joins("JOIN families AS f ON f.id = m.family_id")
	if query.FamilyID != 0 {
		db = db.Where("m.family_id = ?", query.FamilyID)
	}
	if query.PublicOnly {
		db = db.Where("m.status = ? AND m.deleted_at IS NULL", "APPROVED")
	} else if query.Status != "" {
		db = db.Where("m.status = ?", query.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []MessageRow
	err := db.Select("m.*, f.family_name").
		Order("COALESCE(m.reviewed_at, m.created_at) DESC, m.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error
	return rows, total, err
}

func (r *GormRepository) UpdateStatus(ctx context.Context, id uint64, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&messagemodel.VisitorMessage{}).Where("id = ?", id).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) WriteLog(ctx context.Context, input operationlog.WriteInput) error {
	return operationlog.NewGormService(r.db).WriteSuccess(ctx, input)
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
