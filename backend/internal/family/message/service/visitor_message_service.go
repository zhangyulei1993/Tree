package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	messagedto "tree/backend/internal/family/message/dto"
	messageenum "tree/backend/internal/family/message/enum"
	messagemodel "tree/backend/internal/family/message/model"
	messagerepo "tree/backend/internal/family/message/repository"
	"tree/backend/internal/family/message/vo"
	publicenum "tree/backend/internal/family/publicdisplay/enum"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	CodeVisitorMessageNotFound       apperrors.Code = 45101
	CodeVisitorMessageInvalidStatus  apperrors.Code = 45102
	CodeVisitorMessageContentEmpty   apperrors.Code = 45103
	CodeVisitorMessageContentTooLong apperrors.Code = 45104
	CodeVisitorMessageRateLimited    apperrors.Code = 45105
	CodeVisitorMessageFamilyPrivate  apperrors.Code = 45106
	CodeVisitorMessageForbidden      apperrors.Code = 45107

	maxMessageContentLength = 1000
	minuteLimitWindow       = time.Minute
	dayLimitWindow          = 24 * time.Hour
)

type AuditInput struct {
	IP        string
	UserAgent string
}

type Service interface {
	CreatePublic(context.Context, uint64, messagedto.CreateVisitorMessageRequest, AuditInput) (*vo.VisitorMessage, *apperrors.BusinessError)
	ListPublic(context.Context, uint64, messagedto.ListMessagesQuery) (*vo.PublicListResult, *apperrors.BusinessError)
	ListAdmin(context.Context, uint64, string, messagedto.ListMessagesQuery) (*vo.ListResult, *apperrors.BusinessError)
	Approve(context.Context, uint64, string, uint64, messagedto.ReviewVisitorMessageRequest, AuditInput) (*vo.VisitorMessage, *apperrors.BusinessError)
	Reject(context.Context, uint64, string, uint64, messagedto.ReviewVisitorMessageRequest, AuditInput) (*vo.VisitorMessage, *apperrors.BusinessError)
	Delete(context.Context, uint64, string, uint64, messagedto.DeleteVisitorMessageRequest, AuditInput) *apperrors.BusinessError
}

type service struct {
	repo messagerepo.Repository
	uow  messagerepo.UnitOfWork
	now  func() time.Time
}

func NewService(repo messagerepo.Repository, uow messagerepo.UnitOfWork) Service {
	return &service{repo: repo, uow: uow, now: time.Now}
}

func (s *service) CreatePublic(ctx context.Context, familyID uint64, req messagedto.CreateVisitorMessageRequest, audit AuditInput) (*vo.VisitorMessage, *apperrors.BusinessError) {
	content := strings.TrimSpace(req.MessageContent)
	if content == "" {
		return nil, messageError(CodeVisitorMessageContentEmpty, "留言内容不能为空")
	}
	if len([]rune(content)) > maxMessageContentLength {
		return nil, messageError(CodeVisitorMessageContentTooLong, "留言内容过长")
	}
	family, err := s.repo.FindFamily(ctx, familyID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, messageError(CodeVisitorMessageFamilyPrivate, "家庭未公开，不能留言")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if !publicFamily(family.Status, family.PublicDisplayStatus) {
		return nil, messageError(CodeVisitorMessageFamilyPrivate, "家庭未公开，不能留言")
	}
	ip := strings.TrimSpace(audit.IP)
	if ip == "" {
		ip = "unknown"
	}
	now := s.now()
	if count, err := s.repo.CountByIP(ctx, familyID, ip, now.Add(-minuteLimitWindow)); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	} else if count >= 1 {
		return nil, messageError(CodeVisitorMessageRateLimited, "游客留言频率过高")
	}
	if count, err := s.repo.CountByIP(ctx, familyID, ip, now.Add(-dayLimitWindow)); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	} else if count >= 20 {
		return nil, messageError(CodeVisitorMessageRateLimited, "游客留言频率过高")
	}
	msg := &messagemodel.VisitorMessage{
		FamilyID: familyID, VisitorName: clean(req.VisitorName),
		VisitorPhone: clean(req.VisitorPhone), VisitorWechat: clean(req.VisitorWechat),
		MessageContent: content, Status: messageenum.StatusPending,
		IP: clean(&ip), UserAgent: clean(&audit.UserAgent),
	}
	if err := s.repo.Create(ctx, msg); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return s.result(ctx, msg.ID)
}

func (s *service) ListPublic(ctx context.Context, familyID uint64, req messagedto.ListMessagesQuery) (*vo.PublicListResult, *apperrors.BusinessError) {
	family, err := s.repo.FindFamily(ctx, familyID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, messageError(CodeVisitorMessageFamilyPrivate, "家庭未公开，不能留言")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if !publicFamily(family.Status, family.PublicDisplayStatus) {
		return nil, messageError(CodeVisitorMessageFamilyPrivate, "家庭未公开，不能留言")
	}
	req.Page, req.PageSize = normalizePage(req.Page, req.PageSize)
	rows, total, err := s.repo.List(ctx, messagerepo.ListQuery{FamilyID: familyID, PublicOnly: true, Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &vo.PublicListResult{Items: publicRowsVO(rows), Page: req.Page, PageSize: req.PageSize, Total: total}, nil
}

func (s *service) ListAdmin(ctx context.Context, adminID uint64, role string, req messagedto.ListMessagesQuery) (*vo.ListResult, *apperrors.BusinessError) {
	if !canAdminHandle(role) {
		return nil, messageError(CodeVisitorMessageForbidden, "无权处理游客留言")
	}
	req.Page, req.PageSize = normalizePage(req.Page, req.PageSize)
	rows, total, err := s.repo.List(ctx, messagerepo.ListQuery{Status: normalizeStatus(req.Status), FamilyID: req.FamilyID, Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &vo.ListResult{Items: rowsVO(rows), Page: req.Page, PageSize: req.PageSize, Total: total}, nil
}

func (s *service) Approve(ctx context.Context, adminID uint64, role string, messageID uint64, req messagedto.ReviewVisitorMessageRequest, audit AuditInput) (*vo.VisitorMessage, *apperrors.BusinessError) {
	resultID, err := s.review(ctx, adminID, role, messageID, messageenum.StatusApproved, req.ReviewComment, "APPROVE_VISITOR_MESSAGE", audit)
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return s.result(ctx, resultID)
}

func (s *service) Reject(ctx context.Context, adminID uint64, role string, messageID uint64, req messagedto.ReviewVisitorMessageRequest, audit AuditInput) (*vo.VisitorMessage, *apperrors.BusinessError) {
	resultID, err := s.review(ctx, adminID, role, messageID, messageenum.StatusRejected, req.ReviewComment, "REJECT_VISITOR_MESSAGE", audit)
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return s.result(ctx, resultID)
}

func (s *service) Delete(ctx context.Context, adminID uint64, role string, messageID uint64, req messagedto.DeleteVisitorMessageRequest, audit AuditInput) *apperrors.BusinessError {
	if !canAdminHandle(role) {
		return messageError(CodeVisitorMessageForbidden, "无权处理游客留言")
	}
	err := s.uow.WithinTransaction(ctx, func(repo messagerepo.Repository) error {
		msg, err := repo.FindByID(ctx, messageID, true)
		if err != nil {
			return errNotFound
		}
		if msg.Status == messageenum.StatusDeleted || msg.DeletedAt != nil {
			return errInvalidStatus
		}
		now := s.now()
		if err := repo.UpdateStatus(ctx, messageID, map[string]any{
			"status":              messageenum.StatusDeleted,
			"deleted_at":          now,
			"deleted_by_admin_id": adminID,
			"delete_reason":       clean(req.DeleteReason),
		}); err != nil {
			return err
		}
		return writeAdminLog(ctx, repo, adminID, role, msg.FamilyID, messageID, "DELETE_VISITOR_MESSAGE", audit)
	})
	return mapError(err)
}

func (s *service) review(ctx context.Context, adminID uint64, role string, messageID uint64, nextStatus string, comment *string, action string, audit AuditInput) (uint64, error) {
	if !canAdminHandle(role) {
		return 0, errForbidden
	}
	err := s.uow.WithinTransaction(ctx, func(repo messagerepo.Repository) error {
		msg, err := repo.FindByID(ctx, messageID, true)
		if err != nil {
			return errNotFound
		}
		if msg.Status == messageenum.StatusDeleted || msg.DeletedAt != nil {
			return errInvalidStatus
		}
		if nextStatus == messageenum.StatusApproved && msg.Status != messageenum.StatusPending && msg.Status != messageenum.StatusRejected {
			return errInvalidStatus
		}
		if nextStatus == messageenum.StatusRejected && msg.Status == messageenum.StatusRejected {
			return errInvalidStatus
		}
		now := s.now()
		if err := repo.UpdateStatus(ctx, messageID, map[string]any{
			"status":               nextStatus,
			"reviewed_by_admin_id": adminID,
			"reviewed_at":          now,
			"review_comment":       clean(comment),
		}); err != nil {
			return err
		}
		return writeAdminLog(ctx, repo, adminID, role, msg.FamilyID, messageID, action, audit)
	})
	return messageID, err
}

func (s *service) result(ctx context.Context, messageID uint64) (*vo.VisitorMessage, *apperrors.BusinessError) {
	msg, err := s.repo.FindByID(ctx, messageID, false)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	row := messagerepo.MessageRow{VisitorMessage: *msg}
	result := messageVO(&row)
	return &result, nil
}

func rowsVO(rows []messagerepo.MessageRow) []vo.VisitorMessage {
	result := make([]vo.VisitorMessage, 0, len(rows))
	for i := range rows {
		result = append(result, messageVO(&rows[i]))
	}
	return result
}

func publicRowsVO(rows []messagerepo.MessageRow) []vo.PublicVisitorMessage {
	result := make([]vo.PublicVisitorMessage, 0, len(rows))
	for i := range rows {
		result = append(result, vo.PublicVisitorMessage{
			MessageID: rows[i].ID, FamilyID: rows[i].FamilyID, VisitorName: rows[i].VisitorName,
			MessageContent: rows[i].MessageContent, CreatedAt: rows[i].CreatedAt, ReviewedAt: rows[i].ReviewedAt,
		})
	}
	return result
}

func messageVO(row *messagerepo.MessageRow) vo.VisitorMessage {
	return vo.VisitorMessage{
		MessageID: row.ID, FamilyID: row.FamilyID, FamilyName: row.FamilyName,
		VisitorName: row.VisitorName, VisitorPhone: row.VisitorPhone, VisitorWechat: row.VisitorWechat,
		MessageContent: row.MessageContent, Status: row.Status,
		ReviewedByAdminID: row.ReviewedByAdminID, ReviewedAt: row.ReviewedAt, ReviewComment: row.ReviewComment,
		DeletedAt: row.DeletedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func publicFamily(status string, displayStatus string) bool {
	return status == string(enums.StatusNormal) && displayStatus == publicenum.PublicApproved
}

func writeAdminLog(ctx context.Context, repo messagerepo.Repository, adminID uint64, role string, familyID uint64, messageID uint64, action string, audit AuditInput) error {
	targetType := "VISITOR_MESSAGE"
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, OperatorRole: clean(&role),
		Module: "VISITOR_MESSAGE", Action: action, TargetType: &targetType,
		TargetID: &messageID, FamilyID: &familyID,
		IP: clean(&audit.IP), UserAgent: clean(&audit.UserAgent),
	})
}

func canAdminHandle(role string) bool {
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

func messageError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}

var (
	errNotFound      = errors.New("visitor message not found")
	errInvalidStatus = errors.New("visitor message invalid status")
	errForbidden     = errors.New("visitor message forbidden")
)

func mapError(err error) *apperrors.BusinessError {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return messageError(CodeVisitorMessageNotFound, "留言不存在")
	case errors.Is(err, errInvalidStatus):
		return messageError(CodeVisitorMessageInvalidStatus, "留言状态不可操作")
	case errors.Is(err, errForbidden):
		return messageError(CodeVisitorMessageForbidden, "无权处理游客留言")
	default:
		return apperrors.New(apperrors.CodeSystemError)
	}
}
