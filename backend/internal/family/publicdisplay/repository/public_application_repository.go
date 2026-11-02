package repository

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	familymodel "tree/backend/internal/family/core/model"
	publicmodel "tree/backend/internal/family/publicdisplay/model"
	operationlog "tree/backend/internal/operationlog/service"
)

type ApplicationRow struct {
	publicmodel.FamilyPublicApplication
	FamilyName                string `gorm:"column:family_name"`
	FamilyPublicDisplayStatus string `gorm:"column:family_public_display_status"`
}

type ListQuery struct {
	FamilyID *uint64
	Status   string
	Page     int
	PageSize int
}

type QualityEligibility struct {
	MemberCount          int64 `gorm:"column:member_count"`
	RelationshipCount    int64 `gorm:"column:relationship_count"`
	BoundMemberCount     int64 `gorm:"column:bound_member_count"`
	ProfileCompleteCount int64 `gorm:"column:profile_complete_count"`
	UnknownGenderCount   int64 `gorm:"column:unknown_gender_count"`
	ProfileComplete      bool
}

type Repository interface {
	WithTx(*gorm.DB) Repository
	FindFamily(context.Context, uint64, bool) (*familymodel.Family, error)
	EvaluateQualityEligibility(context.Context, uint64) (*QualityEligibility, error)
	GetQualityEligibilityConfig(context.Context) (*publicmodel.QualityEligibilityConfig, error)
	UpdateQualityEligibilityConfig(context.Context, map[string]any) error
	UpdateFamilyPublicStatus(context.Context, uint64, map[string]any) error
	FindPendingByFamily(context.Context, uint64) (*publicmodel.FamilyPublicApplication, error)
	Create(context.Context, *publicmodel.FamilyPublicApplication) error
	FindByID(context.Context, uint64, bool) (*publicmodel.FamilyPublicApplication, error)
	List(context.Context, ListQuery) ([]ApplicationRow, int64, error)
	UpdateApplication(context.Context, uint64, string, map[string]any) error
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

func (r *GormRepository) EvaluateQualityEligibility(ctx context.Context, familyID uint64) (*QualityEligibility, error) {
	var result QualityEligibility
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			(SELECT COUNT(*) FROM family_members WHERE family_id = ? AND status = 'ACTIVE' AND deleted_at IS NULL) AS member_count,
			(SELECT COUNT(*) FROM family_relationships WHERE family_id = ? AND status = 'ACTIVE' AND deleted_at IS NULL) AS relationship_count,
			(SELECT COUNT(DISTINCT member_id) FROM family_member_user_links WHERE family_id = ? AND link_status = 'ACTIVE') AS bound_member_count,
			(SELECT COUNT(*) FROM family_members WHERE family_id = ? AND status = 'ACTIVE' AND deleted_at IS NULL AND gender <> 'UNKNOWN' AND birth_date IS NOT NULL) AS profile_complete_count,
			(SELECT COUNT(*) FROM family_members WHERE family_id = ? AND status = 'ACTIVE' AND deleted_at IS NULL AND gender = 'UNKNOWN') AS unknown_gender_count
	`, familyID, familyID, familyID, familyID, familyID).Scan(&result).Error; err != nil {
		return nil, err
	}
	var profile struct {
		Description *string
		RegionText  *string
		NativePlace *string
	}
	if err := r.db.WithContext(ctx).Table("families").Select("description, region_text, native_place").Where("id = ? AND deleted_at IS NULL", familyID).Scan(&profile).Error; err != nil {
		return nil, err
	}
	result.ProfileComplete = strings.TrimSpace(valueOrEmpty(profile.Description)) != "" &&
		(strings.TrimSpace(valueOrEmpty(profile.RegionText)) != "" || strings.TrimSpace(valueOrEmpty(profile.NativePlace)) != "")
	return &result, nil
}

func (r *GormRepository) GetQualityEligibilityConfig(ctx context.Context) (*publicmodel.QualityEligibilityConfig, error) {
	var value publicmodel.QualityEligibilityConfig
	err := r.db.WithContext(ctx).First(&value, 1).Error
	return &value, err
}

func (r *GormRepository) UpdateQualityEligibilityConfig(ctx context.Context, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&publicmodel.QualityEligibilityConfig{}).Where("id = 1").Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (r *GormRepository) UpdateFamilyPublicStatus(ctx context.Context, familyID uint64, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&familymodel.Family{}).Where("id = ? AND deleted_at IS NULL", familyID).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) FindPendingByFamily(ctx context.Context, familyID uint64) (*publicmodel.FamilyPublicApplication, error) {
	var value publicmodel.FamilyPublicApplication
	err := r.db.WithContext(ctx).
		Where("family_id = ? AND application_status = ?", familyID, "PENDING").
		First(&value).Error
	return &value, err
}

func (r *GormRepository) Create(ctx context.Context, value *publicmodel.FamilyPublicApplication) error {
	return r.db.WithContext(ctx).Create(value).Error
}

func (r *GormRepository) FindByID(ctx context.Context, id uint64, lock bool) (*publicmodel.FamilyPublicApplication, error) {
	var value publicmodel.FamilyPublicApplication
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.First(&value, id).Error
	return &value, err
}

func (r *GormRepository) List(ctx context.Context, query ListQuery) ([]ApplicationRow, int64, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Table("family_public_applications AS a").
		Joins("JOIN families AS f ON f.id = a.family_id")
	if query.FamilyID != nil {
		db = db.Where("a.family_id = ?", *query.FamilyID)
	}
	if query.Status != "" {
		db = db.Where("a.application_status = ?", query.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []ApplicationRow
	err := db.Select("a.*, f.family_name, f.public_display_status AS family_public_display_status").
		Order("a.created_at DESC, a.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error
	return rows, total, err
}

func (r *GormRepository) UpdateApplication(ctx context.Context, id uint64, current string, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&publicmodel.FamilyPublicApplication{}).
		Where("id = ? AND application_status = ?", id, current).
		Updates(values)
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
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(u.repo.WithTx(tx))
	})
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
