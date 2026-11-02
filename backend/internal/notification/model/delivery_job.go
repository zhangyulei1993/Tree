package model

import "time"

type DeliveryJob struct {
	ID            uint64     `gorm:"primaryKey;column:id"`
	UserID        uint64     `gorm:"column:user_id"`
	Kind          string     `gorm:"column:kind"`
	Page          *string    `gorm:"column:page"`
	ValuesJSON    []byte     `gorm:"column:values_json"`
	Status        string     `gorm:"column:status"`
	Attempts      int        `gorm:"column:attempts"`
	NextAttemptAt time.Time  `gorm:"column:next_attempt_at"`
	LastError     *string    `gorm:"column:last_error"`
	SentAt        *time.Time `gorm:"column:sent_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (DeliveryJob) TableName() string { return "notification_delivery_jobs" }
