package permission

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"tree/backend/internal/common/enums"
	familyrolemodel "tree/backend/internal/family/role/model"
)

type FamilyPermissionService interface {
	IsFamilyAdmin(ctx context.Context, userID uint64, familyID uint64) (bool, error)
	IsFamilyMember(ctx context.Context, userID uint64, familyID uint64) (bool, error)
	CanManageFamily(ctx context.Context, userID uint64, familyID uint64) (bool, error)
	CanCreateMember(ctx context.Context, userID uint64, familyID uint64) (bool, error)
	CanEditMember(ctx context.Context, userID uint64, familyID uint64, memberID uint64) (bool, error)
	CanDeleteMember(ctx context.Context, userID uint64, familyID uint64, memberID uint64) (bool, error)
	GetActiveLink(ctx context.Context, userID uint64, familyID uint64) (*familyrolemodel.FamilyMemberUserLink, error)
}

type GormFamilyPermissionService struct {
	db *gorm.DB
}

func NewFamilyPermissionService(db *gorm.DB) *GormFamilyPermissionService {
	return &GormFamilyPermissionService{db: db}
}

func (s *GormFamilyPermissionService) IsFamilyAdmin(ctx context.Context, userID uint64, familyID uint64) (bool, error) {
	link, err := s.GetActiveLink(ctx, userID, familyID)
	if err != nil {
		return false, err
	}
	return link.FamilyRole == string(enums.FamilyRoleFounder) ||
		link.FamilyRole == string(enums.FamilyRoleFamilyAdmin), nil
}

func (s *GormFamilyPermissionService) IsFamilyMember(ctx context.Context, userID uint64, familyID uint64) (bool, error) {
	_, err := s.GetActiveLink(ctx, userID, familyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (s *GormFamilyPermissionService) CanManageFamily(ctx context.Context, userID uint64, familyID uint64) (bool, error) {
	return s.IsFamilyAdmin(ctx, userID, familyID)
}

func (s *GormFamilyPermissionService) CanCreateMember(ctx context.Context, userID uint64, familyID uint64) (bool, error) {
	return s.IsFamilyAdmin(ctx, userID, familyID)
}

func (s *GormFamilyPermissionService) CanEditMember(ctx context.Context, userID uint64, familyID uint64, memberID uint64) (bool, error) {
	return s.IsFamilyAdmin(ctx, userID, familyID)
}

func (s *GormFamilyPermissionService) CanDeleteMember(ctx context.Context, userID uint64, familyID uint64, memberID uint64) (bool, error) {
	return s.IsFamilyAdmin(ctx, userID, familyID)
}

func (s *GormFamilyPermissionService) GetActiveLink(ctx context.Context, userID uint64, familyID uint64) (*familyrolemodel.FamilyMemberUserLink, error) {
	var link familyrolemodel.FamilyMemberUserLink
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND family_id = ? AND link_status = ?", userID, familyID, string(enums.StatusActive)).
		First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}
