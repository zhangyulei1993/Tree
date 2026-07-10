package repository

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"tree/backend/internal/common/enums"
	familymodel "tree/backend/internal/family/core/model"
	dissolutionmodel "tree/backend/internal/family/dissolution/model"
	membermodel "tree/backend/internal/family/member/model"
	rolemodel "tree/backend/internal/family/role/model"
	operationmodel "tree/backend/internal/operationlog/model"
	usermodel "tree/backend/internal/user/model"
)

type FamilyListRow struct {
	familymodel.Family
	FamilyRole string `gorm:"column:family_role"`
}

type OperationLogRow struct {
	operationmodel.OperationLog
}

type PublicFamilySearchQuery struct {
	Keyword       string
	FamilySurname string
	RegionText    string
	Page          int
	PageSize      int
}

const publicFamilyListSelect = "id, family_name, family_surname, native_place, region_text, description, avatar_url, " +
	"public_contact_visible, public_contact_name, public_contact_note, public_approved_at, created_at, updated_at"

type FamilyRepository interface {
	WithTx(tx *gorm.DB) FamilyRepository
	FindUserByID(ctx context.Context, userID uint64) (*usermodel.User, error)
	CreateFamily(ctx context.Context, family *familymodel.Family) error
	CreateMember(ctx context.Context, member *membermodel.FamilyMember) error
	CreateLink(ctx context.Context, link *rolemodel.FamilyMemberUserLink) error
	FindActiveLinkForUpdate(ctx context.Context, familyID uint64, userID uint64) (*rolemodel.FamilyMemberUserLink, error)
	DeactivateLink(ctx context.Context, linkID uint64, userID uint64, reason *string, at time.Time) error
	FindFamilyByID(ctx context.Context, familyID uint64) (*familymodel.Family, error)
	FindFamilyByIDForUpdate(ctx context.Context, familyID uint64) (*familymodel.Family, error)
	FindPublicFamilyByID(ctx context.Context, familyID uint64) (*familymodel.Family, error)
	SearchPublicFamilies(ctx context.Context, query PublicFamilySearchQuery) ([]familymodel.Family, int64, error)
	ListByUser(ctx context.Context, userID uint64) ([]FamilyListRow, error)
	ListOperationLogs(ctx context.Context, familyID uint64, page int, pageSize int) ([]OperationLogRow, int64, error)
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

func (r *GormFamilyRepository) FindActiveLinkForUpdate(ctx context.Context, familyID uint64, userID uint64) (*rolemodel.FamilyMemberUserLink, error) {
	var link rolemodel.FamilyMemberUserLink
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("family_id = ? AND user_id = ? AND link_status = ?", familyID, userID, string(enums.StatusActive)).
		First(&link).Error
	return &link, err
}

func (r *GormFamilyRepository) DeactivateLink(ctx context.Context, linkID uint64, userID uint64, reason *string, at time.Time) error {
	return r.db.WithContext(ctx).Model(&rolemodel.FamilyMemberUserLink{}).
		Where("id = ? AND user_id = ? AND link_status = ?", linkID, userID, string(enums.StatusActive)).
		Updates(map[string]any{
			"link_status":         "INACTIVE",
			"family_role":         string(enums.FamilyRoleMember),
			"unlinked_at":         at,
			"unlinked_reason":     reason,
			"unlinked_by_user_id": userID,
			"updated_at":          at,
		}).Error
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
		Where("id = ? AND deleted_at IS NULL AND status = ? AND public_display_status = ? AND public_display_enabled = ?", familyID, "NORMAL", "APPROVED", true).
		First(&family).Error; err != nil {
		return nil, err
	}
	return &family, nil
}

func (r *GormFamilyRepository) SearchPublicFamilies(ctx context.Context, query PublicFamilySearchQuery) ([]familymodel.Family, int64, error) {
	page, pageSize := normalizePublicSearchPage(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Model(&familymodel.Family{}).
		Select(publicFamilyListSelect).
		Where("deleted_at IS NULL").
		Where("status = ?", "NORMAL").
		Where("searchable = ?", true).
		Where("public_display_status = ?", "APPROVED").
		Where("public_display_enabled = ?", true)

	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		db = db.Where(
			"(family_name LIKE ? OR family_surname LIKE ? OR COALESCE(region_text, '') LIKE ? OR COALESCE(native_place, '') LIKE ?)",
			pattern, pattern, pattern, pattern,
		)
	}
	if surname := strings.TrimSpace(query.FamilySurname); surname != "" {
		db = db.Where("family_surname LIKE ?", "%"+surname+"%")
	}
	if region := strings.TrimSpace(query.RegionText); region != "" {
		db = db.Where("COALESCE(region_text, '') LIKE ?", "%"+region+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var families []familymodel.Family
	err := db.
		Order("public_approved_at IS NULL, public_approved_at DESC, updated_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&families).Error
	return families, total, err
}

func normalizePublicSearchPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
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

func (r *GormFamilyRepository) ListOperationLogs(ctx context.Context, familyID uint64, page int, pageSize int) ([]OperationLogRow, int64, error) {
	page, pageSize = normalizeOperationLogPage(page, pageSize)
	db := r.db.WithContext(ctx).Model(&operationmodel.OperationLog{}).
		Where("family_id = ?", familyID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []OperationLogRow
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}

func (r *GormFamilyRepository) UpdateFamily(ctx context.Context, familyID uint64, values map[string]any) error {
	return r.db.WithContext(ctx).Model(&familymodel.Family{}).Where("id = ?", familyID).Updates(values).Error
}

func normalizeOperationLogPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
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
