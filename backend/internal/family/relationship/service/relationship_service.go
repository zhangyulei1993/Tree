package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	membermodel "tree/backend/internal/family/member/model"
	"tree/backend/internal/family/relationship/dto"
	relationshipenum "tree/backend/internal/family/relationship/enum"
	relationshipmodel "tree/backend/internal/family/relationship/model"
	relationshiprepo "tree/backend/internal/family/relationship/repository"
	"tree/backend/internal/family/relationship/vo"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	CodeSiblingParentRequired apperrors.Code = 43301
	CodeRelationshipMember    apperrors.Code = 43302
	CodeRelationshipDuplicate apperrors.Code = 43303
	CodePrimaryFatherExists   apperrors.Code = 43304
	CodePrimaryMotherExists   apperrors.Code = 43305
	CodeRelationshipType      apperrors.Code = 43306
	CodeRelationshipForbidden apperrors.Code = 43307
	CodeRelationshipNotFound  apperrors.Code = 43308
	CodeRelationshipSelf      apperrors.Code = 43309
	CodeRelationshipFamily    apperrors.Code = 43310
)

type AuditInput struct {
	IP        string
	UserAgent string
}

type RelationshipService interface {
	Create(context.Context, uint64, uint64, dto.CreateRelationshipRequest, AuditInput) (*vo.MutationResult, *apperrors.BusinessError)
	Update(context.Context, uint64, uint64, uint64, dto.UpdateRelationshipRequest, AuditInput) (*vo.MutationResult, *apperrors.BusinessError)
	Delete(context.Context, uint64, uint64, uint64, dto.DeleteRelationshipRequest, AuditInput) (*vo.MutationResult, *apperrors.BusinessError)
}

type familyPermission interface {
	CanManageRelationships(context.Context, uint64, uint64) (bool, error)
}

type relationshipService struct {
	repo        relationshiprepo.Repository
	unitOfWork  relationshiprepo.UnitOfWork
	permissions familyPermission
}

func NewRelationshipService(repo relationshiprepo.Repository, unitOfWork relationshiprepo.UnitOfWork, permissions familyPermission) RelationshipService {
	return &relationshipService{repo: repo, unitOfWork: unitOfWork, permissions: permissions}
}

func (s *relationshipService) Create(ctx context.Context, actorID uint64, familyID uint64, req dto.CreateRelationshipRequest, audit AuditInput) (*vo.MutationResult, *apperrors.BusinessError) {
	if allowed, err := s.permissions.CanManageRelationships(ctx, actorID, familyID); err != nil || !allowed {
		return nil, relationshipError(CodeRelationshipForbidden, "无权修改家庭关系")
	}
	addType, businessErr := normalizeAddType(req.AddType)
	if businessErr != nil {
		return nil, businessErr
	}
	member, businessErr := newMember(familyID, actorID, addType, req.NewMember)
	if businessErr != nil {
		return nil, businessErr
	}
	input, businessErr := normalizeRelationshipInput(addType, req.Relationship)
	if businessErr != nil {
		return nil, businessErr
	}

	var createdRelationships []*relationshipmodel.FamilyRelationship
	var graphVersion int64
	err := s.unitOfWork.WithinTransaction(ctx, func(repo relationshiprepo.Repository) error {
		family, err := repo.LockFamily(ctx, familyID)
		if err != nil {
			return err
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyUnavailable
		}
		baseMember, err := repo.FindMemberForUpdate(ctx, familyID, req.BaseMemberID)
		if err != nil {
			return errMemberUnavailable
		}

		var plans []relationshipPlan
		switch addType {
		case relationshipenum.AddTypeFather:
			member.Gender = string(enums.GenderMale)
			plans = []relationshipPlan{{from: member, to: baseMember, relationshipType: string(relationshipenum.RelationshipTypeParentChild), input: input}}
			if err := ensurePrimaryParentAvailable(ctx, repo, familyID, baseMember.ID, member.Gender, input.parentLinkType, 0); err != nil {
				return err
			}
		case relationshipenum.AddTypeMother:
			member.Gender = string(enums.GenderFemale)
			plans = []relationshipPlan{{from: member, to: baseMember, relationshipType: string(relationshipenum.RelationshipTypeParentChild), input: input}}
			if err := ensurePrimaryParentAvailable(ctx, repo, familyID, baseMember.ID, member.Gender, input.parentLinkType, 0); err != nil {
				return err
			}
		case relationshipenum.AddTypeChild:
			plans = []relationshipPlan{{from: baseMember, to: member, relationshipType: string(relationshipenum.RelationshipTypeParentChild), input: input}}
		case relationshipenum.AddTypeSpouse:
			member.MemberType = "SPOUSE"
			plans = []relationshipPlan{{from: baseMember, to: member, relationshipType: string(relationshipenum.RelationshipTypeSpouse), input: input}}
		case relationshipenum.AddTypeSibling:
			parents, err := repo.ListActiveParents(ctx, familyID, baseMember.ID)
			if err != nil {
				return err
			}
			if len(parents) == 0 {
				return errSiblingParentRequired
			}
			plans = make([]relationshipPlan, 0, len(parents))
			for i := range parents {
				parent, err := repo.FindMemberForUpdate(ctx, familyID, parents[i].FromMemberID)
				if err != nil {
					return errMemberUnavailable
				}
				parentInput := input
				parentInput.parentLinkType = parents[i].ParentLinkType
				plans = append(plans, relationshipPlan{
					from: parent, to: member, relationshipType: string(relationshipenum.RelationshipTypeParentChild), input: parentInput,
				})
			}
		}

		if err := repo.CreateMember(ctx, member); err != nil {
			return err
		}
		for i := range plans {
			if plans[i].from.ID == 0 {
				plans[i].from.ID = member.ID
			}
			if plans[i].to.ID == 0 {
				plans[i].to.ID = member.ID
			}
			if plans[i].from.ID == plans[i].to.ID {
				return errSelfRelationship
			}
			if _, err := repo.FindDuplicate(ctx, familyID, plans[i].from.ID, plans[i].to.ID, plans[i].relationshipType); err == nil {
				return errDuplicateRelationship
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			relationship := plans[i].model(familyID, actorID)
			if err := repo.CreateRelationship(ctx, relationship); err != nil {
				return err
			}
			createdRelationships = append(createdRelationships, relationship)
		}
		graphVersion, err = repo.IncrementGraphVersion(ctx, familyID)
		if err != nil {
			return err
		}
		return writeOperationLog(ctx, repo, actorID, familyID, "CREATE_RELATIONSHIP", createdRelationships, audit)
	})
	if businessErr := mapRepositoryError(err); businessErr != nil {
		return nil, businessErr
	}
	return mutationResult(member, createdRelationships, graphVersion), nil
}

func (s *relationshipService) Update(ctx context.Context, actorID uint64, familyID uint64, relationshipID uint64, req dto.UpdateRelationshipRequest, audit AuditInput) (*vo.MutationResult, *apperrors.BusinessError) {
	if allowed, err := s.permissions.CanManageRelationships(ctx, actorID, familyID); err != nil || !allowed {
		return nil, relationshipError(CodeRelationshipForbidden, "无权修改家庭关系")
	}
	if req.FromMemberID != nil || req.ToMemberID != nil || req.RelationshipType != nil {
		return nil, relationshipError(CodeRelationshipType, "不允许修改关系成员或关系类型")
	}
	values, businessErr := updateValues(req)
	if businessErr != nil {
		return nil, businessErr
	}
	if len(values) == 0 {
		return nil, relationshipError(CodeRelationshipType, "没有可更新的关系字段")
	}

	var relationship *relationshipmodel.FamilyRelationship
	var graphVersion int64
	err := s.unitOfWork.WithinTransaction(ctx, func(repo relationshiprepo.Repository) error {
		family, err := repo.LockFamily(ctx, familyID)
		if err != nil {
			return err
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyUnavailable
		}
		relationship, err = repo.FindActiveRelationship(ctx, familyID, relationshipID)
		if err != nil {
			return errRelationshipNotFound
		}
		if relationship.RelationshipType == string(relationshipenum.RelationshipTypeSpouse) && req.ParentLinkType != nil {
			return errUnsupportedRelationship
		}
		normalizedParentLink, hasParentLink := values["parent_link_type"].(string)
		if hasParentLink && normalizedParentLink == string(relationshipenum.ParentLinkTypePrimary) {
			parent, err := repo.FindMemberForUpdate(ctx, familyID, relationship.FromMemberID)
			if err != nil {
				return errMemberUnavailable
			}
			if err := ensurePrimaryParentAvailable(ctx, repo, familyID, relationship.ToMemberID, parent.Gender, &normalizedParentLink, relationship.ID); err != nil {
				return err
			}
		}
		if err := repo.UpdateRelationship(ctx, familyID, relationshipID, values); err != nil {
			return err
		}
		applyRelationshipValues(relationship, values)
		graphVersion, err = repo.IncrementGraphVersion(ctx, familyID)
		if err != nil {
			return err
		}
		return writeOperationLog(ctx, repo, actorID, familyID, "UPDATE_RELATIONSHIP", []*relationshipmodel.FamilyRelationship{relationship}, audit)
	})
	if businessErr := mapRepositoryError(err); businessErr != nil {
		return nil, businessErr
	}
	return mutationResult(nil, []*relationshipmodel.FamilyRelationship{relationship}, graphVersion), nil
}

func (s *relationshipService) Delete(ctx context.Context, actorID uint64, familyID uint64, relationshipID uint64, req dto.DeleteRelationshipRequest, audit AuditInput) (*vo.MutationResult, *apperrors.BusinessError) {
	if allowed, err := s.permissions.CanManageRelationships(ctx, actorID, familyID); err != nil || !allowed {
		return nil, relationshipError(CodeRelationshipForbidden, "无权修改家庭关系")
	}
	var relationship *relationshipmodel.FamilyRelationship
	var graphVersion int64
	err := s.unitOfWork.WithinTransaction(ctx, func(repo relationshiprepo.Repository) error {
		family, err := repo.LockFamily(ctx, familyID)
		if err != nil {
			return err
		}
		if family.Status != string(enums.StatusNormal) {
			return errFamilyUnavailable
		}
		relationship, err = repo.FindActiveRelationship(ctx, familyID, relationshipID)
		if err != nil {
			return errRelationshipNotFound
		}
		now := time.Now()
		if err := repo.SoftDeleteRelationship(ctx, familyID, relationshipID, actorID, cleanString(req.Reason), now); err != nil {
			return err
		}
		relationship.Status = "DELETED"
		relationship.DeletedAt = &now
		relationship.DeletedByUserID = &actorID
		relationship.DeleteReason = cleanString(req.Reason)
		graphVersion, err = repo.IncrementGraphVersion(ctx, familyID)
		if err != nil {
			return err
		}
		return writeOperationLog(ctx, repo, actorID, familyID, "DELETE_RELATIONSHIP", []*relationshipmodel.FamilyRelationship{relationship}, audit)
	})
	if businessErr := mapRepositoryError(err); businessErr != nil {
		return nil, businessErr
	}
	return mutationResult(nil, []*relationshipmodel.FamilyRelationship{relationship}, graphVersion), nil
}

type normalizedRelationshipInput struct {
	parentLinkType   *string
	relationNoteType *string
	relationNote     *string
}

type relationshipPlan struct {
	from             *membermodel.FamilyMember
	to               *membermodel.FamilyMember
	relationshipType string
	input            normalizedRelationshipInput
}

func (p relationshipPlan) model(familyID uint64, actorID uint64) *relationshipmodel.FamilyRelationship {
	return &relationshipmodel.FamilyRelationship{
		FamilyID: familyID, FromMemberID: p.from.ID, ToMemberID: p.to.ID,
		RelationshipType: p.relationshipType, ParentLinkType: p.input.parentLinkType,
		RelationNoteType: p.input.relationNoteType, RelationNote: p.input.relationNote,
		Status: string(enums.StatusActive), CreatedByUserID: &actorID,
	}
}

func normalizeAddType(value string) (relationshipenum.AddType, *apperrors.BusinessError) {
	addType := relationshipenum.AddType(strings.ToUpper(strings.TrimSpace(value)))
	switch addType {
	case relationshipenum.AddTypeFather, relationshipenum.AddTypeMother, relationshipenum.AddTypeChild,
		relationshipenum.AddTypeSpouse, relationshipenum.AddTypeSibling:
		return addType, nil
	default:
		return "", relationshipError(CodeRelationshipType, "不支持的 addType")
	}
}

func normalizeRelationshipInput(addType relationshipenum.AddType, input dto.RelationshipInput) (normalizedRelationshipInput, *apperrors.BusinessError) {
	expectedType := string(relationshipenum.RelationshipTypeParentChild)
	if addType == relationshipenum.AddTypeSpouse {
		expectedType = string(relationshipenum.RelationshipTypeSpouse)
	}
	if input.RelationshipType != nil && strings.ToUpper(strings.TrimSpace(*input.RelationshipType)) != expectedType {
		return normalizedRelationshipInput{}, relationshipError(CodeRelationshipType, "关系类型与 addType 不匹配")
	}
	result := normalizedRelationshipInput{
		relationNoteType: cleanLimited(input.RelationNoteType, 80),
		relationNote:     cleanLimited(input.RelationNote, 500),
	}
	if input.RelationNoteType != nil && result.relationNoteType == nil && strings.TrimSpace(*input.RelationNoteType) != "" {
		return normalizedRelationshipInput{}, relationshipError(CodeRelationshipType, "关系备注类型过长")
	}
	if input.RelationNote != nil && result.relationNote == nil && strings.TrimSpace(*input.RelationNote) != "" {
		return normalizedRelationshipInput{}, relationshipError(CodeRelationshipType, "关系备注过长")
	}
	if expectedType == string(relationshipenum.RelationshipTypeParentChild) {
		parentLinkType := string(relationshipenum.ParentLinkTypePrimary)
		if input.ParentLinkType != nil {
			parentLinkType = strings.ToUpper(strings.TrimSpace(*input.ParentLinkType))
		}
		if !validParentLinkType(parentLinkType) {
			return normalizedRelationshipInput{}, relationshipError(CodeRelationshipType, "parentLinkType 不支持")
		}
		result.parentLinkType = &parentLinkType
	} else if input.ParentLinkType != nil {
		return normalizedRelationshipInput{}, relationshipError(CodeRelationshipType, "配偶关系不支持 parentLinkType")
	}
	return result, nil
}

func newMember(familyID uint64, actorID uint64, addType relationshipenum.AddType, input dto.NewMemberInput) (*membermodel.FamilyMember, *apperrors.BusinessError) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, relationshipError(CodeRelationshipMember, "新成员姓名不能为空")
	}
	gender := string(enums.GenderUnknown)
	if input.Gender != nil {
		gender = strings.ToUpper(strings.TrimSpace(*input.Gender))
		if gender != string(enums.GenderMale) && gender != string(enums.GenderFemale) && gender != string(enums.GenderUnknown) {
			return nil, relationshipError(CodeRelationshipMember, "新成员性别错误")
		}
	}
	if addType == relationshipenum.AddTypeFather && gender == string(enums.GenderFemale) {
		return nil, relationshipError(CodeRelationshipMember, "父亲成员性别不能为 FEMALE")
	}
	if addType == relationshipenum.AddTypeMother && gender == string(enums.GenderMale) {
		return nil, relationshipError(CodeRelationshipMember, "母亲成员性别不能为 MALE")
	}
	birthDate, err := parseDateOrYear(input.BirthDate, input.BirthYear)
	if err != nil {
		return nil, relationshipError(CodeRelationshipMember, "出生日期或年份错误")
	}
	deathDate, err := parseDateOrYear(input.DeathDate, input.DeathYear)
	if err != nil || birthDate != nil && deathDate != nil && deathDate.Before(*birthDate) {
		return nil, relationshipError(CodeRelationshipMember, "去世日期或年份错误")
	}
	policy := string(enums.UserBindingOptional)
	if input.UserBindingPolicy != nil {
		policy = strings.ToUpper(strings.TrimSpace(*input.UserBindingPolicy))
		if policy != string(enums.UserBindingOptional) && policy != string(enums.UserBindingRequired) && policy != string(enums.UserBindingNotRequired) {
			return nil, relationshipError(CodeRelationshipMember, "用户绑定策略错误")
		}
	}
	return &membermodel.FamilyMember{
		FamilyID: familyID, MemberType: "LINEAGE_MEMBER", DisplayName: name,
		Gender: gender, BirthDate: birthDate, DeathDate: deathDate, IsLiving: input.IsAlive,
		UserBindingPolicy: policy, Status: string(enums.StatusActive), CreatedByUserID: &actorID,
	}, nil
}

func ensurePrimaryParentAvailable(ctx context.Context, repo relationshiprepo.Repository, familyID uint64, childID uint64, gender string, parentLinkType *string, excludeRelationshipID uint64) error {
	if parentLinkType == nil || *parentLinkType != string(relationshipenum.ParentLinkTypePrimary) {
		return nil
	}
	if gender != string(enums.GenderMale) && gender != string(enums.GenderFemale) {
		return nil
	}
	if _, err := repo.FindPrimaryParentByGender(ctx, familyID, childID, gender, excludeRelationshipID); err == nil {
		if gender == string(enums.GenderMale) {
			return errPrimaryFatherExists
		}
		return errPrimaryMotherExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

func updateValues(req dto.UpdateRelationshipRequest) (map[string]any, *apperrors.BusinessError) {
	values := map[string]any{}
	if req.ParentLinkType != nil {
		value := strings.ToUpper(strings.TrimSpace(*req.ParentLinkType))
		if !validParentLinkType(value) {
			return nil, relationshipError(CodeRelationshipType, "parentLinkType 不支持")
		}
		values["parent_link_type"] = value
	}
	if req.RelationNoteType != nil {
		value := cleanLimited(req.RelationNoteType, 80)
		if value == nil && strings.TrimSpace(*req.RelationNoteType) != "" {
			return nil, relationshipError(CodeRelationshipType, "关系备注类型过长")
		}
		values["relation_note_type"] = value
	}
	if req.RelationNote != nil {
		value := cleanLimited(req.RelationNote, 500)
		if value == nil && strings.TrimSpace(*req.RelationNote) != "" {
			return nil, relationshipError(CodeRelationshipType, "关系备注过长")
		}
		values["relation_note"] = value
	}
	return values, nil
}

func validParentLinkType(value string) bool {
	switch relationshipenum.ParentLinkType(value) {
	case relationshipenum.ParentLinkTypePrimary, relationshipenum.ParentLinkTypeStep,
		relationshipenum.ParentLinkTypeAdoptive, relationshipenum.ParentLinkTypeSuccession,
		relationshipenum.ParentLinkTypeNoteOnly, relationshipenum.ParentLinkTypeOther:
		return true
	default:
		return false
	}
}

func applyRelationshipValues(relationship *relationshipmodel.FamilyRelationship, values map[string]any) {
	if value, ok := values["parent_link_type"].(string); ok {
		relationship.ParentLinkType = &value
	}
	if value, ok := values["relation_note_type"].(*string); ok {
		relationship.RelationNoteType = value
	}
	if value, ok := values["relation_note"].(*string); ok {
		relationship.RelationNote = value
	}
	relationship.UpdatedAt = time.Now()
}

func mutationResult(member *membermodel.FamilyMember, relationships []*relationshipmodel.FamilyRelationship, graphVersion int64) *vo.MutationResult {
	result := &vo.MutationResult{Relationships: make([]vo.RelationshipVO, 0, len(relationships)), GraphVersion: graphVersion}
	if member != nil {
		result.CreatedMember = &vo.CreatedMember{
			MemberID: member.ID, FamilyID: member.FamilyID, Name: member.DisplayName, Gender: member.Gender, Status: member.Status,
		}
	}
	for _, relationship := range relationships {
		result.Relationships = append(result.Relationships, relationshipVO(relationship))
	}
	return result
}

func relationshipVO(relationship *relationshipmodel.FamilyRelationship) vo.RelationshipVO {
	return vo.RelationshipVO{
		RelationshipID: relationship.ID, FamilyID: relationship.FamilyID,
		FromMemberID: relationship.FromMemberID, ToMemberID: relationship.ToMemberID,
		RelationshipType: relationship.RelationshipType, ParentLinkType: relationship.ParentLinkType,
		RelationNoteType: relationship.RelationNoteType, RelationNote: relationship.RelationNote,
		Status: relationship.Status, CreatedAt: relationship.CreatedAt,
		UpdatedAt: relationship.UpdatedAt, DeletedAt: relationship.DeletedAt,
	}
}

func writeOperationLog(ctx context.Context, repo relationshiprepo.Repository, actorID uint64, familyID uint64, action string, relationships []*relationshipmodel.FamilyRelationship, audit AuditInput) error {
	targetType := "FAMILY_RELATIONSHIP"
	var targetID *uint64
	if len(relationships) > 0 {
		targetID = &relationships[0].ID
	}
	detail, _ := json.Marshal(map[string]any{"relationshipCount": len(relationships)})
	return repo.WriteOperationLog(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &actorID,
		Module: "FAMILY_RELATIONSHIP", Action: action, TargetType: &targetType,
		TargetID: targetID, FamilyID: &familyID, DetailJSON: detail,
		IP: cleanString(&audit.IP), UserAgent: cleanString(&audit.UserAgent),
	})
}

func mapRepositoryError(err error) *apperrors.BusinessError {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errSiblingParentRequired):
		return relationshipError(CodeSiblingParentRequired, "请先创建父亲或母亲节点，再添加兄弟姐妹")
	case errors.Is(err, errMemberUnavailable), errors.Is(err, gorm.ErrRecordNotFound):
		return relationshipError(CodeRelationshipMember, "关系成员不存在或不属于该家庭")
	case errors.Is(err, errDuplicateRelationship):
		return relationshipError(CodeRelationshipDuplicate, "当前关系已存在")
	case errors.Is(err, errPrimaryFatherExists):
		return relationshipError(CodePrimaryFatherExists, "PRIMARY 父亲已存在")
	case errors.Is(err, errPrimaryMotherExists):
		return relationshipError(CodePrimaryMotherExists, "PRIMARY 母亲已存在")
	case errors.Is(err, errUnsupportedRelationship):
		return relationshipError(CodeRelationshipType, "不支持的关系类型或修改")
	case errors.Is(err, errRelationshipNotFound):
		return relationshipError(CodeRelationshipNotFound, "关系不存在")
	case errors.Is(err, errSelfRelationship):
		return relationshipError(CodeRelationshipSelf, "不能与自己建立关系")
	case errors.Is(err, errFamilyUnavailable):
		return relationshipError(CodeRelationshipFamily, "家庭不存在或状态不允许操作")
	default:
		return apperrors.New(apperrors.CodeSystemError)
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

func cleanLimited(value *string, max int) *string {
	if value == nil {
		return nil
	}
	cleaned := strings.TrimSpace(*value)
	if cleaned == "" {
		return nil
	}
	if len([]byte(cleaned)) > max {
		return nil
	}
	return &cleaned
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

func relationshipError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}

var (
	errSiblingParentRequired   = errors.New("sibling parent required")
	errMemberUnavailable       = errors.New("relationship member unavailable")
	errDuplicateRelationship   = errors.New("duplicate relationship")
	errPrimaryFatherExists     = errors.New("primary father exists")
	errPrimaryMotherExists     = errors.New("primary mother exists")
	errUnsupportedRelationship = errors.New("unsupported relationship")
	errRelationshipNotFound    = errors.New("relationship not found")
	errSelfRelationship        = errors.New("self relationship")
	errFamilyUnavailable       = errors.New("family unavailable")
)
