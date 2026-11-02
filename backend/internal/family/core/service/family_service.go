package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	quotaservice "tree/backend/internal/accountquota/service"
	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/permission"
	commonredis "tree/backend/internal/common/redis"
	"tree/backend/internal/family/core/dto"
	familymodel "tree/backend/internal/family/core/model"
	familyrepo "tree/backend/internal/family/core/repository"
	"tree/backend/internal/family/core/vo"
	dissolutionmodel "tree/backend/internal/family/dissolution/model"
	membermodel "tree/backend/internal/family/member/model"
	relationshipmodel "tree/backend/internal/family/relationship/model"
	rolemodel "tree/backend/internal/family/role/model"
	notificationservice "tree/backend/internal/notification/service"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	familyStatusNormal              = "NORMAL"
	familyStatusDissolutionPending  = "DISSOLUTION_PENDING"
	familyStatusDissolutionCooldown = "DISSOLUTION_COOLDOWN"
	familyStatusDissolved           = "DISSOLVED"
	familyPublicPrivate             = "PRIVATE"
	dissolutionStatusPending        = "PENDING"
	dissolutionStatusCancelled      = "CANCELLED"

	CodeFamilySurnameRequired           apperrors.Code = 42004
	CodeFamilyCreateStatusDenied        apperrors.Code = 42006
	CodeFamilyNotFound                  apperrors.Code = 42101
	CodeFamilyNotPublic                 apperrors.Code = 42103
	CodeFamilyDetailForbidden           apperrors.Code = 42104
	CodeFamilyUpdateForbidden           apperrors.Code = 42201
	CodeFamilyUpdateStatusDenied        apperrors.Code = 42202
	CodeFamilyLeaveForbidden            apperrors.Code = 42203
	CodeFamilyLeaveNotLinked            apperrors.Code = 42204
	CodeFamilyDissolutionForbidden      apperrors.Code = 46301
	CodeFamilyDissolutionStatusDenied   apperrors.Code = 46302
	CodeFamilyDissolutionPending        apperrors.Code = 46303
	CodeFamilyDissolutionFinalizeDenied apperrors.Code = 46304
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
	PublicPlatformStats(context.Context) (*vo.PublicPlatformStats, *apperrors.BusinessError)
	ListOperationLogs(context.Context, uint64, uint64, int, int) (*vo.ListOperationLogsResult, *apperrors.BusinessError)
	Update(context.Context, uint64, uint64, dto.UpdateFamilyRequest, AuditInput) (*vo.FamilyDetail, *apperrors.BusinessError)
	ExportArchive(context.Context, uint64, uint64, string, AuditInput) (*vo.FamilyArchiveExport, *apperrors.BusinessError)
	CreateDissolutionRequest(context.Context, uint64, uint64, dto.CreateDissolutionRequest, AuditInput) (*vo.DissolutionRequest, *apperrors.BusinessError)
	CurrentDissolutionRequest(context.Context, uint64, uint64) (*vo.DissolutionRequest, *apperrors.BusinessError)
	CancelDissolutionRequest(context.Context, uint64, uint64, uint64, dto.CancelDissolutionRequest, AuditInput) (*vo.DissolutionRequest, *apperrors.BusinessError)
	FinalizeDissolution(context.Context, uint64, uint64, AuditInput) (*vo.FamilyDetail, *apperrors.BusinessError)
	RestoreDissolution(context.Context, uint64, uint64, AuditInput) (*vo.FamilyDetail, *apperrors.BusinessError)
	Leave(context.Context, uint64, uint64, dto.LeaveFamilyRequest, AuditInput) (*vo.LeaveFamilyResult, *apperrors.BusinessError)
}

func (s *familyService) ExportArchive(ctx context.Context, userID, familyID uint64, scope string, audit AuditInput) (*vo.FamilyArchiveExport, *apperrors.BusinessError) {
	if allowed, err := s.permissions.CanManageFamily(ctx, userID, familyID); err != nil || !allowed {
		return nil, familyError(CodeFamilyUpdateForbidden, "无权导出家庭档案")
	}
	scope = strings.ToUpper(strings.TrimSpace(scope))
	if scope == "" {
		scope = "INTERNAL"
	}
	if scope != "INTERNAL" && scope != "SHARE" && scope != "PUBLIC" {
		return nil, familyError(apperrors.CodeInvalidParams, "导出范围不正确")
	}
	family, err := s.repo.FindFamilyByID(ctx, familyID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	var members []membermodel.FamilyMember
	if err := s.db.WithContext(ctx).Where("family_id = ? AND status = ? AND deleted_at IS NULL", familyID, "ACTIVE").Order("id").Find(&members).Error; err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	items := make([]vo.ArchiveMember, 0, len(members))
	missing := map[string]bool{}
	incompleteCount := 0
	for _, member := range members {
		incomplete := false
		if strings.TrimSpace(member.DisplayName) == "" {
			missing["姓名"] = true
			incomplete = true
		}
		if member.Gender == "" || member.Gender == string(enums.GenderUnknown) {
			missing["性别"] = true
			incomplete = true
		}
		if member.BirthDate == nil {
			missing["出生年份"] = true
			incomplete = true
		}
		if incomplete {
			incompleteCount++
		}
		item := vo.ArchiveMember{MemberID: member.ID, Name: member.DisplayName}
		if scope != "PUBLIC" {
			item.Gender = member.Gender
		}
		if member.BirthDate != nil {
			year := member.BirthDate.Year()
			item.BirthYear = &year
		}
		if member.DeathDate != nil {
			year := member.DeathDate.Year()
			item.DeathYear = &year
		}
		items = append(items, item)
	}
	missingFields := make([]string, 0, len(missing))
	for field := range missing {
		missingFields = append(missingFields, field)
	}
	sort.Strings(missingFields)
	if scope == "PUBLIC" && incompleteCount > 0 {
		return nil, familyError(CodeFamilyUpdateStatusDenied, fmt.Sprintf("仍有 %d 位成员缺少%s，暂不能生成公开长卷", incompleteCount, strings.Join(missingFields, "、")))
	}
	var relationships []relationshipmodel.FamilyRelationship
	if err := s.db.WithContext(ctx).Where("family_id = ? AND status = ? AND deleted_at IS NULL", familyID, "ACTIVE").Order("id").Find(&relationships).Error; err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	edges := make([]vo.ArchiveRelationship, 0, len(relationships))
	for _, relationship := range relationships {
		edges = append(edges, vo.ArchiveRelationship{FromMemberID: relationship.FromMemberID, ToMemberID: relationship.ToMemberID, RelationshipType: relationship.RelationshipType})
	}
	link, _ := s.permissions.GetActiveLink(ctx, userID, familyID)
	result := &vo.FamilyArchiveExport{Scope: scope, GeneratedAt: time.Now(), Family: *familyDetail(family, link.FamilyRole), Members: items, Relationships: edges, Quality: vo.ArchiveQuality{IncompleteMemberCount: incompleteCount, MissingFields: missingFields}}
	if scope == "PUBLIC" {
		result.Family.Description = nil
		result.Family.NativePlace = nil
	}
	_ = operationlog.NewGormService(s.db).WriteSuccess(ctx, userOperationLog(userID, "EXPORT_FAMILY_ARCHIVE_"+scope, "FAMILY", familyID, familyID, nil, audit))
	return result, nil
}

type familyService struct {
	db            *gorm.DB
	repo          familyrepo.FamilyRepository
	permissions   permission.FamilyPermissionService
	quota         quotaservice.Service
	contentSafety contentsafety.Service
	notifications notificationservice.Service
	treeCache     commonredis.TreeCache
}

func NewFamilyService(db *gorm.DB, repo familyrepo.FamilyRepository, permissions permission.FamilyPermissionService, quota quotaservice.Service, contentSafety contentsafety.Service) FamilyService {
	return newFamilyService(db, repo, permissions, quota, contentSafety, nil)
}

func NewFamilyServiceWithNotifications(db *gorm.DB, repo familyrepo.FamilyRepository, permissions permission.FamilyPermissionService, quota quotaservice.Service, contentSafety contentsafety.Service, notifications notificationservice.Service, treeCaches ...commonredis.TreeCache) FamilyService {
	return newFamilyService(db, repo, permissions, quota, contentSafety, notifications, treeCaches...)
}

func newFamilyService(db *gorm.DB, repo familyrepo.FamilyRepository, permissions permission.FamilyPermissionService, quota quotaservice.Service, contentSafety contentsafety.Service, notifications notificationservice.Service, treeCaches ...commonredis.TreeCache) FamilyService {
	if contentSafety == nil {
		contentSafety = contentsafety.FailClosed()
	}
	treeCache := commonredis.TreeCache(commonredis.NoopTreeCache{})
	if len(treeCaches) > 0 && treeCaches[0] != nil {
		treeCache = treeCaches[0]
	}
	return &familyService{db: db, repo: repo, permissions: permissions, quota: quota, contentSafety: contentSafety, notifications: notifications, treeCache: treeCache}
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
	displayName := strings.TrimSpace(req.FounderName)
	// HTTP callers must provide the founder profile. Keep the historical default for
	// internal callers during the transition from the previous request contract.
	if displayName == "" {
		displayName = surname + "氏创建者"
	}
	founderGender := string(enums.GenderUnknown)
	if req.FounderGender != nil {
		founderGender = strings.ToUpper(strings.TrimSpace(*req.FounderGender))
		if founderGender != string(enums.GenderMale) && founderGender != string(enums.GenderFemale) {
			return nil, familyError(CodeFamilyCreateStatusDenied, "创建者性别必须选择男或女")
		}
	}
	var founderBirthDate *time.Time
	if req.FounderBirthYear != nil {
		currentYear := time.Now().Year()
		if *req.FounderBirthYear < 1 || *req.FounderBirthYear > currentYear {
			return nil, familyError(CodeFamilyCreateStatusDenied, "创建者出生年份不正确")
		}
		value := time.Date(*req.FounderBirthYear, time.January, 1, 0, 0, 0, 0, time.UTC)
		founderBirthDate = &value
	}
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: userID,
		Scene:  contentsafety.SceneProfile,
		Fields: familyCreateTextFields(surname, displayName, familyName, req.NativePlace, req.RegionText, req.Description, req.PublicContact),
		IP:     audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
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
			BirthDate:         founderBirthDate,
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
			Status: row.Status, DissolutionCooldownUntil: row.DissolutionCooldownUntil,
			DissolutionCooldownDays: row.DissolutionCooldownDays, PublicDisplayStatus: row.PublicDisplayStatus,
			PublicDisplayEnabled: row.PublicDisplayEnabled, Role: row.FamilyRole,
		})
	}
	return result, nil
}

func (s *familyService) ListOperationLogs(ctx context.Context, userID uint64, familyID uint64, page int, pageSize int) (*vo.ListOperationLogsResult, *apperrors.BusinessError) {
	allowed, err := s.permissions.CanManageFamily(ctx, userID, familyID)
	if err != nil || !allowed {
		return nil, familyError(CodeFamilyDetailForbidden, "无权查看家庭操作记录")
	}
	page, pageSize = normalizeOperationLogPage(page, pageSize)
	rows, total, err := s.repo.ListOperationLogs(ctx, familyID, page, pageSize)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	items := make([]vo.OperationLogItem, 0, len(rows))
	for i := range rows {
		row := rows[i]
		items = append(items, vo.OperationLogItem{
			ID: row.ID, OperatorType: row.OperatorType, OperatorAdminID: row.OperatorAdminID,
			OperatorUserID: row.OperatorUserID, Module: row.Module, Action: row.Action,
			TargetType: row.TargetType, TargetID: row.TargetID, MemberID: row.MemberID,
			UserID: row.UserID, Result: row.Result, ErrorMessage: row.ErrorMessage,
			CreatedAt: row.CreatedAt,
		})
	}
	return &vo.ListOperationLogsResult{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
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
	result := familyDetail(family, link.FamilyRole)
	if reader, ok := s.quota.(quotaservice.FamilyCapacityReader); ok {
		capacity, capacityErr := reader.GetFamilyCapacity(ctx, s.db, familyID)
		if capacityErr != nil {
			return nil, apperrors.New(apperrors.CodeSystemError)
		}
		result.MemberCount = capacity.UsedMembers
		result.MemberLimit = capacity.EffectiveLimit
	}
	return result, nil
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

func (s *familyService) PublicPlatformStats(ctx context.Context) (*vo.PublicPlatformStats, *apperrors.BusinessError) {
	stats, err := s.repo.PublicPlatformStats(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &vo.PublicPlatformStats{
		FamilyCount:       stats.FamilyCount,
		MemberCount:       stats.MemberCount,
		RelationshipCount: stats.RelationshipCount,
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
	if req.ExpectedRevision != nil && *req.ExpectedRevision != family.Revision {
		return nil, familyError(apperrors.CodeDataConflict, "家庭资料已被家人更新，请刷新后再保存")
	}
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID:   userID,
		Scene:    contentsafety.SceneProfile,
		Fields:   familyUpdateTextFields(req),
		FamilyID: &familyID,
		IP:       audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	values := updateValues(req)
	if len(values) > 0 {
		values["revision"] = gorm.Expr("revision + 1")
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
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: userID, Scene: contentsafety.SceneSocial,
		Fields:   contentsafety.OptionalField("leave_reason", req.Reason),
		FamilyID: &familyID, IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	var memberID uint64
	var memberName string
	var notificationRecipientIDs []uint64
	var leftAt time.Time
	var cacheGraphVersion int64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		family, err := repo.FindFamilyByIDForUpdate(ctx, familyID)
		if err != nil {
			return err
		}
		if family.Status != familyStatusNormal {
			return errFamilyLeaveForbidden
		}
		cacheGraphVersion = family.GraphVersion
		link, err := repo.FindActiveLinkForUpdate(ctx, familyID, userID)
		if err != nil {
			return err
		}
		if link.FamilyRole == string(enums.FamilyRoleFounder) {
			return errFamilyFounderMustTransfer
		}
		memberID = link.MemberID
		member, err := repo.FindMemberByID(ctx, familyID, memberID)
		if err != nil {
			return err
		}
		memberName = strings.TrimSpace(member.DisplayName)
		if memberName == "" {
			memberName = "一位成员"
		}
		notificationRecipientIDs, err = repo.ListFamilyNotificationRecipientUserIDs(ctx, familyID, userID)
		if err != nil {
			return err
		}
		now := time.Now()
		leftAt = now
		if err := repo.DeactivateLink(ctx, link.ID, userID, cleanString(req.Reason), now); err != nil {
			return err
		}
		return operationlog.NewGormService(tx).WriteSuccess(ctx, userOperationLog(
			userID, "LEAVE_FAMILY", "FAMILY_MEMBER_USER_LINK", link.ID, familyID, &memberID, audit,
		))
	})
	if err == nil {
		// Binding state is part of the tree payload but leaving does not change its structure.
		// Evict the current snapshot without incrementing graph_version.
		_ = s.treeCache.Delete(ctx, fmt.Sprintf("family:%d:tree:v%d", familyID, cacheGraphVersion))
		if s.notifications != nil {
			reason := cleanString(req.Reason)
			if reason == nil {
				defaultReason := "成员已解除家庭绑定。"
				reason = &defaultReason
			}
			for _, recipientID := range notificationRecipientIDs {
				s.notifications.Create(ctx, notificationservice.Input{
					UserID: recipientID, Category: "FAMILY_MEMBER_CHANGE", Title: "家庭成员已退出",
					Content: fmt.Sprintf("成员「%s」已退出家庭。", memberName), Route: fmt.Sprintf("/pages/family/detail?familyId=%d&memberId=%d", familyID, memberID), RelatedType: "FAMILY_MEMBER_CHANGE", RelatedID: familyID,
					Subscribe: &notificationservice.SubscribeInput{Kind: "MEMBER_CHANGE", Page: fmt.Sprintf("pages/family/detail?familyId=%d&memberId=%d", familyID, memberID), Values: []string{memberName, "退出家庭", leftAt.Format("2006-01-02 15:04"), *reason}},
				})
			}
		}
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
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: userID, Scene: contentsafety.SceneSocial,
		Fields:   contentsafety.OptionalField("dissolution_request_reason", req.RequestReason),
		FamilyID: &familyID, IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
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
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: userID, Scene: contentsafety.SceneSocial,
		Fields:   contentsafety.OptionalField("dissolution_cancel_reason", req.CancelReason),
		FamilyID: &familyID, IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
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

func (s *familyService) FinalizeDissolution(ctx context.Context, userID uint64, familyID uint64, audit AuditInput) (*vo.FamilyDetail, *apperrors.BusinessError) {
	allowed, err := s.permissions.CanManageFamily(ctx, userID, familyID)
	if err != nil || !allowed {
		return nil, familyError(CodeFamilyDissolutionFinalizeDenied, "无权跳过冷静期")
	}
	link, err := s.permissions.GetActiveLink(ctx, userID, familyID)
	if err != nil || link.FamilyRole != string(enums.FamilyRoleFounder) {
		return nil, familyError(CodeFamilyDissolutionFinalizeDenied, "只有创建者可以跳过冷静期")
	}
	now := time.Now()
	var family *familymodel.Family
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		value, err := txRepo.FindFamilyByIDForUpdate(ctx, familyID)
		if err != nil {
			return err
		}
		if value.Status != familyStatusDissolutionCooldown {
			return errDissolutionFamilyStatus
		}
		if err := txRepo.UpdateFamily(ctx, familyID, map[string]any{
			"status":                   familyStatusDissolved,
			"dissolution_completed_at": now,
		}); err != nil {
			return err
		}
		if err := operationlog.NewGormService(tx).WriteSuccess(ctx, userOperationLog(
			userID, "FINALIZE_DISSOLUTION", "FAMILY", familyID, familyID, &link.MemberID, audit,
		)); err != nil {
			return err
		}
		value.Status = familyStatusDissolved
		value.DissolutionCompletedAt = &now
		family = value
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, familyError(CodeFamilyNotFound, "家庭不存在")
	}
	if errors.Is(err, errDissolutionFamilyStatus) {
		return nil, familyError(CodeFamilyDissolutionStatusDenied, "当前家庭不处于冷静期")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return familyDetail(family, link.FamilyRole), nil
}

func (s *familyService) RestoreDissolution(ctx context.Context, userID uint64, familyID uint64, audit AuditInput) (*vo.FamilyDetail, *apperrors.BusinessError) {
	allowed, err := s.permissions.CanManageFamily(ctx, userID, familyID)
	if err != nil || !allowed {
		return nil, familyError(CodeFamilyDissolutionFinalizeDenied, "无权恢复家庭")
	}
	link, err := s.permissions.GetActiveLink(ctx, userID, familyID)
	if err != nil || link.FamilyRole != string(enums.FamilyRoleFounder) {
		return nil, familyError(CodeFamilyDissolutionFinalizeDenied, "只有创建者可以恢复家庭")
	}
	now := time.Now()
	var family *familymodel.Family
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		value, err := txRepo.FindFamilyByIDForUpdate(ctx, familyID)
		if err != nil {
			return err
		}
		if value.Status != familyStatusDissolutionCooldown {
			return errDissolutionFamilyStatus
		}
		if s.quota != nil {
			if err := s.quota.AssertCanRestoreFamily(ctx, tx, familyID); err != nil {
				return err
			}
		}
		if err := txRepo.UpdateFamily(ctx, familyID, map[string]any{
			"status":                     familyStatusNormal,
			"public_display_status":      string(enums.StatusPrivate),
			"public_display_enabled":     false,
			"searchable":                 true,
			"dissolution_cooldown_until": nil,
			"dissolution_cooldown_days":  0,
			"dissolution_hidden_at":      nil,
			"dissolution_completed_at":   nil,
			"restored_at":                now,
		}); err != nil {
			return err
		}
		if err := operationlog.NewGormService(tx).WriteSuccess(ctx, userOperationLog(
			userID, "RESTORE_FAMILY_DURING_COOLDOWN", "FAMILY", familyID, familyID, &link.MemberID, audit,
		)); err != nil {
			return err
		}
		value.Status = familyStatusNormal
		value.PublicDisplayStatus = string(enums.StatusPrivate)
		value.PublicDisplayEnabled = false
		value.Searchable = true
		value.DissolutionCooldownUntil = nil
		value.DissolutionCooldownDays = 0
		value.DissolutionHiddenAt = nil
		value.DissolutionCompletedAt = nil
		family = value
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, familyError(CodeFamilyNotFound, "家庭不存在")
	}
	if errors.Is(err, errDissolutionFamilyStatus) {
		return nil, familyError(CodeFamilyDissolutionStatusDenied, "当前家庭不处于冷静期")
	}
	if err != nil {
		if s.quota != nil {
			if businessErr := s.quota.MapQuotaError(err); businessErr != nil && businessErr.Code != apperrors.CodeSystemError {
				return nil, businessErr
			}
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return familyDetail(family, link.FamilyRole), nil
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

func familyDetail(family *familymodel.Family, role string) *vo.FamilyDetail {
	return &vo.FamilyDetail{
		ID: family.ID, FamilyName: family.FamilyName, FamilySurname: family.FamilySurname,
		NativePlace: family.NativePlace, RegionCode: family.RegionCode, RegionText: family.RegionText,
		Description: family.Description, AvatarURL: family.AvatarURL, Status: family.Status,
		DissolutionCooldownUntil: family.DissolutionCooldownUntil,
		DissolutionCooldownDays:  family.DissolutionCooldownDays,
		DissolutionCompletedAt:   family.DissolutionCompletedAt,
		Searchable:               family.Searchable, PublicDisplayStatus: family.PublicDisplayStatus,
		PublicDisplayEnabled: family.PublicDisplayEnabled,
		PublicContactName:    family.PublicContactName, PublicContactPhone: family.PublicContactPhone,
		PublicContactWechat: family.PublicContactWechat, PublicContactNote: family.PublicContactNote,
		PublicContactVisible: family.PublicContactVisible, CurrentFounderMemberID: family.CurrentFounderMemberID,
		GraphVersion: family.GraphVersion, Revision: family.Revision, Role: role,
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

func familyCreateTextFields(surname, founderName, familyName string, nativePlace, regionText, description *string, contact *dto.PublicContact) []contentsafety.Field {
	fields := contentsafety.MergeFields(
		contentsafety.StringField("surname", surname),
		contentsafety.StringField("founder_name", founderName),
		contentsafety.StringField("family_name", familyName),
		contentsafety.OptionalField("native_place", nativePlace),
		contentsafety.OptionalField("region_text", regionText),
		contentsafety.OptionalField("description", description),
	)
	if contact != nil {
		fields = contentsafety.MergeFields(fields,
			contentsafety.OptionalField("public_contact_name", contact.Name),
			contentsafety.OptionalField("public_contact_wechat", contact.Wechat),
			contentsafety.OptionalField("public_contact_note", contact.Note),
		)
	}
	return fields
}

func familyUpdateTextFields(req dto.UpdateFamilyRequest) []contentsafety.Field {
	return contentsafety.MergeFields(
		contentsafety.OptionalField("family_name", req.FamilyName),
		contentsafety.OptionalField("native_place", req.NativePlace),
		contentsafety.OptionalField("region_text", req.RegionText),
		contentsafety.OptionalField("description", req.Description),
		contentsafety.OptionalField("public_contact_name", req.PublicContactName),
		contentsafety.OptionalField("public_contact_wechat", req.PublicContactWechat),
		contentsafety.OptionalField("public_contact_note", req.PublicContactNote),
	)
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
