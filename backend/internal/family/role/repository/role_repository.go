package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	familymodel "tree/backend/internal/family/core/model"
	membermodel "tree/backend/internal/family/member/model"
	rolemodel "tree/backend/internal/family/role/model"
	operationlog "tree/backend/internal/operationlog/service"
)

type Repository interface {
	WithTx(*gorm.DB) Repository
	FindFamily(context.Context, uint64, bool) (*familymodel.Family, error)
	FindMember(context.Context, uint64, uint64, bool) (*membermodel.FamilyMember, error)
	FindActiveLinkByUser(context.Context, uint64, uint64, bool) (*rolemodel.FamilyMemberUserLink, error)
	FindActiveLinkByMember(context.Context, uint64, uint64, bool) (*rolemodel.FamilyMemberUserLink, error)
	UpdateLinkRole(context.Context, uint64, map[string]any) error
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

func (r *GormRepository) UpdateLinkRole(ctx context.Context, linkID uint64, values map[string]any) error {
	values["updated_at"] = time.Now()
	result := r.db.WithContext(ctx).Model(&rolemodel.FamilyMemberUserLink{}).Where("id = ?", linkID).Updates(values)
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
