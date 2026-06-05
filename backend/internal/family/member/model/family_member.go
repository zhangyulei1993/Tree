package model

import "time"

type FamilyMember struct {
	ID                    uint64     `gorm:"primaryKey;column:id"`
	FamilyID              uint64     `gorm:"column:family_id;not null"`
	MemberType            string     `gorm:"column:member_type;size:40;not null;default:LINEAGE_MEMBER"`
	Surname               *string    `gorm:"column:surname;size:40"`
	GenerationCharacter   *string    `gorm:"column:generation_character;size:40"`
	GivenName             *string    `gorm:"column:given_name;size:100"`
	DisplayName           string     `gorm:"column:display_name;size:150;not null"`
	Gender                string     `gorm:"column:gender;size:40;not null;default:UNKNOWN"`
	BirthDate             *time.Time `gorm:"column:birth_date;type:date"`
	DeathDate             *time.Time `gorm:"column:death_date;type:date"`
	BirthOrder            *int       `gorm:"column:birth_order"`
	ManualOrder           *int       `gorm:"column:manual_order"`
	NativePlace           *string    `gorm:"column:native_place;size:200"`
	RegionText            *string    `gorm:"column:region_text;size:200"`
	IsLiving              *bool      `gorm:"column:is_living"`
	UserBindingPolicy     string     `gorm:"column:user_binding_policy;size:40;not null;default:OPTIONAL"`
	UnboundReason         *string    `gorm:"column:unbound_reason;size:80"`
	UnboundNote           *string    `gorm:"column:unbound_note;size:500"`
	ManualLineageOverride bool       `gorm:"column:manual_lineage_override;not null;default:0"`
	LineageNoteType       *string    `gorm:"column:lineage_note_type;size:80"`
	LineageNote           *string    `gorm:"column:lineage_note;size:500"`
	IsTerminalNode        bool       `gorm:"column:is_terminal_node;not null;default:0"`
	TerminalReason        *string    `gorm:"column:terminal_reason;size:100"`
	Status                string     `gorm:"column:status;size:40;not null;default:ACTIVE"`
	CreatedByUserID       *uint64    `gorm:"column:created_by_user_id"`
	CreatedByAdminID      *uint64    `gorm:"column:created_by_admin_id"`
	DeletedAt             *time.Time `gorm:"column:deleted_at"`
	DeletedByUserID       *uint64    `gorm:"column:deleted_by_user_id"`
	DeletedByAdminID      *uint64    `gorm:"column:deleted_by_admin_id"`
	DeleteReason          *string    `gorm:"column:delete_reason;size:500"`
	CreatedAt             time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt             time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (FamilyMember) TableName() string {
	return "family_members"
}

// TODO(service): members with descendants cannot be deleted directly.
// TODO(service): NOT_REQUIRED members cannot be invited, bound, or set as FAMILY_ADMIN.
