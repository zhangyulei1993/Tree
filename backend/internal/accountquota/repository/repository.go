package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	quotaenum "tree/backend/internal/accountquota/enum"
	quotamodel "tree/backend/internal/accountquota/model"
	usermodel "tree/backend/internal/user/model"
)

type Limits struct {
	MaxOwnedFamilies         int
	MaxMembersPerOwnedFamily int
	MaxJoinedFamilies        int
	SupportsGenerationNaming bool
}

var ErrFounderNotFound = errors.New("family founder not found")

type Repository interface {
	WithDB(db *gorm.DB) Repository
	ListConfigs(ctx context.Context) ([]quotamodel.AccountQuotaConfig, error)
	FindConfigByTier(ctx context.Context, tier string) (*quotamodel.AccountQuotaConfig, error)
	LockConfigByTier(ctx context.Context, tier string) (*quotamodel.AccountQuotaConfig, error)
	LockAllConfigsInOrder(ctx context.Context) (wechat quotamodel.AccountQuotaConfig, phone quotamodel.AccountQuotaConfig, err error)
	UpdateConfig(ctx context.Context, tier string, values map[string]any) error
	ListFeatureOverrides(ctx context.Context, featureKey string) ([]quotamodel.AccountFeatureOverride, error)
	ReplaceFeatureOverrides(ctx context.Context, featureKey string, rows []quotamodel.AccountFeatureOverride) error
	DeleteFeatureOverride(ctx context.Context, featureKey string, id uint64) (int64, error)
	HasFeatureOverride(ctx context.Context, featureKey string, phoneHash string) (bool, error)
	LockUser(ctx context.Context, userID uint64) (*usermodel.User, error)
	ResolveTrustTier(user *usermodel.User) string
	GetLimitsForTier(ctx context.Context, tier string) (Limits, error)
	CountOwnedFamilies(ctx context.Context, userID uint64) (int, error)
	CountJoinedFamilies(ctx context.Context, userID uint64) (int, error)
	CountActiveMembers(ctx context.Context, familyID uint64) (int, error)
	FindFounderUserID(ctx context.Context, familyID uint64) (uint64, error)
	ListJoinedUserIDsByFamily(ctx context.Context, familyID uint64) ([]uint64, error)
	ListOwnedFamilyIDs(ctx context.Context, userID uint64) ([]uint64, error)
	CountUsersExceedingOwned(ctx context.Context, tier string, maxOwned int) (int, error)
	CountUsersExceedingJoined(ctx context.Context, tier string, maxJoined int) (int, error)
	CountUsersExceedingDistinct(ctx context.Context, tier string, maxOwned, maxJoined int) (int, error)
	CountFamiliesExceedingMembers(ctx context.Context, tier string, maxMembers int) (int, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) WithDB(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) ListConfigs(ctx context.Context) ([]quotamodel.AccountQuotaConfig, error) {
	var rows []quotamodel.AccountQuotaConfig
	err := r.db.WithContext(ctx).Order("id").Find(&rows).Error
	return rows, err
}

func (r *GormRepository) FindConfigByTier(ctx context.Context, tier string) (*quotamodel.AccountQuotaConfig, error) {
	var row quotamodel.AccountQuotaConfig
	err := r.db.WithContext(ctx).Where("trust_tier = ?", tier).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *GormRepository) LockConfigByTier(ctx context.Context, tier string) (*quotamodel.AccountQuotaConfig, error) {
	var row quotamodel.AccountQuotaConfig
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("trust_tier = ?", tier).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *GormRepository) LockAllConfigsInOrder(ctx context.Context) (quotamodel.AccountQuotaConfig, quotamodel.AccountQuotaConfig, error) {
	wechat, err := r.LockConfigByTier(ctx, quotaenum.TrustTierWechatOnly)
	if err != nil {
		return quotamodel.AccountQuotaConfig{}, quotamodel.AccountQuotaConfig{}, err
	}
	phone, err := r.LockConfigByTier(ctx, quotaenum.TrustTierPhoneBound)
	if err != nil {
		return quotamodel.AccountQuotaConfig{}, quotamodel.AccountQuotaConfig{}, err
	}
	return *wechat, *phone, nil
}

func (r *GormRepository) UpdateConfig(ctx context.Context, tier string, values map[string]any) error {
	return r.db.WithContext(ctx).Model(&quotamodel.AccountQuotaConfig{}).Where("trust_tier = ?", tier).Updates(values).Error
}

func (r *GormRepository) ListFeatureOverrides(ctx context.Context, featureKey string) ([]quotamodel.AccountFeatureOverride, error) {
	var rows []quotamodel.AccountFeatureOverride
	err := r.db.WithContext(ctx).
		Where("feature_key = ?", featureKey).
		Order("id").
		Find(&rows).Error
	return rows, err
}

func (r *GormRepository) ReplaceFeatureOverrides(ctx context.Context, featureKey string, rows []quotamodel.AccountFeatureOverride) error {
	if err := r.db.WithContext(ctx).Where("feature_key = ?", featureKey).Delete(&quotamodel.AccountFeatureOverride{}).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&rows).Error
}

func (r *GormRepository) DeleteFeatureOverride(ctx context.Context, featureKey string, id uint64) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND feature_key = ?", id, featureKey).
		Delete(&quotamodel.AccountFeatureOverride{})
	return result.RowsAffected, result.Error
}

func (r *GormRepository) HasFeatureOverride(ctx context.Context, featureKey string, phoneHash string) (bool, error) {
	if phoneHash == "" {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).
		Model(&quotamodel.AccountFeatureOverride{}).
		Where("feature_key = ? AND phone_hash = ?", featureKey, phoneHash).
		Count(&count).Error
	return count > 0, err
}

func (r *GormRepository) LockUser(ctx context.Context, userID uint64) (*usermodel.User, error) {
	var user usermodel.User
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND deleted_at IS NULL", userID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormRepository) ResolveTrustTier(user *usermodel.User) string {
	if user != nil && user.PhoneLoginEnabled {
		return quotaenum.TrustTierPhoneBound
	}
	return quotaenum.TrustTierWechatOnly
}

func defaultLimits(tier string) Limits {
	switch tier {
	case quotaenum.TrustTierPhoneBound:
		return Limits{MaxOwnedFamilies: 1, MaxMembersPerOwnedFamily: 20, MaxJoinedFamilies: 5, SupportsGenerationNaming: true}
	default:
		return Limits{MaxOwnedFamilies: 1, MaxMembersPerOwnedFamily: 10, MaxJoinedFamilies: 1, SupportsGenerationNaming: false}
	}
}

func (r *GormRepository) GetLimitsForTier(ctx context.Context, tier string) (Limits, error) {
	row, err := r.FindConfigByTier(ctx, tier)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return defaultLimits(tier), nil
		}
		return Limits{}, err
	}
	return Limits{
		MaxOwnedFamilies:         row.MaxOwnedFamilies,
		MaxMembersPerOwnedFamily: row.MaxMembersPerOwnedFamily,
		MaxJoinedFamilies:        row.MaxJoinedFamilies,
		SupportsGenerationNaming: row.SupportsGenerationNaming,
	}, nil
}

func (r *GormRepository) CountOwnedFamilies(ctx context.Context, userID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("family_member_user_links AS link").
		Joins("JOIN families AS family ON family.id = link.family_id").
		Where("link.user_id = ? AND link.link_status = 'ACTIVE' AND link.family_role = 'FOUNDER'", userID).
		Where("family.deleted_at IS NULL AND family.status <> 'DISSOLVED'").
		Count(&count).Error
	return int(count), err
}

func joinedFamilyStatusClause() string {
	return "family.deleted_at IS NULL AND family.status <> 'DISSOLVED'"
}

func (r *GormRepository) CountJoinedFamilies(ctx context.Context, userID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("family_member_user_links AS link").
		Joins("JOIN families AS family ON family.id = link.family_id").
		Where("link.user_id = ? AND link.link_status = 'ACTIVE'", userID).
		Where("link.family_role IN ('MEMBER', 'FAMILY_ADMIN')").
		Where(joinedFamilyStatusClause()).
		Count(&count).Error
	return int(count), err
}

func (r *GormRepository) CountActiveMembers(ctx context.Context, familyID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("family_members").
		Where("family_id = ? AND status = 'ACTIVE' AND deleted_at IS NULL", familyID).
		Count(&count).Error
	return int(count), err
}

func (r *GormRepository) FindFounderUserID(ctx context.Context, familyID uint64) (uint64, error) {
	var userID uint64
	err := r.db.WithContext(ctx).
		Table("family_member_user_links").
		Select("user_id").
		Where("family_id = ? AND link_status = 'ACTIVE' AND family_role = 'FOUNDER'", familyID).
		Limit(1).
		Scan(&userID).Error
	if err != nil {
		return 0, err
	}
	if userID == 0 {
		return 0, ErrFounderNotFound
	}
	return userID, nil
}

func (r *GormRepository) ListJoinedUserIDsByFamily(ctx context.Context, familyID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).
		Table("family_member_user_links").
		Select("DISTINCT user_id").
		Where("family_id = ? AND link_status = 'ACTIVE'", familyID).
		Where("family_role IN ('MEMBER', 'FAMILY_ADMIN')").
		Scan(&ids).Error
	return ids, err
}

func (r *GormRepository) ListOwnedFamilyIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).
		Table("family_member_user_links AS link").
		Select("link.family_id").
		Joins("JOIN families AS family ON family.id = link.family_id").
		Where("link.user_id = ? AND link.link_status = 'ACTIVE' AND link.family_role = 'FOUNDER'", userID).
		Where("family.deleted_at IS NULL AND family.status <> 'DISSOLVED'").
		Scan(&ids).Error
	return ids, err
}

func (r *GormRepository) CountUsersExceedingOwned(ctx context.Context, tier string, maxOwned int) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw(`
SELECT COUNT(*) FROM (
  SELECT u.id
  FROM users u
  JOIN family_member_user_links link ON link.user_id = u.id AND link.link_status = 'ACTIVE' AND link.family_role = 'FOUNDER'
  JOIN families family ON family.id = link.family_id AND family.deleted_at IS NULL AND family.status <> 'DISSOLVED'
  WHERE u.deleted_at IS NULL
    AND ((? = 'PHONE_BOUND' AND u.phone_login_enabled = 1) OR (? = 'WECHAT_ONLY' AND u.phone_login_enabled = 0))
  GROUP BY u.id
  HAVING COUNT(DISTINCT family.id) > ?
) t`, tier, tier, maxOwned).Scan(&count).Error
	return int(count), err
}

func (r *GormRepository) CountUsersExceedingJoined(ctx context.Context, tier string, maxJoined int) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw(`
SELECT COUNT(*) FROM (
  SELECT u.id
  FROM users u
  JOIN family_member_user_links link ON link.user_id = u.id AND link.link_status = 'ACTIVE' AND link.family_role IN ('MEMBER', 'FAMILY_ADMIN')
  JOIN families family ON family.id = link.family_id AND family.deleted_at IS NULL AND family.status <> 'DISSOLVED'
  WHERE u.deleted_at IS NULL
    AND ((? = 'PHONE_BOUND' AND u.phone_login_enabled = 1) OR (? = 'WECHAT_ONLY' AND u.phone_login_enabled = 0))
  GROUP BY u.id
  HAVING COUNT(DISTINCT family.id) > ?
) t`, tier, tier, maxJoined).Scan(&count).Error
	return int(count), err
}

func (r *GormRepository) CountUsersExceedingDistinct(ctx context.Context, tier string, maxOwned, maxJoined int) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw(`
SELECT COUNT(DISTINCT user_id) FROM (
  SELECT u.id AS user_id
  FROM users u
  JOIN family_member_user_links link ON link.user_id = u.id AND link.link_status = 'ACTIVE' AND link.family_role = 'FOUNDER'
  JOIN families family ON family.id = link.family_id AND family.deleted_at IS NULL AND family.status <> 'DISSOLVED'
  WHERE u.deleted_at IS NULL
    AND ((? = 'PHONE_BOUND' AND u.phone_login_enabled = 1) OR (? = 'WECHAT_ONLY' AND u.phone_login_enabled = 0))
  GROUP BY u.id
  HAVING COUNT(DISTINCT family.id) > ?
  UNION
  SELECT u.id AS user_id
  FROM users u
  JOIN family_member_user_links link ON link.user_id = u.id AND link.link_status = 'ACTIVE' AND link.family_role IN ('MEMBER', 'FAMILY_ADMIN')
  JOIN families family ON family.id = link.family_id AND family.deleted_at IS NULL AND family.status <> 'DISSOLVED'
  WHERE u.deleted_at IS NULL
    AND ((? = 'PHONE_BOUND' AND u.phone_login_enabled = 1) OR (? = 'WECHAT_ONLY' AND u.phone_login_enabled = 0))
  GROUP BY u.id
  HAVING COUNT(DISTINCT family.id) > ?
) affected_users`, tier, tier, maxOwned, tier, tier, maxJoined).Scan(&count).Error
	return int(count), err
}

func (r *GormRepository) CountFamiliesExceedingMembers(ctx context.Context, tier string, maxMembers int) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw(`
SELECT COUNT(*) FROM (
  SELECT family.id
  FROM families family
  JOIN family_member_user_links founder_link ON founder_link.family_id = family.id AND founder_link.link_status = 'ACTIVE' AND founder_link.family_role = 'FOUNDER'
  JOIN users founder_user ON founder_user.id = founder_link.user_id AND founder_user.deleted_at IS NULL
  WHERE family.deleted_at IS NULL AND family.status <> 'DISSOLVED'
    AND ((? = 'PHONE_BOUND' AND founder_user.phone_login_enabled = 1) OR (? = 'WECHAT_ONLY' AND founder_user.phone_login_enabled = 0))
  GROUP BY family.id
  HAVING (
    SELECT COUNT(*) FROM family_members m
    WHERE m.family_id = family.id AND m.status = 'ACTIVE' AND m.deleted_at IS NULL
  ) > ?
) t`, tier, tier, maxMembers).Scan(&count).Error
	return int(count), err
}
