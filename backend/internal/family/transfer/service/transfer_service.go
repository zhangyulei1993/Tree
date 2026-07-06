package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	quotaservice "tree/backend/internal/accountquota/service"
	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	transferdto "tree/backend/internal/family/transfer/dto"
	transferenum "tree/backend/internal/family/transfer/enum"
	transfermodel "tree/backend/internal/family/transfer/model"
	transferrepo "tree/backend/internal/family/transfer/repository"
	"tree/backend/internal/family/transfer/vo"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	CodeTransferNotFound      apperrors.Code = 46101
	CodeTransferInvalidStatus apperrors.Code = 46102
	CodeTransferDuplicate     apperrors.Code = 46103
	CodeTransferForbidden     apperrors.Code = 46104
	CodeTransferTargetInvalid apperrors.Code = 46105
	CodeTransferAdminDenied   apperrors.Code = 46106
)

type AuditInput struct {
	IP        string
	UserAgent string
}

type Service interface {
	Create(context.Context, uint64, uint64, transferdto.CreateTransferRequest, AuditInput) (*vo.TransferRequest, *apperrors.BusinessError)
	Current(context.Context, uint64, uint64) (*vo.TransferRequest, *apperrors.BusinessError)
	Cancel(context.Context, uint64, uint64, uint64, transferdto.CancelTransferRequest, AuditInput) (*vo.TransferRequest, *apperrors.BusinessError)
	ListAdmin(context.Context, uint64, string, transferdto.ListTransferQuery) (*vo.ListResult, *apperrors.BusinessError)
	Approve(context.Context, uint64, string, uint64, transferdto.ReviewTransferRequest, AuditInput) (*vo.TransferRequest, *apperrors.BusinessError)
	Reject(context.Context, uint64, string, uint64, transferdto.ReviewTransferRequest, AuditInput) (*vo.TransferRequest, *apperrors.BusinessError)
}

type service struct {
	repo          transferrepo.Repository
	uow           transferrepo.UnitOfWork
	quota         quotaservice.Service
	contentSafety contentsafety.Service
	now           func() time.Time
}

func NewService(repo transferrepo.Repository, uow transferrepo.UnitOfWork, quota quotaservice.Service, contentSafety contentsafety.Service) Service {
	if contentSafety == nil {
		contentSafety = contentsafety.FailClosed()
	}
	return &service{repo: repo, uow: uow, quota: quota, contentSafety: contentSafety, now: time.Now}
}

func (s *service) Create(ctx context.Context, actorID uint64, familyID uint64, req transferdto.CreateTransferRequest, audit AuditInput) (*vo.TransferRequest, *apperrors.BusinessError) {
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: actorID, Scene: contentsafety.SceneSocial,
		Fields: contentsafety.OptionalField("request_reason", req.RequestReason),
		FamilyID: &familyID, IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	var requestID uint64
	err := s.uow.WithinTransaction(ctx, func(repo transferrepo.Repository) error {
		family, err := repo.FindFamily(ctx, familyID, true)
		if err != nil || family.Status != string(enums.StatusNormal) {
			return errTargetInvalid
		}
		fromLink, err := repo.FindActiveLinkByUser(ctx, familyID, actorID, true)
		if err != nil || fromLink.FamilyRole != string(enums.FamilyRoleFounder) {
			return errForbidden
		}
		toMember, err := repo.FindMember(ctx, familyID, req.ToMemberID, true)
		if err != nil || toMember.Status != string(enums.StatusActive) || toMember.DeletedAt != nil ||
			toMember.UserBindingPolicy == string(enums.UserBindingNotRequired) {
			return errTargetInvalid
		}
		toLink, err := repo.FindActiveLinkByMember(ctx, familyID, req.ToMemberID, true)
		if err != nil || toLink.FamilyRole == string(enums.FamilyRoleFounder) {
			return errTargetInvalid
		}
		if _, err := repo.FindPendingByFamily(ctx, familyID); err == nil {
			return errDuplicate
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if s.quota != nil {
			if err := s.quota.AssertProfileComplete(ctx, repo.DB(), actorID); err != nil {
				return err
			}
			if err := s.quota.AssertFounderTransferReceiver(ctx, repo.DB(), toLink.UserID, familyID); err != nil {
				return err
			}
		}
		request := &transfermodel.FamilyFounderTransferRequest{
			FamilyID: familyID, FromMemberID: fromLink.MemberID, FromUserID: actorID,
			ToMemberID: req.ToMemberID, ToUserID: toLink.UserID,
			RequestStatus: transferenum.StatusPending, RequestReason: clean(req.RequestReason),
		}
		if err := repo.Create(ctx, request); err != nil {
			return err
		}
		requestID = request.ID
		return writeUserLog(ctx, repo, actorID, familyID, requestID, fromLink.MemberID, "CREATE_FOUNDER_TRANSFER_REQUEST", audit)
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
	return s.result(ctx, requestID)
}

func (s *service) Current(ctx context.Context, actorID uint64, familyID uint64) (*vo.TransferRequest, *apperrors.BusinessError) {
	link, err := s.repo.FindActiveLinkByUser(ctx, familyID, actorID, false)
	if err != nil || link.LinkStatus != string(enums.StatusActive) {
		return nil, transferError(CodeTransferForbidden, "无权查看创始人转让申请")
	}
	request, err := s.repo.FindPendingByFamily(ctx, familyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, transferError(CodeTransferNotFound, "当前没有待审核创始人转让申请")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	row := transferrepo.TransferRow{FamilyFounderTransferRequest: *request}
	result := requestVO(&row)
	return &result, nil
}

func (s *service) Cancel(ctx context.Context, actorID uint64, familyID uint64, requestID uint64, req transferdto.CancelTransferRequest, audit AuditInput) (*vo.TransferRequest, *apperrors.BusinessError) {
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: actorID, Scene: contentsafety.SceneSocial,
		Fields: contentsafety.OptionalField("cancel_reason", req.CancelReason),
		FamilyID: &familyID, IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	err := s.uow.WithinTransaction(ctx, func(repo transferrepo.Repository) error {
		request, err := repo.FindByID(ctx, requestID, true)
		if err != nil || request.FamilyID != familyID {
			return errNotFound
		}
		if request.RequestStatus != transferenum.StatusPending {
			return errInvalidStatus
		}
		link, err := repo.FindActiveLinkByUser(ctx, familyID, actorID, true)
		if err != nil || (request.FromUserID != actorID && link.FamilyRole != string(enums.FamilyRoleFounder)) {
			return errForbidden
		}
		now := s.now()
		if err := repo.UpdateRequest(ctx, requestID, transferenum.StatusPending, map[string]any{
			"request_status": transferenum.StatusCancelled,
			"cancelled_at":   now,
			"cancel_reason":  clean(req.CancelReason),
		}); err != nil {
			return err
		}
		return writeUserLog(ctx, repo, actorID, familyID, requestID, request.FromMemberID, "CANCEL_FOUNDER_TRANSFER_REQUEST", audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return s.result(ctx, requestID)
}

func (s *service) ListAdmin(ctx context.Context, adminID uint64, role string, req transferdto.ListTransferQuery) (*vo.ListResult, *apperrors.BusinessError) {
	if !canAdminReview(role) {
		return nil, transferError(CodeTransferAdminDenied, "后台无权审核创始人转让")
	}
	req.Page, req.PageSize = normalizePage(req.Page, req.PageSize)
	rows, total, err := s.repo.List(ctx, transferrepo.ListQuery{Status: normalizeStatus(req.Status), FamilyID: req.FamilyID, Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &vo.ListResult{Items: rowsVO(rows), Page: req.Page, PageSize: req.PageSize, Total: total}, nil
}

func (s *service) Approve(ctx context.Context, adminID uint64, role string, requestID uint64, req transferdto.ReviewTransferRequest, audit AuditInput) (*vo.TransferRequest, *apperrors.BusinessError) {
	return s.review(ctx, adminID, role, requestID, transferenum.StatusApproved, "APPROVE_FOUNDER_TRANSFER_REQUEST", req.ReviewComment, audit)
}

func (s *service) Reject(ctx context.Context, adminID uint64, role string, requestID uint64, req transferdto.ReviewTransferRequest, audit AuditInput) (*vo.TransferRequest, *apperrors.BusinessError) {
	return s.review(ctx, adminID, role, requestID, transferenum.StatusRejected, "REJECT_FOUNDER_TRANSFER_REQUEST", req.ReviewComment, audit)
}

func (s *service) review(ctx context.Context, adminID uint64, role string, requestID uint64, nextStatus string, action string, comment *string, audit AuditInput) (*vo.TransferRequest, *apperrors.BusinessError) {
	if !canAdminReview(role) {
		return nil, transferError(CodeTransferAdminDenied, "后台无权审核创始人转让")
	}
	err := s.uow.WithinTransaction(ctx, func(repo transferrepo.Repository) error {
		request, err := repo.FindByID(ctx, requestID, true)
		if err != nil {
			return errNotFound
		}
		if request.RequestStatus != transferenum.StatusPending {
			return errInvalidStatus
		}
		family, err := repo.FindFamily(ctx, request.FamilyID, true)
		if err != nil || family.Status != string(enums.StatusNormal) {
			return errTargetInvalid
		}
		now := s.now()
		if nextStatus == transferenum.StatusApproved {
			fromLink, err := repo.FindActiveLinkByMember(ctx, request.FamilyID, request.FromMemberID, true)
			if err != nil || fromLink.FamilyRole != string(enums.FamilyRoleFounder) {
				return errInvalidStatus
			}
			toMember, err := repo.FindMember(ctx, request.FamilyID, request.ToMemberID, true)
			if err != nil || toMember.Status != string(enums.StatusActive) || toMember.UserBindingPolicy == string(enums.UserBindingNotRequired) || toMember.DeletedAt != nil {
				return errTargetInvalid
			}
			toLink, err := repo.FindActiveLinkByMember(ctx, request.FamilyID, request.ToMemberID, true)
			if err != nil {
				return errTargetInvalid
			}
			if s.quota != nil {
				if err := s.quota.AssertPostFounderTransfer(ctx, repo.DB(), request.FromUserID, toLink.UserID, request.FamilyID); err != nil {
					return err
				}
			}
			if err := repo.UpdateLinkRole(ctx, fromLink.ID, map[string]any{"family_role": string(enums.FamilyRoleMember)}); err != nil {
				return err
			}
			if err := repo.UpdateLinkRole(ctx, toLink.ID, map[string]any{
				"family_role":              string(enums.FamilyRoleFounder),
				"role_granted_at":          now,
				"role_granted_by_admin_id": adminID,
			}); err != nil {
				return err
			}
			if err := repo.UpdateFamily(ctx, request.FamilyID, map[string]any{"current_founder_member_id": request.ToMemberID}); err != nil {
				return err
			}
			count, err := repo.CountActiveFounders(ctx, request.FamilyID)
			if err != nil || count != 1 {
				return errFounderCount
			}
		}
		values := map[string]any{
			"request_status":       nextStatus,
			"review_result":        nextStatus,
			"reviewed_by_admin_id": adminID,
			"reviewed_at":          now,
			"review_comment":       clean(comment),
		}
		if nextStatus == transferenum.StatusApproved {
			values["completed_at"] = now
		}
		if err := repo.UpdateRequest(ctx, requestID, transferenum.StatusPending, values); err != nil {
			return err
		}
		return writeAdminLog(ctx, repo, adminID, role, request.FamilyID, requestID, action, audit)
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
	return s.result(ctx, requestID)
}

func (s *service) result(ctx context.Context, requestID uint64) (*vo.TransferRequest, *apperrors.BusinessError) {
	request, err := s.repo.FindByID(ctx, requestID, false)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	row := transferrepo.TransferRow{FamilyFounderTransferRequest: *request}
	result := requestVO(&row)
	return &result, nil
}

func rowsVO(rows []transferrepo.TransferRow) []vo.TransferRequest {
	result := make([]vo.TransferRequest, 0, len(rows))
	for i := range rows {
		result = append(result, requestVO(&rows[i]))
	}
	return result
}

func requestVO(row *transferrepo.TransferRow) vo.TransferRequest {
	return vo.TransferRequest{
		RequestID: row.ID, FamilyID: row.FamilyID, FamilyName: row.FamilyName,
		FromMemberID: row.FromMemberID, FromUserID: row.FromUserID, ToMemberID: row.ToMemberID, ToUserID: row.ToUserID,
		RequestStatus: row.RequestStatus, RequestReason: row.RequestReason, ReviewResult: row.ReviewResult,
		ReviewedByAdminID: row.ReviewedByAdminID, ReviewedAt: row.ReviewedAt, ReviewComment: row.ReviewComment,
		CompletedAt: row.CompletedAt, CancelledAt: row.CancelledAt, CancelReason: row.CancelReason,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func writeUserLog(ctx context.Context, repo transferrepo.Repository, actorID uint64, familyID uint64, requestID uint64, memberID uint64, action string, audit AuditInput) error {
	targetType := "FOUNDER_TRANSFER_REQUEST"
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FOUNDER_TRANSFER", Action: action, TargetType: &targetType,
		TargetID: &requestID, FamilyID: &familyID, MemberID: &memberID,
		IP: clean(&audit.IP), UserAgent: clean(&audit.UserAgent),
	})
}

func writeAdminLog(ctx context.Context, repo transferrepo.Repository, adminID uint64, role string, familyID uint64, requestID uint64, action string, audit AuditInput) error {
	targetType := "FOUNDER_TRANSFER_REQUEST"
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, OperatorRole: clean(&role),
		Module: "FOUNDER_TRANSFER", Action: action, TargetType: &targetType,
		TargetID: &requestID, FamilyID: &familyID,
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

func transferError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}

var (
	errNotFound      = errors.New("transfer not found")
	errInvalidStatus = errors.New("transfer invalid status")
	errDuplicate     = errors.New("transfer duplicate")
	errForbidden     = errors.New("transfer forbidden")
	errTargetInvalid = errors.New("transfer target invalid")
	errFounderCount  = errors.New("invalid founder count")
)

func mapError(err error) *apperrors.BusinessError {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return transferError(CodeTransferNotFound, "创始人转让申请不存在")
	case errors.Is(err, errInvalidStatus):
		return transferError(CodeTransferInvalidStatus, "创始人转让申请状态不可操作")
	case errors.Is(err, errDuplicate):
		return transferError(CodeTransferDuplicate, "已有待审核创始人转让申请")
	case errors.Is(err, errForbidden):
		return transferError(CodeTransferForbidden, "无权发起创始人转让")
	case errors.Is(err, errTargetInvalid), errors.Is(err, errFounderCount):
		return transferError(CodeTransferTargetInvalid, "目标成员不可成为创始人")
	default:
		return apperrors.New(apperrors.CodeSystemError)
	}
}
