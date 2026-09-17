package repository

import (
	"context"

	"gorm.io/gorm"
	sharemodel "tree/backend/internal/shareconfig/model"
)

type Repository interface {
	List(context.Context) ([]sharemodel.ShareConfig, error)
	FindByKey(context.Context, string) (*sharemodel.ShareConfig, error)
	Update(context.Context, string, map[string]any) error
}

type GormRepository struct{ db *gorm.DB }

func New(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (r *GormRepository) List(ctx context.Context) ([]sharemodel.ShareConfig, error) {
	var rows []sharemodel.ShareConfig
	err := r.db.WithContext(ctx).Order("id ASC").Find(&rows).Error
	return rows, err
}

func (r *GormRepository) FindByKey(ctx context.Context, key string) (*sharemodel.ShareConfig, error) {
	var row sharemodel.ShareConfig
	err := r.db.WithContext(ctx).Where("config_key = ?", key).First(&row).Error
	return &row, err
}

func (r *GormRepository) Update(ctx context.Context, key string, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&sharemodel.ShareConfig{}).Where("config_key = ?", key).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
