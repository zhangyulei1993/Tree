package model

import "time"

type VerificationCode struct {
	ID         uint64     `gorm:"primaryKey;column:id"`
	Phone      string     `gorm:"column:phone;size:30;not null"`
	PhoneHash  string     `gorm:"column:phone_hash;size:128;not null"`
	CodeHash   string     `gorm:"column:code_hash;size:255;not null"`
	Scene      string     `gorm:"column:scene;size:80;not null"`
	ClientType string     `gorm:"column:client_type;size:80;not null"`
	Status     string     `gorm:"column:status;size:40;not null;default:PENDING"`
	ExpiredAt  time.Time  `gorm:"column:expired_at;not null"`
	UsedAt     *time.Time `gorm:"column:used_at"`
	IP         *string    `gorm:"column:ip;size:80"`
	UserAgent  *string    `gorm:"column:user_agent;size:500"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (VerificationCode) TableName() string {
	return "verification_codes"
}

// TODO(service): store only code_hash; never store or log plain verification codes.
