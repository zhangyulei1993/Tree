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
	dissolutiondto "tree/backend/internal/family/dissolution/dto"
	dissolutionenum "tree/backend/internal/family/dissolution/enum"
	dissolutionrepo "tree/backend/internal/family/dissolution/repository"
	"tree/backend/internal/family/dissolution/vo"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	CodeDissolutionNotFound      apperrors.Code = 46201
	CodeDissolutionInvalidStatus apperrors.Code = 46202
	CodeDissolutionDuplicate     apperrors.Code = 46203
	CodeDissolutionForbidden     apperrors.Code = 46204
	CodeDissolutionAdminDenied   apperrors.Code = 46205
	CodeFamilyRestoreStatus      apperrors.Code = 46206
	CodeFamilyRestoreAdminDenied apperrors.Code = 46207
)

type AuditInput struct {
	IP        string
	UserAgent string
}

type Service interface {
	ListAdmin(context.Context, uint64, string, dissolutiondto.ListDissolutionQuery) (*vo.ListResult, *apperrors.BusinessError)
	Approve(context.Context, uint64, string, uint64, dissolutiondto.ReviewDissolutionRequest, AuditInput) (*vo.DissolutionRequest, *apperrors.BusinessError)
	Reject(context.Context, uint64, string, uint64, dissolutiondto.ReviewDissolutionRequest, AuditInput) (*vo.DissolutionRequest, *apperrors.BusinessError)
	Restore(context.Context, uint64, string, uint64, dissolutiondto.RestoreFamilyRequest, AuditInput) (*vo.RestoreResult, *apperrors.BusinessError)
}

type service struct {
	repo  dissolutionrepo.Repository
	uow   dissolutionrepo.UnitOfWork
	quota quotaservice.Service
	now   func() time.Time
}

func NewService(repo dissolutionrepo.Repository, uow dissolutionrepo.UnitOfWork, quota quotaservice.Service) Service {
	return &service{repo: repo, uow: uow, quota: quota, now: time.Now}
}

func (s *service) ListAdmin(ctx context.Context, adminID uint64, role string, req dissolutiondto.ListDissolutionQuery) (*vo.ListResult, *apperrors.BusinessError) {
	if !canAdminReview(role) {
		return nil, dissolutionError(CodeDissolutionAdminDenied, "后台无权审核解散申请")
	}
	req.Page, req.PageSize = normalizePage(req.Page, req.PageSize)
	rows, total, err := s.repo.List(ctx, dissolutionrepo.ListQuery{Status: normalizeStatus(req.Status), FamilyID: req.FamilyID, Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &vo.ListResult{Items: rowsVO(rows), Page: req.Page, PageSize: req.PageSize, Total: total}, nil
}

func (s *service) Approve(ctx context.Context, adminID uint64, role string, requestID uint64, req dissolutiondto.ReviewDissolutionRequest, audit AuditInput) (*vo.DissolutionRequest, *apperrors.BusinessError) {
	return s.review(ctx, adminID, role, requestID, dissolutionenum.StatusApproved, "APPROVE_DISSOLUTION_REQUEST", req.ReviewComment, audit)
}

func (s *service) Reject(ctx context.Context, adminID uint64, role string, requestID uint64, req dissolutiondto.ReviewDissolutionRequest, audit AuditInput) (*vo.DissolutionRequest, *apperrors.BusinessError) {
	return s.review(ctx, adminID, role, requestID, dissolutionenum.StatusRejected, "REJECT_DISSOLUTION_REQUEST", req.ReviewComment, audit)
}

func (s *service) review(ctx context.Context, adminID uint64, role string, requestID uint64, nextStatus string, action string, comment *string, audit AuditInput) (*vo.DissolutionRequest, *apperrors.BusinessError) {
	if !canAdminReview(role) {
		return nil, dissolutionError(CodeDissolutionAdminDenied, "后台无权审核解散申请")
	}
	err := s.uow.WithinTransaction(ctx, func(repo dissolutionrepo.Repository) error {
		request, err := repo.FindByID(ctx, requestID, true)
		if err != nil {
			return errNotFound
		}
		if request.RequestStatus != dissolutionenum.StatusPending {
			return errInvalidStatus
		}
		family, err := repo.FindFamily(ctx, request.FamilyID, true)
		if err != nil {
			return errNotFound
		}
		if nextStatus == dissolutionenum.StatusApproved && family.Status != "DISSOLUTION_PENDING" {
			return errInvalidStatus
		}
		now := s.now()
		values := map[string]any{
			"request_status":       nextStatus,
			"review_result":        nextStatus,
			"reviewed_by_admin_id": adminID,
			"reviewed_at":          now,
			"review_comment":       clean(comment),
		}
		if nextStatus == dissolutionenum.StatusApproved {
			values["completed_at"] = now
		}
		if err := repo.UpdateRequest(ctx, requestID, dissolutionenum.StatusPending, values); err != nil {
			return err
		}
		familyValues := map[string]any{"status": string(enums.StatusNormal)}
		if nextStatus == dissolutionenum.StatusApproved {
			familyValues = map[string]any{
				"status":                string(enums.StatusDissolved),
				"dissolved_at":          now,
				"searchable":            false,
				"public_display_status": string(enums.StatusPrivate),
			}
		}
		if err := repo.UpdateFamily(ctx, request.FamilyID, familyValues); err != nil {
			return err
		}
		return writeAdminLog(ctx, repo, adminID, role, request.FamilyID, requestID, action, audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return s.result(ctx, requestID)
}

func (s *service) Restore(ctx context.Context, adminID uint64, role string, familyID uint64, req dissolutiondto.RestoreFamilyRequest, audit AuditInput) (*vo.RestoreResult, *apperrors.BusinessError) {
	if !canAdminReview(role) {
		return nil, dissolutionError(CodeFamilyRestoreAdminDenied, "后台无权恢复家庭")
	}
	var result *vo.RestoreResult
	err := s.uow.WithinTransaction(ctx, func(repo dissolutionrepo.Repository) error {
		family, err := repo.FindFamily(ctx, familyID, true)
		if err != nil {
			return errNotFound
		}
		if family.Status != string(enums.StatusDissolved) {
			return errRestoreStatus
		}
		if s.quota != nil {
			if err := s.quota.AssertCanRestoreFamily(ctx, repo.DB(), familyID); err != nil {
				return err
			}
		}
		searchable := true
		if req.Searchable != nil {
			searchable = *req.Searchable
		}
		now := s.now()
		if err := repo.UpdateFamily(ctx, familyID, map[string]any{
			"status":                string(enums.StatusNormal),
			"public_display_status": string(enums.StatusPrivate),
			"searchable":            searchable,
			"restored_at":           now,
		}); err != nil {
			return err
		}
		if err := writeAdminLog(ctx, repo, adminID, role, familyID, familyID, "RESTORE_FAMILY", audit); err != nil {
			return err
		}
		result = &vo.RestoreResult{FamilyID: familyID, Status: string(enums.StatusNormal), PublicDisplayStatus: string(enums.StatusPrivate), Searchable: searchable, GraphVersion: family.GraphVersion, RestoredAt: &now}
		return nil
	})
	if err != nil {
		if s.quota != nil {
			if businessErr := s.quota.MapQuotaError(err); businessErr != nil && businessErr.Code != apperrors.CodeSystemError {
				return nil, businessErr
			}
		}
	}
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return result, nil
}

func (s *service) result(ctx context.Context, requestID uint64) (*vo.DissolutionRequest, *apperrors.BusinessError) {
	request, err := s.repo.FindByID(ctx, requestID, false)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	row := dissolutionrepo.DissolutionRow{FamilyDissolutionRequest: *request}
	result := requestVO(&row)
	return &result, nil
}

func rowsVO(rows []dissolutionrepo.DissolutionRow) []vo.DissolutionRequest {
	result := make([]vo.DissolutionRequest, 0, len(rows))
	for i := range rows {
		result = append(result, requestVO(&rows[i]))
	}
	return result
}

func requestVO(row *dissolutionrepo.DissolutionRow) vo.DissolutionRequest {
	return vo.DissolutionRequest{
		RequestID: row.ID, FamilyID: row.FamilyID, FamilyName: row.FamilyName,
		FamilyStatus:      row.FamilyStatus,
		RequesterMemberID: row.RequesterMemberID, RequesterUserID: row.RequesterUserID,
		RequestStatus: row.RequestStatus, RequestReason: row.RequestReason, ReviewResult: row.ReviewResult,
		ReviewedByAdminID: row.ReviewedByAdminID, ReviewedAt: row.ReviewedAt, ReviewComment: row.ReviewComment,
		CompletedAt: row.CompletedAt, CancelledAt: row.CancelledAt, CancelReason: row.CancelReason,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func writeAdminLog(ctx context.Context, repo dissolutionrepo.Repository, adminID uint64, role string, familyID uint64, targetID uint64, action string, audit AuditInput) error {
	targetType := "FAMILY_DISSOLUTION_REQUEST"
	if action == "RESTORE_FAMILY" {
		targetType = "FAMILY"
	}
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, OperatorRole: clean(&role),
		Module: "FAMILY_DISSOLUTION", Action: action, TargetType: &targetType,
		TargetID: &targetID, FamilyID: &familyID,
		IP: clean(&audit.IP), UserAgent: clean(&audit.UserAgent),
	})
}

func canAdminReview(role string) bool {
	return role == string(enums.AdminRoleRootAdmin) || role == string(enums.AdminRoleSuperAdmin)
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

func dissolutionError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}

var (
	errNotFound      = errors.New("dissolution not found")
	errInvalidStatus = errors.New("dissolution invalid status")
	errRestoreStatus = errors.New("family restore invalid status")
)

func mapError(err error) *apperrors.BusinessError {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return dissolutionError(CodeDissolutionNotFound, "解散申请不存在")
	case errors.Is(err, errInvalidStatus):
		return dissolutionError(CodeDissolutionInvalidStatus, "解散申请状态不可操作")
	case errors.Is(err, errRestoreStatus):
		return dissolutionError(CodeFamilyRestoreStatus, "家庭状态不可恢复")
	default:
		return apperrors.New(apperrors.CodeSystemError)
	}
}
