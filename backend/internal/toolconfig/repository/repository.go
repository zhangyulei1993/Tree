package repository

import (
	"context"

	"gorm.io/gorm"
	toolmodel "tree/backend/internal/toolconfig/model"
)

type Repository interface {
	List(context.Context) ([]toolmodel.ToolConfig, error)
	FindByKey(context.Context, string) (*toolmodel.ToolConfig, error)
	Create(context.Context, *toolmodel.ToolConfig) error
	Update(context.Context, string, map[string]any) error
}

type GormRepository struct{ db *gorm.DB }

func New(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (r *GormRepository) List(ctx context.Context) ([]toolmodel.ToolConfig, error) {
	var rows []toolmodel.ToolConfig
	err := r.db.WithContext(ctx).Order("pinned DESC, sort_order ASC, highlighted DESC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *GormRepository) FindByKey(ctx context.Context, key string) (*toolmodel.ToolConfig, error) {
	var row toolmodel.ToolConfig
	err := r.db.WithContext(ctx).Where("tool_key = ?", key).First(&row).Error
	return &row, err
}

func (r *GormRepository) Create(ctx context.Context, row *toolmodel.ToolConfig) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *GormRepository) Update(ctx context.Context, key string, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&toolmodel.ToolConfig{}).Where("tool_key = ?", key).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
