package model

import "time"

type FamilyRelationship struct {
	ID               uint64     `gorm:"primaryKey;column:id"`
	FamilyID         uint64     `gorm:"column:family_id;not null"`
	FromMemberID     uint64     `gorm:"column:from_member_id;not null"`
	ToMemberID       uint64     `gorm:"column:to_member_id;not null"`
	RelationshipType string     `gorm:"column:relationship_type;size:40;not null"`
	ParentLinkType   *string    `gorm:"column:parent_link_type;size:40"`
	RelationNoteType *string    `gorm:"column:relation_note_type;size:80"`
	RelationNote     *string    `gorm:"column:relation_note;size:500"`
	Status           string     `gorm:"column:status;size:40;not null;default:ACTIVE"`
	CreatedByUserID  *uint64    `gorm:"column:created_by_user_id"`
	CreatedByAdminID *uint64    `gorm:"column:created_by_admin_id"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
	DeletedByUserID  *uint64    `gorm:"column:deleted_by_user_id"`
	DeletedByAdminID *uint64    `gorm:"column:deleted_by_admin_id"`
	DeleteReason     *string    `gorm:"column:delete_reason;size:500"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (FamilyRelationship) TableName() string {
	return "family_relationships"
}

// TODO(service): allow only PARENT_CHILD and SPOUSE; never persist SIBLING.
// TODO(service): enforce PRIMARY father/mother uniqueness for each child.
