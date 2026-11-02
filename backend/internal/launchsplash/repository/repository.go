package repository

import (
	"context"

	"gorm.io/gorm"
	"tree/backend/internal/launchsplash/model"
)

type Repository interface {
	Find(context.Context, string) (*model.Config, error)
	Update(context.Context, string, map[string]any) error
}

type GormRepository struct{ db *gorm.DB }

func New(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (r *GormRepository) Find(ctx context.Context, key string) (*model.Config, error) {
	var row model.Config
	err := r.db.WithContext(ctx).Where("config_key = ?", key).First(&row).Error
	return &row, err
}

func (r *GormRepository) Update(ctx context.Context, key string, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&model.Config{}).Where("config_key = ?", key).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
