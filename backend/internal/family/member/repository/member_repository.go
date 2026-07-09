package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	familymodel "tree/backend/internal/family/core/model"
	membermodel "tree/backend/internal/family/member/model"
	rolemodel "tree/backend/internal/family/role/model"
	usermodel "tree/backend/internal/user/model"
)

type MemberRow struct {
	membermodel.FamilyMember
	BoundUserID     *uint64 `gorm:"column:bound_user_id"`
	BoundFamilyRole *string `gorm:"column:bound_family_role"`
}

type MemberRepository interface {
	WithTx(*gorm.DB) MemberRepository
	LockFamily(context.Context, uint64) (*familymodel.Family, error)
	IncrementGraphVersion(context.Context, uint64) error
	Create(context.Context, *membermodel.FamilyMember) error
	List(context.Context, uint64) ([]MemberRow, error)
	Find(context.Context, uint64, uint64) (*MemberRow, error)
	FindForUpdate(context.Context, uint64, uint64) (*membermodel.FamilyMember, error)
	Update(context.Context, uint64, uint64, map[string]any) error
	SoftDelete(context.Context, uint64, uint64, uint64, *string, time.Time) error
	CountBlockingRelationships(context.Context, uint64, uint64) (int64, error)
	SoftDeleteDeletableRelationships(context.Context, uint64, uint64, uint64, *string, time.Time) (int64, error)
	FindUserForUpdate(context.Context, uint64) (*usermodel.User, error)
	FindActiveLinkByMember(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error)
	FindActiveLinkByUser(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error)
	CreateLink(context.Context, *rolemodel.FamilyMemberUserLink) error
	UnbindLink(context.Context, uint64, uint64, *string, time.Time) error
}

type GormMemberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) *GormMemberRepository {
	return &GormMemberRepository{db: db}
}

func (r *GormMemberRepository) WithTx(tx *gorm.DB) MemberRepository {
	return &GormMemberRepository{db: tx}
}

func (r *GormMemberRepository) LockFamily(ctx context.Context, familyID uint64) (*familymodel.Family, error) {
	var family familymodel.Family
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND deleted_at IS NULL", familyID).First(&family).Error
	return &family, err
}

func (r *GormMemberRepository) IncrementGraphVersion(ctx context.Context, familyID uint64) error {
	return r.db.WithContext(ctx).Model(&familymodel.Family{}).
		Where("id = ? AND deleted_at IS NULL", familyID).
		UpdateColumn("graph_version", gorm.Expr("graph_version + 1")).Error
}

func (r *GormMemberRepository) Create(ctx context.Context, member *membermodel.FamilyMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *GormMemberRepository) List(ctx context.Context, familyID uint64) ([]MemberRow, error) {
	var rows []MemberRow
	err := r.db.WithContext(ctx).Table("family_members AS fm").
		Select("fm.*, links.user_id AS bound_user_id, links.family_role AS bound_family_role").
		Joins("LEFT JOIN family_member_user_links AS links ON links.member_id = fm.id AND links.family_id = fm.family_id AND links.link_status = ?", "ACTIVE").
		Where("fm.family_id = ? AND fm.deleted_at IS NULL AND fm.status = ?", familyID, "ACTIVE").
		Order("fm.manual_order IS NULL, fm.manual_order, fm.id").
		Scan(&rows).Error
	return rows, err
}

func (r *GormMemberRepository) Find(ctx context.Context, familyID uint64, memberID uint64) (*MemberRow, error) {
	var row MemberRow
	err := r.db.WithContext(ctx).Table("family_members AS fm").
		Select("fm.*, links.user_id AS bound_user_id, links.family_role AS bound_family_role").
		Joins("LEFT JOIN family_member_user_links AS links ON links.member_id = fm.id AND links.family_id = fm.family_id AND links.link_status = ?", "ACTIVE").
		Where("fm.id = ? AND fm.family_id = ? AND fm.deleted_at IS NULL AND fm.status = ?", memberID, familyID, "ACTIVE").
		First(&row).Error
	return &row, err
}

func (r *GormMemberRepository) FindForUpdate(ctx context.Context, familyID uint64, memberID uint64) (*membermodel.FamilyMember, error) {
	var member membermodel.FamilyMember
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND family_id = ? AND deleted_at IS NULL AND status = ?", memberID, familyID, "ACTIVE").
		First(&member).Error
	return &member, err
}

func (r *GormMemberRepository) Update(ctx context.Context, familyID uint64, memberID uint64, values map[string]any) error {
	return r.db.WithContext(ctx).Model(&membermodel.FamilyMember{}).
		Where("id = ? AND family_id = ? AND deleted_at IS NULL", memberID, familyID).
		Updates(values).Error
}

func (r *GormMemberRepository) SoftDelete(ctx context.Context, familyID uint64, memberID uint64, userID uint64, reason *string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&membermodel.FamilyMember{}).
		Where("id = ? AND family_id = ? AND deleted_at IS NULL", memberID, familyID).
		Updates(map[string]any{
			"status":             "DELETED",
			"deleted_at":         now,
			"deleted_by_user_id": userID,
			"delete_reason":      reason,
			"updated_at":         now,
		}).Error
}

func (r *GormMemberRepository) CountBlockingRelationships(ctx context.Context, familyID uint64, memberID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("family_relationships").
		Where("family_id = ? AND status = ? AND deleted_at IS NULL", familyID, "ACTIVE").
		Where("relationship_type = ? AND from_member_id = ?", "PARENT_CHILD", memberID).
		Count(&count).Error
	return count, err
}

func (r *GormMemberRepository) SoftDeleteDeletableRelationships(ctx context.Context, familyID uint64, memberID uint64, userID uint64, reason *string, now time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Table("family_relationships").
		Where("family_id = ? AND status = ? AND deleted_at IS NULL", familyID, "ACTIVE").
		Where("(relationship_type = ? AND to_member_id = ?) OR (relationship_type = ? AND (from_member_id = ? OR to_member_id = ?))",
			"PARENT_CHILD", memberID, "SPOUSE", memberID, memberID).
		Updates(map[string]any{
			"status":             "DELETED",
			"deleted_at":         now,
			"deleted_by_user_id": userID,
			"delete_reason":      reason,
			"updated_at":         now,
		})
	return result.RowsAffected, result.Error
}

func (r *GormMemberRepository) FindUserForUpdate(ctx context.Context, userID uint64) (*usermodel.User, error) {
	var user usermodel.User
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	return &user, err
}

func (r *GormMemberRepository) FindActiveLinkByMember(ctx context.Context, familyID uint64, memberID uint64) (*rolemodel.FamilyMemberUserLink, error) {
	var link rolemodel.FamilyMemberUserLink
	err := r.db.WithContext(ctx).
		Where("family_id = ? AND member_id = ? AND link_status = ?", familyID, memberID, "ACTIVE").
		First(&link).Error
	return &link, err
}

func (r *GormMemberRepository) FindActiveLinkByUser(ctx context.Context, familyID uint64, userID uint64) (*rolemodel.FamilyMemberUserLink, error) {
	var link rolemodel.FamilyMemberUserLink
	err := r.db.WithContext(ctx).
		Where("family_id = ? AND user_id = ? AND link_status = ?", familyID, userID, "ACTIVE").
		First(&link).Error
	return &link, err
}

func (r *GormMemberRepository) CreateLink(ctx context.Context, link *rolemodel.FamilyMemberUserLink) error {
	return r.db.WithContext(ctx).Create(link).Error
}

func (r *GormMemberRepository) UnbindLink(ctx context.Context, linkID uint64, userID uint64, reason *string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&rolemodel.FamilyMemberUserLink{}).
		Where("id = ? AND link_status = ?", linkID, "ACTIVE").
		Updates(map[string]any{
			"link_status":         "UNLINKED",
			"unlinked_at":         now,
			"unlinked_reason":     reason,
			"unlinked_by_user_id": userID,
			"updated_at":          now,
		}).Error
}
