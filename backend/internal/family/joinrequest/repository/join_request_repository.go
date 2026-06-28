package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	familymodel "tree/backend/internal/family/core/model"
	joinmodel "tree/backend/internal/family/joinrequest/model"
	membermodel "tree/backend/internal/family/member/model"
	relationshipmodel "tree/backend/internal/family/relationship/model"
	relationshiprepo "tree/backend/internal/family/relationship/repository"
	rolemodel "tree/backend/internal/family/role/model"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"
)

type JoinRequestRow struct {
	joinmodel.FamilyJoinRequest
	FamilyName string `gorm:"column:family_name"`
}

type Repository interface {
	WithTx(*gorm.DB) Repository
	FindFamily(context.Context, uint64, bool) (*familymodel.Family, error)
	FindUser(context.Context, uint64, bool) (*usermodel.User, error)
	FindMember(context.Context, uint64, uint64, bool) (*membermodel.FamilyMember, error)
	FindActiveLinkByUser(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error)
	FindActiveLinkByMember(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error)
	FindPending(context.Context, uint64, uint64) (*joinmodel.FamilyJoinRequest, error)
	Create(context.Context, *joinmodel.FamilyJoinRequest) error
	FindByID(context.Context, uint64, uint64, bool) (*joinmodel.FamilyJoinRequest, error)
	ListMine(context.Context, uint64) ([]JoinRequestRow, error)
	ListFamily(context.Context, uint64) ([]JoinRequestRow, error)
	UpdateStatus(context.Context, uint64, string, map[string]any) error
	CreateMember(context.Context, *membermodel.FamilyMember) error
	CreateLink(context.Context, *rolemodel.FamilyMemberUserLink) error
	IncrementGraphVersion(context.Context, uint64) (int64, error)
	WriteLog(context.Context, operationlog.WriteInput) error
}

type GormRepository struct {
	db            *gorm.DB
	relationships *relationshiprepo.GormRepository
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db, relationships: relationshiprepo.NewRepository(db)}
}
func (r *GormRepository) WithTx(tx *gorm.DB) Repository { return NewRepository(tx) }

func (r *GormRepository) FindFamily(ctx context.Context, id uint64, lock bool) (*familymodel.Family, error) {
	var value familymodel.Family
	q := r.db.WithContext(ctx)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.Where("id = ? AND deleted_at IS NULL", id).First(&value).Error
	return &value, err
}
func (r *GormRepository) FindUser(ctx context.Context, id uint64, lock bool) (*usermodel.User, error) {
	var value usermodel.User
	q := r.db.WithContext(ctx)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.Where("id = ? AND deleted_at IS NULL", id).First(&value).Error
	return &value, err
}
func (r *GormRepository) FindMember(ctx context.Context, familyID, memberID uint64, lock bool) (*membermodel.FamilyMember, error) {
	var value membermodel.FamilyMember
	q := r.db.WithContext(ctx)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.Where("id = ? AND family_id = ? AND status = ? AND deleted_at IS NULL", memberID, familyID, "ACTIVE").First(&value).Error
	return &value, err
}
func (r *GormRepository) FindActiveLinkByUser(ctx context.Context, familyID, userID uint64) (*rolemodel.FamilyMemberUserLink, error) {
	var value rolemodel.FamilyMemberUserLink
	err := r.db.WithContext(ctx).Where("family_id = ? AND user_id = ? AND link_status = ?", familyID, userID, "ACTIVE").First(&value).Error
	return &value, err
}
func (r *GormRepository) FindActiveLinkByMember(ctx context.Context, familyID, memberID uint64) (*rolemodel.FamilyMemberUserLink, error) {
	var value rolemodel.FamilyMemberUserLink
	err := r.db.WithContext(ctx).Where("family_id = ? AND member_id = ? AND link_status = ?", familyID, memberID, "ACTIVE").First(&value).Error
	return &value, err
}
func (r *GormRepository) FindPending(ctx context.Context, familyID, userID uint64) (*joinmodel.FamilyJoinRequest, error) {
	var value joinmodel.FamilyJoinRequest
	err := r.db.WithContext(ctx).Where("family_id = ? AND applicant_user_id = ? AND request_status = ?", familyID, userID, "PENDING").First(&value).Error
	return &value, err
}
func (r *GormRepository) Create(ctx context.Context, value *joinmodel.FamilyJoinRequest) error {
	return r.db.WithContext(ctx).Create(value).Error
}
func (r *GormRepository) FindByID(ctx context.Context, familyID, id uint64, lock bool) (*joinmodel.FamilyJoinRequest, error) {
	var value joinmodel.FamilyJoinRequest
	q := r.db.WithContext(ctx)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.Where("id = ? AND family_id = ?", id, familyID).First(&value).Error
	return &value, err
}
func (r *GormRepository) ListMine(ctx context.Context, userID uint64) ([]JoinRequestRow, error) {
	var values []JoinRequestRow
	err := r.db.WithContext(ctx).Table("family_join_requests AS j").Select("j.*, f.family_name").
		Joins("JOIN families AS f ON f.id = j.family_id").Where("j.applicant_user_id = ?", userID).
		Order("j.id DESC").Scan(&values).Error
	return values, err
}
func (r *GormRepository) ListFamily(ctx context.Context, familyID uint64) ([]JoinRequestRow, error) {
	var values []JoinRequestRow
	err := r.db.WithContext(ctx).Table("family_join_requests AS j").Select("j.*, f.family_name").
		Joins("JOIN families AS f ON f.id = j.family_id").Where("j.family_id = ?", familyID).
		Order("j.id DESC").Scan(&values).Error
	return values, err
}
func (r *GormRepository) UpdateStatus(ctx context.Context, id uint64, current string, values map[string]any) error {
	result := r.db.WithContext(ctx).Model(&joinmodel.FamilyJoinRequest{}).Where("id = ? AND request_status = ?", id, current).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (r *GormRepository) CreateMember(ctx context.Context, value *membermodel.FamilyMember) error {
	return r.db.WithContext(ctx).Create(value).Error
}

func (r *GormRepository) FindMemberForUpdate(ctx context.Context, familyID, memberID uint64) (*membermodel.FamilyMember, error) {
	return r.relationships.FindMemberForUpdate(ctx, familyID, memberID)
}

func (r *GormRepository) FindDuplicate(ctx context.Context, familyID, fromMemberID, toMemberID uint64, relationshipType string) (*relationshipmodel.FamilyRelationship, error) {
	return r.relationships.FindDuplicate(ctx, familyID, fromMemberID, toMemberID, relationshipType)
}

func (r *GormRepository) ListActiveParents(ctx context.Context, familyID, childMemberID uint64) ([]relationshipmodel.FamilyRelationship, error) {
	return r.relationships.ListActiveParents(ctx, familyID, childMemberID)
}

func (r *GormRepository) FindPrimaryParentByGender(ctx context.Context, familyID, childMemberID uint64, gender string, excludeRelationshipID uint64) (*relationshipmodel.FamilyRelationship, error) {
	return r.relationships.FindPrimaryParentByGender(ctx, familyID, childMemberID, gender, excludeRelationshipID)
}

func (r *GormRepository) FindPrimaryParentWithMemberByGender(ctx context.Context, familyID, childMemberID uint64, gender string, excludeRelationshipID uint64) (*relationshipmodel.FamilyRelationship, *membermodel.FamilyMember, error) {
	return r.relationships.FindPrimaryParentWithMemberByGender(ctx, familyID, childMemberID, gender, excludeRelationshipID)
}

func (r *GormRepository) ListActiveSpouseRelationshipsByGender(ctx context.Context, familyID, baseMemberID uint64, gender string) ([]relationshiprepo.SpouseRelationshipRow, error) {
	return r.relationships.ListActiveSpouseRelationshipsByGender(ctx, familyID, baseMemberID, gender)
}

func (r *GormRepository) CreateRelationship(ctx context.Context, value *relationshipmodel.FamilyRelationship) error {
	return r.relationships.CreateRelationship(ctx, value)
}

func (r *GormRepository) UpdateRelationship(ctx context.Context, familyID, relationshipID uint64, values map[string]any) error {
	return r.relationships.UpdateRelationship(ctx, familyID, relationshipID, values)
}
func (r *GormRepository) CreateLink(ctx context.Context, value *rolemodel.FamilyMemberUserLink) error {
	return r.db.WithContext(ctx).Create(value).Error
}
func (r *GormRepository) IncrementGraphVersion(ctx context.Context, familyID uint64) (int64, error) {
	if err := r.db.WithContext(ctx).Model(&familymodel.Family{}).Where("id = ?", familyID).
		UpdateColumn("graph_version", gorm.Expr("graph_version + 1")).Error; err != nil {
		return 0, err
	}
	var family familymodel.Family
	err := r.db.WithContext(ctx).Select("graph_version").First(&family, familyID).Error
	return family.GraphVersion, err
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
