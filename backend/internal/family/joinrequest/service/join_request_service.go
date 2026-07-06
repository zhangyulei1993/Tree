package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	quotaservice "tree/backend/internal/accountquota/service"
	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/permission"
	"tree/backend/internal/family/joinrequest/dto"
	joinenum "tree/backend/internal/family/joinrequest/enum"
	joinmodel "tree/backend/internal/family/joinrequest/model"
	joinrepo "tree/backend/internal/family/joinrequest/repository"
	"tree/backend/internal/family/joinrequest/vo"
	membermodel "tree/backend/internal/family/member/model"
	relationshipmodel "tree/backend/internal/family/relationship/model"
	relationshipservice "tree/backend/internal/family/relationship/service"
	rolemodel "tree/backend/internal/family/role/model"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	CodeJoinRequestNotFound      apperrors.Code = 44101
	CodeJoinRequestInvalidStatus apperrors.Code = 44102
	CodeJoinRequestDuplicate     apperrors.Code = 44103
	CodeJoinRequestForbidden     apperrors.Code = 44104
	CodeApproveModeInvalid       apperrors.Code = 44105
	CodeApplicantUnavailable     apperrors.Code = 44106
	CodeTargetMemberUnavailable  apperrors.Code = 44107
)

type AuditInput struct{ IP, UserAgent string }

type Service interface {
	Create(context.Context, uint64, uint64, dto.CreateJoinRequest, AuditInput) (*vo.JoinRequest, *apperrors.BusinessError)
	ListMine(context.Context, uint64) ([]vo.JoinRequest, *apperrors.BusinessError)
	ListFamily(context.Context, uint64, uint64) ([]vo.JoinRequest, *apperrors.BusinessError)
	Approve(context.Context, uint64, uint64, uint64, dto.ApproveJoinRequest, AuditInput) (*vo.JoinRequest, *apperrors.BusinessError)
	Reject(context.Context, uint64, uint64, uint64, dto.RejectJoinRequest, AuditInput) (*vo.JoinRequest, *apperrors.BusinessError)
	Cancel(context.Context, uint64, uint64, uint64, dto.CancelJoinRequest, AuditInput) (*vo.JoinRequest, *apperrors.BusinessError)
}

type service struct {
	repo          joinrepo.Repository
	uow           joinrepo.UnitOfWork
	permissions   permission.FamilyPermissionService
	quota         quotaservice.Service
	contentSafety contentsafety.Service
	now           func() time.Time
}

func NewService(repo joinrepo.Repository, uow joinrepo.UnitOfWork, permissions permission.FamilyPermissionService, quota quotaservice.Service, contentSafety contentsafety.Service) Service {
	if contentSafety == nil {
		contentSafety = contentsafety.FailClosed()
	}
	return &service{repo: repo, uow: uow, permissions: permissions, quota: quota, contentSafety: contentSafety, now: time.Now}
}

func (s *service) Create(ctx context.Context, actorID, familyID uint64, req dto.CreateJoinRequest, audit AuditInput) (*vo.JoinRequest, *apperrors.BusinessError) {
	applicantGender, genderErr := normalizeApplicantGender(req.ApplicantGender)
	if genderErr != nil {
		return nil, joinError(CodeApproveModeInvalid, "申请人性别必须选择男或女")
	}
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: actorID,
		Scene:  contentsafety.SceneSocial,
		Fields: contentsafety.MergeFields(
			contentsafety.OptionalField("applicant_real_name", req.ApplicantRealName),
			contentsafety.OptionalField("applicant_message", req.ApplicantMessage),
		),
		FamilyID: &familyID,
		IP:       audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	value := &joinmodel.FamilyJoinRequest{
		FamilyID: familyID, ApplicantUserID: actorID,
		ApplicantRealName: clean(req.ApplicantRealName), ApplicantGender: &applicantGender,
		ApplicantMessage: clean(req.ApplicantMessage),
		RequestStatus:    joinenum.StatusPending,
	}
	err := s.uow.WithinTransaction(ctx, func(repo joinrepo.Repository) error {
		family, err := repo.FindFamily(ctx, familyID, true)
		if err != nil || family.Status != string(enums.StatusNormal) {
			return errApplicant
		}
		user, err := repo.FindUser(ctx, actorID, true)
		if err != nil || user.Status != string(enums.StatusActive) {
			return errApplicant
		}
		linked, err := linkExists(repo.FindActiveLinkByUser(ctx, familyID, actorID))
		if err != nil {
			return err
		}
		if linked {
			return errApplicantLinked
		}
		if s.quota != nil {
			if err := s.quota.AssertProfileComplete(ctx, repo.DB(), actorID); err != nil {
				return err
			}
		}
		if _, err := repo.FindPending(ctx, familyID, actorID); err == nil {
			return errDuplicate
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := repo.Create(ctx, value); err != nil {
			return err
		}
		return writeLog(ctx, repo, actorID, familyID, value.ID, "CREATE_JOIN_REQUEST", audit)
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
	result := requestVO(&joinrepo.JoinRequestRow{FamilyJoinRequest: *value})
	return &result, nil
}

func (s *service) ListMine(ctx context.Context, actorID uint64) ([]vo.JoinRequest, *apperrors.BusinessError) {
	rows, err := s.repo.ListMine(ctx, actorID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return rowsVO(rows), nil
}

func (s *service) ListFamily(ctx context.Context, actorID, familyID uint64) ([]vo.JoinRequest, *apperrors.BusinessError) {
	if ok, err := s.permissions.CanManageFamily(ctx, actorID, familyID); err != nil || !ok {
		return nil, joinError(CodeJoinRequestForbidden, "无权处理加入申请")
	}
	rows, err := s.repo.ListFamily(ctx, familyID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return rowsVO(rows), nil
}

func (s *service) Approve(ctx context.Context, actorID, familyID, requestID uint64, req dto.ApproveJoinRequest, audit AuditInput) (*vo.JoinRequest, *apperrors.BusinessError) {
	if ok, err := s.permissions.CanManageFamily(ctx, actorID, familyID); err != nil || !ok {
		return nil, joinError(CodeJoinRequestForbidden, "无权处理加入申请")
	}
	mode := strings.ToUpper(strings.TrimSpace(req.ApproveMode))
	if mode != joinenum.ApproveBindExisting && mode != joinenum.ApproveCreateNew {
		return nil, joinError(CodeApproveModeInvalid, "批准模式不合法")
	}
	if mode == joinenum.ApproveBindExisting && req.Location != nil {
		return nil, joinError(CodeApproveModeInvalid, "绑定已有成员时不能重复定位")
	}
	if mode == joinenum.ApproveCreateNew && req.NewMember != nil {
		if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
			UserID:   actorID,
			Scene:    contentsafety.SceneProfile,
			Fields:   contentsafety.StringField("member_name", req.NewMember.Name),
			FamilyID: &familyID,
			IP:       audit.IP, UserAgent: audit.UserAgent,
		}); businessErr != nil {
			return nil, businessErr
		}
	}
	if req.Location != nil {
		if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
			UserID:   actorID,
			Scene:    contentsafety.SceneSocial,
			Fields:   contentsafety.OptionalField("relation_note", req.Location.Relationship.RelationNote),
			FamilyID: &familyID,
			IP:       audit.IP, UserAgent: audit.UserAgent,
		}); businessErr != nil {
			return nil, businessErr
		}
	}
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: actorID, Scene: contentsafety.SceneSocial,
		Fields: contentsafety.OptionalField("handle_comment", req.HandleComment),
		FamilyID: &familyID, IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	var graphVersion *int64
	var memberID uint64
	var createdMember *membermodel.FamilyMember
	var createdRelationships []*relationshipmodel.FamilyRelationship
	err := s.uow.WithinTransaction(ctx, func(repo joinrepo.Repository) error {
		family, err := repo.FindFamily(ctx, familyID, true)
		if err != nil || family.Status != string(enums.StatusNormal) {
			return errApplicant
		}
		request, err := repo.FindByID(ctx, familyID, requestID, true)
		if err != nil {
			return errRequestMissing
		}
		if request.RequestStatus != joinenum.StatusPending {
			return errInvalidStatus
		}
		user, err := repo.FindUser(ctx, request.ApplicantUserID, true)
		if err != nil || user.Status != string(enums.StatusActive) {
			return errApplicant
		}
		linked, err := linkExists(repo.FindActiveLinkByUser(ctx, familyID, request.ApplicantUserID))
		if err != nil {
			return err
		}
		if linked {
			return errApplicantLinked
		}
		if s.quota != nil {
			if err := s.quota.AssertProfileComplete(ctx, repo.DB(), actorID); err != nil {
				return err
			}
			if err := s.quota.AssertProfileComplete(ctx, repo.DB(), request.ApplicantUserID); err != nil {
				return err
			}
			if err := s.quota.AssertCanJoinFamily(ctx, repo.DB(), request.ApplicantUserID); err != nil {
				return err
			}
		}

		if mode == joinenum.ApproveBindExisting {
			if req.MemberID == nil || *req.MemberID == 0 {
				return errMode
			}
			member, err := repo.FindMember(ctx, familyID, *req.MemberID, true)
			if err != nil || member.UserBindingPolicy == string(enums.UserBindingNotRequired) {
				return errMember
			}
			if request.ApplicantGender == nil ||
				member.Gender != strings.ToUpper(strings.TrimSpace(*request.ApplicantGender)) {
				return errApplicantGenderMismatch
			}
			linked, err := linkExists(repo.FindActiveLinkByMember(ctx, familyID, member.ID))
			if err != nil {
				return err
			}
			if linked {
				return errMember
			}
			memberID = member.ID
		} else {
			if req.NewMember == nil || strings.TrimSpace(req.NewMember.Name) == "" {
				return errMode
			}
			member, err := newMember(familyID, actorID, *req.NewMember)
			if err != nil {
				return errMode
			}
			if request.ApplicantGender != nil && member.Gender != strings.ToUpper(strings.TrimSpace(*request.ApplicantGender)) {
				return errApplicantGenderMismatch
			}
			if s.quota != nil {
				if err := s.quota.AssertCanAddMember(ctx, repo.DB(), familyID, 1); err != nil {
					return err
				}
			}
			if err := repo.CreateMember(ctx, member); err != nil {
				return err
			}
			createdMember = member
			memberID = member.ID
			if req.Location != nil {
				placementRepo, ok := repo.(relationshipservice.PlacementRepository)
				if !ok {
					return errPlacementUnavailable
				}
				createdRelationships, err = relationshipservice.PlaceExistingMember(
					ctx,
					placementRepo,
					familyID,
					family.FamilySurname,
					actorID,
					req.Location.BaseMemberID,
					member,
					req.Location.AddType,
					req.Location.MemberType,
					req.Location.Relationship,
				)
				if err != nil {
					return relationshipservice.MapPlacementError(err)
				}
			}
			version, err := repo.IncrementGraphVersion(ctx, familyID)
			if err != nil {
				return err
			}
			graphVersion = &version
		}
		now := s.now()
		link := &rolemodel.FamilyMemberUserLink{
			FamilyID: familyID, MemberID: memberID, UserID: request.ApplicantUserID,
			LinkStatus: string(enums.StatusActive), LinkSource: "JOIN_REQUEST_APPROVED",
			FamilyRole: string(enums.FamilyRoleMember), JoinRequestID: &request.ID,
			RoleGrantedAt: &now, RoleGrantedByUserID: &actorID,
		}
		if err := repo.CreateLink(ctx, link); err != nil {
			return err
		}
		values := map[string]any{
			"request_status": joinenum.StatusApproved, "approve_mode": mode,
			"bound_member_id": memberID, "handle_result": "APPROVED",
			"handled_by_user_id": actorID, "handled_at": now, "handle_comment": clean(req.HandleComment),
		}
		if mode == joinenum.ApproveCreateNew {
			values["created_member_id"] = memberID
		}
		if err := repo.UpdateStatus(ctx, request.ID, joinenum.StatusPending, values); err != nil {
			return err
		}
		cancelledInvitations, err := repo.CancelPendingInvitations(ctx, familyID, request.ApplicantUserID, memberID, now)
		if err != nil {
			return err
		}
		if cancelledInvitations > 0 {
			if err := writeAutoCancelledInvitationsLog(ctx, repo, actorID, familyID, cancelledInvitations, audit); err != nil {
				return err
			}
		}
		if createdMember != nil {
			if err := writeCreatedMemberLog(ctx, repo, actorID, familyID, createdMember.ID, audit); err != nil {
				return err
			}
		}
		if len(createdRelationships) > 0 {
			if err := writeCreatedRelationshipLog(ctx, repo, actorID, familyID, createdRelationships, audit); err != nil {
				return err
			}
		}
		return writeLog(ctx, repo, actorID, familyID, request.ID, "APPROVE_JOIN_REQUEST", audit)
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
	return s.result(ctx, familyID, requestID, graphVersion)
}

func (s *service) Reject(ctx context.Context, actorID, familyID, requestID uint64, req dto.RejectJoinRequest, audit AuditInput) (*vo.JoinRequest, *apperrors.BusinessError) {
	if ok, err := s.permissions.CanManageFamily(ctx, actorID, familyID); err != nil || !ok {
		return nil, joinError(CodeJoinRequestForbidden, "无权处理加入申请")
	}
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: actorID, Scene: contentsafety.SceneSocial,
		Fields: contentsafety.OptionalField("handle_comment", req.HandleComment),
		FamilyID: &familyID, IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	err := s.uow.WithinTransaction(ctx, func(repo joinrepo.Repository) error {
		value, err := repo.FindByID(ctx, familyID, requestID, true)
		if err != nil {
			return errRequestMissing
		}
		if value.RequestStatus != joinenum.StatusPending {
			return errInvalidStatus
		}
		now := s.now()
		if err := repo.UpdateStatus(ctx, value.ID, joinenum.StatusPending, map[string]any{
			"request_status": joinenum.StatusRejected, "handle_result": "REJECTED",
			"handled_by_user_id": actorID, "handled_at": now, "handle_comment": clean(req.HandleComment),
		}); err != nil {
			return err
		}
		return writeLog(ctx, repo, actorID, familyID, value.ID, "REJECT_JOIN_REQUEST", audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return s.result(ctx, familyID, requestID, nil)
}

func (s *service) Cancel(ctx context.Context, actorID, familyID, requestID uint64, req dto.CancelJoinRequest, audit AuditInput) (*vo.JoinRequest, *apperrors.BusinessError) {
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: actorID, Scene: contentsafety.SceneSocial,
		Fields: contentsafety.OptionalField("cancel_reason", req.CancelReason),
		FamilyID: &familyID, IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	err := s.uow.WithinTransaction(ctx, func(repo joinrepo.Repository) error {
		value, err := repo.FindByID(ctx, familyID, requestID, true)
		if err != nil {
			return errRequestMissing
		}
		if value.ApplicantUserID != actorID {
			return errForbidden
		}
		if value.RequestStatus != joinenum.StatusPending {
			return errInvalidStatus
		}
		now := s.now()
		if err := repo.UpdateStatus(ctx, value.ID, joinenum.StatusPending, map[string]any{
			"request_status": joinenum.StatusCancelled, "cancelled_at": now, "cancel_reason": clean(req.CancelReason),
		}); err != nil {
			return err
		}
		return writeLog(ctx, repo, actorID, familyID, value.ID, "CANCEL_JOIN_REQUEST", audit)
	})
	if businessErr := mapError(err); businessErr != nil {
		return nil, businessErr
	}
	return s.result(ctx, familyID, requestID, nil)
}

func (s *service) result(ctx context.Context, familyID, requestID uint64, version *int64) (*vo.JoinRequest, *apperrors.BusinessError) {
	value, err := s.repo.FindByID(ctx, familyID, requestID, false)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := requestVO(&joinrepo.JoinRequestRow{FamilyJoinRequest: *value})
	result.GraphVersion = version
	return &result, nil
}

func newMember(familyID, actorID uint64, input dto.NewMemberInput) (*membermodel.FamilyMember, error) {
	gender := string(enums.GenderUnknown)
	if input.Gender != nil {
		gender = strings.ToUpper(strings.TrimSpace(*input.Gender))
		if gender != string(enums.GenderMale) && gender != string(enums.GenderFemale) && gender != string(enums.GenderUnknown) {
			return nil, errMode
		}
	}
	policy := string(enums.UserBindingOptional)
	if input.UserBindingPolicy != nil {
		policy = strings.ToUpper(strings.TrimSpace(*input.UserBindingPolicy))
		if policy == string(enums.UserBindingNotRequired) || (policy != string(enums.UserBindingOptional) && policy != string(enums.UserBindingRequired)) {
			return nil, errMode
		}
	}
	birth, err := parseDate(input.BirthDate, input.BirthYear)
	if err != nil {
		return nil, err
	}
	return &membermodel.FamilyMember{
		FamilyID: familyID, MemberType: "LINEAGE_MEMBER", DisplayName: strings.TrimSpace(input.Name),
		Gender: gender, BirthDate: birth, IsLiving: input.IsAlive, UserBindingPolicy: policy,
		Status: string(enums.StatusActive), CreatedByUserID: &actorID,
	}, nil
}

func normalizeApplicantGender(value *string) (string, error) {
	if value == nil {
		return "", errMode
	}
	gender := strings.ToUpper(strings.TrimSpace(*value))
	if gender != string(enums.GenderMale) && gender != string(enums.GenderFemale) {
		return "", errMode
	}
	return gender, nil
}

func parseDate(value *string, year *int) (*time.Time, error) {
	if value != nil && strings.TrimSpace(*value) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*value))
		return &parsed, err
	}
	if year != nil {
		if *year < 1 || *year > 9999 {
			return nil, errMode
		}
		parsed := time.Date(*year, 1, 1, 0, 0, 0, 0, time.UTC)
		return &parsed, nil
	}
	return nil, nil
}

func requestVO(row *joinrepo.JoinRequestRow) vo.JoinRequest {
	return vo.JoinRequest{
		RequestID: row.ID, FamilyID: row.FamilyID, FamilyName: row.FamilyName,
		ApplicantUserID: row.ApplicantUserID, ApplicantRealName: row.ApplicantRealName,
		ApplicantGender: row.ApplicantGender, ApplicantMessage: row.ApplicantMessage, RequestStatus: row.RequestStatus,
		ApproveMode: row.ApproveMode, BoundMemberID: row.BoundMemberID,
		CreatedMemberID: row.CreatedMemberID, HandleComment: row.HandleComment,
		HandledAt: row.HandledAt, CancelledAt: row.CancelledAt,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}
func rowsVO(rows []joinrepo.JoinRequestRow) []vo.JoinRequest {
	result := make([]vo.JoinRequest, 0, len(rows))
	for i := range rows {
		result = append(result, requestVO(&rows[i]))
	}
	return result
}
func writeLog(ctx context.Context, repo joinrepo.Repository, actorID, familyID, requestID uint64, action string, audit AuditInput) error {
	targetType := "FAMILY_JOIN_REQUEST"
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FAMILY_JOIN_REQUEST", Action: action, TargetType: &targetType,
		TargetID: &requestID, FamilyID: &familyID, IP: clean(&audit.IP), UserAgent: clean(&audit.UserAgent),
	})
}

func writeCreatedMemberLog(ctx context.Context, repo joinrepo.Repository, actorID, familyID, memberID uint64, audit AuditInput) error {
	targetType := "FAMILY_MEMBER"
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FAMILY_MEMBER", Action: "CREATE_MEMBER", TargetType: &targetType,
		TargetID: &memberID, FamilyID: &familyID, MemberID: &memberID,
		IP: clean(&audit.IP), UserAgent: clean(&audit.UserAgent),
	})
}

func writeCreatedRelationshipLog(ctx context.Context, repo joinrepo.Repository, actorID, familyID uint64, relationships []*relationshipmodel.FamilyRelationship, audit AuditInput) error {
	targetType := "FAMILY_RELATIONSHIP"
	targetID := relationships[0].ID
	detail, _ := json.Marshal(map[string]any{"relationshipCount": len(relationships), "source": "JOIN_REQUEST_APPROVAL"})
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FAMILY_RELATIONSHIP", Action: "CREATE_RELATIONSHIP", TargetType: &targetType,
		TargetID: &targetID, FamilyID: &familyID, DetailJSON: detail,
		IP: clean(&audit.IP), UserAgent: clean(&audit.UserAgent),
	})
}

func writeAutoCancelledInvitationsLog(ctx context.Context, repo joinrepo.Repository, actorID, familyID uint64, count int64, audit AuditInput) error {
	targetType := "FAMILY"
	detail, _ := json.Marshal(map[string]any{"resolvedCount": count})
	return repo.WriteLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FAMILY_JOIN_REQUEST", Action: "AUTO_CANCEL_INVITATIONS", TargetType: &targetType,
		TargetID: &familyID, FamilyID: &familyID, DetailJSON: detail,
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
func linkExists[T any](value *T, err error) (bool, error) {
	if err == nil {
		return value != nil, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}
func joinError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}

var (
	errRequestMissing          = errors.New("join request missing")
	errInvalidStatus           = errors.New("invalid join request status")
	errDuplicate               = errors.New("duplicate join request")
	errForbidden               = errors.New("forbidden")
	errMode                    = errors.New("invalid approval mode")
	errApplicant               = errors.New("applicant unavailable")
	errApplicantLinked         = errors.New("applicant linked")
	errApplicantGenderMismatch = errors.New("applicant gender mismatch")
	errMember                  = errors.New("member unavailable")
	errPlacementUnavailable    = errors.New("relationship placement unavailable")
)

func mapError(err error) *apperrors.BusinessError {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errRequestMissing), errors.Is(err, gorm.ErrRecordNotFound):
		return joinError(CodeJoinRequestNotFound, "加入申请不存在")
	case errors.Is(err, errInvalidStatus):
		return joinError(CodeJoinRequestInvalidStatus, "加入申请状态不可操作")
	case errors.Is(err, errDuplicate):
		return joinError(CodeJoinRequestDuplicate, "已存在待处理加入申请")
	case errors.Is(err, errForbidden):
		return joinError(CodeJoinRequestForbidden, "无权处理加入申请")
	case errors.Is(err, errMode):
		return joinError(CodeApproveModeInvalid, "批准模式不合法")
	case errors.Is(err, errApplicantGenderMismatch):
		return joinError(CodeApproveModeInvalid, "成员节点性别必须与申请人性别一致")
	case errors.Is(err, errPlacementUnavailable):
		return apperrors.New(apperrors.CodeSystemError)
	case errors.Is(err, errApplicant), errors.Is(err, errApplicantLinked):
		return joinError(CodeApplicantUnavailable, "申请用户不可用或已加入家庭")
	case errors.Is(err, errMember):
		return joinError(CodeTargetMemberUnavailable, "目标成员不可绑定")
	default:
		var businessErr *apperrors.BusinessError
		if errors.As(err, &businessErr) {
			return businessErr
		}
		return apperrors.New(apperrors.CodeSystemError)
	}
}
