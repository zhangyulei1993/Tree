package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	authmodel "tree/backend/internal/auth/model"
)

type AssistanceCodeRepository interface {
	Create(context.Context, *authmodel.AssistanceCode) error
	ListByOwner(context.Context, uint64, int) ([]authmodel.AssistanceCode, error)
	FindByHashForUpdate(context.Context, string) (*authmodel.AssistanceCode, error)
	FindByIDForUpdate(context.Context, uint64) (*authmodel.AssistanceCode, error)
	ReplaceActiveForOwner(context.Context, uint64, time.Time) ([]authmodel.AssistanceCode, error)
	Update(context.Context, uint64, map[string]any) error
	WithTx(*gorm.DB) AssistanceCodeRepository
}

type GormAssistanceCodeRepository struct{ db *gorm.DB }

func NewGormAssistanceCodeRepository(db *gorm.DB) *GormAssistanceCodeRepository {
	return &GormAssistanceCodeRepository{db: db}
}
func (r *GormAssistanceCodeRepository) WithTx(tx *gorm.DB) AssistanceCodeRepository {
	return &GormAssistanceCodeRepository{db: tx}
}
func (r *GormAssistanceCodeRepository) Create(ctx context.Context, code *authmodel.AssistanceCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}
func (r *GormAssistanceCodeRepository) ListByOwner(ctx context.Context, ownerID uint64, limit int) ([]authmodel.AssistanceCode, error) {
	var items []authmodel.AssistanceCode
	if limit <= 0 || limit > 20 {
		limit = 20
	}
	err := r.db.WithContext(ctx).Where("owner_user_id = ?", ownerID).Order("created_at DESC").Limit(limit).Find(&items).Error
	return items, err
}
func (r *GormAssistanceCodeRepository) FindByHashForUpdate(ctx context.Context, hash string) (*authmodel.AssistanceCode, error) {
	var item authmodel.AssistanceCode
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("code_hash = ?", hash).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
func (r *GormAssistanceCodeRepository) FindByIDForUpdate(ctx context.Context, id uint64) (*authmodel.AssistanceCode, error) {
	var item authmodel.AssistanceCode
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
func (r *GormAssistanceCodeRepository) ReplaceActiveForOwner(ctx context.Context, ownerID uint64, now time.Time) ([]authmodel.AssistanceCode, error) {
	var items []authmodel.AssistanceCode
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_user_id = ? AND code_status = ?", ownerID, authmodel.AssistanceCodeActive).Find(&items).Error; err != nil {
		return nil, err
	}
	if len(items) > 0 {
		if err := r.db.WithContext(ctx).Model(&authmodel.AssistanceCode{}).Where("owner_user_id = ? AND code_status = ?", ownerID, authmodel.AssistanceCodeActive).Updates(map[string]any{"code_status": authmodel.AssistanceCodeReplaced, "replaced_at": now}).Error; err != nil {
			return nil, err
		}
	}
	return items, nil
}
func (r *GormAssistanceCodeRepository) Update(ctx context.Context, id uint64, values map[string]any) error {
	return r.db.WithContext(ctx).Model(&authmodel.AssistanceCode{}).Where("id = ?", id).Updates(values).Error
}
