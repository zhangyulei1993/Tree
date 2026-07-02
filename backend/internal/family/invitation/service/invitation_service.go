package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/permission"
	"tree/backend/internal/family/invitation/dto"
	inviteenum "tree/backend/internal/family/invitation/enum"
	invitationmodel "tree/backend/internal/family/invitation/model"
	inviterepo "tree/backend/internal/family/invitation/repository"
	"tree/backend/internal/family/invitation/vo"
	rolemodel "tree/backend/internal/family/role/model"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	CodeInvitationNotFound       apperrors.Code = 44001
	CodeInvitationExpired        apperrors.Code = 44002
	CodeInvitationInvalidStatus  apperrors.Code = 44003
	CodeMemberNotInvitable       apperrors.Code = 44004
	CodeMemberAlreadyLinked      apperrors.Code = 44005
	CodeUserAlreadyLinked        apperrors.Code = 44006
	CodeInvitationForbidden      apperrors.Code = 44007
	CodeInvitationTargetMismatch apperrors.Code = 44008
	CodeInvitationDuplicate      apperrors.Code = 44009

	invitationTTL = 7 * 24 * time.Hour
)

type AuditInput struct {
	IP        string
	UserAgent string
}

type Service interface {
	Create(context.Context, uint64, uint64, uint64, dto.CreateInvitationRequest, AuditInput) (*vo.CreatedInvitation, *apperrors.BusinessError)
	Detail(context.Context, string) (*vo.Invitation, *apperrors.BusinessError)
	Accept(context.Context, uint64, uint64, AuditInput) (*vo.Invitation, *apperrors.BusinessError)
	Reject(context.Context, uint64, uint64, dto.RejectInvitationRequest, AuditInput) (*vo.Invitation, *apperrors.BusinessError)
	Cancel(context.Context, uint64, uint64, dto.CancelInvitationRequest, AuditInput) (*vo.Invitation, *apperrors.BusinessError)
	ListMine(context.Context, uint64) ([]vo.Invitation, *apperrors.BusinessError)
	ListFamily(context.Context, uint64, uint64) ([]vo.Invitation, *apperrors.BusinessError)
	Regenerate(context.Context, uint64, uint64, AuditInput) (*vo.CreatedInvitation, *apperrors.BusinessError)
}

type service struct {
	repo        inviterepo.Repository
	uow         inviterepo.UnitOfWork
	permissions permission.FamilyPermissionService
	now         func() time.Time
	token       func() (string, error)
}

func NewService(repo inviterepo.Repository, uow inviterepo.UnitOfWork, permissions permission.FamilyPermissionService) Service {
	return &service{repo: repo, uow: uow, permissions: permissions, now: time.Now, token: randomToken}
}

func (s *service) Create(ctx context.Context, actorID, familyID, memberID uint64, req dto.CreateInvitationRequest, audit AuditInput) (*vo.CreatedInvitation, *apperrors.BusinessError) {
	allowed, err := s.permissions.CanManageFamily(ctx, actorID, familyID)
	if err != nil || !allowed {
		return nil, inviteError(CodeInvitationForbidden, "无权创建邀请")
	}
	channel := strings.ToUpper(strings.TrimSpace(req.InviteChannel))
	if channel != inviteenum.ChannelShareLink && channel != inviteenum.ChannelInApp {
		return nil, inviteError(CodeMemberNotInvitable, "邀请方式不支持")
	}
	if channel == inviteenum.ChannelInApp && (req.TargetUserID == nil || *req.TargetUserID == 0) {
		return nil, inviteError(CodeMemberNotInvitable, "站内邀请必须指定用户")
	}
	if channel == inviteenum.ChannelShareLink && req.TargetUserID != nil {
		return nil, inviteError(CodeMemberNotInvitable, "分享邀请不能指定站内用户")
	}
	if req.FamilyRoleAfterAccept != nil &&
		strings.ToUpper(strings.TrimSpace(*req.FamilyRoleAfterAccept)) != string(enums.FamilyRoleMember) {
		return nil, inviteError(CodeMemberNotInvitable, "邀请接受后的家庭角色只能是 MEMBER")
	}
	rawToken, err := s.token()
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	now := s.now()
	role := string(enums.FamilyRoleMember)
	actorType := inviteenum.ActorFamilyAdmin
	if link, err := s.permissions.GetActiveLink(ctx, actorID, familyID); err == nil && link.FamilyRole == string(enums.FamilyRoleFounder) {
		actorType = inviteenum.ActorFamilyFounder
	}
	invitation := &invitationmodel.FamilyInvitation{
		FamilyID: familyID, TargetMemberID: memberID, InviterUserID: &actorID,
		TargetUserID: req.TargetUserID, InviteType: inviteenum.TypeClaimExistingMember,
		InviteChannel: channel, InviteActorType: actorType, FamilyRoleAfterAccept: role,
		InviteToken: stringPtr(tokenHash(rawToken)), InviteMessage: clean(req.InviteMessage),
		Status: inviteenum.StatusPending, ExpiredAt: now.Add(invitationTTL),
	}
	err = s.uow.WithinTransaction(ctx, func(repo inviterepo.Repository) error {
		family, err := repo.FindFamily(ctx, familyID, true)
		if err != nil || family.Status != string(enums.StatusNormal) {
			return errMemberUnavailable
		}
		member, err := repo.FindMember(ctx, familyID, memberID, true)
		if err != nil || member.UserBindingPolicy == string(enums.UserBindingNotRequired) {
			return errMemberUnavailable
		}
		linked, err := linkExists(repo.FindActiveLinkByMember(ctx, familyID, memberID))
		if err != nil {
			return err
		}
		if linked {
			return errMemberLinked
		}
		if _, err := repo.FindPendingByMember(ctx, familyID, memberID, now); err == nil {
			return errDuplicate
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if req.TargetUserID != nil {
			user, err := repo.FindUser(ctx, *req.TargetUserID, true)
			if err != nil || user.Status != string(enums.StatusActive) || !user.PhoneVerified {
				return errMemberUnavailable
			}
			linked, err := linkExists(repo.FindActiveLinkByUser(ctx, familyID, *req.TargetUserID))
			if err != nil {
				return err
			}
			if linked {
				return errUserLinked
			}
		}
		if err := repo.Create(ctx, invitation); err != nil {
			return err
		}
		return writeLog(ctx, repo, actorID, familyID, memberID, invitation.ID, "CREATE_INVITATION", audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	row, err := s.repo.FindByTokenHash(ctx, tokenHash(rawToken))
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &vo.CreatedInvitation{Invitation: invitationVO(row), InviteToken: rawToken}, nil
}

func (s *service) Detail(ctx context.Context, rawToken string) (*vo.Invitation, *apperrors.BusinessError) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, inviteError(CodeInvitationNotFound, "邀请不存在")
	}
	row, err := s.repo.FindByTokenHash(ctx, tokenHash(rawToken))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, inviteError(CodeInvitationNotFound, "邀请不存在")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := invitationVO(row)
	if row.Status == inviteenum.StatusPending && !row.ExpiredAt.After(s.now()) {
		result.Status = inviteenum.StatusExpired
	}
	return &result, nil
}

func (s *service) Accept(ctx context.Context, actorID, invitationID uint64, audit AuditInput) (*vo.Invitation, *apperrors.BusinessError) {
	current, err := s.repo.FindByID(ctx, invitationID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		businessErr := inviteError(CodeInvitationNotFound, "邀请不存在")
		s.writeAcceptFailureLog(ctx, actorID, nil, invitationID, 0, businessErr, audit)
		return nil, businessErr
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	txErr := s.uow.WithinTransaction(ctx, func(repo inviterepo.Repository) error {
		family, err := repo.FindFamily(ctx, current.FamilyID, true)
		if err != nil || family.Status != string(enums.StatusNormal) {
			return errMemberUnavailable
		}
		invitation, err := repo.FindByID(ctx, invitationID, true)
		if err != nil {
			return errInvitationMissing
		}
		if invitation.FamilyID != current.FamilyID {
			return errInvitationMissing
		}
		if invitation.Status != inviteenum.StatusPending {
			return errInvalidStatus
		}
		now := s.now()
		if !invitation.ExpiredAt.After(now) {
			return errExpired
		}
		if invitation.InviteChannel == inviteenum.ChannelInApp && (invitation.TargetUserID == nil || *invitation.TargetUserID != actorID) {
			return errTargetMismatch
		}
		member, err := repo.FindMember(ctx, invitation.FamilyID, invitation.TargetMemberID, true)
		if err != nil || member.UserBindingPolicy == string(enums.UserBindingNotRequired) {
			return errMemberUnavailable
		}
		user, err := repo.FindUser(ctx, actorID, true)
		if err != nil || user.Status != string(enums.StatusActive) || !user.PhoneVerified {
			return errUserUnavailable
		}
		linked, err := linkExists(repo.FindActiveLinkByMember(ctx, invitation.FamilyID, invitation.TargetMemberID))
		if err != nil {
			return err
		}
		if linked {
			return errMemberLinked
		}
		linked, err = linkExists(repo.FindActiveLinkByUser(ctx, invitation.FamilyID, actorID))
		if err != nil {
			return err
		}
		if linked {
			return errUserLinked
		}
		link := &rolemodel.FamilyMemberUserLink{
			FamilyID: invitation.FamilyID, MemberID: invitation.TargetMemberID, UserID: actorID,
			LinkStatus: string(enums.StatusActive), LinkSource: "INVITATION_ACCEPTED",
			FamilyRole: string(enums.FamilyRoleMember), InvitationID: &invitation.ID, RoleGrantedAt: &now,
		}
		if err := repo.CreateLink(ctx, link); err != nil {
			return err
		}
		if err := repo.UpdateStatus(ctx, invitation.ID, inviteenum.StatusPending, map[string]any{
			"status": inviteenum.StatusAccepted, "accepted_by_user_id": actorID, "accepted_at": now,
		}); err != nil {
			return err
		}
		cancelled, err := repo.CancelPendingJoinRequestsForUser(ctx, invitation.FamilyID, actorID, now)
		if err != nil {
			return err
		}
		if cancelled > 0 {
			if err := writeAutoResolutionLog(ctx, repo, actorID, invitation.FamilyID, "AUTO_CANCEL_JOIN_REQUESTS", cancelled, audit); err != nil {
				return err
			}
		}
		return writeLog(ctx, repo, actorID, invitation.FamilyID, invitation.TargetMemberID, invitation.ID, "ACCEPT_INVITATION", audit)
	})
	if businessErr := mapError(txErr); businessErr != nil {
		s.writeAcceptFailureLog(ctx, actorID, current, invitationID, current.TargetMemberID, businessErr, audit)
		return nil, businessErr
	}
	return s.resultByID(ctx, invitationID)
}

func (s *service) Reject(ctx context.Context, actorID, invitationID uint64, req dto.RejectInvitationRequest, audit AuditInput) (*vo.Invitation, *apperrors.BusinessError) {
	err := s.uow.WithinTransaction(ctx, func(repo inviterepo.Repository) error {
		value, err := repo.FindByID(ctx, invitationID, true)
		if err != nil {
			return errInvitationMissing
		}
		if value.Status != inviteenum.StatusPending {
			return errInvalidStatus
		}
		if !value.ExpiredAt.After(s.now()) {
			return errExpired
		}
		if value.InviteChannel == inviteenum.ChannelInApp && (value.TargetUserID == nil || *value.TargetUserID != actorID) {
			return errTargetMismatch
		}
		now := s.now()
		if err := repo.UpdateStatus(ctx, value.ID, inviteenum.StatusPending, map[string]any{
			"status": inviteenum.StatusRejected, "rejected_by_user_id": actorID,
			"rejected_at": now, "reject_reason": clean(req.Reason),
		}); err != nil {
			return err
		}
		return writeLog(ctx, repo, actorID, value.FamilyID, value.TargetMemberID, value.ID, "REJECT_INVITATION", audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return s.resultByID(ctx, invitationID)
}

func (s *service) Cancel(ctx context.Context, actorID, invitationID uint64, req dto.CancelInvitationRequest, audit AuditInput) (*vo.Invitation, *apperrors.BusinessError) {
	err := s.uow.WithinTransaction(ctx, func(repo inviterepo.Repository) error {
		value, err := repo.FindByID(ctx, invitationID, true)
		if err != nil {
			return errInvitationMissing
		}
		if value.Status != inviteenum.StatusPending {
			return errInvalidStatus
		}
		if !value.ExpiredAt.After(s.now()) {
			return errExpired
		}
		allowed, err := s.permissions.CanManageFamily(ctx, actorID, value.FamilyID)
		if err != nil || (!allowed && (value.InviterUserID == nil || *value.InviterUserID != actorID)) {
			return errForbidden
		}
		now := s.now()
		if err := repo.UpdateStatus(ctx, value.ID, inviteenum.StatusPending, map[string]any{
			"status": inviteenum.StatusCancelled, "cancelled_by_user_id": actorID,
			"cancelled_at": now, "cancel_reason": clean(req.Reason),
		}); err != nil {
			return err
		}
		return writeLog(ctx, repo, actorID, value.FamilyID, value.TargetMemberID, value.ID, "CANCEL_INVITATION", audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return s.resultByID(ctx, invitationID)
}

func (s *service) ListMine(ctx context.Context, actorID uint64) ([]vo.Invitation, *apperrors.BusinessError) {
	rows, err := s.repo.ListForUser(ctx, actorID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := make([]vo.Invitation, 0, len(rows))
	for i := range rows {
		item := invitationVO(&rows[i])
		if item.Status == inviteenum.StatusPending && !item.ExpiredAt.After(s.now()) {
			item.Status = inviteenum.StatusExpired
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *service) ListFamily(ctx context.Context, actorID, familyID uint64) ([]vo.Invitation, *apperrors.BusinessError) {
	allowed, err := s.permissions.CanManageFamily(ctx, actorID, familyID)
	if err != nil || !allowed {
		return nil, inviteError(CodeInvitationForbidden, "无权查看家庭邀请")
	}
	rows, err := s.repo.ListForFamily(ctx, familyID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := make([]vo.Invitation, 0, len(rows))
	for i := range rows {
		item := invitationVO(&rows[i])
		if item.Status == inviteenum.StatusPending && !item.ExpiredAt.After(s.now()) {
			item.Status = inviteenum.StatusExpired
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *service) Regenerate(ctx context.Context, actorID, invitationID uint64, audit AuditInput) (*vo.CreatedInvitation, *apperrors.BusinessError) {
	current, err := s.repo.FindByID(ctx, invitationID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, inviteError(CodeInvitationNotFound, "邀请不存在")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	allowed, err := s.permissions.CanManageFamily(ctx, actorID, current.FamilyID)
	if err != nil || !allowed {
		return nil, inviteError(CodeInvitationForbidden, "无权重新生成邀请")
	}
	if current.InviteChannel != inviteenum.ChannelShareLink {
		return nil, inviteError(CodeMemberNotInvitable, "仅分享邀请可以重新生成")
	}
	rawToken, err := s.token()
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	now := s.now()
	replacement := &invitationmodel.FamilyInvitation{}
	err = s.uow.WithinTransaction(ctx, func(repo inviterepo.Repository) error {
		value, err := repo.FindByID(ctx, invitationID, true)
		if err != nil {
			return errInvitationMissing
		}
		if value.Status != inviteenum.StatusPending {
			return errInvalidStatus
		}
		if err := repo.UpdateStatus(ctx, value.ID, inviteenum.StatusPending, map[string]any{
			"status":               inviteenum.StatusCancelled,
			"cancelled_by_user_id": actorID,
			"cancelled_at":         now,
			"cancel_reason":        "重新生成分享邀请",
		}); err != nil {
			return err
		}
		replacement = &invitationmodel.FamilyInvitation{
			FamilyID: value.FamilyID, TargetMemberID: value.TargetMemberID,
			InviterUserID: &actorID, InviteType: value.InviteType,
			InviteChannel: inviteenum.ChannelShareLink, InviteActorType: value.InviteActorType,
			FamilyRoleAfterAccept: value.FamilyRoleAfterAccept,
			InviteToken:           stringPtr(tokenHash(rawToken)), InviteMessage: value.InviteMessage,
			Status: inviteenum.StatusPending, ExpiredAt: now.Add(invitationTTL),
		}
		if err := repo.Create(ctx, replacement); err != nil {
			return err
		}
		return writeLog(ctx, repo, actorID, value.FamilyID, value.TargetMemberID, replacement.ID, "REGENERATE_INVITATION", audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	row, err := s.repo.FindByTokenHash(ctx, tokenHash(rawToken))
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &vo.CreatedInvitation{Invitation: invitationVO(row), InviteToken: rawToken}, nil
}

func (s *service) resultByID(ctx context.Context, id uint64) (*vo.Invitation, *apperrors.BusinessError) {
	row, err := s.repo.FindRowByID(ctx, id)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := invitationVO(row)
	return &result, nil
}

func invitationVO(row *inviterepo.InvitationRow) vo.Invitation {
	return vo.Invitation{
		InvitationID: row.ID, FamilyID: row.FamilyID, FamilyName: row.FamilyName,
		TargetMemberID: row.TargetMemberID, TargetMemberName: row.TargetMemberName,
		InviterDisplayName: row.InviterDisplayName, InviterRole: row.InviteActorType,
		InviteChannel: row.InviteChannel, InviteMessage: row.InviteMessage,
		FamilyRoleAfterAccept: row.FamilyRoleAfterAccept, Status: row.Status,
		ExpiredAt: row.ExpiredAt, AcceptedAt: row.AcceptedAt, RejectedAt: row.RejectedAt,
		CancelledAt: row.CancelledAt, CreatedAt: row.CreatedAt,
	}
}

func writeLog(ctx context.Context, repo inviterepo.Repository, actorID, familyID, memberID, invitationID uint64, action string, audit AuditInput) error {
	targetType := "FAMILY_INVITATION"
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FAMILY_INVITATION", Action: action, TargetType: &targetType,
		TargetID: &invitationID, FamilyID: &familyID, MemberID: &memberID,
		IP: clean(&audit.IP), UserAgent: clean(&audit.UserAgent),
	})
}

func (s *service) writeAcceptFailureLog(
	ctx context.Context,
	actorID uint64,
	invitation *invitationmodel.FamilyInvitation,
	invitationID uint64,
	targetMemberID uint64,
	businessErr *apperrors.BusinessError,
	audit AuditInput,
) {
	if s == nil || s.repo == nil || businessErr == nil {
		return
	}
	targetType := "FAMILY_INVITATION"
	var familyIDPtr *uint64
	if invitation != nil {
		familyIDPtr = &invitation.FamilyID
		if targetMemberID == 0 {
			targetMemberID = invitation.TargetMemberID
		}
	}
	var memberIDPtr *uint64
	if targetMemberID != 0 {
		memberIDPtr = &targetMemberID
	}
	errMsg := fmt.Sprintf("%d:%s", businessErr.Code, businessErr.Message)
	_ = s.repo.WriteFailedLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FAMILY_INVITATION", Action: "ACCEPT_INVITATION", TargetType: &targetType,
		TargetID: &invitationID, FamilyID: familyIDPtr, MemberID: memberIDPtr,
		ErrorMessage: &errMsg,
		IP:           clean(&audit.IP), UserAgent: clean(&audit.UserAgent),
	})
}

func writeAutoResolutionLog(ctx context.Context, repo inviterepo.Repository, actorID, familyID uint64, action string, count int64, audit AuditInput) error {
	targetType := "FAMILY"
	detail, _ := json.Marshal(map[string]any{"resolvedCount": count})
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FAMILY_INVITATION", Action: action, TargetType: &targetType,
		TargetID: &familyID, FamilyID: &familyID, DetailJSON: detail,
		IP: clean(&audit.IP), UserAgent: clean(&audit.UserAgent),
	})
}

func randomToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func tokenHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func linkExists[T any](value *T, err error) (bool, error) {
	if err == nil {
		return value != nil, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
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

func stringPtr(value string) *string { return &value }
func inviteError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}

var (
	errInvitationMissing = errors.New("invitation missing")
	errExpired           = errors.New("invitation expired")
	errInvalidStatus     = errors.New("invalid invitation status")
	errMemberUnavailable = errors.New("member unavailable")
	errMemberLinked      = errors.New("member linked")
	errUserLinked        = errors.New("user linked")
	errUserUnavailable   = errors.New("user unavailable")
	errTargetMismatch    = errors.New("target mismatch")
	errDuplicate         = errors.New("duplicate invitation")
	errForbidden         = errors.New("forbidden")
)

func mapError(err error) *apperrors.BusinessError {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errInvitationMissing), errors.Is(err, gorm.ErrRecordNotFound):
		return inviteError(CodeInvitationNotFound, "邀请不存在")
	case errors.Is(err, errExpired):
		return inviteError(CodeInvitationExpired, "邀请已过期")
	case errors.Is(err, errInvalidStatus):
		return inviteError(CodeInvitationInvalidStatus, "邀请状态不可操作")
	case errors.Is(err, errMemberUnavailable), errors.Is(err, errUserUnavailable):
		return inviteError(CodeMemberNotInvitable, "成员或目标用户不可邀请")
	case errors.Is(err, errMemberLinked):
		return inviteError(CodeMemberAlreadyLinked, "成员已绑定用户")
	case errors.Is(err, errUserLinked):
		return inviteError(CodeUserAlreadyLinked, "用户已在该家庭绑定成员")
	case errors.Is(err, errTargetMismatch):
		return inviteError(CodeInvitationTargetMismatch, "站内邀请目标用户不匹配")
	case errors.Is(err, errDuplicate):
		return inviteError(CodeInvitationDuplicate, "成员已存在待处理邀请")
	case errors.Is(err, errForbidden):
		return inviteError(CodeInvitationForbidden, "无权操作邀请")
	default:
		return apperrors.New(apperrors.CodeSystemError)
	}
}
