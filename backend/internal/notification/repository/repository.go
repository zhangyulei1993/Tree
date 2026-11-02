package repository

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"tree/backend/internal/notification/model"
)

type DeliveryJobStore interface {
	CreateDeliveryJob(context.Context, uint64, string, string, []string) (uint64, error)
	ListReadyDeliveryJobs(context.Context, int) ([]model.DeliveryJob, error)
	MarkDeliverySent(context.Context, uint64) error
	MarkDeliveryRetry(context.Context, uint64, string, time.Time) error
}

type Repository interface {
	Create(context.Context, *model.Notification) error
	List(context.Context, uint64, int, int) ([]model.Notification, int64, error)
	CountUnread(context.Context, uint64) (int64, error)
	MarkRead(context.Context, uint64, uint64, time.Time) error
	MarkAllRead(context.Context, uint64, time.Time) error
	Exists(context.Context, uint64, string, uint64) (bool, error)
}

func (r *gormRepository) CountUnread(ctx context.Context, userID uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&total).Error
	return total, err
}

type gormRepository struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &gormRepository{db: db} }

func (r *gormRepository) Create(ctx context.Context, row *model.Notification) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *gormRepository) List(ctx context.Context, userID uint64, page, pageSize int) ([]model.Notification, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ?", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Notification
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}

func (r *gormRepository) MarkRead(ctx context.Context, userID, id uint64, now time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).Where("id = ? AND user_id = ? AND read_at IS NULL", id, userID).Update("read_at", now).Error
}

func (r *gormRepository) MarkAllRead(ctx context.Context, userID uint64, now time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ? AND read_at IS NULL", userID).Update("read_at", now).Error
}

func (r *gormRepository) Exists(ctx context.Context, userID uint64, relatedType string, relatedID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ? AND related_type = ? AND related_id = ?", userID, relatedType, relatedID).Count(&count).Error
	return count > 0, err
}

func (r *gormRepository) CreateDeliveryJob(ctx context.Context, userID uint64, kind, page string, values []string) (uint64, error) {
	payload, err := json.Marshal(values)
	if err != nil {
		return 0, err
	}
	row := model.DeliveryJob{UserID: userID, Kind: kind, ValuesJSON: payload, Status: "PENDING", NextAttemptAt: time.Now()}
	if page != "" {
		row.Page = &page
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *gormRepository) ListReadyDeliveryJobs(ctx context.Context, limit int) ([]model.DeliveryJob, error) {
	if limit <= 0 || limit > 256 {
		limit = 256
	}
	var rows []model.DeliveryJob
	err := r.db.WithContext(ctx).Where("status IN ? AND next_attempt_at <= ?", []string{"PENDING", "RETRY"}, time.Now()).Order("id").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *gormRepository) MarkDeliverySent(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.DeliveryJob{}).Where("id = ?", id).Updates(map[string]any{"status": "SENT", "sent_at": now, "updated_at": now}).Error
}

func (r *gormRepository) MarkDeliveryRetry(ctx context.Context, id uint64, message string, next time.Time) error {
	return r.db.WithContext(ctx).Model(&model.DeliveryJob{}).Where("id = ?", id).Updates(map[string]any{"status": "RETRY", "attempts": gorm.Expr("attempts + 1"), "last_error": message, "next_attempt_at": next, "updated_at": time.Now()}).Error
}
