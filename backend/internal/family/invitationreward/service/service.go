package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	quotaservice "tree/backend/internal/accountquota/service"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/permission"
	"tree/backend/internal/family/invitationreward/dto"
	"tree/backend/internal/family/invitationreward/model"
	"tree/backend/internal/family/invitationreward/repository"
	operationlog "tree/backend/internal/operationlog/service"
)

type Service interface {
	Summary(context.Context, uint64, uint64) (*dto.Summary, *apperrors.BusinessError)
	CreateShare(context.Context, uint64, uint64) (*dto.ShareResult, *apperrors.BusinessError)
	Qualify(context.Context, uint64, string) (*dto.Summary, *apperrors.BusinessError)
	AdminGet(context.Context) (*dto.Config, *apperrors.BusinessError)
	AdminUpdate(context.Context, uint64, dto.Config) (*dto.Config, *apperrors.BusinessError)
}
type service struct {
	db          *gorm.DB
	repo        repository.Repository
	permissions permission.FamilyPermissionService
	quota       quotaservice.Service
	logs        operationlog.Service
}

func New(db *gorm.DB, repo repository.Repository, permissions permission.FamilyPermissionService, quota quotaservice.Service, logs operationlog.Service) Service {
	return &service{db: db, repo: repo, permissions: permissions, quota: quota, logs: logs}
}

func (s *service) Summary(ctx context.Context, userID, familyID uint64) (*dto.Summary, *apperrors.BusinessError) {
	if _, err := s.permissions.GetActiveLink(ctx, userID, familyID); err != nil {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	return s.summary(ctx, s.repo, familyID)
}
func (s *service) AdminGet(ctx context.Context) (*dto.Config, *apperrors.BusinessError) {
	c, err := s.repo.GetConfig(ctx, false)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	tiers, err := s.repo.ListTiers(ctx, c.ID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return configVO(c, tiers), nil
}
func (s *service) AdminUpdate(ctx context.Context, adminID uint64, req dto.Config) (*dto.Config, *apperrors.BusinessError) {
	if adminID == 0 || !validConfig(req) {
		return nil, apperrors.New(apperrors.CodeInvalidParams)
	}
	var result *dto.Config
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		r := s.repo.WithDB(tx)
		c, err := r.GetConfig(ctx, true)
		if err != nil {
			return err
		}
		c.Enabled = req.Enabled
		c.ActivationRemainingSlots = req.ActivationRemainingSlots
		c.MaxRewardMembers = req.MaxRewardMembers
		c.RequirePhoneBound = req.RequirePhoneBound
		c.UpdatedByAdminID = &adminID
		tiers := make([]model.Tier, 0, len(req.Tiers))
		for i, t := range req.Tiers {
			tiers = append(tiers, model.Tier{ConfigID: c.ID, RequiredInvites: t.RequiredInvites, CumulativeRewardMembers: t.CumulativeRewardMembers, SortOrder: (i + 1) * 10})
		}
		if err := r.ReplaceConfig(ctx, c, tiers); err != nil {
			return err
		}
		result = configVO(c, tiers)
		return nil
	})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if s.logs != nil {
		_ = s.logs.WriteSuccess(ctx, operationlog.WriteInput{OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, Module: "FAMILY_INVITATION_REWARD", Action: "UPDATE_INVITATION_REWARD_CONFIG", TargetType: strptr("FAMILY_INVITATION_REWARD_CONFIG"), DetailJSON: []byte(fmt.Sprintf(`{"tierCount":%d,"maxRewardMembers":%d}`, len(req.Tiers), req.MaxRewardMembers))})
	}
	return result, nil
}

func (s *service) CreateShare(ctx context.Context, userID, familyID uint64) (*dto.ShareResult, *apperrors.BusinessError) {
	if _, err := s.permissions.GetActiveLink(ctx, userID, familyID); err != nil {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	var result *dto.ShareResult
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		r := s.repo.WithDB(tx)
		if err := r.LockFamily(ctx, familyID); err != nil {
			return err
		}
		summary, biz := s.summary(ctx, r, familyID)
		if biz != nil {
			return biz
		}
		if !summary.Enabled || !summary.Active {
			return errors.New("invitation reward is not active")
		}
		if existing, findErr := r.FindReusableReferral(ctx, familyID, userID, time.Now()); findErr == nil {
			result = &dto.ShareResult{ReferralCode: existing.ReferralCode, Summary: *summary}
			return nil
		} else if !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}
		code, err := randomCode()
		if err != nil {
			return err
		}
		expiresAt := time.Now().AddDate(0, 0, 7)
		if err = r.CreateReferral(ctx, &model.Referral{FamilyID: familyID, InviterUserID: userID, ReferralCode: code, Status: model.ReferralStatusPending, ExpiresAt: &expiresAt}); err != nil {
			return err
		}
		result = &dto.ShareResult{ReferralCode: code, Summary: *summary}
		return nil
	})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeInvalidParams)
	}
	return result, nil
}

func (s *service) Qualify(ctx context.Context, userID uint64, code string) (*dto.Summary, *apperrors.BusinessError) {
	code = strings.TrimSpace(code)
	if userID == 0 || code == "" {
		return nil, apperrors.New(apperrors.CodeInvalidParams)
	}
	var result *dto.Summary
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		r := s.repo.WithDB(tx)
		ref, err := r.FindReferral(ctx, code, true)
		if err != nil {
			return err
		}
		if ref.Status == model.ReferralStatusQualified {
			if ref.InviteeUserID != nil && *ref.InviteeUserID == userID {
				summary, biz := s.summary(ctx, r, ref.FamilyID)
				if biz != nil {
					return biz
				}
				result = summary
				return nil
			}
			return errors.New("referral already used")
		}
		if ref.ExpiresAt == nil || !ref.ExpiresAt.After(time.Now()) {
			return errors.New("referral expired")
		}
		if ref.InviterUserID == userID {
			return errors.New("self referral")
		}
		if err := r.LockFamily(ctx, ref.FamilyID); err != nil {
			return err
		}
		c, err := r.GetConfig(ctx, true)
		if err != nil {
			return err
		}
		if !c.Enabled {
			return errors.New("reward disabled")
		}
		if c.RequirePhoneBound {
			has, err := r.HasPhone(ctx, userID)
			if err != nil || !has {
				return errors.New("phone required")
			}
		}
		now := time.Now()
		if err := tx.Model(&model.Referral{}).Where("id = ? AND status = ?", ref.ID, model.ReferralStatusPending).Updates(map[string]any{"invitee_user_id": userID, "status": model.ReferralStatusQualified, "qualified_at": now}).Error; err != nil {
			return err
		}
		tiers, err := r.ListTiers(ctx, c.ID)
		if err != nil {
			return err
		}
		count, err := r.CountQualified(ctx, ref.FamilyID)
		if err != nil {
			return err
		}
		granted, err := r.ListGrantedTierIDs(ctx, ref.FamilyID)
		if err != nil {
			return err
		}
		seen := map[uint64]bool{}
		for _, id := range granted {
			seen[id] = true
		}
		for _, tier := range tiers {
			if count >= tier.RequiredInvites && !seen[tier.ID] {
				previous := 0
				for _, other := range tiers {
					if other.RequiredInvites < tier.RequiredInvites && other.CumulativeRewardMembers > previous {
						previous = other.CumulativeRewardMembers
					}
				}
				add := tier.CumulativeRewardMembers - previous
				if add > 0 {
					if err := r.CreateGrant(ctx, &model.Grant{FamilyID: ref.FamilyID, TierID: tier.ID, AddedMembers: add, Status: model.GrantStatusActive, GrantedAt: now}); err != nil {
						return err
					}
				}
			}
		}
		summary, biz := s.summary(ctx, r, ref.FamilyID)
		if biz != nil {
			return biz
		}
		result = summary
		return nil
	})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeInvalidParams)
	}
	return result, nil
}

func (s *service) summary(ctx context.Context, r repository.Repository, familyID uint64) (*dto.Summary, *apperrors.BusinessError) {
	c, err := r.GetConfig(ctx, false)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	tiers, err := r.ListTiers(ctx, c.ID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	valid, err := r.CountQualified(ctx, familyID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	reward, err := r.SumActiveGrants(ctx, familyID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	reader, ok := s.quota.(quotaservice.FamilyCapacityReader)
	if !ok {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	cap, err := reader.GetFamilyCapacity(ctx, nil, familyID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	item := &dto.Summary{Enabled: c.Enabled, Active: c.Enabled && cap.EffectiveLimit-cap.UsedMembers <= c.ActivationRemainingSlots, FamilyID: familyID, UsedMembers: cap.UsedMembers, EffectiveLimit: cap.EffectiveLimit, ValidInvites: valid, RewardedMembers: reward, MaxRewardMembers: c.MaxRewardMembers, Tiers: tierVOs(tiers)}
	for _, tier := range tiers {
		if valid < tier.RequiredInvites {
			item.NextRequiredInvites = tier.RequiredInvites
			item.NextAddedMembers = tier.CumulativeRewardMembers - reward
			break
		}
	}
	return item, nil
}
func configVO(c *model.Config, tiers []model.Tier) *dto.Config {
	return &dto.Config{Enabled: c.Enabled, ActivationRemainingSlots: c.ActivationRemainingSlots, MaxRewardMembers: c.MaxRewardMembers, RequirePhoneBound: c.RequirePhoneBound, Tiers: tierVOs(tiers)}
}
func tierVOs(rows []model.Tier) []dto.Tier {
	out := make([]dto.Tier, 0, len(rows))
	for _, v := range rows {
		out = append(out, dto.Tier{RequiredInvites: v.RequiredInvites, CumulativeRewardMembers: v.CumulativeRewardMembers})
	}
	return out
}
func validConfig(v dto.Config) bool {
	if v.ActivationRemainingSlots < 0 || v.ActivationRemainingSlots > 20 || v.MaxRewardMembers < 1 || v.MaxRewardMembers > 100 || len(v.Tiers) == 0 || len(v.Tiers) > 10 {
		return false
	}
	lastN, lastR := 0, 0
	for _, t := range v.Tiers {
		if t.RequiredInvites <= lastN || t.CumulativeRewardMembers <= lastR || t.RequiredInvites > 100000 || t.CumulativeRewardMembers > v.MaxRewardMembers {
			return false
		}
		lastN, lastR = t.RequiredInvites, t.CumulativeRewardMembers
	}
	return lastR == v.MaxRewardMembers
}
func randomCode() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func strptr(s string) *string { return &s }
