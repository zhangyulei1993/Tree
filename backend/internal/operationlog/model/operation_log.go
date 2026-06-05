package model

import "time"

type OperationLog struct {
	ID              uint64    `gorm:"primaryKey;column:id"`
	OperatorType    string    `gorm:"column:operator_type;size:40;not null"`
	OperatorAdminID *uint64   `gorm:"column:operator_admin_id"`
	OperatorUserID  *uint64   `gorm:"column:operator_user_id"`
	OperatorRole    *string   `gorm:"column:operator_role;size:60"`
	Module          string    `gorm:"column:module;size:80;not null"`
	Action          string    `gorm:"column:action;size:100;not null"`
	TargetType      *string   `gorm:"column:target_type;size:80"`
	TargetID        *uint64   `gorm:"column:target_id"`
	FamilyID        *uint64   `gorm:"column:family_id"`
	MemberID        *uint64   `gorm:"column:member_id"`
	UserID          *uint64   `gorm:"column:user_id"`
	BeforeJSON      []byte    `gorm:"column:before_json;type:json"`
	AfterJSON       []byte    `gorm:"column:after_json;type:json"`
	DetailJSON      []byte    `gorm:"column:detail_json;type:json"`
	Result          string    `gorm:"column:result;size:30;not null;default:SUCCESS"`
	ErrorMessage    *string   `gorm:"column:error_message;size:500"`
	IP              *string   `gorm:"column:ip;size:80"`
	UserAgent       *string   `gorm:"column:user_agent;size:500"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
}

// TODO(service): filter sensitive fields before writing operation_logs.
