package model

import "time"

type ChoiceScenario struct {
	ID          uint64    `gorm:"primaryKey;column:id"`
	UserID      uint64    `gorm:"column:user_id;not null"`
	Title       string    `gorm:"column:title;size:60;not null"`
	OptionsJSON []byte    `gorm:"column:options_json;type:json;not null"`
	Status      string    `gorm:"column:status;size:30;not null;default:ACTIVE"`
	UseCount    uint      `gorm:"column:use_count;not null;default:0"`
	ExpiresAt   time.Time `gorm:"column:expires_at;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ChoiceScenario) TableName() string { return "choice_scenarios" }
