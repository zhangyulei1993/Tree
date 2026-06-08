package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	roledto "tree/backend/internal/family/role/dto"
	rolerepo "tree/backend/internal/family/role/repository"
	"tree/backend/internal/family/role/vo"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	CodeRoleTargetInvalid  apperrors.Code = 46001
	CodeRoleForbidden      apperrors.Code = 46002
	CodeRoleNoActiveLink   apperrors.Code = 46003
	CodeRoleFounderGuard   apperrors.Code = 46004
	CodeRoleStatusMismatch apperrors.Code = 46005
)

type AuditInput struct {
	IP        string
	UserAgent string
}

type Service interface {
	SetAdmin(context.Context, uint64, uint64, uint64, roledto.RoleChangeRequest, AuditInput) (*vo.RoleChangeResult, *apperrors.BusinessError)
	UnsetAdmin(context.Context, uint64, uint64, uint64, roledto.RoleChangeRequest, AuditInput) (*vo.RoleChangeResult, *apperrors.BusinessError)
}

type service struct {
	repo rolerepo.Repository
	uow  rolerepo.UnitOfWork
	now  func() time.Time
}

func NewService(repo rolerepo.Repository, uow rolerepo.UnitOfWork) Service {
	return &service{repo: repo, uow: uow, now: time.Now}
}

func (s *service) SetAdmin(ctx context.Context, actorID, familyID, memberID uint64, req roledto.RoleChangeRequest, audit AuditInput) (*vo.RoleChangeResult, *apperrors.BusinessError) {
	return s.change(ctx, actorID, familyID, memberID, string(enums.FamilyRoleFamilyAdmin), "SET_FAMILY_ADMIN", req, audit)
}

func (s *service) UnsetAdmin(ctx context.Context, actorID, familyID, memberID uint64, req roledto.RoleChangeRequest, audit AuditInput) (*vo.RoleChangeResult, *apperrors.BusinessError) {
	return s.change(ctx, actorID, familyID, memberID, string(enums.FamilyRoleMember), "UNSET_FAMILY_ADMIN", req, audit)
}

func (s *service) change(ctx context.Context, actorID, familyID, memberID uint64, nextRole string, action string, req roledto.RoleChangeRequest, audit AuditInput) (*vo.RoleChangeResult, *apperrors.BusinessError) {
	var result *vo.RoleChangeResult
	err := s.uow.WithinTransaction(ctx, func(repo rolerepo.Repository) error {
		family, err := repo.FindFamily(ctx, familyID, true)
		if err != nil {
			return errTargetInvalid
		}
		if family.Status != string(enums.StatusNormal) {
			return errTargetInvalid
		}
		actorLink, err := repo.FindActiveLinkByUser(ctx, familyID, actorID, true)
		if err != nil || actorLink.FamilyRole != string(enums.FamilyRoleFounder) {
			return errForbidden
		}
		member, err := repo.FindMember(ctx, familyID, memberID, true)
		if err != nil || member.Status != string(enums.StatusActive) || member.DeletedAt != nil {
			return errTargetInvalid
		}
		if member.UserBindingPolicy == string(enums.UserBindingNotRequired) {
			return errTargetInvalid
		}
		targetLink, err := repo.FindActiveLinkByMember(ctx, familyID, memberID, true)
		if err != nil {
			return errNoActiveLink
		}
		if targetLink.FamilyRole == string(enums.FamilyRoleFounder) {
			return errFounderGuard
		}
		if action == "UNSET_FAMILY_ADMIN" && targetLink.FamilyRole != string(enums.FamilyRoleFamilyAdmin) {
			return errStatusMismatch
		}
		now := s.now()
		values := map[string]any{"family_role": nextRole}
		if action == "SET_FAMILY_ADMIN" {
			values["role_granted_at"] = now
			values["role_granted_by_user_id"] = actorID
		}
		if err := repo.UpdateLinkRole(ctx, targetLink.ID, values); err != nil {
			return err
		}
		if err := writeLog(ctx, repo, actorID, familyID, memberID, targetLink.UserID, action, audit); err != nil {
			return err
		}
		result = &vo.RoleChangeResult{FamilyID: familyID, MemberID: memberID, UserID: targetLink.UserID, FamilyRole: nextRole, GraphVersion: family.GraphVersion, UpdatedAt: now}
		return nil
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return result, nil
}

func writeLog(ctx context.Context, repo rolerepo.Repository, actorID uint64, familyID uint64, memberID uint64, targetUserID uint64, action string, audit AuditInput) error {
	targetType := "FAMILY_MEMBER_USER_LINK"
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FAMILY_ROLE", Action: action, TargetType: &targetType,
		TargetID: &memberID, FamilyID: &familyID, MemberID: &memberID, UserID: &targetUserID,
		IP: clean(&audit.IP), UserAgent: clean(&audit.UserAgent),
	})
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

func roleError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}

var (
	errTargetInvalid  = errors.New("role target invalid")
	errForbidden      = errors.New("role forbidden")
	errNoActiveLink   = errors.New("role no active link")
	errFounderGuard   = errors.New("role founder guard")
	errStatusMismatch = errors.New("role status mismatch")
)

func mapError(err error) *apperrors.BusinessError {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errTargetInvalid):
		return roleError(CodeRoleTargetInvalid, "目标成员不可设置角色")
	case errors.Is(err, errForbidden):
		return roleError(CodeRoleForbidden, "无权管理家庭角色")
	case errors.Is(err, errNoActiveLink), errors.Is(err, gorm.ErrRecordNotFound):
		return roleError(CodeRoleNoActiveLink, "目标成员未绑定用户")
	case errors.Is(err, errFounderGuard):
		return roleError(CodeRoleFounderGuard, "不能操作 FOUNDER 角色")
	case errors.Is(err, errStatusMismatch):
		return roleError(CodeRoleStatusMismatch, "目标角色状态不匹配")
	default:
		return apperrors.New(apperrors.CodeSystemError)
	}
}
