package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	adminmodel "tree/backend/internal/admin/model"
	familymodel "tree/backend/internal/family/core/model"
	membermodel "tree/backend/internal/family/member/model"
	rolemodel "tree/backend/internal/family/role/model"
	transfermodel "tree/backend/internal/family/transfer/model"
	operationlog "tree/backend/internal/operationlog/service"
)

type TransferRow struct {
	transfermodel.FamilyFounderTransferRequest
	FamilyName string `gorm:"column:family_name"`
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
	FindMember(context.Context, uint64, uint64, bool) (*membermodel.FamilyMember, error)
	FindActiveLinkByUser(context.Context, uint64, uint64, bool) (*rolemodel.FamilyMemberUserLink, error)
	FindActiveLinkByMember(context.Context, uint64, uint64, bool) (*rolemodel.FamilyMemberUserLink, error)
	FindPendingByFamily(context.Context, uint64) (*transfermodel.FamilyFounderTransferRequest, error)
	Create(context.Context, *transfermodel.FamilyFounderTransferRequest) error
	FindByID(context.Context, uint64, bool) (*transfermodel.FamilyFounderTransferRequest, error)
	List(context.Context, ListQuery) ([]TransferRow, int64, error)
	UpdateRequest(context.Context, uint64, string, map[string]any) error
	UpdateLinkRole(context.Context, uint64, map[string]any) error
	UpdateFamily(context.Context, uint64, map[string]any) error
	CountActiveFounders(context.Context, uint64) (int64, error)
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

func (r *GormRepository) FindMember(ctx context.Context, familyID uint64, memberID uint64, lock bool) (*membermodel.FamilyMember, error) {
	var member membermodel.FamilyMember
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.Where("id = ? AND family_id = ?", memberID, familyID).First(&member).Error
	return &member, err
}

func (r *GormRepository) FindActiveLinkByUser(ctx context.Context, familyID uint64, userID uint64, lock bool) (*rolemodel.FamilyMemberUserLink, error) {
	return r.findActiveLink(ctx, "family_id = ? AND user_id = ? AND link_status = ?", []any{familyID, userID, "ACTIVE"}, lock)
}

func (r *GormRepository) FindActiveLinkByMember(ctx context.Context, familyID uint64, memberID uint64, lock bool) (*rolemodel.FamilyMemberUserLink, error) {
	return r.findActiveLink(ctx, "family_id = ? AND member_id = ? AND link_status = ?", []any{familyID, memberID, "ACTIVE"}, lock)
}

func (r *GormRepository) findActiveLink(ctx context.Context, where string, args []any, lock bool) (*rolemodel.FamilyMemberUserLink, error) {
	var link rolemodel.FamilyMemberUserLink
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.Where(where, args...).First(&link).Error
	return &link, err
}

func (r *GormRepository) FindPendingByFamily(ctx context.Context, familyID uint64) (*transfermodel.FamilyFounderTransferRequest, error) {
	var request transfermodel.FamilyFounderTransferRequest
	err := r.db.WithContext(ctx).Where("family_id = ? AND request_status = ?", familyID, "PENDING").Order("id DESC").First(&request).Error
	return &request, err
}

func (r *GormRepository) Create(ctx context.Context, request *transfermodel.FamilyFounderTransferRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

func (r *GormRepository) FindByID(ctx context.Context, id uint64, lock bool) (*transfermodel.FamilyFounderTransferRequest, error) {
	var request transfermodel.FamilyFounderTransferRequest
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.First(&request, id).Error
	return &request, err
}

func (r *GormRepository) List(ctx context.Context, query ListQuery) ([]TransferRow, int64, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Table("family_founder_transfer_requests AS t").Joins("JOIN families AS f ON f.id = t.family_id")
	if query.Status != "" {
		db = db.Where("t.request_status = ?", query.Status)
	}
	if query.FamilyID != 0 {
		db = db.Where("t.family_id = ?", query.FamilyID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []TransferRow
	err := db.Select("t.*, f.family_name").Order("t.created_at DESC, t.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error
	return rows, total, err
}

func (r *GormRepository) UpdateRequest(ctx context.Context, id uint64, current string, values map[string]any) error {
	values["updated_at"] = time.Now()
	result := r.db.WithContext(ctx).Model(&transfermodel.FamilyFounderTransferRequest{}).Where("id = ? AND request_status = ?", id, current).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) UpdateLinkRole(ctx context.Context, linkID uint64, values map[string]any) error {
	values["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(&rolemodel.FamilyMemberUserLink{}).Where("id = ?", linkID).Updates(values).Error
}

func (r *GormRepository) UpdateFamily(ctx context.Context, familyID uint64, values map[string]any) error {
	values["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(&familymodel.Family{}).Where("id = ?", familyID).Updates(values).Error
}

func (r *GormRepository) CountActiveFounders(ctx context.Context, familyID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&rolemodel.FamilyMemberUserLink{}).
		Where("family_id = ? AND link_status = ? AND family_role = ?", familyID, "ACTIVE", "FOUNDER").
		Count(&count).Error
	return count, err
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
