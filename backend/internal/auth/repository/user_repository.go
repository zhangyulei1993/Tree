package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/user/model"
)

type UserRepository interface {
	FindByPhoneHash(ctx context.Context, phoneHash string) (*model.User, error)
	FindByID(ctx context.Context, id uint64) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	RecordLoginSuccess(ctx context.Context, userID uint64, ip string, clientType string, now time.Time) error
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) FindByPhoneHash(ctx context.Context, phoneHash string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Where("phone_hash = ? AND deleted_at IS NULL", phoneHash).
		Order("id DESC").
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *GormUserRepository) RecordLoginSuccess(ctx context.Context, userID uint64, ip string, clientType string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{
		"last_login_at":     now,
		"last_login_ip":     ip,
		"last_login_client": clientType,
		"updated_at":        now,
	}).Error
}
