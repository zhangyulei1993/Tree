package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/permission"
	"tree/backend/internal/family/publicdisplay/dto"
	publicenum "tree/backend/internal/family/publicdisplay/enum"
	publicmodel "tree/backend/internal/family/publicdisplay/model"
	publicrepo "tree/backend/internal/family/publicdisplay/repository"
	"tree/backend/internal/family/publicdisplay/vo"
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

func (s *service) TakeDownUser(ctx context.Context, actorID uint64, familyID uint64, req dto.TakeDownPublicFamilyRequest, audit AuditInput) (*vo.FamilyPublicStatus, *apperrors.BusinessError) {
	return s.DisableUser(ctx, actorID, familyID, req, audit)
}

func (s *service) EnableUser(ctx context.Context, actorID uint64, familyID uint64, audit AuditInput) (*vo.FamilyPublicStatus, *apperrors.BusinessError) {
	allowed, err := s.permissions.CanManageFamily(ctx, actorID, familyID)
	if err != nil || !allowed {
		return nil, publicError(CodePublicApplicationForbidden, "无权开启家庭公开展示")
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
		return nil, publicError(CodePublicApplicationForbidden, "无权关闭家庭公开展示")
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
}

func NewService(repo publicrepo.Repository, uow publicrepo.UnitOfWork, permissions permission.FamilyPermissionService, contentSafety contentsafety.Service) Service {
	if contentSafety == nil {
		contentSafety = contentsafety.FailClosed()
	}
	return &service{repo: repo, uow: uow, permissions: permissions, contentSafety: contentSafety, now: time.Now}
}

func (s *service) Submit(ctx context.Context, actorID, familyID uint64, req dto.CreatePublicApplicationRequest, audit AuditInput) (*vo.PublicApplication, *apperrors.BusinessError) {
	if ok, err := s.permissions.CanManageFamily(ctx, actorID, familyID); err != nil || !ok {
		return nil, publicError(CodePublicApplicationForbidden, "无权提交公开申请")
	}
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: actorID, Scene: contentsafety.SceneSocial,
		Fields:   contentsafety.OptionalField("application_reason", req.ApplicationReason),
		FamilyID: &familyID, IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
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
		return nil, publicError(CodePublicApplicationForbidden, "无权查看公开申请")
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
		return nil, publicError(CodePublicApplicationForbidden, "无权查看公开申请")
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
		return nil, publicError(CodePublicApplicationForbidden, "无权审核公开申请")
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
			values["public_display_enabled"] = false
			values["public_enabled_at"] = nil
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
	return s.result(ctx, applicationID)
}

func (s *service) TakeDown(ctx context.Context, adminID uint64, role string, familyID uint64, req dto.TakeDownPublicFamilyRequest, audit AuditInput) (*vo.FamilyPublicStatus, *apperrors.BusinessError) {
	if !canAdminReview(role) {
		return nil, publicError(CodePublicApplicationForbidden, "无权下架公开家庭")
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
	errNotFound          = errors.New("public application not found")
	errInvalidStatus     = errors.New("public application invalid status")
	errDuplicate         = errors.New("public application duplicate")
	errFamilyUnavailable = errors.New("family unavailable")
	errForbidden         = errors.New("forbidden")
	errAlreadyApproved   = errors.New("family already approved")
	errTakeDown          = errors.New("family cannot take down")
)

func mapError(err error) *apperrors.BusinessError {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return publicError(CodePublicApplicationNotFound, "公开申请不存在")
	case errors.Is(err, errInvalidStatus):
		return publicError(CodePublicApplicationInvalidStatus, "公开申请状态不可操作")
	case errors.Is(err, errDuplicate):
		return publicError(CodePublicApplicationDuplicate, "已有待审核公开申请")
	case errors.Is(err, errFamilyUnavailable):
		return publicError(CodePublicFamilyUnavailable, "家庭状态不允许公开申请")
	case errors.Is(err, errForbidden):
		return publicError(CodePublicApplicationForbidden, "无权操作公开申请")
	case errors.Is(err, errAlreadyApproved):
		return publicError(CodePublicFamilyAlreadyApproved, "家庭已公开")
	case errors.Is(err, errTakeDown):
		return publicError(CodePublicFamilyTakeDownDenied, "公开家庭不可下架或状态不允许下架")
	default:
		return apperrors.New(apperrors.CodeSystemError)
	}
}
