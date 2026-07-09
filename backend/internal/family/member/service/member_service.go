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
	"tree/backend/internal/family/member/dto"
	membermodel "tree/backend/internal/family/member/model"
	memberrepo "tree/backend/internal/family/member/repository"
	"tree/backend/internal/family/member/vo"
	rolemodel "tree/backend/internal/family/role/model"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	CodeMemberFamilyUnavailable apperrors.Code = 43002
	CodeMemberCreateForbidden   apperrors.Code = 43003
	CodeMemberNameRequired      apperrors.Code = 43004
	CodeMemberInvalidField      apperrors.Code = 43005
	CodeMemberNotFound          apperrors.Code = 43101
	CodeMemberEditForbidden     apperrors.Code = 43103
	CodeMemberViewForbidden     apperrors.Code = 43104
	CodeMemberHasRelationships  apperrors.Code = 43201
	CodeMemberFounderProtected  apperrors.Code = 43203
	CodeMemberDeleteForbidden   apperrors.Code = 43204
	CodeMemberBindingConflict   apperrors.Code = 43401
	CodeMemberUserUnavailable   apperrors.Code = 43403
	CodeMemberAlreadyBound      apperrors.Code = 43404
	CodeMemberUserAlreadyBound  apperrors.Code = 43405
	CodeMemberNotBound          apperrors.Code = 43406
	CodeMemberUnbindProtected   apperrors.Code = 43407

	profileNoteType = "M8_PROFILE_JSON"
)

var runMemberTransaction = func(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(fn)
}

type profileNote struct {
	Description *string `json:"description,omitempty"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
}

type AuditInput struct {
	IP        string
	UserAgent string
}

type MemberService interface {
	Create(context.Context, uint64, uint64, dto.CreateMemberRequest, AuditInput) (*vo.Member, *apperrors.BusinessError)
	List(context.Context, uint64, uint64) ([]vo.Member, *apperrors.BusinessError)
	Detail(context.Context, uint64, uint64, uint64) (*vo.Member, *apperrors.BusinessError)
	Update(context.Context, uint64, uint64, uint64, dto.UpdateMemberRequest, AuditInput) (*vo.Member, *apperrors.BusinessError)
	Delete(context.Context, uint64, uint64, uint64, dto.DeleteMemberRequest, AuditInput) *apperrors.BusinessError
	BindUser(context.Context, uint64, uint64, uint64, dto.BindUserRequest, AuditInput) (*vo.Member, *apperrors.BusinessError)
	UnbindUser(context.Context, uint64, uint64, uint64, dto.UnbindUserRequest, AuditInput) (*vo.Member, *apperrors.BusinessError)
}

type memberService struct {
	db            *gorm.DB
	repo          memberrepo.MemberRepository
	permissions   permission.FamilyPermissionService
	quota         quotaservice.Service
	contentSafety contentsafety.Service
}

func NewMemberService(db *gorm.DB, repo memberrepo.MemberRepository, permissions permission.FamilyPermissionService, quota quotaservice.Service, contentSafety contentsafety.Service) MemberService {
	if contentSafety == nil {
		contentSafety = contentsafety.FailClosed()
	}
	return &memberService{db: db, repo: repo, permissions: permissions, quota: quota, contentSafety: contentSafety}
}

func (s *memberService) Create(ctx context.Context, actorID uint64, familyID uint64, req dto.CreateMemberRequest, audit AuditInput) (*vo.Member, *apperrors.BusinessError) {
	if allowed, err := s.permissions.CanCreateMember(ctx, actorID, familyID); err != nil || !allowed {
		return nil, memberError(CodeMemberCreateForbidden, "无权创建家庭成员")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, memberError(CodeMemberNameRequired, "成员姓名不能为空")
	}
	values, businessErr := memberValues(req.Gender, req.BirthDate, req.BirthYear, req.DeathDate, req.DeathYear, req.IsAlive, req.AvatarURL, req.Description, req.UserBindingPolicy)
	if businessErr != nil {
		return nil, businessErr
	}
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: actorID,
		Scene:  contentsafety.SceneProfile,
		Fields: contentsafety.MergeFields(
			contentsafety.StringField("member_name", name),
			contentsafety.OptionalField("member_description", req.Description),
		),
		FamilyID: &familyID,
		IP:       audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	member := &membermodel.FamilyMember{
		FamilyID:          familyID,
		MemberType:        "LINEAGE_MEMBER",
		DisplayName:       name,
		Gender:            string(enums.GenderUnknown),
		UserBindingPolicy: string(enums.UserBindingOptional),
		Status:            string(enums.StatusActive),
		CreatedByUserID:   &actorID,
	}
	applyMemberValues(member, values)

	err := runMemberTransaction(ctx, s.db, func(tx *gorm.DB) error {
		if s.quota != nil {
			if err := s.quota.AssertProfileComplete(ctx, tx, actorID); err != nil {
				return err
			}
		}
		txRepo := s.repo.WithTx(tx)
		family, err := txRepo.LockFamily(ctx, familyID)
		if err != nil {
			return err
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyUnavailable
		}
		if s.quota != nil {
			if err := s.quota.AssertCanAddMember(ctx, tx, familyID, 1); err != nil {
				return err
			}
		}
		if err := txRepo.Create(ctx, member); err != nil {
			return err
		}
		if err := txRepo.IncrementGraphVersion(ctx, familyID); err != nil {
			return err
		}
		return writeLog(ctx, tx, actorID, "CREATE_MEMBER", member.ID, familyID, audit)
	})
	if errors.Is(err, errFamilyUnavailable) {
		return nil, memberError(CodeMemberFamilyUnavailable, "家庭状态不允许操作")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, memberError(CodeMemberFamilyUnavailable, "家庭不存在或不可用")
	}
	if err != nil {
		if s.quota != nil {
			if businessErr := s.quota.MapQuotaError(err); businessErr != nil && businessErr.Code != apperrors.CodeSystemError {
				return nil, businessErr
			}
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	row, err := s.repo.Find(ctx, familyID, member.ID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := memberVO(row)
	return &result, nil
}

func (s *memberService) List(ctx context.Context, actorID uint64, familyID uint64) ([]vo.Member, *apperrors.BusinessError) {
	if allowed, err := s.permissions.IsFamilyMember(ctx, actorID, familyID); err != nil || !allowed {
		return nil, memberError(CodeMemberViewForbidden, "无权查看家庭成员")
	}
	rows, err := s.repo.List(ctx, familyID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := make([]vo.Member, 0, len(rows))
	for i := range rows {
		result = append(result, memberVO(&rows[i]))
	}
	return result, nil
}

func (s *memberService) Detail(ctx context.Context, actorID uint64, familyID uint64, memberID uint64) (*vo.Member, *apperrors.BusinessError) {
	if allowed, err := s.permissions.IsFamilyMember(ctx, actorID, familyID); err != nil || !allowed {
		return nil, memberError(CodeMemberViewForbidden, "无权查看家庭成员")
	}
	row, err := s.repo.Find(ctx, familyID, memberID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, memberError(CodeMemberNotFound, "成员不存在")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := memberVO(row)
	return &result, nil
}

func (s *memberService) Update(ctx context.Context, actorID uint64, familyID uint64, memberID uint64, req dto.UpdateMemberRequest, audit AuditInput) (*vo.Member, *apperrors.BusinessError) {
	if allowed, err := s.permissions.CanEditMember(ctx, actorID, familyID, memberID); err != nil || !allowed {
		return nil, memberError(CodeMemberEditForbidden, "无权编辑成员")
	}
	values, businessErr := updateMemberValues(req)
	if businessErr != nil {
		return nil, businessErr
	}
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: actorID,
		Scene:  contentsafety.SceneProfile,
		Fields: contentsafety.MergeFields(
			contentsafety.OptionalField("member_name", req.Name),
			contentsafety.OptionalField("member_description", req.Description),
		),
		FamilyID: &familyID,
		IP:       audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	err := runMemberTransaction(ctx, s.db, func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		family, err := txRepo.LockFamily(ctx, familyID)
		if err != nil {
			return err
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyUnavailable
		}
		member, err := txRepo.FindForUpdate(ctx, familyID, memberID)
		if err != nil {
			return err
		}
		effectiveBirth := member.BirthDate
		if value, ok := values["birth_date"].(*time.Time); ok {
			effectiveBirth = value
		}
		effectiveDeath := member.DeathDate
		if value, ok := values["death_date"].(*time.Time); ok {
			effectiveDeath = value
		}
		if effectiveBirth != nil && effectiveDeath != nil && effectiveDeath.Before(*effectiveBirth) {
			return errInvalidLifeDates
		}
		if req.Description != nil || req.AvatarURL != nil {
			profile := decodeProfile(member)
			if req.Description != nil {
				profile.Description = cleanString(req.Description)
			}
			if req.AvatarURL != nil {
				profile.AvatarURL = cleanString(req.AvatarURL)
			}
			note, noteType, err := encodeProfile(profile.Description, profile.AvatarURL)
			if err != nil {
				return errInvalidProfile
			}
			values["lineage_note"] = note
			values["lineage_note_type"] = noteType
		}
		if policy, ok := values["user_binding_policy"].(string); ok && policy == string(enums.UserBindingNotRequired) {
			if _, err := txRepo.FindActiveLinkByMember(ctx, familyID, memberID); err == nil {
				return errBindingConflict
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		if len(values) == 0 {
			_ = member
			return nil
		}
		if err := txRepo.Update(ctx, familyID, memberID, values); err != nil {
			return err
		}
		if err := txRepo.IncrementGraphVersion(ctx, familyID); err != nil {
			return err
		}
		return writeLog(ctx, tx, actorID, "UPDATE_MEMBER", memberID, familyID, audit)
	})
	if errors.Is(err, errBindingConflict) {
		return nil, memberError(CodeMemberBindingConflict, "当前成员已有绑定，不能标记无需绑定")
	}
	if errors.Is(err, errInvalidProfile) {
		return nil, memberError(CodeMemberInvalidField, "成员描述或头像地址过长")
	}
	if errors.Is(err, errInvalidLifeDates) {
		return nil, memberError(CodeMemberInvalidField, "去世日期不能早于出生日期")
	}
	if errors.Is(err, errFamilyUnavailable) {
		return nil, memberError(CodeMemberFamilyUnavailable, "家庭状态不允许操作")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, memberError(CodeMemberNotFound, "成员不存在")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return s.Detail(ctx, actorID, familyID, memberID)
}

func (s *memberService) Delete(ctx context.Context, actorID uint64, familyID uint64, memberID uint64, req dto.DeleteMemberRequest, audit AuditInput) *apperrors.BusinessError {
	if allowed, err := s.permissions.CanDeleteMember(ctx, actorID, familyID, memberID); err != nil || !allowed {
		return memberError(CodeMemberDeleteForbidden, "无权删除成员")
	}
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: actorID, Scene: contentsafety.SceneSocial,
		Fields:   contentsafety.OptionalField("delete_reason", req.Reason),
		FamilyID: &familyID, IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return businessErr
	}
	err := runMemberTransaction(ctx, s.db, func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		family, err := txRepo.LockFamily(ctx, familyID)
		if err != nil {
			return err
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyUnavailable
		}
		if family.CurrentFounderMemberID != nil && *family.CurrentFounderMemberID == memberID {
			return errFounderProtected
		}
		if _, err := txRepo.FindForUpdate(ctx, familyID, memberID); err != nil {
			return err
		}
		if count, err := txRepo.CountBlockingRelationships(ctx, familyID, memberID); err != nil {
			return err
		} else if count > 0 {
			return errHasRelationships
		}
		if link, err := txRepo.FindActiveLinkByMember(ctx, familyID, memberID); err == nil {
			if link.FamilyRole == string(enums.FamilyRoleFounder) || link.FamilyRole == string(enums.FamilyRoleFamilyAdmin) {
				return errPrivilegedLink
			}
			if err := txRepo.UnbindLink(ctx, link.ID, actorID, stringPtr("MEMBER_DELETED"), time.Now()); err != nil {
				return err
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		now := time.Now()
		relationshipDeletedCount, err := txRepo.SoftDeleteDeletableRelationships(ctx, familyID, memberID, actorID, cleanString(req.Reason), now)
		if err != nil {
			return err
		}
		if err := txRepo.SoftDelete(ctx, familyID, memberID, actorID, cleanString(req.Reason), now); err != nil {
			return err
		}
		if err := txRepo.IncrementGraphVersion(ctx, familyID); err != nil {
			return err
		}
		if relationshipDeletedCount > 0 {
			if err := writeRelationshipCleanupLog(ctx, tx, actorID, memberID, familyID, relationshipDeletedCount, audit); err != nil {
				return err
			}
		}
		return writeLog(ctx, tx, actorID, "DELETE_MEMBER", memberID, familyID, audit)
	})
	switch {
	case errors.Is(err, errFounderProtected), errors.Is(err, errPrivilegedLink):
		return memberError(CodeMemberFounderProtected, "家庭创始人或管理员成员不能直接删除")
	case errors.Is(err, errHasRelationships):
		return memberError(CodeMemberHasRelationships, "该成员已有子女关系，不能直接删除")
	case errors.Is(err, errFamilyUnavailable):
		return memberError(CodeMemberFamilyUnavailable, "家庭状态不允许操作")
	case errors.Is(err, gorm.ErrRecordNotFound):
		return memberError(CodeMemberNotFound, "成员不存在")
	case err != nil:
		return apperrors.New(apperrors.CodeSystemError)
	default:
		return nil
	}
}

func (s *memberService) BindUser(ctx context.Context, actorID uint64, familyID uint64, memberID uint64, req dto.BindUserRequest, audit AuditInput) (*vo.Member, *apperrors.BusinessError) {
	if allowed, err := s.permissions.CanManageFamily(ctx, actorID, familyID); err != nil || !allowed {
		return nil, memberError(CodeMemberEditForbidden, "无权绑定成员用户")
	}
	err := runMemberTransaction(ctx, s.db, func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		family, err := txRepo.LockFamily(ctx, familyID)
		if err != nil {
			return err
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyUnavailable
		}
		member, err := txRepo.FindForUpdate(ctx, familyID, memberID)
		if err != nil {
			return err
		}
		if member.UserBindingPolicy == string(enums.UserBindingNotRequired) {
			return errBindingConflict
		}
		user, err := txRepo.FindUserForUpdate(ctx, req.UserID)
		if err != nil {
			return errUserUnavailable
		}
		if user.Status != string(enums.StatusActive) {
			return errUserUnavailable
		}
		if _, err := txRepo.FindActiveLinkByMember(ctx, familyID, memberID); err == nil {
			return errMemberAlreadyBound
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if _, err := txRepo.FindActiveLinkByUser(ctx, familyID, req.UserID); err == nil {
			return errUserAlreadyBound
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		now := time.Now()
		link := &rolemodel.FamilyMemberUserLink{
			FamilyID: familyID, MemberID: memberID, UserID: req.UserID,
			LinkStatus: string(enums.StatusActive), LinkSource: "MANUAL_BIND",
			FamilyRole: string(enums.FamilyRoleMember), RoleGrantedAt: &now,
			RoleGrantedByUserID: &actorID,
		}
		if err := txRepo.CreateLink(ctx, link); err != nil {
			return err
		}
		return writeLogWithUser(ctx, tx, actorID, req.UserID, "BIND_MEMBER_USER", memberID, familyID, audit)
	})
	switch {
	case errors.Is(err, errBindingConflict):
		return nil, memberError(CodeMemberBindingConflict, "无需绑定用户的成员不能绑定")
	case errors.Is(err, errUserUnavailable):
		return nil, memberError(CodeMemberUserUnavailable, "目标用户不存在或状态不可用")
	case errors.Is(err, errMemberAlreadyBound):
		return nil, memberError(CodeMemberAlreadyBound, "当前成员已绑定用户")
	case errors.Is(err, errUserAlreadyBound):
		return nil, memberError(CodeMemberUserAlreadyBound, "目标用户已在该家庭绑定成员")
	case errors.Is(err, errFamilyUnavailable):
		return nil, memberError(CodeMemberFamilyUnavailable, "家庭状态不允许操作")
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, memberError(CodeMemberNotFound, "成员不存在")
	case err != nil:
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return s.Detail(ctx, actorID, familyID, memberID)
}

func (s *memberService) UnbindUser(ctx context.Context, actorID uint64, familyID uint64, memberID uint64, req dto.UnbindUserRequest, audit AuditInput) (*vo.Member, *apperrors.BusinessError) {
	if allowed, err := s.permissions.CanManageFamily(ctx, actorID, familyID); err != nil || !allowed {
		return nil, memberError(CodeMemberEditForbidden, "无权解绑成员用户")
	}
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: actorID, Scene: contentsafety.SceneSocial,
		Fields:   contentsafety.OptionalField("unbind_reason", req.Reason),
		FamilyID: &familyID, IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	err := runMemberTransaction(ctx, s.db, func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		family, err := txRepo.LockFamily(ctx, familyID)
		if err != nil {
			return err
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyUnavailable
		}
		if _, err := txRepo.FindForUpdate(ctx, familyID, memberID); err != nil {
			return err
		}
		link, err := txRepo.FindActiveLinkByMember(ctx, familyID, memberID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errMemberNotBound
		}
		if err != nil {
			return err
		}
		if link.FamilyRole == string(enums.FamilyRoleFounder) || link.FamilyRole == string(enums.FamilyRoleFamilyAdmin) {
			return errPrivilegedLink
		}
		if err := txRepo.UnbindLink(ctx, link.ID, actorID, cleanString(req.Reason), time.Now()); err != nil {
			return err
		}
		return writeLogWithUser(ctx, tx, actorID, link.UserID, "UNBIND_MEMBER_USER", memberID, familyID, audit)
	})
	switch {
	case errors.Is(err, errMemberNotBound):
		return nil, memberError(CodeMemberNotBound, "当前成员未绑定用户")
	case errors.Is(err, errPrivilegedLink):
		return nil, memberError(CodeMemberUnbindProtected, "创始人或家庭管理员不能直接解绑")
	case errors.Is(err, errFamilyUnavailable):
		return nil, memberError(CodeMemberFamilyUnavailable, "家庭状态不允许操作")
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, memberError(CodeMemberNotFound, "成员不存在")
	case err != nil:
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return s.Detail(ctx, actorID, familyID, memberID)
}

var (
	errFamilyUnavailable  = errors.New("family unavailable")
	errBindingConflict    = errors.New("member binding policy conflict")
	errFounderProtected   = errors.New("founder protected")
	errHasRelationships   = errors.New("member has relationships")
	errPrivilegedLink     = errors.New("privileged link protected")
	errUserUnavailable    = errors.New("user unavailable")
	errMemberAlreadyBound = errors.New("member already bound")
	errUserAlreadyBound   = errors.New("user already bound")
	errMemberNotBound     = errors.New("member not bound")
	errInvalidProfile     = errors.New("invalid member profile")
	errInvalidLifeDates   = errors.New("invalid member life dates")
)

func memberValues(gender *string, birthDate *string, birthYear *int, deathDate *string, deathYear *int, isAlive *bool, avatarURL *string, description *string, policy *string) (map[string]any, *apperrors.BusinessError) {
	values := map[string]any{}
	if gender != nil {
		normalized := strings.ToUpper(strings.TrimSpace(*gender))
		if normalized != string(enums.GenderMale) && normalized != string(enums.GenderFemale) && normalized != string(enums.GenderUnknown) {
			return nil, memberError(CodeMemberInvalidField, "成员性别错误")
		}
		values["gender"] = normalized
	}
	birth, err := parseDateOrYear(birthDate, birthYear)
	if err != nil {
		return nil, memberError(CodeMemberInvalidField, "出生日期或年份错误")
	}
	if birthDate != nil || birthYear != nil {
		values["birth_date"] = birth
	}
	death, err := parseDateOrYear(deathDate, deathYear)
	if err != nil {
		return nil, memberError(CodeMemberInvalidField, "去世日期或年份错误")
	}
	if deathDate != nil || deathYear != nil {
		values["death_date"] = death
	}
	if birth != nil && death != nil && death.Before(*birth) {
		return nil, memberError(CodeMemberInvalidField, "去世日期不能早于出生日期")
	}
	if isAlive != nil {
		values["is_living"] = *isAlive
	}
	if policy != nil {
		normalized := strings.ToUpper(strings.TrimSpace(*policy))
		if normalized != string(enums.UserBindingOptional) && normalized != string(enums.UserBindingRequired) && normalized != string(enums.UserBindingNotRequired) {
			return nil, memberError(CodeMemberInvalidField, "用户绑定策略错误")
		}
		values["user_binding_policy"] = normalized
	}
	note, noteType, err := encodeProfile(description, avatarURL)
	if err != nil {
		return nil, memberError(CodeMemberInvalidField, "成员描述或头像地址过长")
	}
	if description != nil || avatarURL != nil {
		values["lineage_note"] = note
		values["lineage_note_type"] = noteType
	}
	return values, nil
}

func updateMemberValues(req dto.UpdateMemberRequest) (map[string]any, *apperrors.BusinessError) {
	values, businessErr := memberValues(req.Gender, req.BirthDate, req.BirthYear, req.DeathDate, req.DeathYear, req.IsAlive, nil, nil, req.UserBindingPolicy)
	if businessErr != nil {
		return nil, businessErr
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, memberError(CodeMemberNameRequired, "成员姓名不能为空")
		}
		values["display_name"] = name
	}
	return values, nil
}

func applyMemberValues(member *membermodel.FamilyMember, values map[string]any) {
	if value, ok := values["gender"].(string); ok {
		member.Gender = value
	}
	if value, ok := values["birth_date"].(*time.Time); ok {
		member.BirthDate = value
	}
	if value, ok := values["death_date"].(*time.Time); ok {
		member.DeathDate = value
	}
	if value, ok := values["is_living"].(bool); ok {
		member.IsLiving = &value
	}
	if value, ok := values["user_binding_policy"].(string); ok {
		member.UserBindingPolicy = value
	}
	if value, ok := values["lineage_note"].(*string); ok {
		member.LineageNote = value
	}
	if value, ok := values["lineage_note_type"].(*string); ok {
		member.LineageNoteType = value
	}
}

func parseDateOrYear(dateValue *string, yearValue *int) (*time.Time, error) {
	if dateValue != nil && strings.TrimSpace(*dateValue) != "" {
		value, err := time.Parse("2006-01-02", strings.TrimSpace(*dateValue))
		return &value, err
	}
	if yearValue != nil {
		if *yearValue < 1 || *yearValue > 9999 {
			return nil, errors.New("invalid year")
		}
		value := time.Date(*yearValue, time.January, 1, 0, 0, 0, 0, time.UTC)
		return &value, nil
	}
	return nil, nil
}

func encodeProfile(description *string, avatarURL *string) (*string, *string, error) {
	profile := profileNote{Description: cleanString(description), AvatarURL: cleanString(avatarURL)}
	data, err := json.Marshal(profile)
	if err != nil {
		return nil, nil, err
	}
	if len(data) > 500 {
		return nil, nil, errors.New("profile too long")
	}
	value := string(data)
	noteType := profileNoteType
	return &value, &noteType, nil
}

func decodeProfile(member *membermodel.FamilyMember) profileNote {
	if member.LineageNote == nil {
		return profileNote{}
	}
	if member.LineageNoteType == nil || *member.LineageNoteType != profileNoteType {
		return profileNote{Description: member.LineageNote}
	}
	var profile profileNote
	if err := json.Unmarshal([]byte(*member.LineageNote), &profile); err != nil {
		return profileNote{Description: member.LineageNote}
	}
	return profile
}

func memberVO(row *memberrepo.MemberRow) vo.Member {
	profile := decodeProfile(&row.FamilyMember)
	result := vo.Member{
		MemberID: row.ID, FamilyID: row.FamilyID, Name: row.DisplayName, Gender: row.Gender,
		IsAlive: row.IsLiving, AvatarURL: profile.AvatarURL, Description: profile.Description,
		Status: row.Status, UserBindingPolicy: row.UserBindingPolicy,
		BoundUserID: row.BoundUserID, BoundFamilyRole: row.BoundFamilyRole,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.BirthDate != nil {
		value := row.BirthDate.Format("2006-01-02")
		year := row.BirthDate.Year()
		result.BirthDate, result.BirthYear = &value, &year
	}
	if row.DeathDate != nil {
		value := row.DeathDate.Format("2006-01-02")
		year := row.DeathDate.Year()
		result.DeathDate, result.DeathYear = &value, &year
	}
	return result
}

func writeLog(ctx context.Context, tx *gorm.DB, actorID uint64, action string, memberID uint64, familyID uint64, audit AuditInput) error {
	return writeLogWithUser(ctx, tx, actorID, 0, action, memberID, familyID, audit)
}

func writeLogWithUser(ctx context.Context, tx *gorm.DB, actorID uint64, targetUserID uint64, action string, memberID uint64, familyID uint64, audit AuditInput) error {
	targetType := "FAMILY_MEMBER"
	input := operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FAMILY_MEMBER", Action: action, TargetType: &targetType,
		TargetID: &memberID, FamilyID: &familyID, MemberID: &memberID,
		IP: cleanString(&audit.IP), UserAgent: cleanString(&audit.UserAgent),
	}
	if targetUserID != 0 {
		input.UserID = &targetUserID
	}
	return operationlog.NewGormService(tx).WriteSuccess(ctx, input)
}

func writeRelationshipCleanupLog(ctx context.Context, tx *gorm.DB, actorID uint64, memberID uint64, familyID uint64, count int64, audit AuditInput) error {
	targetType := "FAMILY_RELATIONSHIP"
	detail, _ := json.Marshal(map[string]any{
		"relationshipCount":   count,
		"deletedWithMemberID": memberID,
	})
	return operationlog.NewGormService(tx).WriteSuccess(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FAMILY_RELATIONSHIP", Action: "DELETE_RELATIONSHIP_WITH_MEMBER", TargetType: &targetType,
		FamilyID: &familyID, MemberID: &memberID, DetailJSON: detail,
		IP: cleanString(&audit.IP), UserAgent: cleanString(&audit.UserAgent),
	})
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

func stringPtr(value string) *string {
	return &value
}

func memberError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}
