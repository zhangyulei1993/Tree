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
	PlaceExisting(context.Context, uint64, uint64, dto.PlaceExistingMemberRequest, AuditInput) (*vo.MutationResult, *apperrors.BusinessError)
	Update(context.Context, uint64, uint64, uint64, dto.UpdateRelationshipRequest, AuditInput) (*vo.MutationResult, *apperrors.BusinessError)
	Delete(context.Context, uint64, uint64, uint64, dto.DeleteRelationshipRequest, AuditInput) (*vo.MutationResult, *apperrors.BusinessError)
}

func (s *relationshipService) PlaceExisting(ctx context.Context, actorID uint64, familyID uint64, req dto.PlaceExistingMemberRequest, audit AuditInput) (*vo.MutationResult, *apperrors.BusinessError) {
	if allowed, err := s.permissions.CanManageRelationships(ctx, actorID, familyID); err != nil || !allowed {
		return nil, relationshipError(CodeRelationshipForbidden, "无权修改家庭关系")
	}
	if req.BaseMemberID == req.MemberID {
		return nil, relationshipError(CodeRelationshipSelf, "不能与自己建立关系")
	}

	var member *membermodel.FamilyMember
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
		member, err = repo.FindMemberForUpdate(ctx, familyID, req.MemberID)
		if err != nil {
			return errMemberUnavailable
		}
		located, err := repo.HasActiveRelationships(ctx, familyID, req.MemberID)
		if err != nil {
			return err
		}
		if located {
			return errMemberAlreadyLocated
		}
		createdRelationships, err = PlaceExistingMember(
			ctx, repo, familyID, family.FamilySurname, actorID,
			req.BaseMemberID, member, req.AddType, req.MemberType, req.Relationship,
		)
		if err != nil {
			return err
		}
		graphVersion, err = repo.IncrementGraphVersion(ctx, familyID)
		if err != nil {
			return err
		}
		return writeOperationLog(ctx, repo, actorID, familyID, "PLACE_EXISTING_MEMBER", createdRelationships, audit)
	})
	if businessErr := mapRepositoryError(err); businessErr != nil {
		return nil, businessErr
	}
	return mutationResult(member, createdRelationships, graphVersion), nil
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
		if err := validatePlacementMembers(baseMember, member, addType); err != nil {
			return err
		}
		if err := repo.CreateMember(ctx, member); err != nil {
			return err
		}
		var requestedMemberType *string
		if addType == relationshipenum.AddTypeFather || addType == relationshipenum.AddTypeMother {
			requestedMemberType = req.NewMember.MemberType
		}
		createdRelationships, err = placeExistingMember(
			ctx, repo, familyID, family.FamilySurname, actorID, req.BaseMemberID, member, addType, requestedMemberType, input,
		)
		if err != nil {
			return err
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
	if addType == relationshipenum.AddTypeFather && gender != string(enums.GenderMale) {
		return nil, relationshipError(CodeRelationshipMember, "父亲成员性别必须为男")
	}
	if addType == relationshipenum.AddTypeMother && gender != string(enums.GenderFemale) {
		return nil, relationshipError(CodeRelationshipMember, "母亲成员性别必须为女")
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

func ensurePrimaryParentAvailable(ctx context.Context, repo PlacementRepository, familyID uint64, childID uint64, gender string, parentLinkType *string, excludeRelationshipID uint64) error {
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

func applyParentPlacement(ctx context.Context, repo PlacementRepository, familyID uint64, familySurname string, childID uint64, gender string, input normalizedRelationshipInput, noteType string, note string) normalizedRelationshipInput {
	if input.parentLinkType == nil || *input.parentLinkType != string(relationshipenum.ParentLinkTypePrimary) {
		return input
	}
	_, existingParent, err := repo.FindPrimaryParentWithMemberByGender(ctx, familyID, childID, gender, 0)
	if err != nil || existingParent == nil || isLineageMember(existingParent, familySurname) {
		return input
	}
	step := string(relationshipenum.ParentLinkTypeStep)
	input.parentLinkType = &step
	if input.relationNoteType == nil {
		input.relationNoteType = &noteType
	}
	if input.relationNote == nil {
		input.relationNote = &note
	}
	return input
}

func applySpousePlacement(ctx context.Context, repo PlacementRepository, familyID uint64, familySurname string, baseMemberID uint64, spouseGender string, input normalizedRelationshipInput) (normalizedRelationshipInput, error) {
	spouses, err := repo.ListActiveSpouseRelationshipsByGender(ctx, familyID, baseMemberID, spouseGender)
	if err != nil || len(spouses) == 0 {
		return input, err
	}
	for i := range spouses {
		if isLineageMember(&spouses[i].Spouse, familySurname) {
			if spouseGender == string(enums.GenderMale) {
				return input, errPrimaryHusbandExists
			}
			return input, errPrimaryWifeExists
		}
	}
	noteType := "SECOND_WIFE"
	note := "再婚妻子"
	if spouseGender == string(enums.GenderMale) {
		noteType = "SECOND_HUSBAND"
		note = "再婚丈夫"
	}
	if input.relationNoteType == nil {
		input.relationNoteType = &noteType
	}
	if input.relationNote == nil {
		input.relationNote = &note
	}
	previousType := "EX_WIFE"
	previousNote := "前妻"
	if spouseGender == string(enums.GenderMale) {
		previousType = "EX_HUSBAND"
		previousNote = "前夫"
	}
	for i := range spouses {
		values := map[string]any{"relation_note_type": &previousType, "relation_note": &previousNote}
		if err := repo.UpdateRelationship(ctx, familyID, spouses[i].Relationship.ID, values); err != nil {
			return input, err
		}
	}
	return input, nil
}

func oppositeGender(gender string) string {
	if gender == string(enums.GenderMale) {
		return string(enums.GenderFemale)
	}
	return string(enums.GenderMale)
}

func isLineageMember(member *membermodel.FamilyMember, familySurname string) bool {
	if strings.ToUpper(strings.TrimSpace(member.MemberType)) == "SPOUSE" {
		return false
	}
	expected := strings.TrimSpace(familySurname)
	if expected == "" {
		return true
	}
	if member.Surname != nil && strings.TrimSpace(*member.Surname) != "" {
		return strings.TrimSpace(*member.Surname) == expected
	}
	return firstRune(member.DisplayName) == expected
}

func firstRune(value string) string {
	for _, r := range strings.TrimSpace(value) {
		return string(r)
	}
	return ""
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
	if businessErr := asPlacementBusinessError(err); businessErr != nil {
		return businessErr
	}
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errSiblingParentRequired):
		return relationshipError(CodeSiblingParentRequired, "请先创建父亲或母亲节点，再添加兄弟姐妹")
	case errors.Is(err, errMemberUnavailable), errors.Is(err, gorm.ErrRecordNotFound):
		return relationshipError(CodeRelationshipMember, "关系成员不存在或不属于该家庭")
	case errors.Is(err, errMemberAlreadyLocated):
		return relationshipError(CodeRelationshipDuplicate, "该成员已在家谱中，无需重复定位")
	case errors.Is(err, errDuplicateRelationship):
		return relationshipError(CodeRelationshipDuplicate, "当前关系已存在")
	case errors.Is(err, errPrimaryFatherExists):
		return relationshipError(CodePrimaryFatherExists, "PRIMARY 父亲已存在")
	case errors.Is(err, errPrimaryMotherExists):
		return relationshipError(CodePrimaryMotherExists, "PRIMARY 母亲已存在")
	case errors.Is(err, errPrimaryHusbandExists):
		return relationshipError(CodeRelationshipType, "已有本家族丈夫，不能重复添加")
	case errors.Is(err, errPrimaryWifeExists):
		return relationshipError(CodeRelationshipType, "已有本家族妻子，不能重复添加")
	case errors.Is(err, errBaseGenderRequired):
		return relationshipError(CodeRelationshipMember, "请先完善基准成员性别")
	case errors.Is(err, errFatherGenderRequired):
		return relationshipError(CodeRelationshipMember, "父亲成员性别必须为男")
	case errors.Is(err, errMotherGenderRequired):
		return relationshipError(CodeRelationshipMember, "母亲成员性别必须为女")
	case errors.Is(err, errSpouseGenderMismatch):
		return relationshipError(CodeRelationshipMember, "配偶性别必须与基准成员相对")
	case errors.Is(err, errChildGenderRequired):
		return relationshipError(CodeRelationshipMember, "子女性别必须选择男或女")
	case errors.Is(err, errSiblingGenderRequired):
		return relationshipError(CodeRelationshipMember, "兄弟姐妹性别必须选择男或女")
	case errors.Is(err, errUnsupportedRelationship):
		return relationshipError(CodeRelationshipType, "不支持的关系类型或修改")
	case errors.Is(err, errRelationshipNotFound):
		return relationshipError(CodeRelationshipNotFound, "关系不存在")
	case errors.Is(err, errSelfRelationship):
		return relationshipError(CodeRelationshipSelf, "不能与自己建立关系")
	case errors.Is(err, errSpouseBaseExpansion):
		return relationshipError(CodeRelationshipForbidden, "配偶节点不能作为家谱主干扩展，请从族内成员节点操作")
	case errors.Is(err, errParentMemberTypeRequired):
		return relationshipError(CodeRelationshipMember, "添加父母时必须选择成员身份")
	case errors.Is(err, errParentMemberTypeInvalid):
		return relationshipError(CodeRelationshipMember, "父母成员身份仅支持本家成员或本家成员的配偶")
	case errors.Is(err, errLineageParentRequired):
		return relationshipError(CodeRelationshipType, "请先录入本家成员作为父母，再添加配偶父母")
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
	errSiblingParentRequired    = errors.New("sibling parent required")
	errMemberUnavailable        = errors.New("relationship member unavailable")
	errMemberAlreadyLocated     = errors.New("member already located")
	errDuplicateRelationship    = errors.New("duplicate relationship")
	errPrimaryFatherExists      = errors.New("primary father exists")
	errPrimaryMotherExists      = errors.New("primary mother exists")
	errPrimaryHusbandExists     = errors.New("primary husband exists")
	errPrimaryWifeExists        = errors.New("primary wife exists")
	errBaseGenderRequired       = errors.New("base gender required")
	errFatherGenderRequired     = errors.New("father gender required")
	errMotherGenderRequired     = errors.New("mother gender required")
	errSpouseGenderMismatch     = errors.New("spouse gender mismatch")
	errChildGenderRequired      = errors.New("child gender required")
	errSiblingGenderRequired    = errors.New("sibling gender required")
	errUnsupportedRelationship  = errors.New("unsupported relationship")
	errRelationshipNotFound     = errors.New("relationship not found")
	errSelfRelationship         = errors.New("self relationship")
	errSpouseBaseExpansion      = errors.New("spouse base expansion forbidden")
	errFamilyUnavailable        = errors.New("family unavailable")
	errParentMemberTypeRequired = errors.New("parent member type required")
	errParentMemberTypeInvalid  = errors.New("parent member type invalid")
	errLineageParentRequired    = errors.New("lineage parent required before spouse parent")
)
