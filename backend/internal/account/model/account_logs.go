package model

import "time"

type UserPhoneHistory struct {
	ID              uint64    `gorm:"primaryKey;column:id"`
	UserID          uint64    `gorm:"column:user_id;not null"`
	Phone           *string   `gorm:"column:phone;size:30"`
	PhoneHash       *string   `gorm:"column:phone_hash;size:128"`
	ActionType      string    `gorm:"column:action_type;size:80;not null"`
	OperatorType    *string   `gorm:"column:operator_type;size:40"`
	OperatorUserID  *uint64   `gorm:"column:operator_user_id"`
	OperatorAdminID *uint64   `gorm:"column:operator_admin_id"`
	DetailJSON      []byte    `gorm:"column:detail_json;type:json"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (UserPhoneHistory) TableName() string {
	return "user_phone_history"
}

type UserAccountMergeLog struct {
	ID                          uint64    `gorm:"primaryKey;column:id"`
	SourceUserID                uint64    `gorm:"column:source_user_id;not null"`
	TargetUserID                uint64    `gorm:"column:target_user_id;not null"`
	MergeReason                 *string   `gorm:"column:merge_reason;size:500"`
	MergeSource                 string    `gorm:"column:merge_source;size:80;not null"`
	MemberBindingResolutionJSON []byte    `gorm:"column:member_binding_resolution_json;type:json"`
	BeforeJSON                  []byte    `gorm:"column:before_json;type:json"`
	AfterJSON                   []byte    `gorm:"column:after_json;type:json"`
	OperatorType                *string   `gorm:"column:operator_type;size:40"`
	OperatorUserID              *uint64   `gorm:"column:operator_user_id"`
	OperatorAdminID             *uint64   `gorm:"column:operator_admin_id"`
	CreatedAt                   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (UserAccountMergeLog) TableName() string {
	return "user_account_merge_logs"
}

type UserAccountClaimLog struct {
	ID            uint64    `gorm:"primaryKey;column:id"`
	ClaimedUserID uint64    `gorm:"column:claimed_user_id;not null"`
	TempUserID    *uint64   `gorm:"column:temp_user_id"`
	Phone         *string   `gorm:"column:phone;size:30"`
	PhoneHash     *string   `gorm:"column:phone_hash;size:128"`
	ClaimVia      string    `gorm:"column:claim_via;size:80;not null"`
	DetailJSON    []byte    `gorm:"column:detail_json;type:json"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (UserAccountClaimLog) TableName() string {
	return "user_account_claim_logs"
}

// TODO(service): account merge and account claim must be transactional and must write operation_logs.
