package repository

import (
	"context"

	"gorm.io/gorm"

	familymodel "tree/backend/internal/family/core/model"
	membermodel "tree/backend/internal/family/member/model"
	relationshipmodel "tree/backend/internal/family/relationship/model"
)

type MemberRow struct {
	membermodel.FamilyMember
	HasActiveLink bool `gorm:"column:has_active_link"`
}

type Snapshot struct {
	Family        familymodel.Family
	Members       []MemberRow
	Relationships []relationshipmodel.FamilyRelationship
}

type Repository interface {
	FindFamily(context.Context, uint64) (*familymodel.Family, error)
	LoadSnapshot(context.Context, uint64) (*Snapshot, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindFamily(ctx context.Context, familyID uint64) (*familymodel.Family, error) {
	var family familymodel.Family
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", familyID).
		First(&family).Error
	return &family, err
}

func (r *GormRepository) LoadSnapshot(ctx context.Context, familyID uint64) (*Snapshot, error) {
	var snapshot Snapshot
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND deleted_at IS NULL", familyID).
			First(&snapshot.Family).Error; err != nil {
			return err
		}
		if err := tx.Table("family_members AS fm").
			Select("fm.*, EXISTS(SELECT 1 FROM family_member_user_links AS links WHERE links.family_id = fm.family_id AND links.member_id = fm.id AND links.link_status = ?) AS has_active_link", "ACTIVE").
			Where("fm.family_id = ? AND fm.status = ? AND fm.deleted_at IS NULL", familyID, "ACTIVE").
			Order("fm.manual_order IS NULL, fm.manual_order, fm.id").
			Scan(&snapshot.Members).Error; err != nil {
			return err
		}
		return tx.Where("family_id = ? AND status = ? AND deleted_at IS NULL", familyID, "ACTIVE").
			Where("relationship_type IN ?", []string{"PARENT_CHILD", "SPOUSE"}).
			Order("id").
			Find(&snapshot.Relationships).Error
	})
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}
