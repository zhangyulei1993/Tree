package repository

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"

	adminmodel "tree/backend/internal/admin/model"
	operationmodel "tree/backend/internal/operationlog/model"
	usermodel "tree/backend/internal/user/model"
)

type PageQuery struct {
	Keyword  string
	Status   string
	Role     string
	FamilyID uint64
	Result   string
	Page     int
	PageSize int
}

type Dashboard struct {
	Users                     int64 `json:"users"`
	Families                  int64 `json:"families"`
	PendingPublicApplications int64 `json:"pendingPublicApplications"`
	PendingVisitorMessages    int64 `json:"pendingVisitorMessages"`
	PendingFounderTransfers   int64 `json:"pendingFounderTransfers"`
	PendingDissolutions       int64 `json:"pendingDissolutions"`
}

type UserRow struct {
	ID                  uint64     `json:"id"`
	Phone               *string    `json:"phone,omitempty"`
	Nickname            *string    `json:"nickname,omitempty"`
	RealName            *string    `json:"realName,omitempty"`
	AccountOrigin       string     `json:"accountOrigin"`
	RegisterClient      string     `json:"registerClient"`
	PhoneVerified       bool       `json:"phoneVerified"`
	PhoneLoginEnabled   bool       `json:"phoneLoginEnabled"`
	HasWechatLogin      bool       `json:"hasWechatLogin"`
	CanUnbindPhoneLogin bool       `json:"canUnbindPhoneLogin"`
	TrustTier           string     `json:"trustTier"`
	LoginMethod         string     `json:"loginMethod"`
	Status              string     `json:"status"`
	LastLoginAt         *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
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
		{&result.PendingVisitorMessages, "visitor_messages", "status = 'PENDING' AND deleted_at IS NULL"},
		{&result.PendingFounderTransfers, "family_founder_transfer_requests", "request_status = 'PENDING'"},
		{&result.PendingDissolutions, "family_dissolution_requests", "request_status = 'PENDING'"},
	}
	for _, query := range queries {
		if err := r.db.WithContext(ctx).Table(query.table).Where(query.where).Count(query.target).Error; err != nil {
			return nil, err
		}
	}
	return result, nil
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

func resolveLoginMethod(hasWechatLogin, phoneLoginEnabled bool) string {
	switch {
	case hasWechatLogin && phoneLoginEnabled:
		return "微信 + 手机号"
	case hasWechatLogin:
		return "仅微信"
	case phoneLoginEnabled:
		return "仅手机号"
	default:
		return "未设置"
	}
}

func userRow(value usermodel.User, hasWechatLogin bool) UserRow {
	trustTier := "WECHAT_ONLY"
	if value.PhoneLoginEnabled {
		trustTier = "PHONE_BOUND"
	}
	return UserRow{
		ID: value.ID, Phone: maskPhone(value.Phone), Nickname: value.Nickname, RealName: value.RealName,
		AccountOrigin: value.AccountOrigin, RegisterClient: value.RegisterClient, PhoneVerified: value.PhoneVerified,
		PhoneLoginEnabled: value.PhoneLoginEnabled, HasWechatLogin: hasWechatLogin,
		CanUnbindPhoneLogin: value.PhoneLoginEnabled && hasWechatLogin,
		TrustTier:           trustTier, LoginMethod: resolveLoginMethod(hasWechatLogin, value.PhoneLoginEnabled),
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

func (r *ManagementRepository) Logs(ctx context.Context, query PageQuery) ([]LogRow, int64, error) {
	p, size := page(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Model(&operationmodel.OperationLog{})
	if query.Result != "" {
		db = db.Where("result = ?", query.Result)
	}
	if query.FamilyID > 0 {
		db = db.Where("family_id = ?", query.FamilyID)
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
