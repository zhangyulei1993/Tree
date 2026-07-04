package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	quotaservice "tree/backend/internal/accountquota/service"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/permission"
	"tree/backend/internal/family/core/dto"
	familymodel "tree/backend/internal/family/core/model"
	familyrepo "tree/backend/internal/family/core/repository"
	"tree/backend/internal/family/core/vo"
	dissolutionmodel "tree/backend/internal/family/dissolution/model"
	membermodel "tree/backend/internal/family/member/model"
	rolemodel "tree/backend/internal/family/role/model"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	familyStatusNormal             = "NORMAL"
	familyStatusDissolutionPending = "DISSOLUTION_PENDING"
	familyPublicPrivate            = "PRIVATE"
	dissolutionStatusPending       = "PENDING"
	dissolutionStatusCancelled     = "CANCELLED"

	CodeFamilySurnameRequired         apperrors.Code = 42004
	CodeFamilyCreateStatusDenied      apperrors.Code = 42006
	CodeFamilyNotFound                apperrors.Code = 42101
	CodeFamilyNotPublic               apperrors.Code = 42103
	CodeFamilyDetailForbidden         apperrors.Code = 42104
	CodeFamilyUpdateForbidden         apperrors.Code = 42201
	CodeFamilyUpdateStatusDenied      apperrors.Code = 42202
	CodeFamilyLeaveForbidden          apperrors.Code = 42203
	CodeFamilyLeaveNotLinked          apperrors.Code = 42204
	CodeFamilyDissolutionForbidden    apperrors.Code = 46301
	CodeFamilyDissolutionStatusDenied apperrors.Code = 46302
	CodeFamilyDissolutionPending      apperrors.Code = 46303
)

var (
	errDissolutionFamilyStatus   = errors.New("family status disallows dissolution")
	errDissolutionPending        = errors.New("pending dissolution request exists")
	errFamilyLeaveForbidden      = errors.New("family status disallows leave")
	errFamilyFounderMustTransfer = errors.New("founder must transfer before leave")
)

type AuditInput struct {
	IP        string
	UserAgent string
}

type FamilyService interface {
	Create(context.Context, uint64, dto.CreateFamilyRequest, AuditInput) (*vo.FamilyDetail, *apperrors.BusinessError)
	List(context.Context, uint64) ([]vo.FamilySummary, *apperrors.BusinessError)
	Detail(context.Context, uint64, uint64) (*vo.FamilyDetail, *apperrors.BusinessError)
	PublicDetail(context.Context, uint64) (*vo.PublicFamily, *apperrors.BusinessError)
	ListPublicFamilies(context.Context, dto.ListPublicFamiliesQuery) (*vo.ListPublicFamiliesResult, *apperrors.BusinessError)
	ListPublicFamilyShowcase(context.Context, dto.ListPublicFamilyShowcaseQuery) (*vo.ListPublicFamilyShowcaseResult, *apperrors.BusinessError)
	Update(context.Context, uint64, uint64, dto.UpdateFamilyRequest, AuditInput) (*vo.FamilyDetail, *apperrors.BusinessError)
	CreateDissolutionRequest(context.Context, uint64, uint64, dto.CreateDissolutionRequest, AuditInput) (*vo.DissolutionRequest, *apperrors.BusinessError)
	CurrentDissolutionRequest(context.Context, uint64, uint64) (*vo.DissolutionRequest, *apperrors.BusinessError)
	CancelDissolutionRequest(context.Context, uint64, uint64, uint64, dto.CancelDissolutionRequest, AuditInput) (*vo.DissolutionRequest, *apperrors.BusinessError)
	Leave(context.Context, uint64, uint64, dto.LeaveFamilyRequest, AuditInput) (*vo.LeaveFamilyResult, *apperrors.BusinessError)
}

type familyService struct {
	db          *gorm.DB
	repo        familyrepo.FamilyRepository
	permissions permission.FamilyPermissionService
	quota       quotaservice.Service
}

func NewFamilyService(db *gorm.DB, repo familyrepo.FamilyRepository, permissions permission.FamilyPermissionService, quota quotaservice.Service) FamilyService {
	return &familyService{db: db, repo: repo, permissions: permissions, quota: quota}
}

func (s *familyService) Create(ctx context.Context, userID uint64, req dto.CreateFamilyRequest, audit AuditInput) (*vo.FamilyDetail, *apperrors.BusinessError) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil || user.Status != string(enums.StatusActive) {
		return nil, familyError(CodeFamilyCreateStatusDenied, "当前账号状态不允许创建家庭")
	}
	surname := strings.TrimSpace(req.Surname)
	if surname == "" {
		return nil, familyError(CodeFamilySurnameRequired, "家庭姓氏不能为空")
	}
	familyName := firstNonBlank(req.FamilyName, req.Name)
	if familyName == "" {
		familyName = surname + "氏家族"
	}
	displayName := surname + "氏创建者"
	if user.Nickname != nil && strings.TrimSpace(*user.Nickname) != "" {
		displayName = strings.TrimSpace(*user.Nickname)
	}
	founderGender := string(enums.GenderUnknown)
	if req.FounderGender != nil {
		founderGender = strings.ToUpper(strings.TrimSpace(*req.FounderGender))
		if founderGender != string(enums.GenderMale) && founderGender != string(enums.GenderFemale) {
			return nil, familyError(CodeFamilyCreateStatusDenied, "创建者性别必须选择男或女")
		}
	}

	family := &familymodel.Family{
		FamilyName:           familyName,
		FamilySurname:        surname,
		NativePlace:          cleanString(req.NativePlace),
		RegionCode:           cleanString(req.RegionCode),
		RegionText:           cleanString(req.RegionText),
		Description:          cleanString(req.Description),
		AvatarURL:            cleanString(req.AvatarURL),
		CreatorUserID:        &userID,
		Status:               familyStatusNormal,
		Searchable:           true,
		PublicDisplayStatus:  familyPublicPrivate,
		PublicContactVisible: true,
		TreeMode:             string(enums.TreeModeListTree),
		GraphVersion:         1,
	}
	if req.PublicContact != nil {
		family.PublicContactName = cleanString(req.PublicContact.Name)
		family.PublicContactPhone = cleanString(req.PublicContact.Phone)
		family.PublicContactWechat = cleanString(req.PublicContact.Wechat)
		family.PublicContactNote = cleanString(req.PublicContact.Note)
		if req.PublicContact.Visible != nil {
			family.PublicContactVisible = *req.PublicContact.Visible
		}
	}

	role := string(enums.FamilyRoleFounder)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if s.quota != nil {
			if err := s.quota.AssertProfileComplete(ctx, tx, userID); err != nil {
				return err
			}
			if err := s.quota.AssertCanCreateFamily(ctx, tx, userID); err != nil {
				return err
			}
		}
		txRepo := s.repo.WithTx(tx)
		if err := txRepo.CreateFamily(ctx, family); err != nil {
			return err
		}
		member := &membermodel.FamilyMember{
			FamilyID:          family.ID,
			MemberType:        string(enums.StatusLineageMember),
			Surname:           &surname,
			DisplayName:       displayName,
			Gender:            founderGender,
			UserBindingPolicy: string(enums.UserBindingOptional),
			Status:            string(enums.StatusActive),
			CreatedByUserID:   &userID,
		}
		if err := txRepo.CreateMember(ctx, member); err != nil {
			return err
		}
		now := time.Now()
		link := &rolemodel.FamilyMemberUserLink{
			FamilyID:            family.ID,
			MemberID:            member.ID,
			UserID:              userID,
			LinkStatus:          string(enums.StatusActive),
			LinkSource:          "FAMILY_CREATE",
			FamilyRole:          role,
			RoleGrantedAt:       &now,
			RoleGrantedByUserID: &userID,
		}
		if err := txRepo.CreateLink(ctx, link); err != nil {
			return err
		}
		if err := txRepo.UpdateFamily(ctx, family.ID, map[string]any{"current_founder_member_id": member.ID}); err != nil {
			return err
		}
		family.CurrentFounderMemberID = &member.ID
		return operationlog.NewGormService(tx).WriteSuccess(ctx, userOperationLog(
			userID, "CREATE_FAMILY", "FAMILY", family.ID, family.ID, &member.ID, audit,
		))
	})
	if err != nil {
		if s.quota != nil {
			if businessErr := s.quota.MapQuotaError(err); businessErr != nil && businessErr.Code != apperrors.CodeSystemError {
				return nil, businessErr
			}
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return familyDetail(family, role), nil
}

func (s *familyService) List(ctx context.Context, userID uint64) ([]vo.FamilySummary, *apperrors.BusinessError) {
	rows, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := make([]vo.FamilySummary, 0, len(rows))
	for _, row := range rows {
		result = append(result, vo.FamilySummary{
			ID: row.ID, FamilyName: row.FamilyName, FamilySurname: row.FamilySurname,
			NativePlace: row.NativePlace, RegionText: row.RegionText, AvatarURL: row.AvatarURL,
			Status: row.Status, PublicDisplayStatus: row.PublicDisplayStatus, Role: row.FamilyRole,
		})
	}
	return result, nil
}

func (s *familyService) Detail(ctx context.Context, userID uint64, familyID uint64) (*vo.FamilyDetail, *apperrors.BusinessError) {
	member, err := s.permissions.IsFamilyMember(ctx, userID, familyID)
	if err != nil || !member {
		return nil, familyError(CodeFamilyDetailForbidden, "无权查看家庭详情")
	}
	family, err := s.repo.FindFamilyByID(ctx, familyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, familyError(CodeFamilyNotFound, "家庭不存在")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	link, err := s.permissions.GetActiveLink(ctx, userID, familyID)
	if err != nil {
		return nil, familyError(CodeFamilyDetailForbidden, "无权查看家庭详情")
	}
	return familyDetail(family, link.FamilyRole), nil
}

func (s *familyService) PublicDetail(ctx context.Context, familyID uint64) (*vo.PublicFamily, *apperrors.BusinessError) {
	family, err := s.repo.FindPublicFamilyByID(ctx, familyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, familyError(CodeFamilyNotPublic, "家庭公开信息不可访问")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := &vo.PublicFamily{
		ID: family.ID, FamilyName: family.FamilyName, FamilySurname: family.FamilySurname,
		NativePlace: family.NativePlace, RegionText: family.RegionText, Description: family.Description,
		AvatarURL: family.AvatarURL, PublicContactVisible: family.PublicContactVisible,
	}
	if family.PublicContactVisible {
		result.PublicContactName = family.PublicContactName
		result.PublicContactPhone = family.PublicContactPhone
		result.PublicContactWechat = family.PublicContactWechat
	}
	return vo.MaskPublicFamily(result), nil
}

func (s *familyService) ListPublicFamilies(ctx context.Context, req dto.ListPublicFamiliesQuery) (*vo.ListPublicFamiliesResult, *apperrors.BusinessError) {
	page, pageSize := normalizePublicSearchPage(req.Page, req.PageSize)
	rows, total, err := s.repo.SearchPublicFamilies(ctx, familyrepo.PublicFamilySearchQuery{
		Keyword:       req.Keyword,
		FamilySurname: req.FamilySurname,
		RegionText:    req.RegionText,
		Page:          page,
		PageSize:      pageSize,
	})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	items := make([]vo.PublicFamilyListItem, 0, len(rows))
	for i := range rows {
		items = append(items, publicFamilyListItem(&rows[i]))
	}
	return &vo.ListPublicFamiliesResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *familyService) ListPublicFamilyShowcase(ctx context.Context, req dto.ListPublicFamilyShowcaseQuery) (*vo.ListPublicFamilyShowcaseResult, *apperrors.BusinessError) {
	page, pageSize := normalizePublicSearchPage(req.Page, req.PageSize)
	rows, total, err := s.repo.SearchPublicFamilies(ctx, familyrepo.PublicFamilySearchQuery{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	items := make([]vo.PublicFamilyShowcaseItem, 0, len(rows))
	for i := range rows {
		items = append(items, publicFamilyShowcaseItem(&rows[i]))
	}
	return &vo.ListPublicFamilyShowcaseResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *familyService) Update(ctx context.Context, userID uint64, familyID uint64, req dto.UpdateFamilyRequest, audit AuditInput) (*vo.FamilyDetail, *apperrors.BusinessError) {
	allowed, err := s.permissions.CanManageFamily(ctx, userID, familyID)
	if err != nil || !allowed {
		return nil, familyError(CodeFamilyUpdateForbidden, "无权修改家庭信息")
	}
	family, err := s.repo.FindFamilyByID(ctx, familyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, familyError(CodeFamilyNotFound, "家庭不存在")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if family.Status != familyStatusNormal {
		return nil, familyError(CodeFamilyUpdateStatusDenied, "当前家庭状态不允许修改")
	}
	values := updateValues(req)
	if len(values) > 0 {
		err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := s.repo.WithTx(tx).UpdateFamily(ctx, familyID, values); err != nil {
				return err
			}
			return operationlog.NewGormService(tx).WriteSuccess(ctx, userOperationLog(
				userID, "UPDATE_FAMILY", "FAMILY", familyID, familyID, nil, audit,
			))
		})
		if err != nil {
			return nil, apperrors.New(apperrors.CodeSystemError)
		}
		family, err = s.repo.FindFamilyByID(ctx, familyID)
		if err != nil {
			return nil, apperrors.New(apperrors.CodeSystemError)
		}
	}
	link, _ := s.permissions.GetActiveLink(ctx, userID, familyID)
	return familyDetail(family, link.FamilyRole), nil
}

func (s *familyService) Leave(ctx context.Context, userID uint64, familyID uint64, req dto.LeaveFamilyRequest, audit AuditInput) (*vo.LeaveFamilyResult, *apperrors.BusinessError) {
	var memberID uint64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		family, err := repo.FindFamilyByIDForUpdate(ctx, familyID)
		if err != nil {
			return err
		}
		if family.Status != familyStatusNormal {
			return errFamilyLeaveForbidden
		}
		link, err := repo.FindActiveLinkForUpdate(ctx, familyID, userID)
		if err != nil {
			return err
		}
		if link.FamilyRole == string(enums.FamilyRoleFounder) {
			return errFamilyFounderMustTransfer
		}
		memberID = link.MemberID
		now := time.Now()
		if err := repo.DeactivateLink(ctx, link.ID, userID, cleanString(req.Reason), now); err != nil {
			return err
		}
		return operationlog.NewGormService(tx).WriteSuccess(ctx, userOperationLog(
			userID, "LEAVE_FAMILY", "FAMILY_MEMBER_USER_LINK", link.ID, familyID, &memberID, audit,
		))
	})
	if err == nil {
		return &vo.LeaveFamilyResult{FamilyID: familyID, MemberID: memberID, Status: "LEFT"}, nil
	}
	switch {
	case errors.Is(err, errFamilyFounderMustTransfer):
		return nil, familyError(CodeFamilyLeaveForbidden, "家庭创建者需先转让创建者身份后才能退出")
	case errors.Is(err, errFamilyLeaveForbidden):
		return nil, familyError(CodeFamilyLeaveForbidden, "当前家庭状态不允许退出")
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, familyError(CodeFamilyLeaveNotLinked, "当前账号不在该家庭中")
	default:
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
}

func (s *familyService) CreateDissolutionRequest(ctx context.Context, userID uint64, familyID uint64, req dto.CreateDissolutionRequest, audit AuditInput) (*vo.DissolutionRequest, *apperrors.BusinessError) {
	allowed, err := s.permissions.CanManageFamily(ctx, userID, familyID)
	if err != nil || !allowed {
		return nil, familyError(CodeFamilyDissolutionForbidden, "无权操作家庭解散申请")
	}
	link, err := s.permissions.GetActiveLink(ctx, userID, familyID)
	if err != nil {
		return nil, familyError(CodeFamilyDissolutionForbidden, "无权操作家庭解散申请")
	}
	if link.FamilyRole != string(enums.FamilyRoleFounder) {
		return nil, familyError(CodeFamilyDissolutionForbidden, "无权操作家庭解散申请")
	}
	request := &dissolutionmodel.FamilyDissolutionRequest{
		FamilyID: familyID, RequesterMemberID: link.MemberID, RequesterUserID: userID,
		RequestStatus: dissolutionStatusPending, RequestReason: cleanString(req.RequestReason),
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		family, err := txRepo.FindFamilyByIDForUpdate(ctx, familyID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		} else if err != nil {
			return err
		}
		if _, err := txRepo.FindCurrentDissolutionRequest(ctx, familyID); err == nil {
			return errDissolutionPending
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if family.Status != familyStatusNormal {
			return errDissolutionFamilyStatus
		}
		if err := txRepo.CreateDissolutionRequest(ctx, request); err != nil {
			return err
		}
		if err := txRepo.UpdateFamily(ctx, familyID, map[string]any{"status": familyStatusDissolutionPending}); err != nil {
			return err
		}
		return operationlog.NewGormService(tx).WriteSuccess(ctx, userOperationLog(
			userID, "CREATE_DISSOLUTION_REQUEST", "FAMILY_DISSOLUTION_REQUEST", request.ID, familyID, &link.MemberID, audit,
		))
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, familyError(CodeFamilyNotFound, "家庭不存在")
	}
	if errors.Is(err, errDissolutionPending) {
		return nil, familyError(CodeFamilyDissolutionPending, "家庭已有待处理的解散申请")
	}
	if errors.Is(err, errDissolutionFamilyStatus) {
		return nil, familyError(CodeFamilyDissolutionStatusDenied, "当前家庭状态不允许申请解散")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return dissolutionVO(request), nil
}

func (s *familyService) CurrentDissolutionRequest(ctx context.Context, userID uint64, familyID uint64) (*vo.DissolutionRequest, *apperrors.BusinessError) {
	member, err := s.permissions.IsFamilyMember(ctx, userID, familyID)
	if err != nil || !member {
		return nil, familyError(CodeFamilyDetailForbidden, "无权查看家庭详情")
	}
	request, err := s.repo.FindCurrentDissolutionRequest(ctx, familyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return dissolutionVO(request), nil
}

func (s *familyService) CancelDissolutionRequest(ctx context.Context, userID uint64, familyID uint64, requestID uint64, req dto.CancelDissolutionRequest, audit AuditInput) (*vo.DissolutionRequest, *apperrors.BusinessError) {
	request, err := s.repo.FindDissolutionRequest(ctx, familyID, requestID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.New(apperrors.CodeResourceNotFound)
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if request.RequestStatus != dissolutionStatusPending {
		return nil, apperrors.New(apperrors.CodeInvalidDataStatus)
	}
	admin, adminErr := s.permissions.IsFamilyAdmin(ctx, userID, familyID)
	if request.RequesterUserID != userID && (adminErr != nil || !admin) {
		return nil, familyError(CodeFamilyDissolutionForbidden, "无权操作家庭解散申请")
	}
	now := time.Now()
	reason := cleanString(req.CancelReason)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		if err := txRepo.CancelDissolutionRequest(ctx, requestID, reason, now); err != nil {
			return err
		}
		if err := txRepo.UpdateFamily(ctx, familyID, map[string]any{"status": familyStatusNormal}); err != nil {
			return err
		}
		return operationlog.NewGormService(tx).WriteSuccess(ctx, userOperationLog(
			userID, "CANCEL_DISSOLUTION_REQUEST", "FAMILY_DISSOLUTION_REQUEST", requestID, familyID, &request.RequesterMemberID, audit,
		))
	})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	request.RequestStatus = dissolutionStatusCancelled
	request.CancelledAt = &now
	request.CancelReason = reason
	return dissolutionVO(request), nil
}

func publicFamilyListItem(family *familymodel.Family) vo.PublicFamilyListItem {
	item := vo.PublicFamilyListItem{
		ID:                   family.ID,
		FamilyName:           family.FamilyName,
		FamilySurname:        family.FamilySurname,
		NativePlace:          family.NativePlace,
		RegionText:           family.RegionText,
		Description:          family.Description,
		AvatarURL:            family.AvatarURL,
		PublicContactVisible: family.PublicContactVisible,
		PublicApprovedAt:     family.PublicApprovedAt,
		CreatedAt:            family.CreatedAt,
		UpdatedAt:            family.UpdatedAt,
	}
	if family.PublicContactVisible {
		item.PublicContactName = family.PublicContactName
		item.PublicContactNote = family.PublicContactNote
	}
	return item
}

func publicFamilyShowcaseItem(family *familymodel.Family) vo.PublicFamilyShowcaseItem {
	return vo.PublicFamilyShowcaseItem{
		ID:               family.ID,
		FamilyName:       family.FamilyName,
		FamilySurname:    family.FamilySurname,
		NativePlace:      family.NativePlace,
		RegionText:       family.RegionText,
		Description:      family.Description,
		AvatarURL:        family.AvatarURL,
		PublicApprovedAt: family.PublicApprovedAt,
		CreatedAt:        family.CreatedAt,
		UpdatedAt:        family.UpdatedAt,
	}
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

func familyDetail(family *familymodel.Family, role string) *vo.FamilyDetail {
	return &vo.FamilyDetail{
		ID: family.ID, FamilyName: family.FamilyName, FamilySurname: family.FamilySurname,
		NativePlace: family.NativePlace, RegionCode: family.RegionCode, RegionText: family.RegionText,
		Description: family.Description, AvatarURL: family.AvatarURL, Status: family.Status,
		Searchable: family.Searchable, PublicDisplayStatus: family.PublicDisplayStatus,
		PublicContactName: family.PublicContactName, PublicContactPhone: family.PublicContactPhone,
		PublicContactWechat: family.PublicContactWechat, PublicContactNote: family.PublicContactNote,
		PublicContactVisible: family.PublicContactVisible, CurrentFounderMemberID: family.CurrentFounderMemberID,
		GraphVersion: family.GraphVersion, Role: role,
	}
}

func dissolutionVO(request *dissolutionmodel.FamilyDissolutionRequest) *vo.DissolutionRequest {
	return &vo.DissolutionRequest{
		ID: request.ID, FamilyID: request.FamilyID, RequesterMemberID: request.RequesterMemberID,
		RequesterUserID: request.RequesterUserID, RequestStatus: request.RequestStatus,
		RequestReason: request.RequestReason, CancelledAt: request.CancelledAt,
		CancelReason: request.CancelReason, CreatedAt: request.CreatedAt,
	}
}

func updateValues(req dto.UpdateFamilyRequest) map[string]any {
	values := map[string]any{}
	assignString := func(column string, value *string) {
		if value != nil {
			values[column] = cleanString(value)
		}
	}
	assignString("family_name", req.FamilyName)
	assignString("native_place", req.NativePlace)
	assignString("region_code", req.RegionCode)
	assignString("region_text", req.RegionText)
	assignString("description", req.Description)
	assignString("avatar_url", req.AvatarURL)
	assignString("public_contact_name", req.PublicContactName)
	assignString("public_contact_phone", req.PublicContactPhone)
	assignString("public_contact_wechat", req.PublicContactWechat)
	assignString("public_contact_note", req.PublicContactNote)
	if req.Searchable != nil {
		values["searchable"] = *req.Searchable
	}
	if req.PublicContactVisible != nil {
		values["public_contact_visible"] = *req.PublicContactVisible
	}
	return values
}

func firstNonBlank(values ...*string) string {
	for _, value := range values {
		if value != nil && strings.TrimSpace(*value) != "" {
			return strings.TrimSpace(*value)
		}
	}
	return ""
}

func cleanString(value *string) *string {
	if value == nil {
		return nil
	}
	cleaned := strings.TrimSpace(*value)
	if cleaned == "" {
		return nil
	}
	return &cleaned
}

func userOperationLog(userID uint64, action string, targetType string, targetID uint64, familyID uint64, memberID *uint64, audit AuditInput) operationlog.WriteInput {
	return operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &userID, Module: "FAMILY_CORE", Action: action,
		TargetType: &targetType, TargetID: &targetID, FamilyID: &familyID, MemberID: memberID,
		IP: cleanString(&audit.IP), UserAgent: cleanString(&audit.UserAgent),
	}
}

func familyError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}
