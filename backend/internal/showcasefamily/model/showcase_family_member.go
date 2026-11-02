package model

import (
	"time"

	"gorm.io/gorm"
)

type ShowcaseFamilyMember struct {
	ID               uint64         `gorm:"primaryKey;column:id"`
	ShowcaseFamilyID uint64         `gorm:"column:showcase_family_id;not null"`
	MemberType       string         `gorm:"column:member_type;size:40;not null;default:LINEAGE_MEMBER"`
	DisplayName      string         `gorm:"column:display_name;size:80;not null"`
	IdentityText     string         `gorm:"column:identity_text;size:120;not null"`
	Gender           string         `gorm:"column:gender;size:20;not null;default:UNKNOWN"`
	BirthYear        *int           `gorm:"column:birth_year"`
	SortOrder        int            `gorm:"column:sort_order;not null;default:0"`
	CreatedAt        time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (ShowcaseFamilyMember) TableName() string { return "showcase_family_members" }
