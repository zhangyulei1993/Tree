package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	familymodel "tree/backend/internal/family/core/model"
	dissolutionmodel "tree/backend/internal/family/dissolution/model"
	membermodel "tree/backend/internal/family/member/model"
	rolemodel "tree/backend/internal/family/role/model"
	usermodel "tree/backend/internal/user/model"
)

type FamilyListRow struct {
	familymodel.Family
	FamilyRole string `gorm:"column:family_role"`
}

type FamilyRepository interface {
	WithTx(tx *gorm.DB) FamilyRepository
	FindUserByID(ctx context.Context, userID uint64) (*usermodel.User, error)
	CreateFamily(ctx context.Context, family *familymodel.Family) error
	CreateMember(ctx context.Context, member *membermodel.FamilyMember) error
	CreateLink(ctx context.Context, link *rolemodel.FamilyMemberUserLink) error
	FindFamilyByID(ctx context.Context, familyID uint64) (*familymodel.Family, error)
	FindFamilyByIDForUpdate(ctx context.Context, familyID uint64) (*familymodel.Family, error)
	FindPublicFamilyByID(ctx context.Context, familyID uint64) (*familymodel.Family, error)
	ListByUser(ctx context.Context, userID uint64) ([]FamilyListRow, error)
	UpdateFamily(ctx context.Context, familyID uint64, values map[string]any) error
	CreateDissolutionRequest(ctx context.Context, request *dissolutionmodel.FamilyDissolutionRequest) error
	FindCurrentDissolutionRequest(ctx context.Context, familyID uint64) (*dissolutionmodel.FamilyDissolutionRequest, error)
	FindDissolutionRequest(ctx context.Context, familyID uint64, requestID uint64) (*dissolutionmodel.FamilyDissolutionRequest, error)
	CancelDissolutionRequest(ctx context.Context, requestID uint64, reason *string, now time.Time) error
}

type GormFamilyRepository struct {
	db *gorm.DB
}

func NewFamilyRepository(db *gorm.DB) *GormFamilyRepository {
	return &GormFamilyRepository{db: db}
}

func (r *GormFamilyRepository) WithTx(tx *gorm.DB) FamilyRepository {
	return &GormFamilyRepository{db: tx}
}

func (r *GormFamilyRepository) FindUserByID(ctx context.Context, userID uint64) (*usermodel.User, error) {
	var user usermodel.User
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormFamilyRepository) CreateFamily(ctx context.Context, family *familymodel.Family) error {
	return r.db.WithContext(ctx).Create(family).Error
}

func (r *GormFamilyRepository) CreateMember(ctx context.Context, member *membermodel.FamilyMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *GormFamilyRepository) CreateLink(ctx context.Context, link *rolemodel.FamilyMemberUserLink) error {
	return r.db.WithContext(ctx).Create(link).Error
}

func (r *GormFamilyRepository) FindFamilyByID(ctx context.Context, familyID uint64) (*familymodel.Family, error) {
	var family familymodel.Family
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", familyID).First(&family).Error; err != nil {
		return nil, err
	}
	return &family, nil
}

func (r *GormFamilyRepository) FindFamilyByIDForUpdate(ctx context.Context, familyID uint64) (*familymodel.Family, error) {
	var family familymodel.Family
	if err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND deleted_at IS NULL", familyID).
		First(&family).Error; err != nil {
		return nil, err
	}
	return &family, nil
}

func (r *GormFamilyRepository) FindPublicFamilyByID(ctx context.Context, familyID uint64) (*familymodel.Family, error) {
	var family familymodel.Family
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL AND status = ? AND public_display_status = ?", familyID, "NORMAL", "APPROVED").
		First(&family).Error; err != nil {
		return nil, err
	}
	return &family, nil
}

func (r *GormFamilyRepository) ListByUser(ctx context.Context, userID uint64) ([]FamilyListRow, error) {
	var rows []FamilyListRow
	err := r.db.WithContext(ctx).
		Table("families").
		Select("families.*, family_member_user_links.family_role").
		Joins("JOIN family_member_user_links ON family_member_user_links.family_id = families.id").
		Where("family_member_user_links.user_id = ? AND family_member_user_links.link_status = ?", userID, "ACTIVE").
		Where("families.deleted_at IS NULL").
		Order("families.updated_at DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *GormFamilyRepository) UpdateFamily(ctx context.Context, familyID uint64, values map[string]any) error {
	return r.db.WithContext(ctx).Model(&familymodel.Family{}).Where("id = ?", familyID).Updates(values).Error
}

func (r *GormFamilyRepository) CreateDissolutionRequest(ctx context.Context, request *dissolutionmodel.FamilyDissolutionRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

func (r *GormFamilyRepository) FindCurrentDissolutionRequest(ctx context.Context, familyID uint64) (*dissolutionmodel.FamilyDissolutionRequest, error) {
	var request dissolutionmodel.FamilyDissolutionRequest
	if err := r.db.WithContext(ctx).
		Where("family_id = ? AND request_status = ?", familyID, "PENDING").
		Order("id DESC").
		First(&request).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *GormFamilyRepository) FindDissolutionRequest(ctx context.Context, familyID uint64, requestID uint64) (*dissolutionmodel.FamilyDissolutionRequest, error) {
	var request dissolutionmodel.FamilyDissolutionRequest
	if err := r.db.WithContext(ctx).
		Where("id = ? AND family_id = ?", requestID, familyID).
		First(&request).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *GormFamilyRepository) CancelDissolutionRequest(ctx context.Context, requestID uint64, reason *string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&dissolutionmodel.FamilyDissolutionRequest{}).
		Where("id = ? AND request_status = ?", requestID, "PENDING").
		Updates(map[string]any{
			"request_status": "CANCELLED",
			"cancelled_at":   now,
			"cancel_reason":  reason,
			"updated_at":     now,
		}).Error
}
