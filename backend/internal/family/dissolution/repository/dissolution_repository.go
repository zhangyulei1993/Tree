package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	adminmodel "tree/backend/internal/admin/model"
	familymodel "tree/backend/internal/family/core/model"
	dissolutionmodel "tree/backend/internal/family/dissolution/model"
	operationlog "tree/backend/internal/operationlog/service"
)

type DissolutionRow struct {
	dissolutionmodel.FamilyDissolutionRequest
	FamilyName   string `gorm:"column:family_name"`
	FamilyStatus string `gorm:"column:family_status"`
}

type ListQuery struct {
	Status   string
	FamilyID uint64
	Page     int
	PageSize int
}

type Repository interface {
	WithTx(*gorm.DB) Repository
	FindAdmin(context.Context, uint64) (*adminmodel.AdminUser, error)
	FindFamily(context.Context, uint64, bool) (*familymodel.Family, error)
	FindByID(context.Context, uint64, bool) (*dissolutionmodel.FamilyDissolutionRequest, error)
	List(context.Context, ListQuery) ([]DissolutionRow, int64, error)
	UpdateRequest(context.Context, uint64, string, map[string]any) error
	UpdateFamily(context.Context, uint64, map[string]any) error
	WriteLog(context.Context, operationlog.WriteInput) error
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository         { return &GormRepository{db: db} }
func (r *GormRepository) WithTx(tx *gorm.DB) Repository { return &GormRepository{db: tx} }

func (r *GormRepository) FindAdmin(ctx context.Context, id uint64) (*adminmodel.AdminUser, error) {
	var admin adminmodel.AdminUser
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&admin).Error
	return &admin, err
}

func (r *GormRepository) FindFamily(ctx context.Context, familyID uint64, lock bool) (*familymodel.Family, error) {
	var family familymodel.Family
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.Where("id = ? AND deleted_at IS NULL", familyID).First(&family).Error
	return &family, err
}

func (r *GormRepository) FindByID(ctx context.Context, id uint64, lock bool) (*dissolutionmodel.FamilyDissolutionRequest, error) {
	var request dissolutionmodel.FamilyDissolutionRequest
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.First(&request, id).Error
	return &request, err
}

func (r *GormRepository) List(ctx context.Context, query ListQuery) ([]DissolutionRow, int64, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Table("family_dissolution_requests AS d").Joins("JOIN families AS f ON f.id = d.family_id")
	if query.Status != "" {
		db = db.Where("d.request_status = ?", query.Status)
	}
	if query.FamilyID != 0 {
		db = db.Where("d.family_id = ?", query.FamilyID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []DissolutionRow
	err := db.Select("d.*, f.family_name, f.status AS family_status").
		Order("d.created_at DESC, d.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error
	return rows, total, err
}

func (r *GormRepository) UpdateRequest(ctx context.Context, id uint64, current string, values map[string]any) error {
	values["updated_at"] = time.Now()
	result := r.db.WithContext(ctx).Model(&dissolutionmodel.FamilyDissolutionRequest{}).Where("id = ? AND request_status = ?", id, current).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) UpdateFamily(ctx context.Context, familyID uint64, values map[string]any) error {
	values["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(&familymodel.Family{}).Where("id = ?", familyID).Updates(values).Error
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
