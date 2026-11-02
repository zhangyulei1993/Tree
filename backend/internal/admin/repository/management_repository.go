package repository

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	adminmodel "tree/backend/internal/admin/model"
	"tree/backend/internal/common/enums"
	familymodel "tree/backend/internal/family/core/model"
	membermodel "tree/backend/internal/family/member/model"
	relationshipenum "tree/backend/internal/family/relationship/enum"
	relationshipmodel "tree/backend/internal/family/relationship/model"
	rolemodel "tree/backend/internal/family/role/model"
	operationmodel "tree/backend/internal/operationlog/model"
	usermodel "tree/backend/internal/user/model"
)

type PageQuery struct {
	Keyword      string
	Status       string
	Role         string
	OperatorType string
	FamilyID     uint64
	Result       string
	StartAt      time.Time
	EndAt        time.Time
	Page         int
	PageSize     int
}

type Dashboard struct {
	Users                        int64 `json:"users"`
	Families                     int64 `json:"families"`
	PendingPublicApplications    int64 `json:"pendingPublicApplications"`
	PendingFounderTransfers      int64 `json:"pendingFounderTransfers"`
	PendingDissolutions          int64 `json:"pendingDissolutions"`
	PaidCapacityOrders           int64 `json:"paidCapacityOrders"`
	ActiveCapacityRevenueFen     int64 `json:"activeCapacityRevenueFen"`
	PendingCapacityRefunds       int64 `json:"pendingCapacityRefunds"`
	ReviewCapacityRefunds        int64 `json:"reviewCapacityRefunds"`
	FailedCapacityRefunds        int64 `json:"failedCapacityRefunds"`
	PendingOperations            int64 `json:"pendingOperations"`
	NewUsersLast7Days            int64 `json:"newUsersLast7Days"`
	NewFamiliesLast7Days         int64 `json:"newFamiliesLast7Days"`
	AcceptedInvitationsLast7Days int64 `json:"acceptedInvitationsLast7Days"`
}

// RecoveryItem is deliberately limited to structural records. Restoring an account
// binding automatically would be unsafe because the account may have changed since deletion.
type RecoveryItem struct {
	Kind         string    `json:"kind"`
	ID           uint64    `json:"id"`
	FamilyID     uint64    `json:"familyId"`
	FamilyName   string    `json:"familyName"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	DeletedAt    time.Time `json:"deletedAt"`
	DeleteReason *string   `json:"deleteReason,omitempty"`
}

type UserRow struct {
	ID             uint64     `json:"id"`
	Phone          *string    `json:"phone,omitempty"`
	Nickname       *string    `json:"nickname,omitempty"`
	RealName       *string    `json:"realName,omitempty"`
	AccountOrigin  string     `json:"accountOrigin"`
	RegisterClient string     `json:"registerClient"`
	PhoneVerified  bool       `json:"phoneVerified"`
	HasWechatLogin bool       `json:"hasWechatLogin"`
	TrustTier      string     `json:"trustTier"`
	LoginMethod    string     `json:"loginMethod"`
	Status         string     `json:"status"`
	LastLoginAt    *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type FamilyRow struct {
	ID                     uint64    `json:"id"`
	FamilyName             string    `json:"familyName"`
	FamilySurname          string    `json:"familySurname"`
	NativePlace            *string   `json:"nativePlace,omitempty"`
	RegionText             *string   `json:"regionText,omitempty"`
	Description            *string   `json:"description,omitempty"`
	Status                 string    `json:"status"`
	Searchable             bool      `json:"searchable"`
	PublicDisplayStatus    string    `json:"publicDisplayStatus"`
	CurrentFounderMemberID *uint64   `json:"currentFounderMemberId,omitempty"`
	GraphVersion           int64     `json:"graphVersion"`
	MemberCount            int64     `json:"memberCount"`
	CreatedAt              time.Time `json:"createdAt"`
}

type MemberRow struct {
	ID                uint64  `json:"id"`
	FamilyID          uint64  `json:"familyId"`
	DisplayName       string  `json:"displayName"`
	Gender            string  `json:"gender"`
	Status            string  `json:"status"`
	IsLiving          *bool   `json:"isLiving,omitempty"`
	UserBindingPolicy string  `json:"userBindingPolicy"`
	BoundUserID       *uint64 `json:"boundUserId,omitempty"`
	FamilyRole        *string `json:"familyRole,omitempty"`
}

type FamilyTreeNode struct {
	MemberID          uint64     `json:"memberId"`
	DisplayName       string     `json:"displayName"`
	Surname           *string    `json:"surname,omitempty"`
	Gender            string     `json:"gender"`
	MemberType        string     `json:"memberType"`
	BirthDate         *time.Time `json:"birthDate,omitempty"`
	IsLiving          *bool      `json:"isLiving,omitempty"`
	UserBindingPolicy string     `json:"userBindingPolicy"`
	UserBindingState  string     `json:"userBindingState"`
}

type FamilyTreeEdge struct {
	RelationshipID   uint64  `json:"relationshipId"`
	FromMemberID     uint64  `json:"fromMemberId"`
	ToMemberID       uint64  `json:"toMemberId"`
	RelationshipType string  `json:"relationshipType"`
	ParentLinkType   *string `json:"parentLinkType,omitempty"`
}

type FamilyTree struct {
	FamilyID     uint64           `json:"familyId"`
	GraphVersion int64            `json:"graphVersion"`
	Nodes        []FamilyTreeNode `json:"nodes"`
	Edges        []FamilyTreeEdge `json:"edges"`
}

type AdminAudit struct {
	AdminID   uint64
	Role      string
	IP        string
	UserAgent string
}

type UpdateFamilyMemberInput struct {
	DisplayName       *string
	Gender            *string
	BirthDate         **time.Time
	IsLiving          **bool
	UserBindingPolicy *string
}

type DeleteFamilyMemberInput struct {
	Reason *string
}

type CreateFamilyRelativeInput struct {
	AddType      string
	NewMember    NewFamilyRelativeMemberInput
	Relationship FamilyRelativeRelationshipInput
}

type NewFamilyRelativeMemberInput struct {
	DisplayName       string
	Gender            string
	BirthDate         *time.Time
	IsLiving          *bool
	UserBindingPolicy string
	MemberType        string
}

type FamilyRelativeRelationshipInput struct {
	ParentLinkType *string
}

type CreatedFamilyRelative struct {
	MemberID        uint64   `json:"memberId"`
	RelationshipIDs []uint64 `json:"relationshipIds"`
	GraphVersion    int64    `json:"graphVersion"`
}

type UserFamilyLink struct {
	FamilyID   uint64 `json:"familyId"`
	FamilyName string `json:"familyName"`
	MemberID   uint64 `json:"memberId"`
	MemberName string `json:"memberName"`
	FamilyRole string `json:"familyRole"`
}

type UserDetail struct {
	User     UserRow          `json:"user"`
	Families []UserFamilyLink `json:"families"`
}

type AdminRow struct {
	ID          uint64     `json:"id"`
	Username    string     `json:"username"`
	DisplayName *string    `json:"displayName,omitempty"`
	Phone       *string    `json:"phone,omitempty"`
	Email       *string    `json:"email,omitempty"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	LockedUntil *time.Time `json:"lockedUntil,omitempty"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type LogRow struct {
	ID              uint64          `json:"id"`
	OperatorType    string          `json:"operatorType"`
	OperatorAdminID *uint64         `json:"operatorAdminId,omitempty"`
	OperatorUserID  *uint64         `json:"operatorUserId,omitempty"`
	OperatorRole    *string         `json:"operatorRole,omitempty"`
	Module          string          `json:"module"`
	Action          string          `json:"action"`
	TargetType      *string         `json:"targetType,omitempty"`
	TargetID        *uint64         `json:"targetId,omitempty"`
	FamilyID        *uint64         `json:"familyId,omitempty"`
	MemberID        *uint64         `json:"memberId,omitempty"`
	UserID          *uint64         `json:"userId,omitempty"`
	DetailJSON      json.RawMessage `json:"detailJson,omitempty"`
	Result          string          `json:"result"`
	ErrorMessage    *string         `json:"errorMessage,omitempty"`
	IP              *string         `json:"ip,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
}

type ManagementRepository struct {
	db          *gorm.DB
	wechatAppID string
}

type CreateAdminInput struct {
	Username     string
	PasswordHash string
	DisplayName  *string
	Phone        *string
	Email        *string
	Role         string
	Remark       *string
}

type UpdateAdminInput struct {
	DisplayName *string
	Phone       *string
	Email       *string
	Role        *string
	Remark      *string
}

type UpdateFamilyInput struct {
	FamilyName  *string
	Description *string
	NativePlace *string
	RegionText  *string
	Searchable  *bool
}

var (
	errBindingConflict       = errors.New("member binding conflict")
	errFamilyStatus          = errors.New("family status does not allow operation")
	errFounderProtected      = errors.New("founder member cannot be deleted")
	errMemberHasChildren     = errors.New("member has children")
	errPrivilegedMember      = errors.New("privileged member cannot be deleted")
	errSiblingNoParents      = errors.New("sibling parent required")
	errDuplicateRelation     = errors.New("duplicate relationship")
	errPrimaryFather         = errors.New("primary father exists")
	errPrimaryMother         = errors.New("primary mother exists")
	errPrimaryParentOpen     = errors.New("primary parent gender unknown")
	errBaseGender            = errors.New("base gender required")
	errChildGender           = errors.New("child gender required")
	errSiblingGender         = errors.New("sibling gender required")
	errSpouseGender          = errors.New("spouse gender mismatch")
	errSpouseBase            = errors.New("spouse base expansion forbidden")
	errParentType            = errors.New("parent member type invalid")
	errLineageParent         = errors.New("lineage parent required")
	errRestoreEndpoints      = errors.New("relationship endpoints unavailable")
	errRestoreMemberConflict = errors.New("member restore position conflict")
	errAdminDuplicate        = errors.New("admin username already exists")
	errAdminNotFound         = errors.New("admin not found")
	errRootProtected         = errors.New("root admin protected")
)

func ErrBindingConflict() error       { return errBindingConflict }
func ErrFamilyStatus() error          { return errFamilyStatus }
func ErrFounderProtected() error      { return errFounderProtected }
func ErrMemberHasChildren() error     { return errMemberHasChildren }
func ErrPrivilegedMember() error      { return errPrivilegedMember }
func ErrSiblingNoParents() error      { return errSiblingNoParents }
func ErrDuplicateRelation() error     { return errDuplicateRelation }
func ErrPrimaryFather() error         { return errPrimaryFather }
func ErrPrimaryMother() error         { return errPrimaryMother }
func ErrPrimaryParentOpen() error     { return errPrimaryParentOpen }
func ErrBaseGender() error            { return errBaseGender }
func ErrChildGender() error           { return errChildGender }
func ErrSiblingGender() error         { return errSiblingGender }
func ErrSpouseGender() error          { return errSpouseGender }
func ErrSpouseBase() error            { return errSpouseBase }
func ErrParentType() error            { return errParentType }
func ErrLineageParent() error         { return errLineageParent }
func ErrRestoreEndpoints() error      { return errRestoreEndpoints }
func ErrRestoreMemberConflict() error { return errRestoreMemberConflict }
func ErrAdminDuplicate() error        { return errAdminDuplicate }
func ErrAdminNotFound() error         { return errAdminNotFound }
func ErrRootProtected() error         { return errRootProtected }

func NewManagementRepository(db *gorm.DB, wechatAppID string) *ManagementRepository {
	return &ManagementRepository{db: db, wechatAppID: wechatAppID}
}

func page(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

func (r *ManagementRepository) Dashboard(ctx context.Context) (*Dashboard, error) {
	result := &Dashboard{}
	queries := []struct {
		target       *int64
		table, where string
	}{
		{&result.Users, "users", "deleted_at IS NULL"},
		{&result.Families, "families", "deleted_at IS NULL"},
		{&result.PendingPublicApplications, "family_public_applications", "application_status = 'PENDING'"},
		{&result.PendingFounderTransfers, "family_founder_transfer_requests", "request_status = 'PENDING'"},
		{&result.PendingDissolutions, "family_dissolution_requests", "request_status = 'PENDING'"},
	}
	for _, query := range queries {
		if err := r.db.WithContext(ctx).Table(query.table).Where(query.where).Count(query.target).Error; err != nil {
			return nil, err
		}
	}
	if err := r.db.WithContext(ctx).Table("family_capacity_orders").
		Where("status IN ?", []string{"PAID", "REFUNDING", "REFUNDED"}).
		Count(&result.PaidCapacityOrders).Error; err != nil {
		return nil, err
	}
	// Only fully retained paid orders are treated as current revenue. Orders in refund flow
	// stay out of this figure, so the dashboard does not overstate available revenue.
	if err := r.db.WithContext(ctx).Table("family_capacity_orders").
		Select("COALESCE(SUM(amount_fen), 0)").
		Where("status = ?", "PAID").
		Scan(&result.ActiveCapacityRevenueFen).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("family_capacity_refunds").
		Where("status = ?", "PENDING").
		Count(&result.PendingCapacityRefunds).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("family_capacity_refunds").
		Where("status = ?", "REVIEW_REQUIRED").
		Count(&result.ReviewCapacityRefunds).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table("family_capacity_refunds").
		Where("status = ?", "FAILED").
		Count(&result.FailedCapacityRefunds).Error; err != nil {
		return nil, err
	}
	result.PendingOperations = result.PendingPublicApplications + result.PendingFounderTransfers + result.PendingDissolutions + result.PendingCapacityRefunds + result.ReviewCapacityRefunds + result.FailedCapacityRefunds
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	for _, query := range []struct {
		table  string
		where  string
		target *int64
	}{
		{"users", "deleted_at IS NULL AND created_at >= ?", &result.NewUsersLast7Days},
		{"families", "deleted_at IS NULL AND created_at >= ?", &result.NewFamiliesLast7Days},
		{"family_invitations", "status = 'ACCEPTED' AND accepted_at >= ?", &result.AcceptedInvitationsLast7Days},
	} {
		if err := r.db.WithContext(ctx).Table(query.table).Where(query.where, sevenDaysAgo).Count(query.target).Error; err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (r *ManagementRepository) RecoveryItems(ctx context.Context, query PageQuery) ([]RecoveryItem, int64, error) {
	// The recovery centre intentionally focuses on the last 30 days. Older structural
	// changes remain auditable in operation logs but should be restored through a manual review.
	cutoff := time.Now().AddDate(0, 0, -30)
	items := make([]RecoveryItem, 0)
	var members []struct {
		ID           uint64
		FamilyID     uint64
		FamilyName   string
		DisplayName  string
		DeletedAt    time.Time
		DeleteReason *string
	}
	if err := r.db.WithContext(ctx).Table("family_members AS m").
		Select("m.id, m.family_id, f.family_name, m.display_name, m.deleted_at, m.delete_reason").
		Joins("JOIN families AS f ON f.id = m.family_id").
		Where("m.status = 'DELETED' AND m.deleted_at IS NOT NULL AND m.recovery_discarded_at IS NULL AND m.deleted_at >= ?", cutoff).
		Order("m.deleted_at DESC").Limit(200).Scan(&members).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range members {
		items = append(items, RecoveryItem{Kind: "MEMBER", ID: row.ID, FamilyID: row.FamilyID, FamilyName: row.FamilyName, Title: row.DisplayName, Description: "成员节点及同次删除的可恢复关系", DeletedAt: row.DeletedAt, DeleteReason: row.DeleteReason})
	}
	var relations []struct {
		ID               uint64
		FamilyID         uint64
		FamilyName       string
		FromName         string
		ToName           string
		RelationshipType string
		DeletedAt        time.Time
		DeleteReason     *string
	}
	if err := r.db.WithContext(ctx).Table("family_relationships AS r").
		Select("r.id, r.family_id, f.family_name, fm.display_name AS from_name, tm.display_name AS to_name, r.relationship_type, r.deleted_at, r.delete_reason").
		Joins("JOIN families AS f ON f.id = r.family_id").
		Joins("JOIN family_members AS fm ON fm.id = r.from_member_id").
		Joins("JOIN family_members AS tm ON tm.id = r.to_member_id").
		Where("r.status = 'DELETED' AND r.deleted_at IS NOT NULL AND r.deleted_at >= ?", cutoff).
		Order("r.deleted_at DESC").Limit(200).Scan(&relations).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range relations {
		description := "父母子女关系"
		if row.RelationshipType == string(enums.RelationshipTypeSpouse) {
			description = "配偶关系"
		}
		items = append(items, RecoveryItem{Kind: "RELATIONSHIP", ID: row.ID, FamilyID: row.FamilyID, FamilyName: row.FamilyName, Title: row.FromName + " - " + row.ToName, Description: description, DeletedAt: row.DeletedAt, DeleteReason: row.DeleteReason})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].DeletedAt.After(items[j].DeletedAt) })
	p, size := page(query.Page, query.PageSize)
	total := int64(len(items))
	start := (p - 1) * size
	if start >= len(items) {
		return []RecoveryItem{}, total, nil
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}

func (r *ManagementRepository) RestoreMember(ctx context.Context, familyID, memberID uint64, audit AdminAudit, beforeRestore func(context.Context, *gorm.DB, uint64) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var family familymodel.Family
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", familyID).First(&family).Error; err != nil {
			return err
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyStatus
		}
		var member membermodel.FamilyMember
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND family_id = ? AND status = 'DELETED' AND deleted_at IS NOT NULL AND recovery_discarded_at IS NULL", memberID, familyID).First(&member).Error; err != nil {
			return err
		}
		conflict, err := hasRestoreMemberPositionConflict(ctx, tx, &member)
		if err != nil {
			return err
		}
		if conflict {
			return errRestoreMemberConflict
		}
		if beforeRestore != nil {
			if err := beforeRestore(ctx, tx, familyID); err != nil {
				return err
			}
		}
		before := member
		if err := tx.Model(&membermodel.FamilyMember{}).Where("id = ?", memberID).Updates(map[string]any{"status": "ACTIVE", "deleted_at": nil, "deleted_by_user_id": nil, "deleted_by_admin_id": nil, "delete_reason": nil, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		// Admin deletion gives members and their linked relationships the same timestamp.
		// Restore only that batch and only when both endpoints are currently active.
		relationQuery := tx.Model(&relationshipmodel.FamilyRelationship{}).Where("family_id = ? AND status = 'DELETED' AND deleted_at = ? AND (from_member_id = ? OR to_member_id = ?)", familyID, before.DeletedAt, memberID, memberID)
		if err := relationQuery.Updates(map[string]any{"status": "ACTIVE", "deleted_at": nil, "deleted_by_user_id": nil, "deleted_by_admin_id": nil, "delete_reason": nil, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		if err := tx.Model(&familymodel.Family{}).Where("id = ?", familyID).UpdateColumn("graph_version", gorm.Expr("graph_version + 1")).Error; err != nil {
			return err
		}
		var after membermodel.FamilyMember
		if err := tx.Where("id = ?", memberID).First(&after).Error; err != nil {
			return err
		}
		detail, _ := json.Marshal(map[string]any{"accountBinding": "not_restored", "relationshipRecovery": "same_delete_batch"})
		return writeAdminMemberLog(ctx, tx, audit, "RESTORE_MEMBER", familyID, memberID, before, after, detail)
	})
}

func hasRestoreMemberPositionConflict(ctx context.Context, tx *gorm.DB, member *membermodel.FamilyMember) (bool, error) {
	var count int64
	err := tx.WithContext(ctx).Table("family_relationships AS historical").
		Joins(`INNER JOIN family_members AS replacement
			ON replacement.family_id = historical.family_id
			AND replacement.status = 'ACTIVE'
			AND replacement.deleted_at IS NULL
			AND replacement.id <> ?
			AND replacement.display_name = ?`, member.ID, member.DisplayName).
		Joins(`INNER JOIN family_relationships AS current
			ON current.family_id = historical.family_id
			AND current.status = 'ACTIVE'
			AND current.deleted_at IS NULL
			AND current.relationship_type = historical.relationship_type
			AND (
				(historical.from_member_id = ? AND current.from_member_id = replacement.id AND current.to_member_id = historical.to_member_id)
				OR
				(historical.to_member_id = ? AND current.to_member_id = replacement.id AND current.from_member_id = historical.from_member_id)
			)`, member.ID, member.ID).
		Where("historical.family_id = ? AND historical.status = ? AND historical.deleted_at = ?", member.FamilyID, "DELETED", member.DeletedAt).
		Where("historical.from_member_id = ? OR historical.to_member_id = ?", member.ID, member.ID).
		Count(&count).Error
	if err != nil || count > 0 {
		return count > 0, err
	}
	err = tx.WithContext(ctx).Model(&membermodel.FamilyMember{}).
		Where("family_id = ? AND status = ? AND deleted_at IS NULL AND id <> ?", member.FamilyID, "ACTIVE", member.ID).
		Where("display_name = ? AND gender = ?", member.DisplayName, member.Gender).
		Where("birth_date <=> ?", member.BirthDate).
		Count(&count).Error
	return count > 0, err
}

func (r *ManagementRepository) RestoreRelationship(ctx context.Context, familyID, relationshipID uint64, audit AdminAudit) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var family familymodel.Family
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", familyID).First(&family).Error; err != nil {
			return err
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyStatus
		}
		var relation relationshipmodel.FamilyRelationship
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND family_id = ? AND status = 'DELETED' AND deleted_at IS NOT NULL", relationshipID, familyID).First(&relation).Error; err != nil {
			return err
		}
		var activeEndpoints int64
		if err := tx.Model(&membermodel.FamilyMember{}).Where("family_id = ? AND id IN ? AND status = 'ACTIVE' AND deleted_at IS NULL", familyID, []uint64{relation.FromMemberID, relation.ToMemberID}).Count(&activeEndpoints).Error; err != nil {
			return err
		}
		if activeEndpoints != 2 {
			return errRestoreEndpoints
		}
		var duplicate int64
		if err := tx.Model(&relationshipmodel.FamilyRelationship{}).Where("family_id = ? AND status = 'ACTIVE' AND deleted_at IS NULL AND from_member_id = ? AND to_member_id = ? AND relationship_type = ?", familyID, relation.FromMemberID, relation.ToMemberID, relation.RelationshipType).Count(&duplicate).Error; err != nil {
			return err
		}
		if duplicate > 0 {
			return errDuplicateRelation
		}
		if err := tx.Model(&relationshipmodel.FamilyRelationship{}).Where("id = ?", relationshipID).Updates(map[string]any{"status": "ACTIVE", "deleted_at": nil, "deleted_by_user_id": nil, "deleted_by_admin_id": nil, "delete_reason": nil, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		if err := tx.Model(&familymodel.Family{}).Where("id = ?", familyID).UpdateColumn("graph_version", gorm.Expr("graph_version + 1")).Error; err != nil {
			return err
		}
		detail, _ := json.Marshal(map[string]any{"relationshipId": relationshipID})
		return writeAdminRelationshipLog(ctx, tx, audit, "RESTORE_RELATIONSHIP", familyID, relation.FromMemberID, []uint64{relationshipID}, detail)
	})
}

func (r *ManagementRepository) Users(ctx context.Context, query PageQuery) ([]UserRow, int64, error) {
	p, size := page(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Model(&usermodel.User{}).Where("deleted_at IS NULL")
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		db = db.Where("CAST(id AS CHAR) LIKE ? OR COALESCE(phone, '') LIKE ? OR COALESCE(nickname, '') LIKE ? OR COALESCE(real_name, '') LIKE ?", pattern, pattern, pattern, pattern)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var models []usermodel.User
	if err := db.Order("id DESC").Offset((p - 1) * size).Limit(size).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]UserRow, 0, len(models))
	wechatFlags, err := r.wechatLoginUserIDs(ctx, models)
	if err != nil {
		return nil, 0, err
	}
	for _, value := range models {
		rows = append(rows, userRow(value, wechatFlags[value.ID]))
	}
	return rows, total, nil
}

const wechatMiniProvider = "WECHAT_MINI"

func (r *ManagementRepository) wechatLoginUserIDs(ctx context.Context, users []usermodel.User) (map[uint64]bool, error) {
	result := make(map[uint64]bool, len(users))
	if len(users) == 0 {
		return result, nil
	}
	ids := make([]uint64, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	var rows []struct {
		UserID uint64 `gorm:"column:user_id"`
	}
	err := r.db.WithContext(ctx).Table("user_auth_identities").
		Select("DISTINCT user_id").
		Where(
			"user_id IN ? AND provider = ? AND provider_app_id = ? AND identity_status = ? AND deleted_at IS NULL",
			ids, wechatMiniProvider, r.wechatAppID, "ACTIVE",
		).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.UserID] = true
	}
	return result, nil
}

func resolveLoginMethod(hasWechatLogin bool) string {
	if hasWechatLogin {
		return "微信登录"
	}
	return "未设置"
}

func userRow(value usermodel.User, hasWechatLogin bool) UserRow {
	trustTier := "WECHAT_ONLY"
	if value.PhoneVerified {
		trustTier = "PHONE_BOUND"
	}
	return UserRow{
		ID: value.ID, Phone: maskPhone(value.Phone), Nickname: value.Nickname, RealName: value.RealName,
		AccountOrigin: value.AccountOrigin, RegisterClient: value.RegisterClient, PhoneVerified: value.PhoneVerified,
		HasWechatLogin: hasWechatLogin,
		TrustTier:      trustTier, LoginMethod: resolveLoginMethod(hasWechatLogin),
		Status: value.Status, LastLoginAt: value.LastLoginAt, CreatedAt: value.CreatedAt,
	}
}
func maskPhone(value *string) *string {
	if value == nil {
		return nil
	}
	phone := *value
	if len(phone) == 11 {
		phone = phone[:3] + "****" + phone[7:]
	}
	return &phone
}

func (r *ManagementRepository) UserDetail(ctx context.Context, id uint64) (*UserDetail, error) {
	var user usermodel.User
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		return nil, err
	}
	var links []UserFamilyLink
	err := r.db.WithContext(ctx).Table("family_member_user_links AS link").Select("link.family_id, family.family_name, link.member_id, member.display_name AS member_name, link.family_role").Joins("JOIN families AS family ON family.id = link.family_id").Joins("JOIN family_members AS member ON member.id = link.member_id").Where("link.user_id = ? AND link.link_status = 'ACTIVE'", id).Scan(&links).Error
	if err != nil {
		return nil, err
	}
	wechatFlags, err := r.wechatLoginUserIDs(ctx, []usermodel.User{user})
	if err != nil {
		return nil, err
	}
	return &UserDetail{User: userRow(user, wechatFlags[user.ID]), Families: links}, nil
}

func (r *ManagementRepository) SetUserStatus(ctx context.Context, audit AdminAudit, id uint64, status, reason string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user usermodel.User
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
			return errAdminNotFound
		}
		now := time.Now()
		values := map[string]any{"status": status, "updated_at": now}
		action := "ENABLE_USER"
		if status == "DISABLED" {
			values["disabled_at"] = now
			values["disabled_by_admin_id"] = audit.AdminID
			values["disabled_reason"] = reason
			action = "DISABLE_USER"
		}
		if status == "DELETED" || status == "CANCELLED" {
			values["status"] = "CANCELLED"
			values["cancelled_at"] = now
			values["cancel_reason"] = reason
			values["deleted_at"] = now
			action = "DELETE_USER"
		}
		if err := tx.Model(&user).Updates(values).Error; err != nil {
			return err
		}
		return writeAdminLifecycleLog(ctx, tx, audit, action, "user", id, user.Status, status, []byte(reason))
	})
}

func (r *ManagementRepository) Families(ctx context.Context, query PageQuery) ([]FamilyRow, int64, error) {
	p, size := page(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Table("families AS family").Where("family.deleted_at IS NULL")
	if query.Status != "" {
		db = db.Where("family.status = ?", query.Status)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		db = db.Where("CAST(family.id AS CHAR) LIKE ? OR family.family_name LIKE ? OR family.family_surname LIKE ? OR COALESCE(family.region_text, '') LIKE ?", pattern, pattern, pattern, pattern)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []FamilyRow
	err := db.Select("family.id, family.family_name, family.family_surname, family.native_place, family.region_text, family.description, family.status, family.searchable, family.public_display_status, family.current_founder_member_id, family.graph_version, family.created_at, (SELECT COUNT(*) FROM family_members m WHERE m.family_id = family.id AND m.status = 'ACTIVE' AND m.deleted_at IS NULL) AS member_count").Order("family.id DESC").Offset((p - 1) * size).Limit(size).Scan(&rows).Error
	return rows, total, err
}

func (r *ManagementRepository) Family(ctx context.Context, id uint64) (*FamilyRow, error) {
	var row FamilyRow
	result := r.db.WithContext(ctx).
		Table("families AS family").
		Select("family.id, family.family_name, family.family_surname, family.native_place, family.region_text, family.description, family.status, family.searchable, family.public_display_status, family.current_founder_member_id, family.graph_version, family.created_at, (SELECT COUNT(*) FROM family_members m WHERE m.family_id = family.id AND m.status = 'ACTIVE' AND m.deleted_at IS NULL) AS member_count").
		Where("family.id = ? AND family.deleted_at IS NULL", id).
		Take(&row)
	if result.Error != nil {
		return nil, result.Error
	}
	return &row, nil
}

func (r *ManagementRepository) Members(ctx context.Context, familyID uint64) ([]MemberRow, error) {
	var rows []MemberRow
	err := r.db.WithContext(ctx).Table("family_members AS member").Select("member.id, member.family_id, member.display_name, member.gender, member.status, member.is_living, member.user_binding_policy, link.user_id AS bound_user_id, link.family_role").Joins("LEFT JOIN family_member_user_links AS link ON link.member_id = member.id AND link.family_id = member.family_id AND link.link_status = 'ACTIVE'").Where("member.family_id = ? AND member.deleted_at IS NULL", familyID).Order("member.id").Scan(&rows).Error
	return rows, err
}

func (r *ManagementRepository) Tree(ctx context.Context, familyID uint64) (*FamilyTree, error) {
	family, err := r.Family(ctx, familyID)
	if err != nil {
		return nil, err
	}
	result := &FamilyTree{FamilyID: family.ID, GraphVersion: family.GraphVersion}
	if err := r.db.WithContext(ctx).
		Table("family_members AS member").
		Select(`member.id AS member_id, member.display_name, member.surname, member.gender, member.member_type, member.birth_date, member.is_living, member.user_binding_policy,
			CASE
				WHEN member.user_binding_policy = 'NOT_REQUIRED' THEN 'NOT_REQUIRED'
				WHEN link.user_id IS NOT NULL THEN 'BOUND'
				ELSE 'UNBOUND'
			END AS user_binding_state`).
		Joins("LEFT JOIN family_member_user_links AS link ON link.member_id = member.id AND link.family_id = member.family_id AND link.link_status = 'ACTIVE'").
		Where("member.family_id = ? AND member.status = 'ACTIVE' AND member.deleted_at IS NULL", familyID).
		Order("member.id").
		Scan(&result.Nodes).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).
		Table("family_relationships AS relation").
		Select("relation.id AS relationship_id, relation.from_member_id, relation.to_member_id, relation.relationship_type, relation.parent_link_type").
		Where("relation.family_id = ? AND relation.status = 'ACTIVE' AND relation.deleted_at IS NULL", familyID).
		Order("relation.id").
		Scan(&result.Edges).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r *ManagementRepository) UpdateFamilyMember(ctx context.Context, familyID uint64, memberID uint64, input UpdateFamilyMemberInput, audit AdminAudit) (*MemberRow, error) {
	var row MemberRow
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var family familymodel.Family
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND deleted_at IS NULL", familyID).
			First(&family).Error; err != nil {
			return err
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyStatus
		}
		var before membermodel.FamilyMember
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND family_id = ? AND status = ? AND deleted_at IS NULL", memberID, familyID, "ACTIVE").
			First(&before).Error; err != nil {
			return err
		}
		values := map[string]any{}
		if input.DisplayName != nil {
			values["display_name"] = *input.DisplayName
		}
		if input.Gender != nil {
			values["gender"] = *input.Gender
		}
		if input.BirthDate != nil {
			values["birth_date"] = *input.BirthDate
		}
		if input.IsLiving != nil {
			values["is_living"] = *input.IsLiving
		}
		if input.UserBindingPolicy != nil {
			if *input.UserBindingPolicy == string(enums.UserBindingNotRequired) {
				var link rolemodel.FamilyMemberUserLink
				err := tx.Where("family_id = ? AND member_id = ? AND link_status = ?", familyID, memberID, "ACTIVE").First(&link).Error
				if err == nil {
					return errBindingConflict
				}
				if err != nil && err != gorm.ErrRecordNotFound {
					return err
				}
			}
			values["user_binding_policy"] = *input.UserBindingPolicy
		}
		if len(values) == 0 {
			row = memberRowFromModel(before)
			return nil
		}
		if err := tx.Model(&membermodel.FamilyMember{}).
			Where("id = ? AND family_id = ? AND status = ? AND deleted_at IS NULL", memberID, familyID, "ACTIVE").
			Updates(values).Error; err != nil {
			return err
		}
		if err := tx.Model(&familymodel.Family{}).
			Where("id = ? AND deleted_at IS NULL", familyID).
			UpdateColumn("graph_version", gorm.Expr("graph_version + 1")).Error; err != nil {
			return err
		}
		var after membermodel.FamilyMember
		if err := tx.Where("id = ? AND family_id = ?", memberID, familyID).First(&after).Error; err != nil {
			return err
		}
		if err := writeAdminMemberLog(ctx, tx, audit, "UPDATE_MEMBER", familyID, memberID, before, after, nil); err != nil {
			return err
		}
		row = memberRowFromModel(after)
		return nil
	})
	return &row, err
}

func (r *ManagementRepository) DeleteFamilyMember(ctx context.Context, familyID uint64, memberID uint64, input DeleteFamilyMemberInput, audit AdminAudit) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var family familymodel.Family
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND deleted_at IS NULL", familyID).
			First(&family).Error; err != nil {
			return err
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyStatus
		}
		if family.CurrentFounderMemberID != nil && *family.CurrentFounderMemberID == memberID {
			return errFounderProtected
		}
		var member membermodel.FamilyMember
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND family_id = ? AND status = ? AND deleted_at IS NULL", memberID, familyID, "ACTIVE").
			First(&member).Error; err != nil {
			return err
		}
		var childCount int64
		if err := tx.Model(&relationshipmodel.FamilyRelationship{}).
			Where("family_id = ? AND status = ? AND deleted_at IS NULL", familyID, "ACTIVE").
			Where("relationship_type = ? AND from_member_id = ?", string(enums.RelationshipTypeParentChild), memberID).
			Count(&childCount).Error; err != nil {
			return err
		}
		if childCount > 0 {
			return errMemberHasChildren
		}
		var link rolemodel.FamilyMemberUserLink
		if err := tx.Where("family_id = ? AND member_id = ? AND link_status = ?", familyID, memberID, "ACTIVE").First(&link).Error; err == nil {
			if link.FamilyRole == string(enums.FamilyRoleFounder) || link.FamilyRole == string(enums.FamilyRoleFamilyAdmin) {
				return errPrivilegedMember
			}
			now := time.Now()
			if err := tx.Model(&rolemodel.FamilyMemberUserLink{}).
				Where("id = ? AND link_status = ?", link.ID, "ACTIVE").
				Updates(map[string]any{
					"link_status":          "UNLINKED",
					"unlinked_at":          now,
					"unlinked_reason":      stringPtr("MEMBER_DELETED_BY_ADMIN"),
					"unlinked_by_admin_id": audit.AdminID,
					"updated_at":           now,
				}).Error; err != nil {
				return err
			}
		} else if err != gorm.ErrRecordNotFound {
			return err
		}
		now := time.Now()
		relationResult := tx.Model(&relationshipmodel.FamilyRelationship{}).
			Where("family_id = ? AND status = ? AND deleted_at IS NULL", familyID, "ACTIVE").
			Where("from_member_id = ? OR to_member_id = ?", memberID, memberID).
			Updates(map[string]any{
				"status":              "DELETED",
				"deleted_at":          now,
				"deleted_by_admin_id": audit.AdminID,
				"delete_reason":       input.Reason,
				"updated_at":          now,
			})
		if relationResult.Error != nil {
			return relationResult.Error
		}
		if err := tx.Model(&membermodel.FamilyMember{}).
			Where("id = ? AND family_id = ? AND status = ? AND deleted_at IS NULL", memberID, familyID, "ACTIVE").
			Updates(map[string]any{
				"status":              "DELETED",
				"deleted_at":          now,
				"deleted_by_admin_id": audit.AdminID,
				"delete_reason":       input.Reason,
				"updated_at":          now,
			}).Error; err != nil {
			return err
		}
		if err := tx.Model(&familymodel.Family{}).
			Where("id = ? AND deleted_at IS NULL", familyID).
			UpdateColumn("graph_version", gorm.Expr("graph_version + 1")).Error; err != nil {
			return err
		}
		detail, _ := json.Marshal(map[string]any{"deletedRelationships": relationResult.RowsAffected})
		return writeAdminMemberLog(ctx, tx, audit, "DELETE_MEMBER", familyID, memberID, member, nil, detail)
	})
}

func (r *ManagementRepository) CreateFamilyRelative(ctx context.Context, familyID uint64, baseMemberID uint64, input CreateFamilyRelativeInput, audit AdminAudit) (*CreatedFamilyRelative, error) {
	result := &CreatedFamilyRelative{RelationshipIDs: []uint64{}}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var family familymodel.Family
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND deleted_at IS NULL", familyID).
			First(&family).Error; err != nil {
			return err
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyStatus
		}
		var base membermodel.FamilyMember
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND family_id = ? AND status = ? AND deleted_at IS NULL", baseMemberID, familyID, "ACTIVE").
			First(&base).Error; err != nil {
			return err
		}
		if strings.ToUpper(strings.TrimSpace(base.MemberType)) == "SPOUSE" {
			return errSpouseBase
		}
		if err := validateAdminRelativePlacement(tx, familyID, &base, input); err != nil {
			return err
		}
		member := membermodel.FamilyMember{
			FamilyID: familyID, MemberType: input.NewMember.MemberType, DisplayName: input.NewMember.DisplayName,
			Gender: input.NewMember.Gender, BirthDate: input.NewMember.BirthDate, IsLiving: input.NewMember.IsLiving,
			UserBindingPolicy: input.NewMember.UserBindingPolicy, Status: string(enums.StatusActive), CreatedByAdminID: &audit.AdminID,
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}
		relationships, err := r.adminRelativeRelationships(ctx, tx, familyID, &base, &member, input, audit.AdminID)
		if err != nil {
			return err
		}
		for i := range relationships {
			if err := tx.Create(relationships[i]).Error; err != nil {
				return err
			}
			result.RelationshipIDs = append(result.RelationshipIDs, relationships[i].ID)
		}
		if err := tx.Model(&familymodel.Family{}).
			Where("id = ? AND deleted_at IS NULL", familyID).
			UpdateColumn("graph_version", gorm.Expr("graph_version + 1")).Error; err != nil {
			return err
		}
		var graphVersion int64
		if err := tx.Model(&familymodel.Family{}).Where("id = ?", familyID).Select("graph_version").Scan(&graphVersion).Error; err != nil {
			return err
		}
		result.MemberID = member.ID
		result.GraphVersion = graphVersion
		detail, _ := json.Marshal(map[string]any{
			"addType":         input.AddType,
			"baseMemberId":    baseMemberID,
			"createdMemberId": member.ID,
			"relationshipIds": result.RelationshipIDs,
		})
		if err := writeAdminMemberLog(ctx, tx, audit, "CREATE_RELATIVE", familyID, member.ID, nil, member, detail); err != nil {
			return err
		}
		return writeAdminRelationshipLog(ctx, tx, audit, "CREATE_RELATIVE_RELATIONSHIPS", familyID, baseMemberID, result.RelationshipIDs, detail)
	})
	return result, err
}

func validateAdminRelativePlacement(tx *gorm.DB, familyID uint64, base *membermodel.FamilyMember, input CreateFamilyRelativeInput) error {
	addType := relationshipenum.AddType(input.AddType)
	switch addType {
	case relationshipenum.AddTypeFather:
		if input.NewMember.Gender != string(enums.GenderMale) {
			return errChildGender
		}
		if input.NewMember.MemberType != "LINEAGE_MEMBER" && input.NewMember.MemberType != "SPOUSE" {
			return errParentType
		}
		if input.NewMember.MemberType == "SPOUSE" {
			if _, err := findAdminPrimaryLineageParent(tx, familyID, base.ID, string(enums.GenderFemale)); err != nil {
				return errLineageParent
			}
		}
		return ensureAdminPrimaryParentAvailable(tx, familyID, base.ID, input.NewMember.Gender, input.Relationship.ParentLinkType)
	case relationshipenum.AddTypeMother:
		if input.NewMember.Gender != string(enums.GenderFemale) {
			return errChildGender
		}
		if input.NewMember.MemberType != "LINEAGE_MEMBER" && input.NewMember.MemberType != "SPOUSE" {
			return errParentType
		}
		if input.NewMember.MemberType == "SPOUSE" {
			if _, err := findAdminPrimaryLineageParent(tx, familyID, base.ID, string(enums.GenderMale)); err != nil {
				return errLineageParent
			}
		}
		return ensureAdminPrimaryParentAvailable(tx, familyID, base.ID, input.NewMember.Gender, input.Relationship.ParentLinkType)
	case relationshipenum.AddTypeChild:
		if base.Gender != string(enums.GenderMale) && base.Gender != string(enums.GenderFemale) {
			return errBaseGender
		}
		if input.NewMember.Gender != string(enums.GenderMale) && input.NewMember.Gender != string(enums.GenderFemale) {
			return errChildGender
		}
	case relationshipenum.AddTypeSpouse:
		if base.Gender != string(enums.GenderMale) && base.Gender != string(enums.GenderFemale) {
			return errBaseGender
		}
		if input.NewMember.Gender != oppositeAdminGender(base.Gender) {
			return errSpouseGender
		}
	case relationshipenum.AddTypeSibling:
		if input.NewMember.Gender != string(enums.GenderMale) && input.NewMember.Gender != string(enums.GenderFemale) {
			return errSiblingGender
		}
		var count int64
		if err := tx.Model(&relationshipmodel.FamilyRelationship{}).
			Where("family_id = ? AND relationship_type = ? AND to_member_id = ? AND status = ? AND deleted_at IS NULL", familyID, string(enums.RelationshipTypeParentChild), base.ID, "ACTIVE").
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errSiblingNoParents
		}
	default:
		return errParentType
	}
	return nil
}

func (r *ManagementRepository) adminRelativeRelationships(ctx context.Context, tx *gorm.DB, familyID uint64, base *membermodel.FamilyMember, member *membermodel.FamilyMember, input CreateFamilyRelativeInput, adminID uint64) ([]*relationshipmodel.FamilyRelationship, error) {
	addType := relationshipenum.AddType(input.AddType)
	parentLinkType := input.Relationship.ParentLinkType
	if parentLinkType == nil {
		primary := string(relationshipenum.ParentLinkTypePrimary)
		parentLinkType = &primary
	}
	switch addType {
	case relationshipenum.AddTypeFather, relationshipenum.AddTypeMother:
		relationship := adminRelationship(familyID, member.ID, base.ID, string(enums.RelationshipTypeParentChild), parentLinkType, adminID)
		relationships := []*relationshipmodel.FamilyRelationship{relationship}
		if member.MemberType == "SPOUSE" {
			oppositeParent, err := findAdminPrimaryLineageParent(tx, familyID, base.ID, oppositeAdminGender(member.Gender))
			if err != nil {
				return nil, errLineageParent
			}
			spouseRelationship := adminRelationship(familyID, oppositeParent.ID, member.ID, string(enums.RelationshipTypeSpouse), nil, adminID)
			relationships = append(relationships, spouseRelationship)
		}
		for _, plan := range relationships {
			if err := r.ensureAdminNoDuplicate(ctx, tx, plan); err != nil {
				return nil, err
			}
		}
		return relationships, nil
	case relationshipenum.AddTypeChild:
		relationship := adminRelationship(familyID, base.ID, member.ID, string(enums.RelationshipTypeParentChild), parentLinkType, adminID)
		return []*relationshipmodel.FamilyRelationship{relationship}, r.ensureAdminNoDuplicate(ctx, tx, relationship)
	case relationshipenum.AddTypeSpouse:
		relationship := adminRelationship(familyID, base.ID, member.ID, string(enums.RelationshipTypeSpouse), nil, adminID)
		return []*relationshipmodel.FamilyRelationship{relationship}, r.ensureAdminNoDuplicate(ctx, tx, relationship)
	case relationshipenum.AddTypeSibling:
		var parents []relationshipmodel.FamilyRelationship
		if err := tx.WithContext(ctx).
			Where("family_id = ? AND relationship_type = ? AND to_member_id = ? AND status = ? AND deleted_at IS NULL", familyID, string(enums.RelationshipTypeParentChild), base.ID, "ACTIVE").
			Order("id").
			Find(&parents).Error; err != nil {
			return nil, err
		}
		relationships := make([]*relationshipmodel.FamilyRelationship, 0, len(parents))
		for i := range parents {
			linkType := parents[i].ParentLinkType
			relationship := adminRelationship(familyID, parents[i].FromMemberID, member.ID, string(enums.RelationshipTypeParentChild), linkType, adminID)
			if err := r.ensureAdminNoDuplicate(ctx, tx, relationship); err != nil {
				return nil, err
			}
			relationships = append(relationships, relationship)
		}
		return relationships, nil
	default:
		return nil, errParentType
	}
}

func adminRelationship(familyID uint64, fromMemberID uint64, toMemberID uint64, relationshipType string, parentLinkType *string, adminID uint64) *relationshipmodel.FamilyRelationship {
	return &relationshipmodel.FamilyRelationship{
		FamilyID: familyID, FromMemberID: fromMemberID, ToMemberID: toMemberID,
		RelationshipType: relationshipType, ParentLinkType: parentLinkType,
		Status: string(enums.StatusActive), CreatedByAdminID: &adminID,
	}
}

func (r *ManagementRepository) ensureAdminNoDuplicate(ctx context.Context, tx *gorm.DB, relationship *relationshipmodel.FamilyRelationship) error {
	if relationship.FromMemberID == relationship.ToMemberID {
		return errDuplicateRelation
	}
	var count int64
	if err := tx.WithContext(ctx).Model(&relationshipmodel.FamilyRelationship{}).
		Where("family_id = ? AND from_member_id = ? AND to_member_id = ? AND relationship_type = ? AND status = ? AND deleted_at IS NULL",
			relationship.FamilyID, relationship.FromMemberID, relationship.ToMemberID, relationship.RelationshipType, "ACTIVE").
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errDuplicateRelation
	}
	return nil
}

func ensureAdminPrimaryParentAvailable(tx *gorm.DB, familyID uint64, childID uint64, gender string, parentLinkType *string) error {
	if parentLinkType == nil || *parentLinkType != string(relationshipenum.ParentLinkTypePrimary) {
		return nil
	}
	var unknownCount int64
	if err := tx.Table("family_relationships AS relation").
		Joins("JOIN family_members AS parent ON parent.id = relation.from_member_id AND parent.family_id = relation.family_id").
		Where("relation.family_id = ? AND relation.to_member_id = ? AND relation.relationship_type = ? AND relation.parent_link_type = ? AND relation.status = ? AND relation.deleted_at IS NULL",
			familyID, childID, string(enums.RelationshipTypeParentChild), string(relationshipenum.ParentLinkTypePrimary), "ACTIVE").
		Where("parent.member_type = ? AND parent.status = ? AND parent.deleted_at IS NULL", "LINEAGE_MEMBER", "ACTIVE").
		Where("parent.gender NOT IN ?", []string{string(enums.GenderMale), string(enums.GenderFemale)}).
		Count(&unknownCount).Error; err != nil {
		return err
	}
	if unknownCount > 0 {
		return errPrimaryParentOpen
	}
	var count int64
	if err := tx.Table("family_relationships AS relation").
		Joins("JOIN family_members AS parent ON parent.id = relation.from_member_id AND parent.family_id = relation.family_id").
		Where("relation.family_id = ? AND relation.to_member_id = ? AND relation.relationship_type = ? AND relation.parent_link_type = ? AND relation.status = ? AND relation.deleted_at IS NULL",
			familyID, childID, string(enums.RelationshipTypeParentChild), string(relationshipenum.ParentLinkTypePrimary), "ACTIVE").
		Where("parent.gender = ? AND parent.status = ? AND parent.deleted_at IS NULL", gender, "ACTIVE").
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	if gender == string(enums.GenderMale) {
		return errPrimaryFather
	}
	return errPrimaryMother
}

func findAdminPrimaryLineageParent(tx *gorm.DB, familyID uint64, childID uint64, gender string) (*membermodel.FamilyMember, error) {
	var member membermodel.FamilyMember
	err := tx.Table("family_members AS parent").
		Select("parent.*").
		Joins("JOIN family_relationships AS relation ON relation.from_member_id = parent.id AND relation.family_id = parent.family_id").
		Where("relation.family_id = ? AND relation.to_member_id = ? AND relation.relationship_type = ? AND relation.parent_link_type = ? AND relation.status = ? AND relation.deleted_at IS NULL",
			familyID, childID, string(enums.RelationshipTypeParentChild), string(relationshipenum.ParentLinkTypePrimary), "ACTIVE").
		Where("parent.gender = ? AND parent.member_type = ? AND parent.status = ? AND parent.deleted_at IS NULL", gender, "LINEAGE_MEMBER", "ACTIVE").
		Order("relation.id").
		Take(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func oppositeAdminGender(gender string) string {
	if gender == string(enums.GenderMale) {
		return string(enums.GenderFemale)
	}
	return string(enums.GenderMale)
}

func memberRowFromModel(member membermodel.FamilyMember) MemberRow {
	return MemberRow{
		ID:                member.ID,
		FamilyID:          member.FamilyID,
		DisplayName:       member.DisplayName,
		Gender:            member.Gender,
		Status:            member.Status,
		IsLiving:          member.IsLiving,
		UserBindingPolicy: member.UserBindingPolicy,
	}
}

func writeAdminMemberLog(ctx context.Context, tx *gorm.DB, audit AdminAudit, action string, familyID uint64, memberID uint64, before any, after any, detail []byte) error {
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)
	targetType := "family_member"
	ip := audit.IP
	userAgent := audit.UserAgent
	role := audit.Role
	record := operationmodel.OperationLog{
		OperatorType:    string(enums.OperatorTypeAdmin),
		OperatorAdminID: &audit.AdminID,
		OperatorRole:    &role,
		Module:          "FAMILY_MEMBER",
		Action:          action,
		TargetType:      &targetType,
		TargetID:        &memberID,
		FamilyID:        &familyID,
		MemberID:        &memberID,
		BeforeJSON:      beforeJSON,
		AfterJSON:       afterJSON,
		DetailJSON:      detail,
		Result:          "SUCCESS",
		IP:              &ip,
		UserAgent:       &userAgent,
	}
	return tx.WithContext(ctx).Create(&record).Error
}

func writeAdminRelationshipLog(ctx context.Context, tx *gorm.DB, audit AdminAudit, action string, familyID uint64, baseMemberID uint64, relationshipIDs []uint64, detail []byte) error {
	targetType := "family_relationship"
	var targetID *uint64
	if len(relationshipIDs) > 0 {
		targetID = &relationshipIDs[0]
	}
	ip := audit.IP
	userAgent := audit.UserAgent
	role := audit.Role
	record := operationmodel.OperationLog{
		OperatorType:    string(enums.OperatorTypeAdmin),
		OperatorAdminID: &audit.AdminID,
		OperatorRole:    &role,
		Module:          "FAMILY_RELATIONSHIP",
		Action:          action,
		TargetType:      &targetType,
		TargetID:        targetID,
		FamilyID:        &familyID,
		MemberID:        &baseMemberID,
		DetailJSON:      detail,
		Result:          "SUCCESS",
		IP:              &ip,
		UserAgent:       &userAgent,
	}
	return tx.WithContext(ctx).Create(&record).Error
}

func stringPtr(value string) *string { return &value }

func (r *ManagementRepository) Admins(ctx context.Context, query PageQuery) ([]AdminRow, int64, error) {
	p, size := page(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Model(&adminmodel.AdminUser{}).Where("deleted_at IS NULL")
	if query.Role != "" {
		db = db.Where("role = ?", query.Role)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var models []adminmodel.AdminUser
	err := db.Select("id, username, display_name, phone, email, role, status, locked_until, last_login_at, created_at").Order("id").Offset((p - 1) * size).Limit(size).Find(&models).Error
	rows := make([]AdminRow, 0, len(models))
	for _, value := range models {
		rows = append(rows, AdminRow{ID: value.ID, Username: value.Username, DisplayName: value.DisplayName, Phone: maskPhone(value.Phone), Email: value.Email, Role: value.Role, Status: value.Status, LockedUntil: value.LockedUntil, LastLoginAt: value.LastLoginAt, CreatedAt: value.CreatedAt})
	}
	return rows, total, err
}

func (r *ManagementRepository) CreateAdmin(ctx context.Context, audit AdminAudit, input CreateAdminInput) (*AdminRow, error) {
	var row adminmodel.AdminUser
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row = adminmodel.AdminUser{Username: input.Username, PasswordHash: input.PasswordHash, DisplayName: input.DisplayName, Phone: input.Phone, Email: input.Email, Role: input.Role, Status: "ACTIVE", Remark: input.Remark, CreatedByAdminID: &audit.AdminID}
		if err := tx.Create(&row).Error; err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
				return errAdminDuplicate
			}
			return err
		}
		return writeAdminLifecycleLog(ctx, tx, audit, "CREATE_ADMIN", "admin_user", row.ID, nil, row.Status, nil)
	})
	if err != nil {
		return nil, err
	}
	return adminRow(row), nil
}

func (r *ManagementRepository) UpdateAdmin(ctx context.Context, audit AdminAudit, id uint64, input UpdateAdminInput) (*AdminRow, error) {
	var row adminmodel.AdminUser
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
			return errAdminNotFound
		}
		if row.Role == string(enums.AdminRoleRootAdmin) {
			return errRootProtected
		}
		before := row.Status
		values := map[string]any{"updated_at": time.Now()}
		if input.DisplayName != nil {
			values["display_name"] = input.DisplayName
		}
		if input.Phone != nil {
			values["phone"] = input.Phone
		}
		if input.Email != nil {
			values["email"] = input.Email
		}
		if input.Role != nil {
			values["role"] = input.Role
		}
		if input.Remark != nil {
			values["remark"] = input.Remark
		}
		if err := tx.Model(&row).Updates(values).Error; err != nil {
			return err
		}
		if err := tx.First(&row, id).Error; err != nil {
			return err
		}
		return writeAdminLifecycleLog(ctx, tx, audit, "UPDATE_ADMIN", "admin_user", id, before, row.Status, nil)
	})
	if err != nil {
		return nil, err
	}
	return adminRow(row), nil
}

func (r *ManagementRepository) SetAdminStatus(ctx context.Context, audit AdminAudit, id uint64, status, reason string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row adminmodel.AdminUser
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
			return errAdminNotFound
		}
		if row.Role == string(enums.AdminRoleRootAdmin) {
			return errRootProtected
		}
		now := time.Now()
		values := map[string]any{"updated_at": now}
		action := "UNLOCK_ADMIN"
		switch status {
		case "ACTIVE":
			values["status"] = "ACTIVE"
			values["disabled_at"] = nil
			values["disabled_by_admin_id"] = nil
			action = "ENABLE_ADMIN"
		case "DISABLED":
			values["status"] = "DISABLED"
			values["disabled_at"] = now
			values["disabled_by_admin_id"] = audit.AdminID
			values["disabled_reason"] = reason
			action = "DISABLE_ADMIN"
		case "DELETED":
			values["status"] = "DELETED"
			values["deleted_at"] = now
			values["deleted_by_admin_id"] = audit.AdminID
			values["disabled_reason"] = reason
			action = "DELETE_ADMIN"
		case "UNLOCKED":
			values["failed_login_count"] = 0
			values["locked_until"] = nil
			values["unlocked_at"] = now
			values["unlocked_by_admin_id"] = audit.AdminID
		}
		if err := tx.Model(&row).Updates(values).Error; err != nil {
			return err
		}
		return writeAdminLifecycleLog(ctx, tx, audit, action, "admin_user", id, row.Status, status, []byte(reason))
	})
}

func adminRow(value adminmodel.AdminUser) *AdminRow {
	return &AdminRow{ID: value.ID, Username: value.Username, DisplayName: value.DisplayName, Phone: maskPhone(value.Phone), Email: value.Email, Role: value.Role, Status: value.Status, LockedUntil: value.LockedUntil, LastLoginAt: value.LastLoginAt, CreatedAt: value.CreatedAt}
}

func (r *ManagementRepository) UpdateFamily(ctx context.Context, audit AdminAudit, id uint64, input UpdateFamilyInput) (*FamilyRow, error) {
	var family familymodel.Family
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&family).Error; err != nil {
			return errAdminNotFound
		}
		if family.Status == "DISSOLVED" || family.Status == "DELETED" {
			return errFamilyStatus
		}
		values := map[string]any{"updated_at": time.Now()}
		if input.FamilyName != nil {
			values["family_name"] = input.FamilyName
		}
		if input.Description != nil {
			values["description"] = input.Description
		}
		if input.NativePlace != nil {
			values["native_place"] = input.NativePlace
		}
		if input.RegionText != nil {
			values["region_text"] = input.RegionText
		}
		if input.Searchable != nil {
			values["searchable"] = input.Searchable
		}
		if err := tx.Model(&family).Updates(values).Error; err != nil {
			return err
		}
		return writeAdminLifecycleLog(ctx, tx, audit, "UPDATE_FAMILY", "family", id, family.Status, family.Status, nil)
	})
	if err != nil {
		return nil, err
	}
	return r.Family(ctx, id)
}

func (r *ManagementRepository) SetFamilyStatus(ctx context.Context, audit AdminAudit, id uint64, status, reason string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var family familymodel.Family
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&family).Error; err != nil {
			return errAdminNotFound
		}
		// Dissolution states have their own audited recovery workflow.  Do not
		// let the generic enable/disable endpoint bypass the cooldown, quota and
		// public-display cleanup handled by RestoreDissolution.
		if family.Status == "DISSOLUTION_PENDING" || family.Status == "DISSOLUTION_COOLDOWN" || family.Status == "DISSOLVED" {
			return errFamilyStatus
		}
		now := time.Now()
		values := map[string]any{"status": status, "updated_at": now}
		action := "ENABLE_FAMILY"
		if status == "DISABLED" {
			values["disabled_at"] = now
			values["disabled_by_admin_id"] = audit.AdminID
			values["disabled_reason"] = reason
			values["public_display_status"] = "PRIVATE"
			values["public_display_enabled"] = false
			action = "DISABLE_FAMILY"
		}
		if status == "DELETED" {
			values["deleted_at"] = now
			values["deleted_at"] = now
			values["public_display_status"] = "PRIVATE"
			values["public_display_enabled"] = false
			action = "DELETE_FAMILY"
		}
		if err := tx.Model(&family).Updates(values).Error; err != nil {
			return err
		}
		return writeAdminLifecycleLog(ctx, tx, audit, action, "family", id, family.Status, status, []byte(reason))
	})
}

func writeAdminLifecycleLog(ctx context.Context, tx *gorm.DB, audit AdminAudit, action, targetType string, targetID uint64, before, after any, detail []byte) error {
	role, ip, ua := audit.Role, audit.IP, audit.UserAgent
	return tx.WithContext(ctx).Create(&operationmodel.OperationLog{OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &audit.AdminID, OperatorRole: &role, Module: "ADMIN_MANAGEMENT", Action: action, TargetType: &targetType, TargetID: &targetID, BeforeJSON: mustJSON(before), AfterJSON: mustJSON(after), DetailJSON: detail, Result: "SUCCESS", IP: &ip, UserAgent: &ua}).Error
}

func mustJSON(value any) []byte {
	if value == nil {
		return nil
	}
	data, _ := json.Marshal(value)
	return data
}

func (r *ManagementRepository) Logs(ctx context.Context, query PageQuery) ([]LogRow, int64, error) {
	p, size := page(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Model(&operationmodel.OperationLog{})
	if query.Result != "" {
		db = db.Where("result = ?", query.Result)
	}
	if query.FamilyID > 0 {
		db = db.Where("family_id = ?", query.FamilyID)
	}
	if operatorType := strings.TrimSpace(query.OperatorType); operatorType != "" {
		db = db.Where("operator_type = ?", operatorType)
	}
	if role := strings.TrimSpace(query.Role); role != "" {
		db = db.Where("operator_role = ?", role)
	}
	if !query.StartAt.IsZero() {
		db = db.Where("created_at >= ?", query.StartAt)
	}
	if !query.EndAt.IsZero() {
		db = db.Where("created_at < ?", query.EndAt.AddDate(0, 0, 1))
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		db = db.Where("module LIKE ? OR action LIKE ? OR COALESCE(target_type, '') LIKE ?", pattern, pattern, pattern)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var models []operationmodel.OperationLog
	if err := db.Order("id DESC").Offset((p - 1) * size).Limit(size).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]LogRow, 0, len(models))
	for _, value := range models {
		rows = append(rows, LogRow{ID: value.ID, OperatorType: value.OperatorType, OperatorAdminID: value.OperatorAdminID, OperatorUserID: value.OperatorUserID, OperatorRole: value.OperatorRole, Module: value.Module, Action: value.Action, TargetType: value.TargetType, TargetID: value.TargetID, FamilyID: value.FamilyID, MemberID: value.MemberID, UserID: value.UserID, DetailJSON: json.RawMessage(value.DetailJSON), Result: value.Result, ErrorMessage: value.ErrorMessage, IP: value.IP, CreatedAt: value.CreatedAt})
	}
	return rows, total, nil
}
