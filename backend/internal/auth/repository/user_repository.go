package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	accountmodel "tree/backend/internal/account/model"
	"tree/backend/internal/common/enums"
	familyrolemodel "tree/backend/internal/family/role/model"
	"tree/backend/internal/user/model"
)

type UserRepository interface {
	FindByPhoneHash(ctx context.Context, phoneHash string) (*model.User, error)
	FindByID(ctx context.Context, id uint64) (*model.User, error)
	LockByID(ctx context.Context, id uint64) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	RecordLoginSuccess(ctx context.Context, userID uint64, ip string, clientType string, now time.Time) error
	UpdateUser(ctx context.Context, userID uint64, values map[string]any) error
	CreateIdentity(ctx context.Context, identity *model.UserAuthIdentity) error
	FindIdentityByOpenIDHash(ctx context.Context, provider string, appID string, openIDHash string) (*model.UserAuthIdentity, error)
	HasActiveWechatMiniIdentity(ctx context.Context, userID uint64, appID string) (bool, error)
	FindActiveWechatMiniOpenID(ctx context.Context, userID uint64, appID string) (string, error)
	UpdateIdentity(ctx context.Context, identityID uint64, values map[string]any) error
	CancelActiveIdentities(ctx context.Context, userID uint64) error
	MoveIdentities(ctx context.Context, sourceUserID uint64, targetUserID uint64) error
	HasMemberBindingConflict(ctx context.Context, sourceUserID uint64, targetUserID uint64) (bool, error)
	MoveFamilyLinks(ctx context.Context, sourceUserID uint64, targetUserID uint64) error
	CountActiveFamilyLinks(ctx context.Context, userID uint64) (int64, error)
	CountFounderLinks(ctx context.Context, userID uint64) (int64, error)
	CountPendingFounderTransfer(ctx context.Context, userID uint64) (int64, error)
	CountPendingDissolution(ctx context.Context, userID uint64) (int64, error)
	CreatePhoneHistory(ctx context.Context, record *accountmodel.UserPhoneHistory) error
	CreateMergeLog(ctx context.Context, record *accountmodel.UserAccountMergeLog) error
	CreateClaimLog(ctx context.Context, record *accountmodel.UserAccountClaimLog) error
	WithTx(tx *gorm.DB) UserRepository
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) WithTx(tx *gorm.DB) UserRepository {
	return &GormUserRepository{db: tx}
}

func (r *GormUserRepository) FindByPhoneHash(ctx context.Context, phoneHash string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Where("phone_hash = ? AND deleted_at IS NULL", phoneHash).
		Order("id DESC").
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) LockByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *GormUserRepository) RecordLoginSuccess(ctx context.Context, userID uint64, ip string, clientType string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{
		"last_login_at":     now,
		"last_login_ip":     ip,
		"last_login_client": clientType,
		"updated_at":        now,
	}).Error
}

func (r *GormUserRepository) UpdateUser(ctx context.Context, userID uint64, values map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(values).Error
}

func (r *GormUserRepository) CreateIdentity(ctx context.Context, identity *model.UserAuthIdentity) error {
	return r.db.WithContext(ctx).Create(identity).Error
}

const wechatMiniProvider = "WECHAT_MINI"

func (r *GormUserRepository) HasActiveWechatMiniIdentity(ctx context.Context, userID uint64, appID string) (bool, error) {
	openid, err := r.FindActiveWechatMiniOpenID(ctx, userID, appID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return openid != "", nil
}

func (r *GormUserRepository) FindActiveWechatMiniOpenID(ctx context.Context, userID uint64, appID string) (string, error) {
	var identity model.UserAuthIdentity
	query := r.db.WithContext(ctx).
		Where(
			"user_id = ? AND provider = ? AND identity_status = ? AND deleted_at IS NULL",
			userID, wechatMiniProvider, string(enums.StatusActive),
		)
	if strings.TrimSpace(appID) != "" {
		query = query.Where("provider_app_id = ?", appID)
	}
	err := query.Order("id DESC").First(&identity).Error
	if err != nil {
		return "", err
	}
	if identity.OpenID == nil || strings.TrimSpace(*identity.OpenID) == "" {
		return "", gorm.ErrRecordNotFound
	}
	return strings.TrimSpace(*identity.OpenID), nil
}

func (r *GormUserRepository) FindIdentityByOpenIDHash(ctx context.Context, provider string, appID string, openIDHash string) (*model.UserAuthIdentity, error) {
	var identity model.UserAuthIdentity
	if err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_app_id = ? AND openid_hash = ? AND identity_status = ? AND deleted_at IS NULL", provider, appID, openIDHash, "ACTIVE").
		Order("id DESC").
		First(&identity).Error; err != nil {
		return nil, err
	}
	return &identity, nil
}

func (r *GormUserRepository) UpdateIdentity(ctx context.Context, identityID uint64, values map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.UserAuthIdentity{}).Where("id = ?", identityID).Updates(values).Error
}

func (r *GormUserRepository) CancelActiveIdentities(ctx context.Context, userID uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.UserAuthIdentity{}).
		Where("user_id = ? AND identity_status = ? AND deleted_at IS NULL", userID, string(enums.StatusActive)).
		Updates(map[string]any{
			"identity_status": string(enums.StatusCancelled),
			"unbound_at":      now,
			"updated_at":      now,
		}).Error
}

func (r *GormUserRepository) MoveIdentities(ctx context.Context, sourceUserID uint64, targetUserID uint64) error {
	return r.db.WithContext(ctx).Model(&model.UserAuthIdentity{}).
		Where("user_id = ? AND deleted_at IS NULL", sourceUserID).
		Update("user_id", targetUserID).Error
}

func (r *GormUserRepository) HasMemberBindingConflict(ctx context.Context, sourceUserID uint64, targetUserID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("family_member_user_links AS source").
		Joins("JOIN family_member_user_links AS target ON source.family_id = target.family_id").
		Where("source.user_id = ? AND target.user_id = ?", sourceUserID, targetUserID).
		Where("source.link_status = ? AND target.link_status = ?", "ACTIVE", "ACTIVE").
		Where("source.member_id <> target.member_id").
		Count(&count).Error
	return count > 0, err
}

func (r *GormUserRepository) MoveFamilyLinks(ctx context.Context, sourceUserID uint64, targetUserID uint64) error {
	return r.db.WithContext(ctx).Model(&familyrolemodel.FamilyMemberUserLink{}).
		Where("user_id = ? AND link_status = ?", sourceUserID, "ACTIVE").
		Update("user_id", targetUserID).Error
}

func (r *GormUserRepository) CountActiveFamilyLinks(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&familyrolemodel.FamilyMemberUserLink{}).
		Where("user_id = ? AND link_status = ?", userID, "ACTIVE").
		Count(&count).Error
	return count, err
}

func (r *GormUserRepository) CountFounderLinks(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&familyrolemodel.FamilyMemberUserLink{}).
		Where("user_id = ? AND link_status = ? AND family_role = ?", userID, "ACTIVE", "FOUNDER").
		Count(&count).Error
	return count, err
}

func (r *GormUserRepository) CountPendingFounderTransfer(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("family_founder_transfer_requests").
		Where("(from_user_id = ? OR to_user_id = ?) AND request_status = ?", userID, userID, "PENDING").
		Count(&count).Error
	return count, err
}

func (r *GormUserRepository) CountPendingDissolution(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("family_dissolution_requests").
		Where("requester_user_id = ? AND request_status = ?", userID, "PENDING").
		Count(&count).Error
	return count, err
}

func (r *GormUserRepository) CreatePhoneHistory(ctx context.Context, record *accountmodel.UserPhoneHistory) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *GormUserRepository) CreateMergeLog(ctx context.Context, record *accountmodel.UserAccountMergeLog) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *GormUserRepository) CreateClaimLog(ctx context.Context, record *accountmodel.UserAccountClaimLog) error {
	return r.db.WithContext(ctx).Create(record).Error
}
