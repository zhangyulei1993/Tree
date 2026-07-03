package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"tree/backend/internal/common/enums"
	familymodel "tree/backend/internal/family/core/model"
	membermodel "tree/backend/internal/family/member/model"
	relationshipmodel "tree/backend/internal/family/relationship/model"
	operationlog "tree/backend/internal/operationlog/service"
)

type SpouseRelationshipRow struct {
	Relationship relationshipmodel.FamilyRelationship
	Spouse       membermodel.FamilyMember
}

type Repository interface {
	WithTx(*gorm.DB) Repository
	DB() *gorm.DB
	LockFamily(context.Context, uint64) (*familymodel.Family, error)
	FindMemberForUpdate(context.Context, uint64, uint64) (*membermodel.FamilyMember, error)
	CreateMember(context.Context, *membermodel.FamilyMember) error
	UpdateMember(context.Context, uint64, uint64, map[string]any) error
	FindActiveRelationship(context.Context, uint64, uint64) (*relationshipmodel.FamilyRelationship, error)
	HasActiveRelationships(context.Context, uint64, uint64) (bool, error)
	FindDuplicate(context.Context, uint64, uint64, uint64, string) (*relationshipmodel.FamilyRelationship, error)
	ListActiveParents(context.Context, uint64, uint64) ([]relationshipmodel.FamilyRelationship, error)
	FindPrimaryParentByGender(context.Context, uint64, uint64, string, uint64) (*relationshipmodel.FamilyRelationship, error)
	FindPrimaryParentWithMemberByGender(context.Context, uint64, uint64, string, uint64) (*relationshipmodel.FamilyRelationship, *membermodel.FamilyMember, error)
	ListActiveSpouseMembersByGender(context.Context, uint64, uint64, string) ([]membermodel.FamilyMember, error)
	ListActiveSpouseRelationshipsByGender(context.Context, uint64, uint64, string) ([]SpouseRelationshipRow, error)
	CreateRelationship(context.Context, *relationshipmodel.FamilyRelationship) error
	UpdateRelationship(context.Context, uint64, uint64, map[string]any) error
	SoftDeleteRelationship(context.Context, uint64, uint64, uint64, *string, time.Time) error
	IncrementGraphVersion(context.Context, uint64) (int64, error)
	WriteOperationLog(context.Context, operationlog.WriteInput) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) WithTx(tx *gorm.DB) Repository {
	return &GormRepository{db: tx}
}

func (r *GormRepository) DB() *gorm.DB {
	return r.db
}

func (r *GormRepository) LockFamily(ctx context.Context, familyID uint64) (*familymodel.Family, error) {
	var family familymodel.Family
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND deleted_at IS NULL", familyID).
		First(&family).Error
	return &family, err
}

func (r *GormRepository) FindMemberForUpdate(ctx context.Context, familyID uint64, memberID uint64) (*membermodel.FamilyMember, error) {
	var member membermodel.FamilyMember
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND family_id = ? AND status = ? AND deleted_at IS NULL", memberID, familyID, string(enums.StatusActive)).
		First(&member).Error
	return &member, err
}

func (r *GormRepository) CreateMember(ctx context.Context, member *membermodel.FamilyMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *GormRepository) UpdateMember(ctx context.Context, familyID uint64, memberID uint64, values map[string]any) error {
	if len(values) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&membermodel.FamilyMember{}).
		Where("id = ? AND family_id = ? AND status = ? AND deleted_at IS NULL", memberID, familyID, string(enums.StatusActive)).
		Updates(values).Error
}

func (r *GormRepository) FindActiveRelationship(ctx context.Context, familyID uint64, relationshipID uint64) (*relationshipmodel.FamilyRelationship, error) {
	var relationship relationshipmodel.FamilyRelationship
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND family_id = ? AND status = ? AND deleted_at IS NULL", relationshipID, familyID, string(enums.StatusActive)).
		First(&relationship).Error
	return &relationship, err
}

func (r *GormRepository) HasActiveRelationships(ctx context.Context, familyID uint64, memberID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&relationshipmodel.FamilyRelationship{}).
		Where("family_id = ? AND status = ? AND deleted_at IS NULL", familyID, string(enums.StatusActive)).
		Where("from_member_id = ? OR to_member_id = ?", memberID, memberID).
		Count(&count).Error
	return count > 0, err
}

func (r *GormRepository) FindDuplicate(ctx context.Context, familyID uint64, fromMemberID uint64, toMemberID uint64, relationshipType string) (*relationshipmodel.FamilyRelationship, error) {
	var relationship relationshipmodel.FamilyRelationship
	query := r.db.WithContext(ctx).
		Where("family_id = ? AND relationship_type = ? AND status = ? AND deleted_at IS NULL", familyID, relationshipType, string(enums.StatusActive))
	if relationshipType == string(enums.RelationshipTypeSpouse) {
		query = query.Where(
			"(from_member_id = ? AND to_member_id = ?) OR (from_member_id = ? AND to_member_id = ?)",
			fromMemberID, toMemberID, toMemberID, fromMemberID,
		)
	} else {
		query = query.Where("from_member_id = ? AND to_member_id = ?", fromMemberID, toMemberID)
	}
	err := query.First(&relationship).Error
	return &relationship, err
}

func (r *GormRepository) ListActiveParents(ctx context.Context, familyID uint64, childMemberID uint64) ([]relationshipmodel.FamilyRelationship, error) {
	var relationships []relationshipmodel.FamilyRelationship
	err := r.db.WithContext(ctx).
		Where("family_id = ? AND to_member_id = ? AND relationship_type = ? AND status = ? AND deleted_at IS NULL",
			familyID, childMemberID, string(enums.RelationshipTypeParentChild), string(enums.StatusActive)).
		Order("id").
		Find(&relationships).Error
	return relationships, err
}

func (r *GormRepository) FindPrimaryParentByGender(ctx context.Context, familyID uint64, childMemberID uint64, gender string, excludeRelationshipID uint64) (*relationshipmodel.FamilyRelationship, error) {
	relationship, _, err := r.FindPrimaryParentWithMemberByGender(ctx, familyID, childMemberID, gender, excludeRelationshipID)
	return relationship, err
}

func (r *GormRepository) FindPrimaryParentWithMemberByGender(ctx context.Context, familyID uint64, childMemberID uint64, gender string, excludeRelationshipID uint64) (*relationshipmodel.FamilyRelationship, *membermodel.FamilyMember, error) {
	var relationship relationshipmodel.FamilyRelationship
	var parent membermodel.FamilyMember
	query := r.db.WithContext(ctx).Table("family_relationships AS fr").
		Select("fr.*").
		Joins("JOIN family_members AS parent ON parent.id = fr.from_member_id AND parent.family_id = fr.family_id").
		Where("fr.family_id = ? AND fr.to_member_id = ? AND fr.relationship_type = ? AND fr.parent_link_type = ?",
			familyID, childMemberID, string(enums.RelationshipTypeParentChild), "PRIMARY").
		Where("fr.status = ? AND fr.deleted_at IS NULL AND parent.status = ? AND parent.deleted_at IS NULL AND parent.gender = ?",
			string(enums.StatusActive), string(enums.StatusActive), gender)
	if excludeRelationshipID != 0 {
		query = query.Where("fr.id <> ?", excludeRelationshipID)
	}
	err := query.First(&relationship).Error
	if err != nil {
		return nil, nil, err
	}
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND family_id = ? AND status = ? AND deleted_at IS NULL",
			relationship.FromMemberID, familyID, string(enums.StatusActive)).
		First(&parent).Error; err != nil {
		return nil, nil, err
	}
	return &relationship, &parent, nil
}

func (r *GormRepository) ListActiveSpouseMembersByGender(ctx context.Context, familyID uint64, baseMemberID uint64, gender string) ([]membermodel.FamilyMember, error) {
	rows, err := r.ListActiveSpouseRelationshipsByGender(ctx, familyID, baseMemberID, gender)
	if err != nil {
		return nil, err
	}
	members := make([]membermodel.FamilyMember, 0, len(rows))
	for i := range rows {
		members = append(members, rows[i].Spouse)
	}
	return members, nil
}

func (r *GormRepository) ListActiveSpouseRelationshipsByGender(ctx context.Context, familyID uint64, baseMemberID uint64, gender string) ([]SpouseRelationshipRow, error) {
	var members []membermodel.FamilyMember
	var relationships []relationshipmodel.FamilyRelationship
	err := r.db.WithContext(ctx).
		Where("family_id = ? AND relationship_type = ? AND status = ? AND deleted_at IS NULL",
			familyID, string(enums.RelationshipTypeSpouse), string(enums.StatusActive)).
		Where("(from_member_id = ? OR to_member_id = ?)", baseMemberID, baseMemberID).
		Order("id").
		Find(&relationships).Error
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).Table("family_relationships AS fr").
		Select("spouse.*").
		Joins(`JOIN family_members AS spouse ON spouse.family_id = fr.family_id
			AND spouse.id = CASE WHEN fr.from_member_id = ? THEN fr.to_member_id ELSE fr.from_member_id END`,
			baseMemberID).
		Where("fr.family_id = ? AND fr.relationship_type = ? AND fr.status = ? AND fr.deleted_at IS NULL",
			familyID, string(enums.RelationshipTypeSpouse), string(enums.StatusActive)).
		Where("(fr.from_member_id = ? OR fr.to_member_id = ?)", baseMemberID, baseMemberID).
		Where("spouse.status = ? AND spouse.deleted_at IS NULL AND spouse.gender = ?",
			string(enums.StatusActive), gender).
		Order("fr.id").
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	result := make([]SpouseRelationshipRow, 0, len(members))
	for i := range relationships {
		for j := range members {
			if relationships[i].FromMemberID == members[j].ID || relationships[i].ToMemberID == members[j].ID {
				result = append(result, SpouseRelationshipRow{Relationship: relationships[i], Spouse: members[j]})
				break
			}
		}
	}
	return result, nil
}

func (r *GormRepository) CreateRelationship(ctx context.Context, relationship *relationshipmodel.FamilyRelationship) error {
	return r.db.WithContext(ctx).Create(relationship).Error
}

func (r *GormRepository) UpdateRelationship(ctx context.Context, familyID uint64, relationshipID uint64, values map[string]any) error {
	return r.db.WithContext(ctx).Model(&relationshipmodel.FamilyRelationship{}).
		Where("id = ? AND family_id = ? AND status = ? AND deleted_at IS NULL",
			relationshipID, familyID, string(enums.StatusActive)).
		Updates(values).Error
}

func (r *GormRepository) SoftDeleteRelationship(ctx context.Context, familyID uint64, relationshipID uint64, actorID uint64, reason *string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&relationshipmodel.FamilyRelationship{}).
		Where("id = ? AND family_id = ? AND status = ? AND deleted_at IS NULL",
			relationshipID, familyID, string(enums.StatusActive)).
		Updates(map[string]any{
			"status":             "DELETED",
			"deleted_at":         now,
			"deleted_by_user_id": actorID,
			"delete_reason":      reason,
			"updated_at":         now,
		}).Error
}

func (r *GormRepository) IncrementGraphVersion(ctx context.Context, familyID uint64) (int64, error) {
	if err := r.db.WithContext(ctx).Model(&familymodel.Family{}).
		Where("id = ? AND deleted_at IS NULL", familyID).
		UpdateColumn("graph_version", gorm.Expr("graph_version + 1")).Error; err != nil {
		return 0, err
	}
	var family familymodel.Family
	if err := r.db.WithContext(ctx).Select("graph_version").First(&family, familyID).Error; err != nil {
		return 0, err
	}
	return family.GraphVersion, nil
}

func (r *GormRepository) WriteOperationLog(ctx context.Context, input operationlog.WriteInput) error {
	return operationlog.NewGormService(r.db).WriteSuccess(ctx, input)
}

type UnitOfWork interface {
	WithinTransaction(context.Context, func(Repository) error) error
}

type GormUnitOfWork struct {
	db   *gorm.DB
	repo Repository
}

func NewUnitOfWork(db *gorm.DB, repo Repository) *GormUnitOfWork {
	return &GormUnitOfWork{db: db, repo: repo}
}

func (u *GormUnitOfWork) WithinTransaction(ctx context.Context, fn func(Repository) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(u.repo.WithTx(tx))
	})
}
