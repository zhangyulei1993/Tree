package repository

import (
	"context"

	"gorm.io/gorm"
	backupmodel "tree/backend/internal/backupconfig/model"
)

type Repository interface {
	Find(context.Context) (*backupmodel.Config, error)
	Update(context.Context, map[string]any) error
}

type GormRepository struct{ db *gorm.DB }

func New(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (r *GormRepository) Find(ctx context.Context) (*backupmodel.Config, error) {
	var row backupmodel.Config
	err := r.db.WithContext(ctx).First(&row, 1).Error
	return &row, err
}

func (r *GormRepository) Update(ctx context.Context, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&backupmodel.Config{}).Where("id = ?", 1).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
