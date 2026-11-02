package repository

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	"tree/backend/internal/family/invitationreward/model"
)

type Repository interface {
	WithDB(*gorm.DB) Repository
	GetConfig(context.Context, bool) (*model.Config, error)
	ListTiers(context.Context, uint64) ([]model.Tier, error)
	ReplaceConfig(context.Context, *model.Config, []model.Tier) error
	LockFamily(context.Context, uint64) error
	CreateReferral(context.Context, *model.Referral) error
	FindReusableReferral(context.Context, uint64, uint64, time.Time) (*model.Referral, error)
	FindReferral(context.Context, string, bool) (*model.Referral, error)
	CountQualified(context.Context, uint64) (int, error)
	SumActiveGrants(context.Context, uint64) (int, error)
	ListGrantedTierIDs(context.Context, uint64) ([]uint64, error)
	CreateGrant(context.Context, *model.Grant) error
	HasPhone(context.Context, uint64) (bool, error)
	HasCompletedProfile(context.Context, uint64) (bool, error)
}
type repo struct{ db *gorm.DB }

func New(db *gorm.DB) Repository              { return &repo{db: db} }
func (r *repo) WithDB(db *gorm.DB) Repository { return &repo{db: db} }
func (r *repo) GetConfig(ctx context.Context, lock bool) (*model.Config, error) {
	var v model.Config
	q := r.db.WithContext(ctx).Order("id")
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.First(&v).Error
	return &v, err
}
func (r *repo) ListTiers(ctx context.Context, id uint64) ([]model.Tier, error) {
	var v []model.Tier
	err := r.db.WithContext(ctx).Where("config_id = ? AND retired = ?", id, false).Order("sort_order,id").Find(&v).Error
	return v, err
}
func (r *repo) ReplaceConfig(ctx context.Context, c *model.Config, tiers []model.Tier) error {
	if err := r.db.WithContext(ctx).Model(&model.Config{}).Where("id = ?", c.ID).Updates(map[string]any{"enabled": c.Enabled, "activation_remaining_slots": c.ActivationRemainingSlots, "max_reward_members": c.MaxRewardMembers, "require_phone_bound": c.RequirePhoneBound, "updated_by_admin_id": c.UpdatedByAdminID}).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Model(&model.Tier{}).Where("config_id = ?", c.ID).Update("retired", true).Error; err != nil {
		return err
	}
	for _, tier := range tiers {
		var existing model.Tier
		err := r.db.WithContext(ctx).Where("config_id = ? AND required_invites = ?", c.ID, tier.RequiredInvites).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := r.db.WithContext(ctx).Create(&tier).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if err := r.db.WithContext(ctx).Model(&existing).Updates(map[string]any{"cumulative_reward_members": tier.CumulativeRewardMembers, "sort_order": tier.SortOrder, "retired": false}).Error; err != nil {
			return err
		}
	}
	return nil
}
func (r *repo) LockFamily(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Table("families").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", id).First(&struct{ ID uint64 }{}).Error
}
func (r *repo) CreateReferral(ctx context.Context, v *model.Referral) error {
	return r.db.WithContext(ctx).Create(v).Error
}
func (r *repo) FindReusableReferral(ctx context.Context, familyID, inviterUserID uint64, now time.Time) (*model.Referral, error) {
	var row model.Referral
	err := r.db.WithContext(ctx).Where("family_id = ? AND inviter_user_id = ? AND status = ? AND expires_at > ?", familyID, inviterUserID, model.ReferralStatusPending, now).Order("id DESC").First(&row).Error
	return &row, err
}
func (r *repo) FindReferral(ctx context.Context, code string, lock bool) (*model.Referral, error) {
	var v model.Referral
	q := r.db.WithContext(ctx).Where("referral_code = ?", code)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.First(&v).Error
	return &v, err
}
func (r *repo) CountQualified(ctx context.Context, fid uint64) (int, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Referral{}).Where("family_id = ? AND status = ?", fid, model.ReferralStatusQualified).Count(&n).Error
	return int(n), err
}
func (r *repo) SumActiveGrants(ctx context.Context, fid uint64) (int, error) {
	var n int
	err := r.db.WithContext(ctx).Model(&model.Grant{}).Select("COALESCE(SUM(added_members),0)").Where("family_id = ? AND status = ?", fid, model.GrantStatusActive).Scan(&n).Error
	return n, err
}
func (r *repo) ListGrantedTierIDs(ctx context.Context, fid uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).Model(&model.Grant{}).Where("family_id = ? AND status = ?", fid, model.GrantStatusActive).Pluck("tier_id", &ids).Error
	return ids, err
}
func (r *repo) CreateGrant(ctx context.Context, v *model.Grant) error {
	return r.db.WithContext(ctx).Create(v).Error
}
func (r *repo) HasPhone(ctx context.Context, userID uint64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("users").Where("id = ? AND phone_verified = 1 AND phone_hash IS NOT NULL", userID).Count(&n).Error
	return n > 0, err
}
func (r *repo) HasCompletedProfile(ctx context.Context, userID uint64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("users").Where("id = ? AND nickname IS NOT NULL AND TRIM(nickname) <> ''", userID).Count(&n).Error
	return n > 0, err
}
