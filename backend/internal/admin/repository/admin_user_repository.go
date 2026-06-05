package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/admin/model"
)

type AdminUserRepository interface {
	FindByUsername(ctx context.Context, username string) (*model.AdminUser, error)
	FindByID(ctx context.Context, id uint64) (*model.AdminUser, error)
	RecordLoginSuccess(ctx context.Context, adminID uint64, ip string, now time.Time) error
	RecordLoginFailure(ctx context.Context, adminID uint64, failedCount int, lockedUntil *time.Time) error
	UpdatePassword(ctx context.Context, adminID uint64, passwordHash string, now time.Time) error
}

type GormAdminUserRepository struct {
	db *gorm.DB
}

func NewGormAdminUserRepository(db *gorm.DB) *GormAdminUserRepository {
	return &GormAdminUserRepository{db: db}
}

func (r *GormAdminUserRepository) FindByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	var admin model.AdminUser
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *GormAdminUserRepository) FindByID(ctx context.Context, id uint64) (*model.AdminUser, error) {
	var admin model.AdminUser
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *GormAdminUserRepository) RecordLoginSuccess(ctx context.Context, adminID uint64, ip string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", adminID).Updates(map[string]any{
		"failed_login_count": 0,
		"locked_until":       nil,
		"last_login_at":      now,
		"last_login_ip":      ip,
	}).Error
}

func (r *GormAdminUserRepository) RecordLoginFailure(ctx context.Context, adminID uint64, failedCount int, lockedUntil *time.Time) error {
	return r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", adminID).Updates(map[string]any{
		"failed_login_count": failedCount,
		"locked_until":       lockedUntil,
	}).Error
}

func (r *GormAdminUserRepository) UpdatePassword(ctx context.Context, adminID uint64, passwordHash string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", adminID).Updates(map[string]any{
		"password_hash":       passwordHash,
		"password_changed_at": now,
		"failed_login_count":  0,
		"locked_until":        nil,
		"updated_at":          now,
	}).Error
}
