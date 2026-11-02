package model

import (
	"time"

	"gorm.io/gorm"
)

type ShowcaseFamilyRelationship struct {
	ID               uint64         `gorm:"primaryKey;column:id"`
	ShowcaseFamilyID uint64         `gorm:"column:showcase_family_id;not null"`
	FromMemberID     uint64         `gorm:"column:from_member_id;not null"`
	ToMemberID       uint64         `gorm:"column:to_member_id;not null"`
	RelationshipType string         `gorm:"column:relationship_type;size:20;not null"`
	CreatedAt        time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (ShowcaseFamilyRelationship) TableName() string { return "showcase_family_relationships" }
