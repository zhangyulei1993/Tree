package repository

import (
	"context"

	"gorm.io/gorm"
	showcasemodel "tree/backend/internal/showcasefamily/model"
)

type Repository interface {
	List(context.Context) ([]showcasemodel.ShowcaseFamily, error)
	ListPublished(context.Context) ([]showcasemodel.ShowcaseFamily, error)
	FindByID(context.Context, uint64) (*showcasemodel.ShowcaseFamily, error)
	Create(context.Context, *showcasemodel.ShowcaseFamily) error
	CreateWithRoot(context.Context, *showcasemodel.ShowcaseFamily, *showcasemodel.ShowcaseFamilyMember) error
	Update(context.Context, uint64, map[string]any) error
	Delete(context.Context, uint64) error
	DeleteWithTree(context.Context, uint64) error
	ListMembers(context.Context, uint64) ([]showcasemodel.ShowcaseFamilyMember, error)
	ListRelationships(context.Context, uint64) ([]showcasemodel.ShowcaseFamilyRelationship, error)
	FindMemberByID(context.Context, uint64, uint64) (*showcasemodel.ShowcaseFamilyMember, error)
	CreateMember(context.Context, *showcasemodel.ShowcaseFamilyMember) error
	UpdateMember(context.Context, uint64, uint64, map[string]any) error
	DeleteMember(context.Context, uint64, uint64) error
	CreateRelationship(context.Context, *showcasemodel.ShowcaseFamilyRelationship) error
	DeleteRelationship(context.Context, uint64, uint64) error
	CreateRelative(context.Context, uint64, uint64, string, *showcasemodel.ShowcaseFamilyMember, uint64, *showcasemodel.ShowcaseFamilyMember) error
	GetDisplaySettings(context.Context) (*showcasemodel.ShowcaseDisplaySettings, error)
	UpdateDisplaySettings(context.Context, map[string]any) error
}

type GormRepository struct{ db *gorm.DB }

func New(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (r *GormRepository) List(ctx context.Context) ([]showcasemodel.ShowcaseFamily, error) {
	var rows []showcasemodel.ShowcaseFamily
	err := r.db.WithContext(ctx).Order("status ASC, sort_order ASC, id DESC").Find(&rows).Error
	return rows, err
}

func (r *GormRepository) ListPublished(ctx context.Context) ([]showcasemodel.ShowcaseFamily, error) {
	var rows []showcasemodel.ShowcaseFamily
	err := r.db.WithContext(ctx).Where("status = ?", "PUBLISHED").Order("sort_order ASC, id DESC").Find(&rows).Error
	return rows, err
}

func (r *GormRepository) FindByID(ctx context.Context, id uint64) (*showcasemodel.ShowcaseFamily, error) {
	var row showcasemodel.ShowcaseFamily
	err := r.db.WithContext(ctx).First(&row, id).Error
	return &row, err
}

func (r *GormRepository) Create(ctx context.Context, row *showcasemodel.ShowcaseFamily) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *GormRepository) CreateWithRoot(ctx context.Context, family *showcasemodel.ShowcaseFamily, root *showcasemodel.ShowcaseFamilyMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(family).Error; err != nil {
			return err
		}
		root.ShowcaseFamilyID = family.ID
		return tx.Create(root).Error
	})
}

func (r *GormRepository) Update(ctx context.Context, id uint64, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&showcasemodel.ShowcaseFamily{}).Where("id = ?", id).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&showcasemodel.ShowcaseFamily{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) DeleteWithTree(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("showcase_family_id = ?", id).Delete(&showcasemodel.ShowcaseFamilyRelationship{})
		if result.Error != nil {
			return result.Error
		}
		result = tx.Where("showcase_family_id = ?", id).Delete(&showcasemodel.ShowcaseFamilyMember{})
		if result.Error != nil {
			return result.Error
		}
		result = tx.Delete(&showcasemodel.ShowcaseFamily{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *GormRepository) ListMembers(ctx context.Context, showcaseFamilyID uint64) ([]showcasemodel.ShowcaseFamilyMember, error) {
	var rows []showcasemodel.ShowcaseFamilyMember
	err := r.db.WithContext(ctx).Where("showcase_family_id = ?", showcaseFamilyID).Order("sort_order ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *GormRepository) ListRelationships(ctx context.Context, showcaseFamilyID uint64) ([]showcasemodel.ShowcaseFamilyRelationship, error) {
	var rows []showcasemodel.ShowcaseFamilyRelationship
	err := r.db.WithContext(ctx).Where("showcase_family_id = ?", showcaseFamilyID).Order("id ASC").Find(&rows).Error
	return rows, err
}

func (r *GormRepository) FindMemberByID(ctx context.Context, showcaseFamilyID, memberID uint64) (*showcasemodel.ShowcaseFamilyMember, error) {
	var row showcasemodel.ShowcaseFamilyMember
	err := r.db.WithContext(ctx).Where("showcase_family_id = ?", showcaseFamilyID).First(&row, memberID).Error
	return &row, err
}

func (r *GormRepository) CreateMember(ctx context.Context, row *showcasemodel.ShowcaseFamilyMember) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *GormRepository) UpdateMember(ctx context.Context, showcaseFamilyID, memberID uint64, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&showcasemodel.ShowcaseFamilyMember{}).Where("showcase_family_id = ? AND id = ?", showcaseFamilyID, memberID).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) DeleteMember(ctx context.Context, showcaseFamilyID, memberID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("showcase_family_id = ? AND id = ?", showcaseFamilyID, memberID).Delete(&showcasemodel.ShowcaseFamilyMember{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Where("showcase_family_id = ? AND (from_member_id = ? OR to_member_id = ?)", showcaseFamilyID, memberID, memberID).Delete(&showcasemodel.ShowcaseFamilyRelationship{}).Error
	})
}

func (r *GormRepository) CreateRelationship(ctx context.Context, row *showcasemodel.ShowcaseFamilyRelationship) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *GormRepository) DeleteRelationship(ctx context.Context, showcaseFamilyID, relationshipID uint64) error {
	result := r.db.WithContext(ctx).Where("showcase_family_id = ?", showcaseFamilyID).Delete(&showcasemodel.ShowcaseFamilyRelationship{}, relationshipID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) CreateRelative(ctx context.Context, showcaseFamilyID, anchorMemberID uint64, relationKind string, member *showcasemodel.ShowcaseFamilyMember, existingSpouseMemberID uint64, newSpouse *showcasemodel.ShowcaseFamilyMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(member).Error; err != nil {
			return err
		}
		otherParentID := existingSpouseMemberID
		if newSpouse != nil {
			if err := tx.Create(newSpouse).Error; err != nil {
				return err
			}
			otherParentID = newSpouse.ID
			if err := tx.Create(&showcasemodel.ShowcaseFamilyRelationship{ShowcaseFamilyID: showcaseFamilyID, FromMemberID: anchorMemberID, ToMemberID: otherParentID, RelationshipType: "SPOUSE"}).Error; err != nil {
				return err
			}
		}
		if relationKind == "SPOUSE" {
			return tx.Create(&showcasemodel.ShowcaseFamilyRelationship{ShowcaseFamilyID: showcaseFamilyID, FromMemberID: anchorMemberID, ToMemberID: member.ID, RelationshipType: "SPOUSE"}).Error
		}
		if err := tx.Create(&showcasemodel.ShowcaseFamilyRelationship{ShowcaseFamilyID: showcaseFamilyID, FromMemberID: anchorMemberID, ToMemberID: member.ID, RelationshipType: "PARENT_CHILD"}).Error; err != nil {
			return err
		}
		if otherParentID != 0 {
			return tx.Create(&showcasemodel.ShowcaseFamilyRelationship{ShowcaseFamilyID: showcaseFamilyID, FromMemberID: otherParentID, ToMemberID: member.ID, RelationshipType: "PARENT_CHILD"}).Error
		}
		return nil
	})
}

func (r *GormRepository) GetDisplaySettings(ctx context.Context) (*showcasemodel.ShowcaseDisplaySettings, error) {
	row := &showcasemodel.ShowcaseDisplaySettings{}
	err := r.db.WithContext(ctx).First(row, 1).Error
	return row, err
}

func (r *GormRepository) UpdateDisplaySettings(ctx context.Context, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&showcasemodel.ShowcaseDisplaySettings{}).Where("id = ?", 1).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
