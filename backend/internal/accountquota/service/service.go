package service

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	quotadto "tree/backend/internal/accountquota/dto"
	quotaenum "tree/backend/internal/accountquota/enum"
	quotamodel "tree/backend/internal/accountquota/model"
	quotarepo "tree/backend/internal/accountquota/repository"
	quotavo "tree/backend/internal/accountquota/vo"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	commonmask "tree/backend/internal/common/mask"
	"tree/backend/internal/common/security"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"
)

const GrayAccess = "GRAY_ACCESS"

type AuditInput struct {
	IP        string
	UserAgent string
}

type Service interface {
	GetCapabilities(ctx context.Context, userID uint64) (*quotavo.Capabilities, *apperrors.BusinessError)
	ListConfigs(ctx context.Context, role string) ([]quotavo.ConfigItem, *apperrors.BusinessError)
	UpdateConfig(ctx context.Context, adminID uint64, role string, tier string, values quotadto.ConfigValues, audit AuditInput) (*quotavo.ConfigItem, *apperrors.BusinessError)
	PreviewImpact(ctx context.Context, role string, tier string, values quotadto.ConfigValues) (*quotavo.ImpactPreview, *apperrors.BusinessError)
	ListFeatureOverrides(ctx context.Context, role string, featureKey string) ([]quotavo.FeatureOverrideItem, *apperrors.BusinessError)
	UpdateFeatureOverrides(ctx context.Context, adminID uint64, role string, featureKey string, phones []string, audit AuditInput) ([]quotavo.FeatureOverrideItem, *apperrors.BusinessError)
	DeleteFeatureOverride(ctx context.Context, adminID uint64, role string, featureKey string, overrideID uint64, audit AuditInput) ([]quotavo.FeatureOverrideItem, *apperrors.BusinessError)

	AssertProfileComplete(ctx context.Context, db *gorm.DB, userID uint64) error
	AssertCanCreateFamily(ctx context.Context, db *gorm.DB, userID uint64) error
	AssertCanJoinFamily(ctx context.Context, db *gorm.DB, userID uint64) error
	AssertCanAddMember(ctx context.Context, db *gorm.DB, familyID uint64, delta int) error
	AssertCanRestoreFamily(ctx context.Context, db *gorm.DB, familyID uint64) error
	AssertFounderTransferReceiver(ctx context.Context, db *gorm.DB, toUserID uint64, familyID uint64) error
	AssertPostFounderTransfer(ctx context.Context, db *gorm.DB, fromUserID, toUserID, familyID uint64) error

	MapQuotaError(err error) *apperrors.BusinessError
}

type quotaService struct {
	db   *gorm.DB
	repo quotarepo.Repository
}

func NewService(db *gorm.DB, repo quotarepo.Repository) Service {
	return &quotaService{db: db, repo: repo}
}

func (s *quotaService) repoFor(db *gorm.DB) quotarepo.Repository {
	if db == nil {
		return s.repo
	}
	return s.repo.WithDB(db)
}

func IsProfileComplete(user *usermodel.User) bool {
	return user != nil && user.Nickname != nil && strings.TrimSpace(*user.Nickname) != ""
}

func (s *quotaService) GetCapabilities(ctx context.Context, userID uint64) (*quotavo.Capabilities, *apperrors.BusinessError) {
	user, err := s.repo.LockUser(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeResourceNotFound)
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return s.capabilitiesForUser(ctx, s.repo, user)
}

func (s *quotaService) capabilitiesForUser(ctx context.Context, repo quotarepo.Repository, user *usermodel.User) (*quotavo.Capabilities, *apperrors.BusinessError) {
	tier := repo.ResolveTrustTier(user)
	limits, err := repo.GetLimitsForTier(ctx, tier)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	grayAccessEnabled := false
	if user.PhoneHash != nil {
		enabled, overrideErr := repo.HasFeatureOverride(ctx, GrayAccess, *user.PhoneHash)
		if overrideErr != nil {
			return nil, apperrors.New(apperrors.CodeSystemError)
		}
		grayAccessEnabled = enabled
	}
	owned, err := repo.CountOwnedFamilies(ctx, user.ID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	joined, err := repo.CountJoinedFamilies(ctx, user.ID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	ownedFamilyIDs, err := repo.ListOwnedFamilyIDs(ctx, user.ID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	membersByFamily := make(map[string]int, len(ownedFamilyIDs))
	for _, familyID := range ownedFamilyIDs {
		count, countErr := repo.CountActiveMembers(ctx, familyID)
		if countErr != nil {
			return nil, apperrors.New(apperrors.CodeSystemError)
		}
		membersByFamily[strconv.FormatUint(familyID, 10)] = count
	}
	return &quotavo.Capabilities{
		TrustTier:         tier,
		GrayAccessEnabled: grayAccessEnabled,
		Limits: quotavo.Limits{
			MaxOwnedFamilies:         limits.MaxOwnedFamilies,
			MaxMembersPerOwnedFamily: limits.MaxMembersPerOwnedFamily,
			MaxJoinedFamilies:        limits.MaxJoinedFamilies,
		},
		Usage: quotavo.Usage{
			OwnedFamilies:         owned,
			JoinedFamilies:        joined,
			MembersPerOwnedFamily: membersByFamily,
		},
	}, nil
}

func (s *quotaService) ListConfigs(ctx context.Context, role string) ([]quotavo.ConfigItem, *apperrors.BusinessError) {
	if !canManageConfig(role) {
		return nil, quotaConfigError(apperrors.CodeQuotaConfigForbidden, "无权查看账号权益配置")
	}
	rows, err := s.repo.ListConfigs(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if len(rows) == 0 {
		return defaultConfigItems(), nil
	}
	result := make([]quotavo.ConfigItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, configItemVO(row))
	}
	return result, nil
}

func (s *quotaService) UpdateConfig(ctx context.Context, adminID uint64, role string, tier string, values quotadto.ConfigValues, audit AuditInput) (*quotavo.ConfigItem, *apperrors.BusinessError) {
	if !canManageConfig(role) {
		return nil, quotaConfigError(apperrors.CodeQuotaConfigForbidden, "无权修改账号权益配置")
	}
	tier = strings.ToUpper(strings.TrimSpace(tier))
	if !quotaenum.IsValidTier(tier) {
		return nil, quotaConfigError(apperrors.CodeQuotaConfigInvalid, "权益等级不合法")
	}
	if businessErr := validateConfigValues(values); businessErr != nil {
		return nil, businessErr
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := s.repoFor(tx)
		wechatRow, phoneRow, err := txRepo.LockAllConfigsInOrder(ctx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errConfigMissing
			}
			return err
		}
		if err := validateTierOrdering(tier, values, wechatRow, phoneRow); err != nil {
			return err
		}
		if err := txRepo.UpdateConfig(ctx, tier, map[string]any{
			"max_owned_families":           values.MaxOwnedFamilies,
			"max_members_per_owned_family": values.MaxMembersPerOwnedFamily,
			"max_joined_families":          values.MaxJoinedFamilies,
			"updated_by_admin_id":          adminID,
		}); err != nil {
			return err
		}
		targetType := "ACCOUNT_QUOTA_CONFIG"
		detail, _ := json.Marshal(map[string]any{
			"trustTier":                tier,
			"maxOwnedFamilies":         values.MaxOwnedFamilies,
			"maxMembersPerOwnedFamily": values.MaxMembersPerOwnedFamily,
			"maxJoinedFamilies":        values.MaxJoinedFamilies,
		})
		return operationlog.NewGormService(tx).WriteSuccess(ctx, operationlog.WriteInput{
			OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, OperatorRole: &role,
			Module: "ACCOUNT_QUOTA", Action: "UPDATE_ACCOUNT_QUOTA_CONFIG", TargetType: &targetType,
			DetailJSON: detail,
			IP:         stringPtr(audit.IP), UserAgent: stringPtr(audit.UserAgent),
		})
	})
	if err != nil {
		if errors.Is(err, errTierOrder) {
			return nil, quotaConfigError(apperrors.CodeQuotaConfigTierOrder, "高等级权益不得低于基础等级")
		}
		if errors.Is(err, errConfigMissing) {
			return nil, apperrors.New(apperrors.CodeSystemError)
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	row, err := s.repo.FindConfigByTier(ctx, tier)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	item := configItemVO(*row)
	return &item, nil
}

func (s *quotaService) PreviewImpact(ctx context.Context, role string, tier string, values quotadto.ConfigValues) (*quotavo.ImpactPreview, *apperrors.BusinessError) {
	if !canManageConfig(role) {
		return nil, quotaConfigError(apperrors.CodeQuotaConfigForbidden, "无权预览权益影响范围")
	}
	tier = strings.ToUpper(strings.TrimSpace(tier))
	if !quotaenum.IsValidTier(tier) {
		return nil, quotaConfigError(apperrors.CodeQuotaConfigInvalid, "权益等级不合法")
	}
	if businessErr := validateConfigValues(values); businessErr != nil {
		return nil, businessErr
	}
	usersAffected, err := s.repo.CountUsersExceedingDistinct(ctx, tier, values.MaxOwnedFamilies, values.MaxJoinedFamilies)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	familiesMembers, err := s.repo.CountFamiliesExceedingMembers(ctx, tier, values.MaxMembersPerOwnedFamily)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &quotavo.ImpactPreview{
		TrustTier:                tier,
		AffectedUsers:            usersAffected,
		AffectedFamilies:         familiesMembers,
		MaxOwnedFamilies:         values.MaxOwnedFamilies,
		MaxMembersPerOwnedFamily: values.MaxMembersPerOwnedFamily,
		MaxJoinedFamilies:        values.MaxJoinedFamilies,
	}, nil
}

func (s *quotaService) ListFeatureOverrides(ctx context.Context, role string, featureKey string) ([]quotavo.FeatureOverrideItem, *apperrors.BusinessError) {
	if !canManageConfig(role) {
		return nil, quotaConfigError(apperrors.CodeQuotaConfigForbidden, "无权查看账号权益配置")
	}
	featureKey, businessErr := normalizeFeatureKey(featureKey)
	if businessErr != nil {
		return nil, businessErr
	}
	rows, err := s.repo.ListFeatureOverrides(ctx, featureKey)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := make([]quotavo.FeatureOverrideItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, featureOverrideVO(row))
	}
	return result, nil
}

func (s *quotaService) UpdateFeatureOverrides(ctx context.Context, adminID uint64, role string, featureKey string, phones []string, audit AuditInput) ([]quotavo.FeatureOverrideItem, *apperrors.BusinessError) {
	if !canManageConfig(role) {
		return nil, quotaConfigError(apperrors.CodeQuotaConfigForbidden, "无权修改账号权益配置")
	}
	featureKey, businessErr := normalizeFeatureKey(featureKey)
	if businessErr != nil {
		return nil, businessErr
	}
	rows, businessErr := buildFeatureOverrideRows(featureKey, phones, adminID)
	if businessErr != nil {
		return nil, businessErr
	}
	if len(rows) == 0 {
		return s.ListFeatureOverrides(ctx, role, featureKey)
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := s.repoFor(tx)
		if err := txRepo.AddFeatureOverrides(ctx, rows); err != nil {
			return err
		}
		targetType := "ACCOUNT_FEATURE_OVERRIDE"
		masks := make([]string, 0, len(rows))
		for _, row := range rows {
			masks = append(masks, row.PhoneMask)
		}
		detail, _ := json.Marshal(map[string]any{
			"featureKey": featureKey,
			"phoneMasks": masks,
			"count":      len(rows),
		})
		return operationlog.NewGormService(tx).WriteSuccess(ctx, operationlog.WriteInput{
			OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, OperatorRole: &role,
			Module: "ACCOUNT_QUOTA", Action: "ADD_ACCOUNT_FEATURE_OVERRIDES", TargetType: &targetType,
			DetailJSON: detail,
			IP:         stringPtr(audit.IP), UserAgent: stringPtr(audit.UserAgent),
		})
	})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return s.ListFeatureOverrides(ctx, role, featureKey)
}

func (s *quotaService) DeleteFeatureOverride(ctx context.Context, adminID uint64, role string, featureKey string, overrideID uint64, audit AuditInput) ([]quotavo.FeatureOverrideItem, *apperrors.BusinessError) {
	if !canManageConfig(role) {
		return nil, quotaConfigError(apperrors.CodeQuotaConfigForbidden, "无权修改账号权益配置")
	}
	if overrideID == 0 {
		return nil, quotaConfigError(apperrors.CodeQuotaConfigInvalid, "灰度名单记录不合法")
	}
	featureKey, businessErr := normalizeFeatureKey(featureKey)
	if businessErr != nil {
		return nil, businessErr
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := s.repoFor(tx)
		rowsAffected, err := txRepo.DeleteFeatureOverride(ctx, featureKey, overrideID)
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		targetType := "ACCOUNT_FEATURE_OVERRIDE"
		detail, _ := json.Marshal(map[string]any{
			"featureKey": featureKey,
			"overrideId": overrideID,
		})
		return operationlog.NewGormService(tx).WriteSuccess(ctx, operationlog.WriteInput{
			OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, OperatorRole: &role,
			Module: "ACCOUNT_QUOTA", Action: "DELETE_ACCOUNT_FEATURE_OVERRIDE", TargetType: &targetType,
			TargetID:   &overrideID,
			DetailJSON: detail,
			IP:         stringPtr(audit.IP), UserAgent: stringPtr(audit.UserAgent),
		})
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeResourceNotFound)
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return s.ListFeatureOverrides(ctx, role, featureKey)
}

func (s *quotaService) AssertProfileComplete(ctx context.Context, db *gorm.DB, userID uint64) error {
	repo := s.repoFor(db)
	user, err := repo.LockUser(ctx, userID)
	if err != nil {
		return err
	}
	if !IsProfileComplete(user) {
		return &quotaViolation{
			code:      apperrors.CodeProfileIncomplete,
			message:   "请先完善昵称",
			trustTier: repo.ResolveTrustTier(user),
			quotaType: "profile",
		}
	}
	return nil
}

func (s *quotaService) AssertCanCreateFamily(ctx context.Context, db *gorm.DB, userID uint64) error {
	repo := s.repoFor(db)
	user, err := repo.LockUser(ctx, userID)
	if err != nil {
		return err
	}
	tier := repo.ResolveTrustTier(user)
	limits, err := repo.GetLimitsForTier(ctx, tier)
	if err != nil {
		return err
	}
	usage, err := repo.CountOwnedFamilies(ctx, userID)
	if err != nil {
		return err
	}
	if usage+1 > limits.MaxOwnedFamilies {
		return exceedError(tier, "ownedFamilies", limits.MaxOwnedFamilies, usage)
	}
	return nil
}

func (s *quotaService) AssertCanJoinFamily(ctx context.Context, db *gorm.DB, userID uint64) error {
	repo := s.repoFor(db)
	user, err := repo.LockUser(ctx, userID)
	if err != nil {
		return err
	}
	tier := repo.ResolveTrustTier(user)
	limits, err := repo.GetLimitsForTier(ctx, tier)
	if err != nil {
		return err
	}
	usage, err := repo.CountJoinedFamilies(ctx, userID)
	if err != nil {
		return err
	}
	if usage+1 > limits.MaxJoinedFamilies {
		return exceedError(tier, "joinedFamilies", limits.MaxJoinedFamilies, usage)
	}
	return nil
}

func (s *quotaService) AssertCanAddMember(ctx context.Context, db *gorm.DB, familyID uint64, delta int) error {
	if delta <= 0 {
		return nil
	}
	repo := s.repoFor(db)
	founderUserID, err := repo.FindFounderUserID(ctx, familyID)
	if err != nil {
		return err
	}
	user, err := repo.LockUser(ctx, founderUserID)
	if err != nil {
		return err
	}
	tier := repo.ResolveTrustTier(user)
	limits, err := repo.GetLimitsForTier(ctx, tier)
	if err != nil {
		return err
	}
	usage, err := repo.CountActiveMembers(ctx, familyID)
	if err != nil {
		return err
	}
	if usage+delta > limits.MaxMembersPerOwnedFamily {
		return exceedError(tier, "membersPerOwnedFamily", limits.MaxMembersPerOwnedFamily, usage)
	}
	return nil
}

func (s *quotaService) AssertCanRestoreFamily(ctx context.Context, db *gorm.DB, familyID uint64) error {
	repo := s.repoFor(db)
	founderUserID, err := repo.FindFounderUserID(ctx, familyID)
	if err != nil {
		return err
	}
	joinedUserIDs, err := repo.ListJoinedUserIDsByFamily(ctx, familyID)
	if err != nil {
		return err
	}
	lockIDs := append([]uint64{founderUserID}, joinedUserIDs...)
	if err := lockUsersSorted(ctx, repo, lockIDs); err != nil {
		return err
	}
	if err := s.assertOwnedWithinLimit(ctx, repo, founderUserID, 1); err != nil {
		return err
	}
	if err := s.assertFamilyMembersWithinLimit(ctx, repo, familyID, founderUserID); err != nil {
		return err
	}
	for _, userID := range joinedUserIDs {
		if err := s.assertJoinedWithinLimit(ctx, repo, userID, 1); err != nil {
			return err
		}
	}
	return nil
}

func (s *quotaService) AssertFounderTransferReceiver(ctx context.Context, db *gorm.DB, toUserID uint64, familyID uint64) error {
	repo := s.repoFor(db)
	user, err := repo.LockUser(ctx, toUserID)
	if err != nil {
		return err
	}
	tier := repo.ResolveTrustTier(user)
	limits, err := repo.GetLimitsForTier(ctx, tier)
	if err != nil {
		return err
	}
	owned, err := repo.CountOwnedFamilies(ctx, toUserID)
	if err != nil {
		return err
	}
	if owned+1 > limits.MaxOwnedFamilies {
		return exceedError(tier, "ownedFamilies", limits.MaxOwnedFamilies, owned)
	}
	return nil
}

func (s *quotaService) AssertPostFounderTransfer(ctx context.Context, db *gorm.DB, fromUserID, toUserID, familyID uint64) error {
	repo := s.repoFor(db)
	if err := s.AssertFounderTransferReceiver(ctx, db, toUserID, familyID); err != nil {
		return err
	}
	fromUser, err := repo.LockUser(ctx, fromUserID)
	if err != nil {
		return err
	}
	fromTier := repo.ResolveTrustTier(fromUser)
	fromLimits, err := repo.GetLimitsForTier(ctx, fromTier)
	if err != nil {
		return err
	}
	fromJoined, err := repo.CountJoinedFamilies(ctx, fromUserID)
	if err != nil {
		return err
	}
	if fromJoined+1 > fromLimits.MaxJoinedFamilies {
		return exceedError(fromTier, "joinedFamilies", fromLimits.MaxJoinedFamilies, fromJoined)
	}
	toUser, err := repo.LockUser(ctx, toUserID)
	if err != nil {
		return err
	}
	toTier := repo.ResolveTrustTier(toUser)
	toLimits, err := repo.GetLimitsForTier(ctx, toTier)
	if err != nil {
		return err
	}
	toOwned, err := repo.CountOwnedFamilies(ctx, toUserID)
	if err != nil {
		return err
	}
	if toOwned+1 > toLimits.MaxOwnedFamilies {
		return exceedError(toTier, "ownedFamilies", toLimits.MaxOwnedFamilies, toOwned)
	}
	return nil
}

func (s *quotaService) MapQuotaError(err error) *apperrors.BusinessError {
	if err == nil {
		return nil
	}
	if errors.Is(err, quotarepo.ErrFounderNotFound) {
		return apperrors.New(apperrors.CodeResourceNotFound)
	}
	var violation *quotaViolation
	if errors.As(err, &violation) {
		return &apperrors.BusinessError{
			Code:    violation.code,
			Message: violation.message,
			Data: map[string]any{
				"trustTier": violation.trustTier,
				"quotaType": violation.quotaType,
				"limit":     violation.limit,
				"usage":     violation.usage,
			},
		}
	}
	return apperrors.New(apperrors.CodeSystemError)
}

type quotaViolation struct {
	code      apperrors.Code
	message   string
	trustTier string
	quotaType string
	limit     int
	usage     int
}

func (e *quotaViolation) Error() string {
	return e.message
}

func exceedError(tier, quotaType string, limit, usage int) error {
	code := apperrors.CodeQuotaMembersExceeded
	message := "已达到家庭成员数量上限"
	switch quotaType {
	case "ownedFamilies":
		code = apperrors.CodeQuotaOwnedFamiliesExceeded
		message = "已达到可创建家庭数量上限"
	case "joinedFamilies":
		code = apperrors.CodeQuotaJoinedFamiliesExceeded
		message = "已达到可加入家庭数量上限"
	}
	return &quotaViolation{
		code: code, message: message, trustTier: tier, quotaType: quotaType, limit: limit, usage: usage,
	}
}

func canManageConfig(role string) bool {
	return role == string(enums.AdminRoleRootAdmin) || role == string(enums.AdminRoleSuperAdmin)
}

func validateConfigValues(values quotadto.ConfigValues) *apperrors.BusinessError {
	if values.MaxOwnedFamilies < 0 || values.MaxMembersPerOwnedFamily < 1 || values.MaxJoinedFamilies < 0 {
		return quotaConfigError(apperrors.CodeQuotaConfigInvalid, "权益配置参数不合法")
	}
	return nil
}

func normalizeFeatureKey(featureKey string) (string, *apperrors.BusinessError) {
	featureKey = strings.ToUpper(strings.TrimSpace(featureKey))
	if featureKey != GrayAccess {
		return "", quotaConfigError(apperrors.CodeQuotaConfigInvalid, "权益功能不合法")
	}
	return featureKey, nil
}

func buildFeatureOverrideRows(featureKey string, phones []string, adminID uint64) ([]quotamodel.AccountFeatureOverride, *apperrors.BusinessError) {
	rows := make([]quotamodel.AccountFeatureOverride, 0, len(phones))
	seen := make(map[string]struct{}, len(phones))
	for _, phone := range phones {
		phone = strings.TrimSpace(phone)
		if phone == "" {
			continue
		}
		if !security.ValidPhone(phone) {
			return nil, quotaConfigError(apperrors.CodeQuotaConfigInvalid, "手机号格式不合法")
		}
		phoneHash := security.PhoneHash(phone)
		if _, ok := seen[phoneHash]; ok {
			continue
		}
		seen[phoneHash] = struct{}{}
		rows = append(rows, quotamodel.AccountFeatureOverride{
			FeatureKey:       featureKey,
			PhoneHash:        phoneHash,
			PhoneMask:        commonmask.Phone(phone),
			UpdatedByAdminID: &adminID,
		})
	}
	return rows, nil
}

func validateTierOrdering(updatingTier string, values quotadto.ConfigValues, wechat, phone quotamodel.AccountQuotaConfig) error {
	if updatingTier == quotaenum.TrustTierWechatOnly {
		wechat.MaxOwnedFamilies = values.MaxOwnedFamilies
		wechat.MaxMembersPerOwnedFamily = values.MaxMembersPerOwnedFamily
		wechat.MaxJoinedFamilies = values.MaxJoinedFamilies
	} else {
		phone.MaxOwnedFamilies = values.MaxOwnedFamilies
		phone.MaxMembersPerOwnedFamily = values.MaxMembersPerOwnedFamily
		phone.MaxJoinedFamilies = values.MaxJoinedFamilies
	}
	if phone.MaxOwnedFamilies < wechat.MaxOwnedFamilies ||
		phone.MaxMembersPerOwnedFamily < wechat.MaxMembersPerOwnedFamily ||
		phone.MaxJoinedFamilies < wechat.MaxJoinedFamilies {
		return errTierOrder
	}
	return nil
}

func lockUsersSorted(ctx context.Context, repo quotarepo.Repository, userIDs []uint64) error {
	if len(userIDs) == 0 {
		return nil
	}
	unique := make([]uint64, 0, len(userIDs))
	seen := make(map[uint64]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID == 0 {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		unique = append(unique, userID)
	}
	sort.Slice(unique, func(i, j int) bool { return unique[i] < unique[j] })
	for _, userID := range unique {
		if _, err := repo.LockUser(ctx, userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *quotaService) assertOwnedWithinLimit(ctx context.Context, repo quotarepo.Repository, userID uint64, delta int) error {
	user, err := repo.LockUser(ctx, userID)
	if err != nil {
		return err
	}
	tier := repo.ResolveTrustTier(user)
	limits, err := repo.GetLimitsForTier(ctx, tier)
	if err != nil {
		return err
	}
	usage, err := repo.CountOwnedFamilies(ctx, userID)
	if err != nil {
		return err
	}
	if usage+delta > limits.MaxOwnedFamilies {
		return exceedError(tier, "ownedFamilies", limits.MaxOwnedFamilies, usage)
	}
	return nil
}

func (s *quotaService) assertJoinedWithinLimit(ctx context.Context, repo quotarepo.Repository, userID uint64, delta int) error {
	user, err := repo.LockUser(ctx, userID)
	if err != nil {
		return err
	}
	tier := repo.ResolveTrustTier(user)
	limits, err := repo.GetLimitsForTier(ctx, tier)
	if err != nil {
		return err
	}
	usage, err := repo.CountJoinedFamilies(ctx, userID)
	if err != nil {
		return err
	}
	if usage+delta > limits.MaxJoinedFamilies {
		return exceedError(tier, "joinedFamilies", limits.MaxJoinedFamilies, usage)
	}
	return nil
}

func (s *quotaService) assertFamilyMembersWithinLimit(ctx context.Context, repo quotarepo.Repository, familyID, founderUserID uint64) error {
	user, err := repo.LockUser(ctx, founderUserID)
	if err != nil {
		return err
	}
	tier := repo.ResolveTrustTier(user)
	limits, err := repo.GetLimitsForTier(ctx, tier)
	if err != nil {
		return err
	}
	usage, err := repo.CountActiveMembers(ctx, familyID)
	if err != nil {
		return err
	}
	if usage > limits.MaxMembersPerOwnedFamily {
		return exceedError(tier, "membersPerOwnedFamily", limits.MaxMembersPerOwnedFamily, usage)
	}
	return nil
}

func configItemVO(row quotamodel.AccountQuotaConfig) quotavo.ConfigItem {
	return quotavo.ConfigItem{
		TrustTier:                row.TrustTier,
		MaxOwnedFamilies:         row.MaxOwnedFamilies,
		MaxMembersPerOwnedFamily: row.MaxMembersPerOwnedFamily,
		MaxJoinedFamilies:        row.MaxJoinedFamilies,
		UpdatedByAdminID:         row.UpdatedByAdminID,
		UpdatedAt:                row.UpdatedAt.Format(time.RFC3339),
	}
}

func featureOverrideVO(row quotamodel.AccountFeatureOverride) quotavo.FeatureOverrideItem {
	return quotavo.FeatureOverrideItem{
		ID:         row.ID,
		FeatureKey: row.FeatureKey,
		PhoneMask:  row.PhoneMask,
		UpdatedAt:  row.UpdatedAt.Format(time.RFC3339),
	}
}

func defaultConfigItems() []quotavo.ConfigItem {
	now := time.Now().Format(time.RFC3339)
	return []quotavo.ConfigItem{
		{TrustTier: quotaenum.TrustTierWechatOnly, MaxOwnedFamilies: 1, MaxMembersPerOwnedFamily: 10, MaxJoinedFamilies: 1, UpdatedAt: now},
		{TrustTier: quotaenum.TrustTierPhoneBound, MaxOwnedFamilies: 1, MaxMembersPerOwnedFamily: 20, MaxJoinedFamilies: 5, UpdatedAt: now},
	}
}

func quotaConfigError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

var (
	errTierOrder     = errors.New("tier order invalid")
	errConfigMissing = errors.New("config missing")
)
