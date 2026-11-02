package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/permission"
	familymodel "tree/backend/internal/family/core/model"
	"tree/backend/internal/family/publicdisplay/dto"
	publicenum "tree/backend/internal/family/publicdisplay/enum"
	publicmodel "tree/backend/internal/family/publicdisplay/model"
	publicrepo "tree/backend/internal/family/publicdisplay/repository"
	"tree/backend/internal/family/publicdisplay/vo"
	notificationservice "tree/backend/internal/notification/service"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	CodePublicApplicationNotFound      apperrors.Code = 45001
	CodePublicApplicationInvalidStatus apperrors.Code = 45002
	CodePublicApplicationDuplicate     apperrors.Code = 45003
	CodePublicFamilyUnavailable        apperrors.Code = 45004
	CodePublicApplicationForbidden     apperrors.Code = 45005
	CodePublicFamilyAlreadyApproved    apperrors.Code = 45006
	CodePublicFamilyTakeDownDenied     apperrors.Code = 45007
)

type AuditInput struct {
	IP        string
	UserAgent string
}

type Service interface {
	QualityEligibility(context.Context, uint64, uint64) (*vo.QualityEligibility, *apperrors.BusinessError)
	GetQualityEligibilityConfig(context.Context, string) (*vo.QualityEligibilityConfig, *apperrors.BusinessError)
	UpdateQualityEligibilityConfig(context.Context, uint64, string, vo.QualityEligibilityConfig) (*vo.QualityEligibilityConfig, *apperrors.BusinessError)
	Submit(context.Context, uint64, uint64, dto.CreatePublicApplicationRequest, AuditInput) (*vo.PublicApplication, *apperrors.BusinessError)
	ListFamily(context.Context, uint64, uint64, dto.ListApplicationsQuery) (*vo.ListResult, *apperrors.BusinessError)
	Cancel(context.Context, uint64, uint64, uint64, dto.CancelPublicApplicationRequest, AuditInput) (*vo.PublicApplication, *apperrors.BusinessError)
	EnableUser(context.Context, uint64, uint64, AuditInput) (*vo.FamilyPublicStatus, *apperrors.BusinessError)
	DisableUser(context.Context, uint64, uint64, dto.TakeDownPublicFamilyRequest, AuditInput) (*vo.FamilyPublicStatus, *apperrors.BusinessError)
	TakeDownUser(context.Context, uint64, uint64, dto.TakeDownPublicFamilyRequest, AuditInput) (*vo.FamilyPublicStatus, *apperrors.BusinessError)
	ListAdmin(context.Context, uint64, string, dto.ListApplicationsQuery) (*vo.ListResult, *apperrors.BusinessError)
	Approve(context.Context, uint64, string, uint64, dto.ReviewPublicApplicationRequest, AuditInput) (*vo.PublicApplication, *apperrors.BusinessError)
	Reject(context.Context, uint64, string, uint64, dto.ReviewPublicApplicationRequest, AuditInput) (*vo.PublicApplication, *apperrors.BusinessError)
	TakeDown(context.Context, uint64, string, uint64, dto.TakeDownPublicFamilyRequest, AuditInput) (*vo.FamilyPublicStatus, *apperrors.BusinessError)
}

func (s *service) QualityEligibility(ctx context.Context, actorID, familyID uint64) (*vo.QualityEligibility, *apperrors.BusinessError) {
	if ok, err := s.permissions.CanManageFamily(ctx, actorID, familyID); err != nil || !ok {
		return nil, publicError(CodePublicApplicationForbidden, "无权查看优质家庭邀请资格")
	}
	family, err := s.repo.FindFamily(ctx, familyID, false)
	if err != nil || family.Status != string(enums.StatusNormal) {
		return nil, publicError(CodePublicFamilyUnavailable, "家庭状态不允许申请优质家庭")
	}
	eligibility, err := s.evaluateQualityEligibility(ctx, s.repo, family)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	eligible := eligibility.Eligible
	if eligible && family.CreatorUserID != nil && s.notifications != nil {
		input := notificationservice.Input{
			UserID: *family.CreatorUserID, Category: "FAMILY", Title: "已获得优质家庭申请资格",
			Content: fmt.Sprintf("“%s”已达到平台优质家庭标准。是否申请成为优质家庭案例，由家庭创建者自行决定。", family.FamilyName),
			Route:   fmt.Sprintf("/pages/family/settings?familyId=%d&section=public", familyID), RelatedType: "QUALITY_FAMILY_ELIGIBILITY", RelatedID: familyID,
			Subscribe: &notificationservice.SubscribeInput{Kind: "AUDIT", Page: fmt.Sprintf("pages/family/settings?familyId=%d", familyID), Values: []string{"优质家庭资格", family.FamilyName, "已获得申请资格", time.Now().Format("2006-01-02 15:04"), "请进入家庭设置提交申请"}},
		}
		if once, ok := s.notifications.(notificationservice.CreateOnceService); ok {
			once.CreateOnce(ctx, input)
		} else {
			s.notifications.Create(ctx, input)
		}
	}
	return eligibility, nil
}

func (s *service) evaluateQualityEligibility(ctx context.Context, repo publicrepo.Repository, family *familymodel.Family) (*vo.QualityEligibility, error) {
	metrics, err := repo.EvaluateQualityEligibility(ctx, family.ID)
	if err != nil {
		return nil, err
	}
	config, err := repo.GetQualityEligibilityConfig(ctx)
	if err != nil {
		return nil, err
	}
	profileRate := percentage(metrics.ProfileCompleteCount, metrics.MemberCount)
	boundRate := percentage(metrics.BoundMemberCount, metrics.MemberCount)
	requirements := make([]string, 0, 6)
	if !config.Enabled {
		requirements = append(requirements, "平台暂未开放优质家庭申请")
	}
	if metrics.MemberCount < int64(config.MinMemberCount) {
		requirements = append(requirements, fmt.Sprintf("至少录入 %d 位活跃成员", config.MinMemberCount))
	}
	if metrics.RelationshipCount < int64(config.MinRelationshipCount) {
		requirements = append(requirements, fmt.Sprintf("至少维护 %d 条有效亲属关系", config.MinRelationshipCount))
	}
	if metrics.BoundMemberCount < int64(config.MinBoundMemberCount) {
		requirements = append(requirements, fmt.Sprintf("至少绑定 %d 位真实用户", config.MinBoundMemberCount))
	}
	if profileRate < config.MinProfileCompleteRate {
		requirements = append(requirements, fmt.Sprintf("节点资料完整度达到 %d%%", config.MinProfileCompleteRate))
	}
	if boundRate < config.MinBoundMemberRate {
		requirements = append(requirements, fmt.Sprintf("节点绑定比例达到 %d%%", config.MinBoundMemberRate))
	}
	if config.RequireDescription && strings.TrimSpace(valueOrEmpty(family.Description)) == "" {
		requirements = append(requirements, "完善家庭简介")
	}
	if config.RequireLocation && strings.TrimSpace(valueOrEmpty(family.RegionText)) == "" && strings.TrimSpace(valueOrEmpty(family.NativePlace)) == "" {
		requirements = append(requirements, "完善家庭所在地")
	}
	descriptionComplete := strings.TrimSpace(valueOrEmpty(family.Description)) != ""
	locationComplete := strings.TrimSpace(valueOrEmpty(family.RegionText)) != "" || strings.TrimSpace(valueOrEmpty(family.NativePlace)) != ""
	checks := []vo.QualityEligibilityCheck{
		{Key: "members", Label: "活跃成员", Current: fmt.Sprintf("%d 位", metrics.MemberCount), Required: fmt.Sprintf("至少 %d 位", config.MinMemberCount), Met: metrics.MemberCount >= int64(config.MinMemberCount)},
		{Key: "relationships", Label: "有效亲属关系", Current: fmt.Sprintf("%d 条", metrics.RelationshipCount), Required: fmt.Sprintf("至少 %d 条", config.MinRelationshipCount), Met: metrics.RelationshipCount >= int64(config.MinRelationshipCount)},
		{Key: "boundMembers", Label: "已绑定用户", Current: fmt.Sprintf("%d 位", metrics.BoundMemberCount), Required: fmt.Sprintf("至少 %d 位", config.MinBoundMemberCount), Met: metrics.BoundMemberCount >= int64(config.MinBoundMemberCount)},
		{Key: "profileRate", Label: "节点资料完整度", Current: fmt.Sprintf("%d%%", profileRate), Required: fmt.Sprintf("至少 %d%%", config.MinProfileCompleteRate), Met: profileRate >= config.MinProfileCompleteRate},
		{Key: "boundRate", Label: "节点绑定比例", Current: fmt.Sprintf("%d%%", boundRate), Required: fmt.Sprintf("至少 %d%%", config.MinBoundMemberRate), Met: boundRate >= config.MinBoundMemberRate},
	}
	if config.RequireDescription {
		checks = append(checks, vo.QualityEligibilityCheck{Key: "description", Label: "家庭简介", Current: map[bool]string{true: "已填写", false: "未填写"}[descriptionComplete], Required: "需要填写", Met: descriptionComplete})
	}
	if config.RequireLocation {
		checks = append(checks, vo.QualityEligibilityCheck{Key: "location", Label: "家庭所在地", Current: map[bool]string{true: "已填写", false: "未填写"}[locationComplete], Required: "需要填写", Met: locationComplete})
	}
	eligible := config.Enabled
	for _, check := range checks {
		eligible = eligible && check.Met
	}
	return &vo.QualityEligibility{
		Eligible:    eligible,
		MemberCount: metrics.MemberCount, RelationshipCount: metrics.RelationshipCount, BoundMemberCount: metrics.BoundMemberCount, ProfileCompleteCount: metrics.ProfileCompleteCount,
		UnknownGenderCount: metrics.UnknownGenderCount, ProfileComplete: metrics.ProfileComplete,
		Requirements: requirements, Checks: checks,
	}, nil
}

func (s *service) GetQualityEligibilityConfig(ctx context.Context, role string) (*vo.QualityEligibilityConfig, *apperrors.BusinessError) {
	if !canManageQualityConfig(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	value, err := s.repo.GetQualityEligibilityConfig(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return qualityConfigVO(value), nil
}

func (s *service) UpdateQualityEligibilityConfig(ctx context.Context, adminID uint64, role string, input vo.QualityEligibilityConfig) (*vo.QualityEligibilityConfig, *apperrors.BusinessError) {
	if !canManageQualityConfig(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	if input.MinMemberCount < 1 || input.MinRelationshipCount < 0 || input.MinBoundMemberCount < 0 || input.MinProfileCompleteRate < 0 || input.MinProfileCompleteRate > 100 || input.MinBoundMemberRate < 0 || input.MinBoundMemberRate > 100 {
		return nil, publicError(CodePublicApplicationForbidden, "优质家庭准入条件不合法")
	}
	if err := s.repo.UpdateQualityEligibilityConfig(ctx, map[string]any{
		"enabled": input.Enabled, "min_member_count": input.MinMemberCount, "min_relationship_count": input.MinRelationshipCount,
		"min_bound_member_count": input.MinBoundMemberCount, "min_profile_complete_rate": input.MinProfileCompleteRate,
		"min_bound_member_rate": input.MinBoundMemberRate, "require_description": input.RequireDescription,
		"require_location": input.RequireLocation, "updated_by_admin_id": adminID,
	}); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	value, err := s.repo.GetQualityEligibilityConfig(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return qualityConfigVO(value), nil
}

func canManageQualityConfig(role string) bool { return role == "ROOT_ADMIN" || role == "SUPER_ADMIN" }
func percentage(part, total int64) int {
	if total <= 0 {
		return 0
	}
	return int(part * 100 / total)
}
func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func qualityConfigVO(value *publicmodel.QualityEligibilityConfig) *vo.QualityEligibilityConfig {
	return &vo.QualityEligibilityConfig{ID: value.ID, Enabled: value.Enabled, MinMemberCount: value.MinMemberCount, MinRelationshipCount: value.MinRelationshipCount, MinBoundMemberCount: value.MinBoundMemberCount, MinProfileCompleteRate: value.MinProfileCompleteRate, MinBoundMemberRate: value.MinBoundMemberRate, RequireDescription: value.RequireDescription, RequireLocation: value.RequireLocation, UpdatedAt: value.UpdatedAt.Format(time.RFC3339)}
}

func (s *service) TakeDownUser(ctx context.Context, actorID uint64, familyID uint64, req dto.TakeDownPublicFamilyRequest, audit AuditInput) (*vo.FamilyPublicStatus, *apperrors.BusinessError) {
	return s.DisableUser(ctx, actorID, familyID, req, audit)
}

func (s *service) EnableUser(ctx context.Context, actorID uint64, familyID uint64, audit AuditInput) (*vo.FamilyPublicStatus, *apperrors.BusinessError) {
	allowed, err := s.permissions.CanManageFamily(ctx, actorID, familyID)
	if err != nil || !allowed {
		return nil, publicError(CodePublicApplicationForbidden, "无权开启优质家庭展示")
	}
	var status *vo.FamilyPublicStatus
	err = s.uow.WithinTransaction(ctx, func(repo publicrepo.Repository) error {
		family, err := repo.FindFamily(ctx, familyID, true)
		if err != nil || family.Status != string(enums.StatusNormal) {
			return errFamilyUnavailable
		}
		if family.PublicDisplayStatus != publicenum.PublicApproved {
			return errTakeDown
		}
		now := s.now()
		if err := repo.UpdateFamilyPublicStatus(ctx, familyID, map[string]any{
			"public_display_enabled": true,
			"public_enabled_at":      now,
		}); err != nil {
			return err
		}
		status = &vo.FamilyPublicStatus{
			FamilyID: familyID, PublicDisplayStatus: publicenum.PublicApproved,
			PublicDisplayEnabled: true, PublicApprovedAt: family.PublicApprovedAt,
			PublicEnabledAt: &now,
		}
		return writeUserFamilyLog(ctx, repo, actorID, familyID, "ENABLE_PUBLIC_DISPLAY", audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return status, nil
}

func (s *service) DisableUser(ctx context.Context, actorID uint64, familyID uint64, req dto.TakeDownPublicFamilyRequest, audit AuditInput) (*vo.FamilyPublicStatus, *apperrors.BusinessError) {
	allowed, err := s.permissions.CanManageFamily(ctx, actorID, familyID)
	if err != nil || !allowed {
		return nil, publicError(CodePublicApplicationForbidden, "无权关闭优质家庭展示")
	}
	var status *vo.FamilyPublicStatus
	err = s.uow.WithinTransaction(ctx, func(repo publicrepo.Repository) error {
		family, err := repo.FindFamily(ctx, familyID, true)
		if err != nil || family.Status != string(enums.StatusNormal) {
			return errFamilyUnavailable
		}
		if family.PublicDisplayStatus != publicenum.PublicApproved {
			return errTakeDown
		}
		if err := repo.UpdateFamilyPublicStatus(ctx, familyID, map[string]any{
			"public_display_enabled": false,
		}); err != nil {
			return err
		}
		status = &vo.FamilyPublicStatus{
			FamilyID: familyID, PublicDisplayStatus: publicenum.PublicApproved,
			PublicDisplayEnabled: false, PublicApprovedAt: family.PublicApprovedAt,
			PublicEnabledAt: family.PublicEnabledAt,
		}
		return writeUserFamilyLog(ctx, repo, actorID, familyID, "DISABLE_PUBLIC_DISPLAY", audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return status, nil
}

type service struct {
	repo          publicrepo.Repository
	uow           publicrepo.UnitOfWork
	permissions   permission.FamilyPermissionService
	contentSafety contentsafety.Service
	now           func() time.Time
	notifications notificationservice.Service
}

func NewService(repo publicrepo.Repository, uow publicrepo.UnitOfWork, permissions permission.FamilyPermissionService, contentSafety contentsafety.Service) Service {
	return newService(repo, uow, permissions, contentSafety, nil)
}

func NewServiceWithNotifications(repo publicrepo.Repository, uow publicrepo.UnitOfWork, permissions permission.FamilyPermissionService, contentSafety contentsafety.Service, notifications notificationservice.Service) Service {
	return newService(repo, uow, permissions, contentSafety, notifications)
}

func newService(repo publicrepo.Repository, uow publicrepo.UnitOfWork, permissions permission.FamilyPermissionService, contentSafety contentsafety.Service, notifications notificationservice.Service) Service {
	if contentSafety == nil {
		contentSafety = contentsafety.FailClosed()
	}
	return &service{repo: repo, uow: uow, permissions: permissions, contentSafety: contentSafety, now: time.Now, notifications: notifications}
}

func (s *service) Submit(ctx context.Context, actorID, familyID uint64, req dto.CreatePublicApplicationRequest, audit AuditInput) (*vo.PublicApplication, *apperrors.BusinessError) {
	if ok, err := s.permissions.CanManageFamily(ctx, actorID, familyID); err != nil || !ok {
		return nil, publicError(CodePublicApplicationForbidden, "无权提交优质家庭申请")
	}
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: actorID, Scene: contentsafety.SceneSocial,
		Fields:   contentsafety.OptionalField("application_reason", req.ApplicationReason),
		FamilyID: &familyID, IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	eligibility, eligibilityErr := s.QualityEligibility(ctx, actorID, familyID)
	if eligibilityErr != nil {
		return nil, eligibilityErr
	}
	if !eligibility.Eligible {
		return nil, publicError(CodePublicApplicationForbidden, "当前家庭尚未获得优质家庭申请资格")
	}
	var applicationID uint64
	err := s.uow.WithinTransaction(ctx, func(repo publicrepo.Repository) error {
		family, err := repo.FindFamily(ctx, familyID, true)
		if err != nil {
			return errFamilyUnavailable
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyUnavailable
		}
		if family.PublicDisplayStatus == publicenum.PublicApproved {
			return errAlreadyApproved
		}
		if _, err := repo.FindPendingByFamily(ctx, familyID); err == nil {
			return errDuplicate
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		now := s.now()
		app := &publicmodel.FamilyPublicApplication{
			FamilyID: familyID, ApplicantUserID: &actorID,
			ApplicationStatus:       publicenum.StatusPending,
			ApplicationReason:       clean(req.ApplicationReason),
			ApplicationSnapshotJSON: snapshotFamily(family),
		}
		if err := repo.Create(ctx, app); err != nil {
			return err
		}
		applicationID = app.ID
		if err := repo.UpdateFamilyPublicStatus(ctx, familyID, map[string]any{
			"public_display_status":  publicenum.PublicPending,
			"public_display_enabled": false,
			"public_applied_at":      now,
		}); err != nil {
			return err
		}
		return writeUserLog(ctx, repo, actorID, familyID, app.ID, "SUBMIT_PUBLIC_APPLICATION", audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return s.result(ctx, applicationID)
}

func (s *service) ListFamily(ctx context.Context, actorID, familyID uint64, req dto.ListApplicationsQuery) (*vo.ListResult, *apperrors.BusinessError) {
	if ok, err := s.permissions.CanManageFamily(ctx, actorID, familyID); err != nil || !ok {
		return nil, publicError(CodePublicApplicationForbidden, "无权查看优质家庭申请")
	}
	req.Page, req.PageSize = normalizePage(req.Page, req.PageSize)
	rows, total, err := s.repo.List(ctx, publicrepo.ListQuery{FamilyID: &familyID, Status: normalizeStatus(req.Status), Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &vo.ListResult{Items: rowsVO(rows), Page: req.Page, PageSize: req.PageSize, Total: total}, nil
}

func (s *service) Cancel(ctx context.Context, actorID, familyID, applicationID uint64, req dto.CancelPublicApplicationRequest, audit AuditInput) (*vo.PublicApplication, *apperrors.BusinessError) {
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: actorID, Scene: contentsafety.SceneSocial,
		Fields:   contentsafety.OptionalField("cancel_reason", req.CancelReason),
		FamilyID: &familyID, IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	err := s.uow.WithinTransaction(ctx, func(repo publicrepo.Repository) error {
		app, err := repo.FindByID(ctx, applicationID, true)
		if err != nil || app.FamilyID != familyID {
			return errNotFound
		}
		if app.ApplicationStatus != publicenum.StatusPending {
			return errInvalidStatus
		}
		allowed, err := s.permissions.CanManageFamily(ctx, actorID, familyID)
		if err != nil || (!allowed && (app.ApplicantUserID == nil || *app.ApplicantUserID != actorID)) {
			return errForbidden
		}
		if _, err := repo.FindFamily(ctx, familyID, true); err != nil {
			return errFamilyUnavailable
		}
		now := s.now()
		if err := repo.UpdateApplication(ctx, applicationID, publicenum.StatusPending, map[string]any{
			"application_status": publicenum.StatusCancelled,
			"cancelled_at":       now,
			"cancel_reason":      clean(req.CancelReason),
		}); err != nil {
			return err
		}
		if err := repo.UpdateFamilyPublicStatus(ctx, familyID, map[string]any{
			"public_display_status":  publicenum.PublicPrivate,
			"public_display_enabled": false,
		}); err != nil {
			return err
		}
		return writeUserLog(ctx, repo, actorID, familyID, applicationID, "CANCEL_PUBLIC_APPLICATION", audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return s.result(ctx, applicationID)
}

func (s *service) ListAdmin(ctx context.Context, adminID uint64, role string, req dto.ListApplicationsQuery) (*vo.ListResult, *apperrors.BusinessError) {
	if !canAdminReview(role) {
		return nil, publicError(CodePublicApplicationForbidden, "无权查看优质家庭申请")
	}
	req.Page, req.PageSize = normalizePage(req.Page, req.PageSize)
	rows, total, err := s.repo.List(ctx, publicrepo.ListQuery{Status: normalizeStatus(req.Status), Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &vo.ListResult{Items: rowsVO(rows), Page: req.Page, PageSize: req.PageSize, Total: total}, nil
}

func (s *service) Approve(ctx context.Context, adminID uint64, role string, applicationID uint64, req dto.ReviewPublicApplicationRequest, audit AuditInput) (*vo.PublicApplication, *apperrors.BusinessError) {
	return s.review(ctx, adminID, role, applicationID, publicenum.StatusApproved, publicenum.PublicApproved, "APPROVE_PUBLIC_APPLICATION", req.ReviewComment, audit)
}

func (s *service) Reject(ctx context.Context, adminID uint64, role string, applicationID uint64, req dto.ReviewPublicApplicationRequest, audit AuditInput) (*vo.PublicApplication, *apperrors.BusinessError) {
	return s.review(ctx, adminID, role, applicationID, publicenum.StatusRejected, publicenum.PublicRejected, "REJECT_PUBLIC_APPLICATION", req.ReviewComment, audit)
}

func (s *service) review(ctx context.Context, adminID uint64, role string, applicationID uint64, appStatus string, familyStatus string, action string, comment *string, audit AuditInput) (*vo.PublicApplication, *apperrors.BusinessError) {
	if !canAdminReview(role) {
		return nil, publicError(CodePublicApplicationForbidden, "无权审核优质家庭申请")
	}
	err := s.uow.WithinTransaction(ctx, func(repo publicrepo.Repository) error {
		app, err := repo.FindByID(ctx, applicationID, true)
		if err != nil {
			return errNotFound
		}
		if app.ApplicationStatus != publicenum.StatusPending {
			return errInvalidStatus
		}
		if role == string(enums.AdminRolePlatformAdmin) && app.ApplicantAdminID != nil && *app.ApplicantAdminID == adminID {
			return errForbidden
		}
		if _, err := repo.FindFamily(ctx, app.FamilyID, true); err != nil {
			return errFamilyUnavailable
		}
		family, err := repo.FindFamily(ctx, app.FamilyID, false)
		if err != nil || family.Status != string(enums.StatusNormal) {
			return errFamilyUnavailable
		}
		if familyStatus == publicenum.PublicApproved {
			eligibility, err := s.evaluateQualityEligibility(ctx, repo, family)
			if err != nil {
				return err
			}
			if !eligibility.Eligible {
				return errEligibilityChanged
			}
		}
		now := s.now()
		if err := repo.UpdateApplication(ctx, applicationID, publicenum.StatusPending, map[string]any{
			"application_status":   appStatus,
			"review_result":        appStatus,
			"reviewed_by_admin_id": adminID,
			"reviewed_at":          now,
			"review_comment":       clean(comment),
		}); err != nil {
			return err
		}
		values := map[string]any{"public_display_status": familyStatus}
		if familyStatus == publicenum.PublicApproved {
			values["public_approved_at"] = now
			values["public_display_enabled"] = true
			values["public_enabled_at"] = now
			values["searchable"] = true
		} else {
			values["public_display_enabled"] = false
		}
		if err := repo.UpdateFamilyPublicStatus(ctx, app.FamilyID, values); err != nil {
			return err
		}
		return writeAdminLog(ctx, repo, adminID, role, app.FamilyID, applicationID, action, audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	result, businessErr := s.result(ctx, applicationID)
	if businessErr == nil && result != nil && s.notifications != nil && result.ApplicantUserID != nil {
		status := "审核通过"
		if appStatus != publicenum.StatusApproved {
			status = "审核拒绝"
		}
		commentText := "无"
		if result.ReviewComment != nil && *result.ReviewComment != "" {
			commentText = *result.ReviewComment
		}
		requestText := "申请成为优质家庭"
		if result.Reason != nil && *result.Reason != "" {
			requestText = *result.Reason
		}
		reviewedAt := time.Now()
		if result.ReviewedAt != nil {
			reviewedAt = *result.ReviewedAt
		}
		s.notifications.Create(ctx, notificationservice.Input{UserID: *result.ApplicantUserID, Category: "AUDIT", Title: "优质家庭申请审核结果", Content: fmt.Sprintf("优质家庭申请%s。", status), Route: fmt.Sprintf("/pages/family/detail?familyId=%d", result.FamilyID), RelatedType: "FAMILY_PUBLIC_APPLICATION", RelatedID: applicationID, Subscribe: &notificationservice.SubscribeInput{Kind: "AUDIT", Page: fmt.Sprintf("pages/family/detail?familyId=%d", result.FamilyID), Values: []string{"优质家庭申请", requestText, status, reviewedAt.Format("2006-01-02 15:04"), commentText}}})
	}
	return result, businessErr
}

func (s *service) TakeDown(ctx context.Context, adminID uint64, role string, familyID uint64, req dto.TakeDownPublicFamilyRequest, audit AuditInput) (*vo.FamilyPublicStatus, *apperrors.BusinessError) {
	if !canAdminReview(role) {
		return nil, publicError(CodePublicApplicationForbidden, "无权移出优质家庭列表")
	}
	var status *vo.FamilyPublicStatus
	err := s.uow.WithinTransaction(ctx, func(repo publicrepo.Repository) error {
		family, err := repo.FindFamily(ctx, familyID, true)
		if err != nil {
			return errFamilyUnavailable
		}
		if family.PublicDisplayStatus != publicenum.PublicApproved {
			return errTakeDown
		}
		now := s.now()
		if err := repo.UpdateFamilyPublicStatus(ctx, familyID, map[string]any{
			"public_display_status":  publicenum.PublicTakenDown,
			"public_display_enabled": false,
			"public_taken_down_at":   now,
		}); err != nil {
			return err
		}
		status = &vo.FamilyPublicStatus{FamilyID: familyID, PublicDisplayStatus: publicenum.PublicTakenDown, PublicDisplayEnabled: false, PublicTakenDownAt: &now}
		return writeAdminLog(ctx, repo, adminID, role, familyID, 0, "TAKE_DOWN_PUBLIC_FAMILY", audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return status, nil
}

func (s *service) result(ctx context.Context, applicationID uint64) (*vo.PublicApplication, *apperrors.BusinessError) {
	app, err := s.repo.FindByID(ctx, applicationID, false)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	row := publicrepo.ApplicationRow{FamilyPublicApplication: *app}
	result := appVO(&row)
	return &result, nil
}

func snapshotFamily(family any) []byte {
	data, _ := json.Marshal(family)
	return data
}

func rowsVO(rows []publicrepo.ApplicationRow) []vo.PublicApplication {
	result := make([]vo.PublicApplication, 0, len(rows))
	for i := range rows {
		result = append(result, appVO(&rows[i]))
	}
	return result
}

func appVO(row *publicrepo.ApplicationRow) vo.PublicApplication {
	return vo.PublicApplication{
		ApplicationID: row.ID, FamilyID: row.FamilyID, FamilyName: row.FamilyName,
		FamilyPublicDisplayStatus: row.FamilyPublicDisplayStatus,
		ApplicantUserID:           row.ApplicantUserID, ApplicantAdminID: row.ApplicantAdminID,
		Status: row.ApplicationStatus, Reason: row.ApplicationReason, ReviewResult: row.ReviewResult,
		ReviewedByAdminID: row.ReviewedByAdminID, ReviewedAt: row.ReviewedAt,
		ReviewComment: row.ReviewComment, CancelledAt: row.CancelledAt,
		CancelReason: row.CancelReason, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func writeUserFamilyLog(ctx context.Context, repo publicrepo.Repository, actorID uint64, familyID uint64, action string, audit AuditInput) error {
	targetType := "FAMILY"
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FAMILY_PUBLIC_APPLICATION", Action: action, TargetType: &targetType,
		TargetID: &familyID, FamilyID: &familyID,
		IP: clean(&audit.IP), UserAgent: clean(&audit.UserAgent),
	})
}

func writeUserLog(ctx context.Context, repo publicrepo.Repository, actorID uint64, familyID uint64, applicationID uint64, action string, audit AuditInput) error {
	targetType := "FAMILY_PUBLIC_APPLICATION"
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FAMILY_PUBLIC_APPLICATION", Action: action, TargetType: &targetType,
		TargetID: &applicationID, FamilyID: &familyID,
		IP: clean(&audit.IP), UserAgent: clean(&audit.UserAgent),
	})
}

func writeAdminLog(ctx context.Context, repo publicrepo.Repository, adminID uint64, role string, familyID uint64, applicationID uint64, action string, audit AuditInput) error {
	targetType := "FAMILY_PUBLIC_APPLICATION"
	if action == "TAKE_DOWN_PUBLIC_FAMILY" {
		targetType = "FAMILY"
		applicationID = familyID
	}
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, OperatorRole: clean(&role),
		Module: "FAMILY_PUBLIC_APPLICATION", Action: action, TargetType: &targetType,
		TargetID: &applicationID, FamilyID: &familyID,
		IP: clean(&audit.IP), UserAgent: clean(&audit.UserAgent),
	})
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func normalizeStatus(status string) string {
	return strings.ToUpper(strings.TrimSpace(status))
}

func canAdminReview(role string) bool {
	switch role {
	case string(enums.AdminRoleRootAdmin), string(enums.AdminRoleSuperAdmin), string(enums.AdminRolePlatformAdmin):
		return true
	default:
		return false
	}
}

func clean(value *string) *string {
	if value == nil {
		return nil
	}
	result := strings.TrimSpace(*value)
	if result == "" {
		return nil
	}
	return &result
}

func publicError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}

var (
	errNotFound           = errors.New("public application not found")
	errInvalidStatus      = errors.New("public application invalid status")
	errDuplicate          = errors.New("public application duplicate")
	errFamilyUnavailable  = errors.New("family unavailable")
	errForbidden          = errors.New("forbidden")
	errAlreadyApproved    = errors.New("family already approved")
	errTakeDown           = errors.New("family cannot take down")
	errEligibilityChanged = errors.New("quality eligibility changed")
)

func mapError(err error) *apperrors.BusinessError {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return publicError(CodePublicApplicationNotFound, "优质家庭申请不存在")
	case errors.Is(err, errInvalidStatus):
		return publicError(CodePublicApplicationInvalidStatus, "优质家庭申请状态不可操作")
	case errors.Is(err, errDuplicate):
		return publicError(CodePublicApplicationDuplicate, "已有待审核优质家庭申请")
	case errors.Is(err, errFamilyUnavailable):
		return publicError(CodePublicFamilyUnavailable, "家庭状态不允许申请优质家庭")
	case errors.Is(err, errForbidden):
		return publicError(CodePublicApplicationForbidden, "无权操作优质家庭申请")
	case errors.Is(err, errAlreadyApproved):
		return publicError(CodePublicFamilyAlreadyApproved, "家庭已公开")
	case errors.Is(err, errTakeDown):
		return publicError(CodePublicFamilyTakeDownDenied, "优质家庭不可移出或当前状态不允许移出")
	case errors.Is(err, errEligibilityChanged):
		return publicError(CodePublicApplicationForbidden, "当前家庭已不满足最新优质家庭标准，请重新核对后再审核")
	default:
		return apperrors.New(apperrors.CodeSystemError)
	}
}
