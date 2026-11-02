package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/user/model"
)

type VerificationCodeRepository interface {
	Create(ctx context.Context, code *model.VerificationCode) error
	FindLatestPending(ctx context.Context, phoneHash string, scene string) (*model.VerificationCode, error)
	FindLatestCreatedAfter(ctx context.Context, phoneHash string, scene string, after time.Time) (*model.VerificationCode, error)
	MarkUsed(ctx context.Context, id uint64, now time.Time) error
	WithTx(tx *gorm.DB) VerificationCodeRepository
}

type GormVerificationCodeRepository struct {
	db *gorm.DB
}

func NewGormVerificationCodeRepository(db *gorm.DB) *GormVerificationCodeRepository {
	return &GormVerificationCodeRepository{db: db}
}

func (r *GormVerificationCodeRepository) WithTx(tx *gorm.DB) VerificationCodeRepository {
	return &GormVerificationCodeRepository{db: tx}
}

func (r *GormVerificationCodeRepository) Create(ctx context.Context, code *model.VerificationCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

func (r *GormVerificationCodeRepository) FindLatestPending(ctx context.Context, phoneHash string, scene string) (*model.VerificationCode, error) {
	var code model.VerificationCode
	if err := r.db.WithContext(ctx).
		Where("phone_hash = ? AND scene = ? AND status = ?", phoneHash, scene, "PENDING").
		Order("id DESC").
		First(&code).Error; err != nil {
		return nil, err
	}
	return &code, nil
}

func (r *GormVerificationCodeRepository) FindLatestCreatedAfter(ctx context.Context, phoneHash string, scene string, after time.Time) (*model.VerificationCode, error) {
	var code model.VerificationCode
	if err := r.db.WithContext(ctx).
		Where("phone_hash = ? AND scene = ? AND created_at > ?", phoneHash, scene, after).
		Order("id DESC").
		First(&code).Error; err != nil {
		return nil, err
	}
	return &code, nil
}

func (r *GormVerificationCodeRepository) MarkUsed(ctx context.Context, id uint64, now time.Time) error {
	return r.db.WithContext(ctx).Model(&model.VerificationCode{}).Where("id = ?", id).Updates(map[string]any{
		"status":  "USED",
		"used_at": now,
	}).Error
}
