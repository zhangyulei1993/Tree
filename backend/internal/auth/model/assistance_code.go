package model

import "time"

const (
	AssistanceCodeActive   = "ACTIVE"
	AssistanceCodeReplaced = "REPLACED"
	AssistanceCodeRevoked  = "REVOKED"
	AssistanceCodeExpired  = "EXPIRED"
	AssistanceCodeLocked   = "LOCKED"
)

// AssistanceCode is a short lived delegated-maintenance credential. Its numeric
// value is never persisted: code_hash is an HMAC derived with a server secret.
type AssistanceCode struct {
	ID             uint64     `gorm:"primaryKey;column:id"`
	OwnerUserID    uint64     `gorm:"column:owner_user_id;not null"`
	CodeHash       string     `gorm:"column:code_hash;size:128;not null"`
	CodeStatus     string     `gorm:"column:code_status;size:32;not null"`
	ExpiresAt      time.Time  `gorm:"column:expires_at;not null"`
	ReplacedAt     *time.Time `gorm:"column:replaced_at"`
	RevokedAt      *time.Time `gorm:"column:revoked_at"`
	FailedAttempts int        `gorm:"column:failed_attempts;not null"`
	LockedAt       *time.Time `gorm:"column:locked_at"`
	SessionTokenID *string    `gorm:"column:session_token_id;size:128"`
	LastUsedAt     *time.Time `gorm:"column:last_used_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (AssistanceCode) TableName() string { return "user_assistance_codes" }
