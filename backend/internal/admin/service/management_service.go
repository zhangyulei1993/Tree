package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	quotaservice "tree/backend/internal/accountquota/service"
	"tree/backend/internal/admin/repository"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/security"
	relationshipenum "tree/backend/internal/family/relationship/enum"
)

type PageResult struct {
	Items    any   `json:"items"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}
type ManagementService struct {
	repo  *repository.ManagementRepository
	quota quotaservice.Service
}

type AdminAuditInput struct {
	AdminID   uint64
	Role      string
	IP        string
	UserAgent string
}

type CreateAdminRequest struct {
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	DisplayName *string `json:"displayName"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
	Role        string  `json:"role"`
	Remark      *string `json:"remark"`
}

type UpdateAdminRequest struct {
	DisplayName *string `json:"displayName"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
	Role        *string `json:"role"`
	Remark      *string `json:"remark"`
}

type UpdateFamilyRequest struct {
	FamilyName  *string `json:"familyName"`
	Description *string `json:"description"`
	NativePlace *string `json:"nativePlace"`
	RegionText  *string `json:"regionText"`
	Searchable  *bool   `json:"searchable"`
}

type UpdateFamilyMemberRequest struct {
	DisplayName       *string `json:"displayName"`
	Gender            *string `json:"gender"`
	BirthDate         *string `json:"birthDate"`
	IsLiving          *bool   `json:"isLiving"`
	UserBindingPolicy *string `json:"userBindingPolicy"`
}

type DeleteFamilyMemberRequest struct {
	Reason *string `json:"reason"`
}

type CreateFamilyRelativeRequest struct {
	AddType      string                       `json:"addType"`
	NewMember    CreateFamilyRelativeMember   `json:"newMember"`
	Relationship CreateFamilyRelativeRelation `json:"relationship"`
}

type CreateFamilyRelativeMember struct {
	DisplayName       string  `json:"displayName"`
	Name              string  `json:"name"`
	Gender            string  `json:"gender"`
	BirthDate         *string `json:"birthDate"`
	IsLiving          *bool   `json:"isLiving"`
	UserBindingPolicy string  `json:"userBindingPolicy"`
	MemberType        string  `json:"memberType"`
}

type CreateFamilyRelativeRelation struct {
	ParentLinkType *string `json:"parentLinkType"`
}

func NewManagementService(repo *repository.ManagementRepository, quotas ...quotaservice.Service) *ManagementService {
	var quota quotaservice.Service
	if len(quotas) > 0 {
		quota = quotas[0]
	}
	return &ManagementService{repo: repo, quota: quota}
}

func canManageAdminAccounts(role string) bool {
	return role == string(enums.AdminRoleRootAdmin) || role == string(enums.AdminRoleSuperAdmin)
}

func canManageFamilyLifecycle(role string) bool {
	return canManageAdminAccounts(role)
}

func (s *ManagementService) CreateAdmin(ctx context.Context, actor AdminAuditInput, req CreateAdminRequest) (any, *apperrors.BusinessError) {
	if !canManageAdminAccounts(actor.Role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	req.Username = strings.TrimSpace(req.Username)
	if len(req.Username) < 3 || len(req.Password) < 8 || !canManageAdminAccounts(req.Role) && req.Role != string(enums.AdminRolePlatformAdmin) {
		return nil, adminError(apperrors.CodeInvalidParams, "管理员账号、密码或角色不合法")
	}
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	row, err := s.repo.CreateAdmin(ctx, repository.AdminAudit(actor), repository.CreateAdminInput{Username: req.Username, PasswordHash: hash, DisplayName: req.DisplayName, Phone: req.Phone, Email: req.Email, Role: req.Role, Remark: req.Remark})
	if errors.Is(err, repository.ErrAdminDuplicate()) {
		return nil, adminError(apperrors.CodeDataConflict, "管理员用户名已存在")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return row, nil
}

func (s *ManagementService) UpdateAdmin(ctx context.Context, actor AdminAuditInput, adminID uint64, req UpdateAdminRequest) (any, *apperrors.BusinessError) {
	if !canManageAdminAccounts(actor.Role) || adminID == actor.AdminID {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	if req.Role != nil && !canManageAdminAccounts(*req.Role) && *req.Role != string(enums.AdminRolePlatformAdmin) {
		return nil, adminError(apperrors.CodeInvalidParams, "管理员角色不合法")
	}
	row, err := s.repo.UpdateAdmin(ctx, repository.AdminAudit(actor), adminID, repository.UpdateAdminInput{DisplayName: req.DisplayName, Phone: req.Phone, Email: req.Email, Role: req.Role, Remark: req.Remark})
	if errors.Is(err, repository.ErrAdminNotFound()) {
		return nil, apperrors.New(apperrors.CodeResourceNotFound)
	}
	if errors.Is(err, repository.ErrRootProtected()) {
		return nil, adminError(apperrors.CodeForbidden, "根管理员不能被修改")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return row, nil
}

func (s *ManagementService) SetAdminStatus(ctx context.Context, actor AdminAuditInput, adminID uint64, status string, reason string) *apperrors.BusinessError {
	if !canManageAdminAccounts(actor.Role) || adminID == actor.AdminID {
		return apperrors.New(apperrors.CodeForbidden)
	}
	if status != "ACTIVE" && status != "DISABLED" && status != "DELETED" && status != "UNLOCKED" {
		return apperrors.New(apperrors.CodeInvalidParams)
	}
	err := s.repo.SetAdminStatus(ctx, repository.AdminAudit(actor), adminID, status, strings.TrimSpace(reason))
	switch {
	case errors.Is(err, repository.ErrAdminNotFound()):
		return apperrors.New(apperrors.CodeResourceNotFound)
	case errors.Is(err, repository.ErrRootProtected()):
		return adminError(apperrors.CodeForbidden, "根管理员不能被禁用或删除")
	case err != nil:
		return apperrors.New(apperrors.CodeSystemError)
	default:
		return nil
	}
}

func (s *ManagementService) UpdateFamily(ctx context.Context, actor AdminAuditInput, familyID uint64, req UpdateFamilyRequest) (any, *apperrors.BusinessError) {
	if !canManageFamilyLifecycle(actor.Role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	if req.FamilyName != nil {
		value := strings.TrimSpace(*req.FamilyName)
		if value == "" || len(value) > 150 {
			return nil, adminError(apperrors.CodeInvalidParams, "家庭名称不合法")
		}
		req.FamilyName = &value
	}
	row, err := s.repo.UpdateFamily(ctx, repository.AdminAudit(actor), familyID, repository.UpdateFamilyInput{FamilyName: req.FamilyName, Description: req.Description, NativePlace: req.NativePlace, RegionText: req.RegionText, Searchable: req.Searchable})
	if errors.Is(err, repository.ErrAdminNotFound()) {
		return nil, apperrors.New(apperrors.CodeResourceNotFound)
	}
	if errors.Is(err, repository.ErrFamilyStatus()) {
		return nil, adminError(apperrors.CodeInvalidDataStatus, "当前家庭状态不允许编辑")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return row, nil
}

func (s *ManagementService) SetFamilyStatus(ctx context.Context, actor AdminAuditInput, familyID uint64, status string, reason string) *apperrors.BusinessError {
	if !canManageFamilyLifecycle(actor.Role) {
		return apperrors.New(apperrors.CodeForbidden)
	}
	if status != "NORMAL" && status != "DISABLED" && status != "DELETED" {
		return apperrors.New(apperrors.CodeInvalidParams)
	}
	err := s.repo.SetFamilyStatus(ctx, repository.AdminAudit(actor), familyID, status, strings.TrimSpace(reason))
	switch {
	case errors.Is(err, repository.ErrAdminNotFound()):
		return apperrors.New(apperrors.CodeResourceNotFound)
	case errors.Is(err, repository.ErrFamilyStatus()):
		return adminError(apperrors.CodeInvalidDataStatus, "解散中的家庭必须通过恢复中心处理")
	case err != nil:
		return apperrors.New(apperrors.CodeSystemError)
	default:
		return nil
	}
}

func (s *ManagementService) SetUserStatus(ctx context.Context, actor AdminAuditInput, userID uint64, status, reason string) *apperrors.BusinessError {
	if !canManageAdminAccounts(actor.Role) {
		return apperrors.New(apperrors.CodeForbidden)
	}
	if status != "ACTIVE" && status != "DISABLED" && status != "DELETED" && status != "CANCELLED" {
		return apperrors.New(apperrors.CodeInvalidParams)
	}
	err := s.repo.SetUserStatus(ctx, repository.AdminAudit(actor), userID, status, strings.TrimSpace(reason))
	if errors.Is(err, repository.ErrAdminNotFound()) {
		return apperrors.New(apperrors.CodeResourceNotFound)
	}
	if err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	return nil
}
func normalize(query repository.PageQuery) repository.PageQuery {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	return query
}
func (s *ManagementService) Dashboard(ctx context.Context) (any, *apperrors.BusinessError) {
	value, err := s.repo.Dashboard(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return value, nil
}
func (s *ManagementService) Users(ctx context.Context, q repository.PageQuery) (any, *apperrors.BusinessError) {
	q = normalize(q)
	rows, total, err := s.repo.Users(ctx, q)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return PageResult{rows, q.Page, q.PageSize, total}, nil
}
func (s *ManagementService) User(ctx context.Context, id uint64) (any, *apperrors.BusinessError) {
	value, err := s.repo.UserDetail(ctx, id)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeResourceNotFound)
	}
	return value, nil
}
func (s *ManagementService) Families(ctx context.Context, q repository.PageQuery) (any, *apperrors.BusinessError) {
	q = normalize(q)
	rows, total, err := s.repo.Families(ctx, q)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return PageResult{rows, q.Page, q.PageSize, total}, nil
}
func (s *ManagementService) Family(ctx context.Context, id uint64) (any, *apperrors.BusinessError) {
	value, err := s.repo.Family(ctx, id)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeResourceNotFound)
	}
	return value, nil
}
func (s *ManagementService) Members(ctx context.Context, id uint64) (any, *apperrors.BusinessError) {
	value, err := s.repo.Members(ctx, id)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return value, nil
}
func (s *ManagementService) Tree(ctx context.Context, id uint64) (any, *apperrors.BusinessError) {
	value, err := s.repo.Tree(ctx, id)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeResourceNotFound)
	}
	return value, nil
}

func (s *ManagementService) UpdateFamilyMember(ctx context.Context, familyID uint64, memberID uint64, req UpdateFamilyMemberRequest, audit AdminAuditInput) (any, *apperrors.BusinessError) {
	input, businessErr := normalizeUpdateFamilyMemberRequest(req)
	if businessErr != nil {
		return nil, businessErr
	}
	value, err := s.repo.UpdateFamilyMember(ctx, familyID, memberID, input, repository.AdminAudit(audit))
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, apperrors.New(apperrors.CodeResourceNotFound)
	case errors.Is(err, repository.ErrBindingConflict()):
		return nil, adminError(apperrors.CodeDataConflict, "当前成员已有账号绑定，不能标记为无需绑定")
	case errors.Is(err, repository.ErrFamilyStatus()):
		return nil, adminError(apperrors.CodeInvalidDataStatus, "家庭状态不允许编辑成员")
	case err != nil:
		return nil, apperrors.New(apperrors.CodeSystemError)
	default:
		return value, nil
	}
}

func (s *ManagementService) DeleteFamilyMember(ctx context.Context, familyID uint64, memberID uint64, req DeleteFamilyMemberRequest, audit AdminAuditInput) *apperrors.BusinessError {
	reason := cleanOptionalString(req.Reason)
	err := s.repo.DeleteFamilyMember(ctx, familyID, memberID, repository.DeleteFamilyMemberInput{Reason: reason}, repository.AdminAudit(audit))
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return apperrors.New(apperrors.CodeResourceNotFound)
	case errors.Is(err, repository.ErrFamilyStatus()):
		return adminError(apperrors.CodeInvalidDataStatus, "家庭状态不允许删除成员")
	case errors.Is(err, repository.ErrFounderProtected()):
		return adminError(apperrors.CodeDataConflict, "家庭创始人不能直接删除")
	case errors.Is(err, repository.ErrMemberHasChildren()):
		return adminError(apperrors.CodeDataConflict, "该成员已有子女关系，不能直接删除")
	case errors.Is(err, repository.ErrPrivilegedMember()):
		return adminError(apperrors.CodeDataConflict, "家庭创始人或管理员成员不能直接删除")
	case err != nil:
		return apperrors.New(apperrors.CodeSystemError)
	default:
		return nil
	}
}

func (s *ManagementService) CreateFamilyRelative(ctx context.Context, familyID uint64, memberID uint64, req CreateFamilyRelativeRequest, audit AdminAuditInput) (any, *apperrors.BusinessError) {
	input, businessErr := normalizeCreateFamilyRelativeRequest(req)
	if businessErr != nil {
		return nil, businessErr
	}
	value, err := s.repo.CreateFamilyRelative(ctx, familyID, memberID, input, repository.AdminAudit(audit))
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, apperrors.New(apperrors.CodeResourceNotFound)
	case errors.Is(err, repository.ErrFamilyStatus()):
		return nil, adminError(apperrors.CodeInvalidDataStatus, "家庭状态不允许添加亲属")
	case errors.Is(err, repository.ErrSiblingNoParents()):
		return nil, adminError(apperrors.CodeDataConflict, "该成员还没有父亲或母亲节点，请先创建父亲或母亲节点")
	case errors.Is(err, repository.ErrDuplicateRelation()):
		return nil, adminError(apperrors.CodeDataConflict, "当前关系已存在")
	case errors.Is(err, repository.ErrPrimaryFather()):
		return nil, adminError(apperrors.CodeDataConflict, "该成员已有主要父亲节点")
	case errors.Is(err, repository.ErrPrimaryMother()):
		return nil, adminError(apperrors.CodeDataConflict, "该成员已有主要母亲节点")
	case errors.Is(err, repository.ErrPrimaryParentOpen()):
		return nil, adminError(apperrors.CodeDataConflict, "该成员已有性别未确定的父母节点，请先完善父母性别后再添加父亲或母亲")
	case errors.Is(err, repository.ErrBaseGender()):
		return nil, adminError(apperrors.CodeInvalidParams, "请先完善基准成员性别")
	case errors.Is(err, repository.ErrChildGender()):
		return nil, adminError(apperrors.CodeInvalidParams, "子女或同辈性别必须选择男或女")
	case errors.Is(err, repository.ErrSiblingGender()):
		return nil, adminError(apperrors.CodeInvalidParams, "兄弟姐妹性别必须选择男或女")
	case errors.Is(err, repository.ErrSpouseGender()):
		return nil, adminError(apperrors.CodeInvalidParams, "配偶性别必须与基准成员相对")
	case errors.Is(err, repository.ErrSpouseBase()):
		return nil, adminError(apperrors.CodeDataConflict, "配偶节点不能作为家谱主干扩展，请从本家成员节点操作")
	case errors.Is(err, repository.ErrParentType()):
		return nil, adminError(apperrors.CodeInvalidParams, "亲属类型或父母成员身份不正确")
	case errors.Is(err, repository.ErrLineageParent()):
		return nil, adminError(apperrors.CodeDataConflict, "请先录入本家成员作为父母，再添加配偶父母")
	case err != nil:
		return nil, apperrors.New(apperrors.CodeSystemError)
	default:
		return value, nil
	}
}
func (s *ManagementService) Admins(ctx context.Context, q repository.PageQuery) (any, *apperrors.BusinessError) {
	q = normalize(q)
	rows, total, err := s.repo.Admins(ctx, q)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return PageResult{rows, q.Page, q.PageSize, total}, nil
}
func (s *ManagementService) Logs(ctx context.Context, q repository.PageQuery) (any, *apperrors.BusinessError) {
	q = normalize(q)
	rows, total, err := s.repo.Logs(ctx, q)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return PageResult{rows, q.Page, q.PageSize, total}, nil
}

func (s *ManagementService) RecoveryItems(ctx context.Context, q repository.PageQuery) (any, *apperrors.BusinessError) {
	q = normalize(q)
	rows, total, err := s.repo.RecoveryItems(ctx, q)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return PageResult{rows, q.Page, q.PageSize, total}, nil
}

func (s *ManagementService) RestoreMember(ctx context.Context, familyID, memberID uint64, audit AdminAuditInput) *apperrors.BusinessError {
	var beforeRestore func(context.Context, *gorm.DB, uint64) error
	if s.quota != nil {
		beforeRestore = func(ctx context.Context, tx *gorm.DB, familyID uint64) error {
			return s.quota.AssertCanAddMember(ctx, tx, familyID, 1)
		}
	}
	err := s.repo.RestoreMember(ctx, familyID, memberID, repository.AdminAudit(audit), beforeRestore)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return apperrors.New(apperrors.CodeResourceNotFound)
	case errors.Is(err, repository.ErrFamilyStatus()):
		return adminError(apperrors.CodeInvalidDataStatus, "家庭状态不允许恢复成员")
	case errors.Is(err, repository.ErrRestoreMemberConflict()):
		return adminError(apperrors.CodeDataConflict, "该成员原有关系位置已被新节点占用，不能恢复")
	case err != nil:
		if s.quota != nil {
			if businessErr := s.quota.MapQuotaError(err); businessErr != nil && businessErr.Code != apperrors.CodeSystemError {
				return businessErr
			}
		}
		return apperrors.New(apperrors.CodeSystemError)
	default:
		return nil
	}
}

func (s *ManagementService) RestoreRelationship(ctx context.Context, familyID, relationshipID uint64, audit AdminAuditInput) *apperrors.BusinessError {
	err := s.repo.RestoreRelationship(ctx, familyID, relationshipID, repository.AdminAudit(audit))
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return apperrors.New(apperrors.CodeResourceNotFound)
	case errors.Is(err, repository.ErrFamilyStatus()):
		return adminError(apperrors.CodeInvalidDataStatus, "家庭状态不允许恢复关系")
	case errors.Is(err, repository.ErrRestoreEndpoints()):
		return adminError(apperrors.CodeDataConflict, "关系两端成员尚未恢复，不能单独恢复关系")
	case errors.Is(err, repository.ErrDuplicateRelation()):
		return adminError(apperrors.CodeDataConflict, "当前家庭已存在相同有效关系")
	case err != nil:
		return apperrors.New(apperrors.CodeSystemError)
	default:
		return nil
	}
}

func normalizeUpdateFamilyMemberRequest(req UpdateFamilyMemberRequest) (repository.UpdateFamilyMemberInput, *apperrors.BusinessError) {
	var input repository.UpdateFamilyMemberInput
	if req.DisplayName != nil {
		name := strings.TrimSpace(*req.DisplayName)
		if name == "" || len([]rune(name)) > 80 {
			return input, adminError(apperrors.CodeInvalidParams, "成员姓名不能为空，且不能超过 80 个字")
		}
		input.DisplayName = &name
	}
	if req.Gender != nil {
		gender := strings.ToUpper(strings.TrimSpace(*req.Gender))
		if gender != string(enums.GenderMale) && gender != string(enums.GenderFemale) && gender != string(enums.GenderUnknown) {
			return input, adminError(apperrors.CodeInvalidParams, "性别参数不正确")
		}
		input.Gender = &gender
	}
	if req.BirthDate != nil {
		value := strings.TrimSpace(*req.BirthDate)
		var parsed *time.Time
		if value != "" {
			t, err := time.Parse("2006-01-02", value)
			if err != nil {
				return input, adminError(apperrors.CodeInvalidParams, "出生日期格式应为 YYYY-MM-DD")
			}
			parsed = &t
		}
		input.BirthDate = &parsed
	}
	if req.IsLiving != nil {
		value := *req.IsLiving
		parsed := &value
		input.IsLiving = &parsed
	}
	if req.UserBindingPolicy != nil {
		policy := strings.ToUpper(strings.TrimSpace(*req.UserBindingPolicy))
		switch policy {
		case string(enums.UserBindingOptional), string(enums.UserBindingRequired), string(enums.UserBindingNotRequired):
			input.UserBindingPolicy = &policy
		default:
			return input, adminError(apperrors.CodeInvalidParams, "绑定策略参数不正确")
		}
	}
	return input, nil
}

func normalizeCreateFamilyRelativeRequest(req CreateFamilyRelativeRequest) (repository.CreateFamilyRelativeInput, *apperrors.BusinessError) {
	var input repository.CreateFamilyRelativeInput
	addType := strings.ToUpper(strings.TrimSpace(req.AddType))
	switch relationshipenum.AddType(addType) {
	case relationshipenum.AddTypeFather, relationshipenum.AddTypeMother, relationshipenum.AddTypeChild, relationshipenum.AddTypeSpouse, relationshipenum.AddTypeSibling:
		input.AddType = addType
	default:
		return input, adminError(apperrors.CodeInvalidParams, "不支持的亲属类型")
	}
	name := strings.TrimSpace(req.NewMember.DisplayName)
	if name == "" {
		name = strings.TrimSpace(req.NewMember.Name)
	}
	if name == "" || len([]rune(name)) > 80 {
		return input, adminError(apperrors.CodeInvalidParams, "成员姓名不能为空，且不能超过 80 个字")
	}
	gender := strings.ToUpper(strings.TrimSpace(req.NewMember.Gender))
	if gender == "" {
		gender = string(enums.GenderUnknown)
	}
	if addType == string(relationshipenum.AddTypeFather) {
		gender = string(enums.GenderMale)
	}
	if addType == string(relationshipenum.AddTypeMother) {
		gender = string(enums.GenderFemale)
	}
	if gender != string(enums.GenderMale) && gender != string(enums.GenderFemale) && gender != string(enums.GenderUnknown) {
		return input, adminError(apperrors.CodeInvalidParams, "性别参数不正确")
	}
	var birthDate *time.Time
	if req.NewMember.BirthDate != nil {
		value := strings.TrimSpace(*req.NewMember.BirthDate)
		if value != "" {
			t, err := time.Parse("2006-01-02", value)
			if err != nil {
				return input, adminError(apperrors.CodeInvalidParams, "出生日期格式应为 YYYY-MM-DD")
			}
			birthDate = &t
		}
	}
	policy := strings.ToUpper(strings.TrimSpace(req.NewMember.UserBindingPolicy))
	if policy == "" {
		policy = string(enums.UserBindingOptional)
	}
	switch policy {
	case string(enums.UserBindingOptional), string(enums.UserBindingRequired), string(enums.UserBindingNotRequired):
	default:
		return input, adminError(apperrors.CodeInvalidParams, "绑定策略参数不正确")
	}
	memberType := strings.ToUpper(strings.TrimSpace(req.NewMember.MemberType))
	if memberType == "" {
		memberType = "LINEAGE_MEMBER"
	}
	if addType == string(relationshipenum.AddTypeSpouse) {
		memberType = "SPOUSE"
	}
	if memberType != "LINEAGE_MEMBER" && memberType != "SPOUSE" {
		return input, adminError(apperrors.CodeInvalidParams, "成员身份参数不正确")
	}
	input.NewMember = repository.NewFamilyRelativeMemberInput{
		DisplayName: name, Gender: gender, BirthDate: birthDate, IsLiving: req.NewMember.IsLiving,
		UserBindingPolicy: policy, MemberType: memberType,
	}
	if addType != string(relationshipenum.AddTypeSpouse) {
		parentLinkType := string(relationshipenum.ParentLinkTypePrimary)
		if req.Relationship.ParentLinkType != nil {
			parentLinkType = strings.ToUpper(strings.TrimSpace(*req.Relationship.ParentLinkType))
		}
		switch relationshipenum.ParentLinkType(parentLinkType) {
		case relationshipenum.ParentLinkTypePrimary, relationshipenum.ParentLinkTypeStep, relationshipenum.ParentLinkTypeAdoptive, relationshipenum.ParentLinkTypeSuccession, relationshipenum.ParentLinkTypeNoteOnly, relationshipenum.ParentLinkTypeOther:
		default:
			return input, adminError(apperrors.CodeInvalidParams, "亲属关系属性不正确")
		}
		input.Relationship.ParentLinkType = &parentLinkType
	}
	return input, nil
}

func cleanOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	cleaned := strings.TrimSpace(*value)
	if cleaned == "" {
		return nil
	}
	if len([]rune(cleaned)) > 500 {
		cleaned = string([]rune(cleaned)[:500])
	}
	return &cleaned
}

func adminError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}
