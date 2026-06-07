package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	familymodel "tree/backend/internal/family/core/model"
	invitationmodel "tree/backend/internal/family/invitation/model"
	membermodel "tree/backend/internal/family/member/model"
	rolemodel "tree/backend/internal/family/role/model"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"
)

type InvitationRow struct {
	invitationmodel.FamilyInvitation
	FamilyName       string `gorm:"column:family_name"`
	TargetMemberName string `gorm:"column:target_member_name"`
}

type Repository interface {
	WithTx(*gorm.DB) Repository
	FindFamily(context.Context, uint64, bool) (*familymodel.Family, error)
	FindMember(context.Context, uint64, uint64, bool) (*membermodel.FamilyMember, error)
	FindUser(context.Context, uint64, bool) (*usermodel.User, error)
	FindActiveLinkByMember(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error)
	FindActiveLinkByUser(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error)
	FindPendingByMember(context.Context, uint64, uint64, time.Time) (*invitationmodel.FamilyInvitation, error)
	Create(context.Context, *invitationmodel.FamilyInvitation) error
	FindByID(context.Context, uint64, bool) (*invitationmodel.FamilyInvitation, error)
	FindRowByID(context.Context, uint64) (*InvitationRow, error)
	FindByTokenHash(context.Context, string) (*InvitationRow, error)
	ListForUser(context.Context, uint64) ([]InvitationRow, error)
	UpdateStatus(context.Context, uint64, string, map[string]any) error
	CreateLink(context.Context, *rolemodel.FamilyMemberUserLink) error
	WriteLog(context.Context, operationlog.WriteInput) error
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository         { return &GormRepository{db: db} }
func (r *GormRepository) WithTx(tx *gorm.DB) Repository { return &GormRepository{db: tx} }

func (r *GormRepository) FindFamily(ctx context.Context, id uint64, lock bool) (*familymodel.Family, error) {
	var value familymodel.Family
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.Where("id = ? AND deleted_at IS NULL", id).First(&value).Error
	return &value, err
}

func (r *GormRepository) FindMember(ctx context.Context, familyID, memberID uint64, lock bool) (*membermodel.FamilyMember, error) {
	var value membermodel.FamilyMember
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.Where("id = ? AND family_id = ? AND status = ? AND deleted_at IS NULL", memberID, familyID, "ACTIVE").First(&value).Error
	return &value, err
}

func (r *GormRepository) FindUser(ctx context.Context, userID uint64, lock bool) (*usermodel.User, error) {
	var value usermodel.User
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.Where("id = ? AND deleted_at IS NULL", userID).First(&value).Error
	return &value, err
}

func (r *GormRepository) FindActiveLinkByMember(ctx context.Context, familyID, memberID uint64) (*rolemodel.FamilyMemberUserLink, error) {
	var value rolemodel.FamilyMemberUserLink
	err := r.db.WithContext(ctx).Where("family_id = ? AND member_id = ? AND link_status = ?", familyID, memberID, "ACTIVE").First(&value).Error
	return &value, err
}

func (r *GormRepository) FindActiveLinkByUser(ctx context.Context, familyID, userID uint64) (*rolemodel.FamilyMemberUserLink, error) {
	var value rolemodel.FamilyMemberUserLink
	err := r.db.WithContext(ctx).Where("family_id = ? AND user_id = ? AND link_status = ?", familyID, userID, "ACTIVE").First(&value).Error
	return &value, err
}

func (r *GormRepository) FindPendingByMember(ctx context.Context, familyID, memberID uint64, now time.Time) (*invitationmodel.FamilyInvitation, error) {
	var value invitationmodel.FamilyInvitation
	err := r.db.WithContext(ctx).
		Where("family_id = ? AND target_member_id = ? AND status = ? AND expired_at > ?", familyID, memberID, "PENDING", now).
		First(&value).Error
	return &value, err
}

func (r *GormRepository) Create(ctx context.Context, value *invitationmodel.FamilyInvitation) error {
	return r.db.WithContext(ctx).Create(value).Error
}

func (r *GormRepository) FindByID(ctx context.Context, id uint64, lock bool) (*invitationmodel.FamilyInvitation, error) {
	var value invitationmodel.FamilyInvitation
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.First(&value, id).Error
	return &value, err
}

func (r *GormRepository) FindRowByID(ctx context.Context, id uint64) (*InvitationRow, error) {
	var value InvitationRow
	err := r.db.WithContext(ctx).Table("family_invitations AS i").
		Select("i.*, f.family_name, m.display_name AS target_member_name").
		Joins("JOIN families AS f ON f.id = i.family_id").
		Joins("JOIN family_members AS m ON m.id = i.target_member_id").
		Where("i.id = ?", id).First(&value).Error
	return &value, err
}

func (r *GormRepository) FindByTokenHash(ctx context.Context, hash string) (*InvitationRow, error) {
	var value InvitationRow
	err := r.db.WithContext(ctx).Table("family_invitations AS i").
		Select("i.*, f.family_name, m.display_name AS target_member_name").
		Joins("JOIN families AS f ON f.id = i.family_id").
		Joins("JOIN family_members AS m ON m.id = i.target_member_id").
		Where("i.invite_token = ?", hash).First(&value).Error
	return &value, err
}

func (r *GormRepository) ListForUser(ctx context.Context, userID uint64) ([]InvitationRow, error) {
	var values []InvitationRow
	err := r.db.WithContext(ctx).Table("family_invitations AS i").
		Select("i.*, f.family_name, m.display_name AS target_member_name").
		Joins("JOIN families AS f ON f.id = i.family_id").
		Joins("JOIN family_members AS m ON m.id = i.target_member_id").
		Where("i.target_user_id = ?", userID).Order("i.id DESC").Scan(&values).Error
	return values, err
}

func (r *GormRepository) UpdateStatus(ctx context.Context, id uint64, current string, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&invitationmodel.FamilyInvitation{}).Where("id = ? AND status = ?", id, current).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) CreateLink(ctx context.Context, value *rolemodel.FamilyMemberUserLink) error {
	return r.db.WithContext(ctx).Create(value).Error
}

func (r *GormRepository) WriteLog(ctx context.Context, input operationlog.WriteInput) error {
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
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return fn(u.repo.WithTx(tx)) })
}
