package model

import "time"

type VisitorMessage struct {
	ID                uint64     `gorm:"primaryKey;column:id"`
	FamilyID          uint64     `gorm:"column:family_id;not null"`
	VisitorUserID     *uint64    `gorm:"column:visitor_user_id"`
	VisitorName       *string    `gorm:"column:visitor_name;size:100"`
	VisitorPhone      *string    `gorm:"column:visitor_phone;size:30"`
	VisitorWechat     *string    `gorm:"column:visitor_wechat;size:120"`
	MessageContent    string     `gorm:"column:message_content;size:1000;not null"`
	Status            string     `gorm:"column:status;size:40;not null;default:PENDING"`
	ReviewedByAdminID *uint64    `gorm:"column:reviewed_by_admin_id"`
	ReviewedAt        *time.Time `gorm:"column:reviewed_at"`
	ReviewComment     *string    `gorm:"column:review_comment;size:500"`
	DeletedAt         *time.Time `gorm:"column:deleted_at"`
	DeletedByAdminID  *uint64    `gorm:"column:deleted_by_admin_id"`
	DeleteReason      *string    `gorm:"column:delete_reason;size:500"`
	IP                *string    `gorm:"column:ip;size:80"`
	UserAgent         *string    `gorm:"column:user_agent;size:500"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (VisitorMessage) TableName() string {
	return "visitor_messages"
}

// TODO(service): public message lists must hide visitor_phone and visitor_wechat.
