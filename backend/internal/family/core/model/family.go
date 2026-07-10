package model

import "time"

type Family struct {
	ID                     uint64     `gorm:"primaryKey;column:id"`
	FamilyName             string     `gorm:"column:family_name;size:150;not null"`
	FamilySurname          string     `gorm:"column:family_surname;size:40;not null"`
	NativePlace            *string    `gorm:"column:native_place;size:200"`
	RegionCode             *string    `gorm:"column:region_code;size:50"`
	RegionText             *string    `gorm:"column:region_text;size:200"`
	Description            *string    `gorm:"column:description;type:text"`
	AvatarURL              *string    `gorm:"column:avatar_url;size:500"`
	CreatorUserID          *uint64    `gorm:"column:creator_user_id"`
	CurrentFounderMemberID *uint64    `gorm:"column:current_founder_member_id"`
	Status                 string     `gorm:"column:status;size:40;not null;default:NORMAL"`
	Searchable             bool       `gorm:"column:searchable;not null;default:1"`
	PublicDisplayStatus    string     `gorm:"column:public_display_status;size:40;not null;default:PRIVATE"`
	PublicDisplayEnabled   bool       `gorm:"column:public_display_enabled;not null;default:0"`
	PublicContactName      *string    `gorm:"column:public_contact_name;size:100"`
	PublicContactPhone     *string    `gorm:"column:public_contact_phone;size:30"`
	PublicContactWechat    *string    `gorm:"column:public_contact_wechat;size:120"`
	PublicContactNote      *string    `gorm:"column:public_contact_note;size:500"`
	PublicContactVisible   bool       `gorm:"column:public_contact_visible;not null;default:1"`
	TreeMode               string     `gorm:"column:tree_mode;size:40;not null;default:LIST_TREE"`
	GraphVersion           int64      `gorm:"column:graph_version;not null;default:1"`
	PublicAppliedAt        *time.Time `gorm:"column:public_applied_at"`
	PublicApprovedAt       *time.Time `gorm:"column:public_approved_at"`
	PublicEnabledAt        *time.Time `gorm:"column:public_enabled_at"`
	PublicTakenDownAt      *time.Time `gorm:"column:public_taken_down_at"`
	DisabledAt             *time.Time `gorm:"column:disabled_at"`
	DisabledByAdminID      *uint64    `gorm:"column:disabled_by_admin_id"`
	DisabledReason         *string    `gorm:"column:disabled_reason;size:500"`
	DissolvedAt            *time.Time `gorm:"column:dissolved_at"`
	RestoredAt             *time.Time `gorm:"column:restored_at"`
	DeletedAt              *time.Time `gorm:"column:deleted_at"`
	CreatedAt              time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (Family) TableName() string {
	return "families"
}

// TODO(service): graph_version must increment for member/relationship structure changes only.
