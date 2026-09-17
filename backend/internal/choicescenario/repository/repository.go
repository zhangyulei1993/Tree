package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	choicemodel "tree/backend/internal/choicescenario/model"
	operationlog "tree/backend/internal/operationlog/service"
)

type Repository interface {
	WithTx(*gorm.DB) Repository
	Create(context.Context, *choicemodel.ChoiceScenario) error
	List(context.Context, time.Time, time.Time, int) ([]choicemodel.ChoiceScenario, error)
	CountCreatedByUserSince(context.Context, uint64, time.Time) (int64, error)
	IncrementUseCount(context.Context, uint64) error
	WriteLog(context.Context, operationlog.WriteInput) error
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (r *GormRepository) WithTx(tx *gorm.DB) Repository { return &GormRepository{db: tx} }

func (r *GormRepository) Create(ctx context.Context, scenario *choicemodel.ChoiceScenario) error {
	return r.db.WithContext(ctx).Create(scenario).Error
}

func (r *GormRepository) List(ctx context.Context, from, now time.Time, limit int) ([]choicemodel.ChoiceScenario, error) {
	var rows []choicemodel.ChoiceScenario
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at >= ? AND created_at <= ? AND expires_at > ?", "ACTIVE", from, now, now).
		Order("created_at DESC, id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *GormRepository) CountCreatedByUserSince(ctx context.Context, userID uint64, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&choicemodel.ChoiceScenario{}).
		Where("user_id = ? AND created_at >= ?", userID, since).Count(&count).Error
	return count, err
}

func (r *GormRepository) IncrementUseCount(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Model(&choicemodel.ChoiceScenario{}).
		Where("id = ? AND status = ?", id, "ACTIVE").
		UpdateColumn("use_count", gorm.Expr("use_count + 1"))
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
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(u.repo.WithTx(tx))
	})
}
