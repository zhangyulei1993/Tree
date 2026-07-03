package service

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	membermodel "tree/backend/internal/family/member/model"
	"tree/backend/internal/family/relationship/dto"
	relationshipenum "tree/backend/internal/family/relationship/enum"
	relationshipmodel "tree/backend/internal/family/relationship/model"
	relationshiprepo "tree/backend/internal/family/relationship/repository"
)

// PlacementRepository is the transaction-scoped subset required to attach an
// existing member to the family tree. Join approval reuses this logic so member
// creation, relationship creation and account binding can commit atomically.
type PlacementRepository interface {
	FindMemberForUpdate(context.Context, uint64, uint64) (*membermodel.FamilyMember, error)
	UpdateMember(context.Context, uint64, uint64, map[string]any) error
	FindDuplicate(context.Context, uint64, uint64, uint64, string) (*relationshipmodel.FamilyRelationship, error)
	ListActiveParents(context.Context, uint64, uint64) ([]relationshipmodel.FamilyRelationship, error)
	FindPrimaryParentByGender(context.Context, uint64, uint64, string, uint64) (*relationshipmodel.FamilyRelationship, error)
	FindPrimaryParentWithMemberByGender(context.Context, uint64, uint64, string, uint64) (*relationshipmodel.FamilyRelationship, *membermodel.FamilyMember, error)
	ListActiveSpouseRelationshipsByGender(context.Context, uint64, uint64, string) ([]relationshiprepo.SpouseRelationshipRow, error)
	CreateRelationship(context.Context, *relationshipmodel.FamilyRelationship) error
	UpdateRelationship(context.Context, uint64, uint64, map[string]any) error
}

func PlaceExistingMember(
	ctx context.Context,
	repo PlacementRepository,
	familyID uint64,
	familySurname string,
	actorID uint64,
	baseMemberID uint64,
	member *membermodel.FamilyMember,
	addTypeValue string,
	memberType *string,
	relationshipInput dto.RelationshipInput,
) ([]*relationshipmodel.FamilyRelationship, error) {
	addType, businessErr := normalizeAddType(addTypeValue)
	if businessErr != nil {
		return nil, businessErr
	}
	input, businessErr := normalizeRelationshipInput(addType, relationshipInput)
	if businessErr != nil {
		return nil, businessErr
	}
	return placeExistingMember(ctx, repo, familyID, familySurname, actorID, baseMemberID, member, addType, memberType, input)
}

func normalizedStoredMemberType(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func placeExistingMember(
	ctx context.Context,
	repo PlacementRepository,
	familyID uint64,
	familySurname string,
	actorID uint64,
	baseMemberID uint64,
	member *membermodel.FamilyMember,
	addType relationshipenum.AddType,
	requestedMemberType *string,
	input normalizedRelationshipInput,
) ([]*relationshipmodel.FamilyRelationship, error) {
	dbMember, err := repo.FindMemberForUpdate(ctx, familyID, member.ID)
	if err != nil {
		return nil, errMemberUnavailable
	}
	originalMemberType := normalizedStoredMemberType(dbMember.MemberType)
	workingMember := *dbMember

	baseMember, err := repo.FindMemberForUpdate(ctx, familyID, baseMemberID)
	if err != nil {
		return nil, errMemberUnavailable
	}
	if err := validatePlacementMembers(baseMember, &workingMember, addType); err != nil {
		return nil, err
	}

	switch addType {
	case relationshipenum.AddTypeFather, relationshipenum.AddTypeMother:
		if err := applyExplicitParentMemberTypeValue(ctx, repo, familyID, baseMember.ID, &workingMember, addType, requestedMemberType); err != nil {
			return nil, err
		}
	case relationshipenum.AddTypeSpouse:
		workingMember.MemberType = "SPOUSE"
	}

	var plans []relationshipPlan
	switch addType {
	case relationshipenum.AddTypeFather:
		input = applyParentPlacement(ctx, repo, familyID, familySurname, baseMember.ID, workingMember.Gender, input, "STEP_FATHER", "继父")
		plans = []relationshipPlan{{from: &workingMember, to: baseMember, relationshipType: string(relationshipenum.RelationshipTypeParentChild), input: input}}
		if err := ensurePrimaryParentAvailable(ctx, repo, familyID, baseMember.ID, workingMember.Gender, input.parentLinkType, 0); err != nil {
			return nil, err
		}
		if extraPlans, err := spousePlansForSpouseParent(ctx, repo, familyID, baseMember.ID, &workingMember, workingMember.Gender); err != nil {
			return nil, err
		} else {
			plans = append(plans, extraPlans...)
		}
	case relationshipenum.AddTypeMother:
		input = applyParentPlacement(ctx, repo, familyID, familySurname, baseMember.ID, workingMember.Gender, input, "STEP_MOTHER", "继母")
		plans = []relationshipPlan{{from: &workingMember, to: baseMember, relationshipType: string(relationshipenum.RelationshipTypeParentChild), input: input}}
		if err := ensurePrimaryParentAvailable(ctx, repo, familyID, baseMember.ID, workingMember.Gender, input.parentLinkType, 0); err != nil {
			return nil, err
		}
		if extraPlans, err := spousePlansForSpouseParent(ctx, repo, familyID, baseMember.ID, &workingMember, workingMember.Gender); err != nil {
			return nil, err
		} else {
			plans = append(plans, extraPlans...)
		}
	case relationshipenum.AddTypeChild:
		plans = []relationshipPlan{{from: baseMember, to: &workingMember, relationshipType: string(relationshipenum.RelationshipTypeParentChild), input: input}}
	case relationshipenum.AddTypeSpouse:
		input, err = applySpousePlacement(ctx, repo, familyID, familySurname, baseMember.ID, workingMember.Gender, input)
		if err != nil {
			return nil, err
		}
		plans = []relationshipPlan{{from: baseMember, to: &workingMember, relationshipType: string(relationshipenum.RelationshipTypeSpouse), input: input}}
	case relationshipenum.AddTypeSibling:
		parents, err := repo.ListActiveParents(ctx, familyID, baseMember.ID)
		if err != nil {
			return nil, err
		}
		if len(parents) == 0 {
			return nil, errSiblingParentRequired
		}
		plans = make([]relationshipPlan, 0, len(parents))
		for i := range parents {
			parent, err := repo.FindMemberForUpdate(ctx, familyID, parents[i].FromMemberID)
			if err != nil {
				return nil, errMemberUnavailable
			}
			parentInput := input
			parentInput.parentLinkType = parents[i].ParentLinkType
			plans = append(plans, relationshipPlan{
				from: parent, to: &workingMember, relationshipType: string(relationshipenum.RelationshipTypeParentChild), input: parentInput,
			})
		}
	}

	created := make([]*relationshipmodel.FamilyRelationship, 0, len(plans))
	for i := range plans {
		if plans[i].from.ID == plans[i].to.ID {
			return nil, errSelfRelationship
		}
		if _, err := repo.FindDuplicate(ctx, familyID, plans[i].from.ID, plans[i].to.ID, plans[i].relationshipType); err == nil {
			if plans[i].relationshipType == string(relationshipenum.RelationshipTypeSpouse) {
				continue
			}
			return nil, errDuplicateRelationship
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		relationship := plans[i].model(familyID, actorID)
		if err := repo.CreateRelationship(ctx, relationship); err != nil {
			return nil, err
		}
		created = append(created, relationship)
	}

	updatedMemberType := normalizedStoredMemberType(workingMember.MemberType)
	if updatedMemberType != originalMemberType {
		if err := repo.UpdateMember(ctx, familyID, workingMember.ID, map[string]any{"member_type": workingMember.MemberType}); err != nil {
			return nil, err
		}
	}

	*member = workingMember
	return created, nil
}

func applyExplicitParentMemberType(
	ctx context.Context,
	repo PlacementRepository,
	familyID uint64,
	childID uint64,
	member *membermodel.FamilyMember,
	addType relationshipenum.AddType,
	input dto.NewMemberInput,
) error {
	return applyExplicitParentMemberTypeValue(ctx, repo, familyID, childID, member, addType, input.MemberType)
}

func applyExplicitParentMemberTypeValue(
	ctx context.Context,
	repo PlacementRepository,
	familyID uint64,
	childID uint64,
	member *membermodel.FamilyMember,
	addType relationshipenum.AddType,
	memberType *string,
) error {
	if addType != relationshipenum.AddTypeFather && addType != relationshipenum.AddTypeMother {
		return nil
	}
	if memberType == nil || strings.TrimSpace(*memberType) == "" {
		return errParentMemberTypeRequired
	}
	requested := strings.ToUpper(strings.TrimSpace(*memberType))
	switch requested {
	case "LINEAGE_MEMBER":
		member.MemberType = "LINEAGE_MEMBER"
		return nil
	case "SPOUSE":
		member.MemberType = "SPOUSE"
		opposite := oppositeGender(member.Gender)
		_, existingParent, err := repo.FindPrimaryParentWithMemberByGender(ctx, familyID, childID, opposite, 0)
		if errors.Is(err, gorm.ErrRecordNotFound) || existingParent == nil {
			return errLineageParentRequired
		}
		if err != nil {
			return err
		}
		if normalizedStoredMemberType(existingParent.MemberType) != "LINEAGE_MEMBER" {
			return errLineageParentRequired
		}
		return nil
	default:
		return errParentMemberTypeInvalid
	}
}

func validatePlacementMembers(baseMember, member *membermodel.FamilyMember, addType relationshipenum.AddType) error {
	if normalizedStoredMemberType(baseMember.MemberType) == "SPOUSE" {
		switch addType {
		case relationshipenum.AddTypeFather, relationshipenum.AddTypeMother, relationshipenum.AddTypeChild,
			relationshipenum.AddTypeSpouse, relationshipenum.AddTypeSibling:
			return errSpouseBaseExpansion
		}
	}
	switch addType {
	case relationshipenum.AddTypeFather:
		if member.Gender != string(enums.GenderMale) {
			return errFatherGenderRequired
		}
	case relationshipenum.AddTypeMother:
		if member.Gender != string(enums.GenderFemale) {
			return errMotherGenderRequired
		}
	case relationshipenum.AddTypeChild:
		if baseMember.Gender != string(enums.GenderMale) && baseMember.Gender != string(enums.GenderFemale) {
			return errBaseGenderRequired
		}
		if member.Gender != string(enums.GenderMale) && member.Gender != string(enums.GenderFemale) {
			return errChildGenderRequired
		}
	case relationshipenum.AddTypeSpouse:
		if baseMember.Gender != string(enums.GenderMale) && baseMember.Gender != string(enums.GenderFemale) {
			return errBaseGenderRequired
		}
		if member.Gender != oppositeGender(baseMember.Gender) {
			return errSpouseGenderMismatch
		}
	case relationshipenum.AddTypeSibling:
		if member.Gender != string(enums.GenderMale) && member.Gender != string(enums.GenderFemale) {
			return errSiblingGenderRequired
		}
	}
	return nil
}

func spousePlansForSpouseParent(
	ctx context.Context,
	repo PlacementRepository,
	familyID uint64,
	childID uint64,
	member *membermodel.FamilyMember,
	gender string,
) ([]relationshipPlan, error) {
	if normalizedStoredMemberType(member.MemberType) != "SPOUSE" {
		return nil, nil
	}
	opposite := oppositeGender(gender)
	_, existingParent, err := repo.FindPrimaryParentWithMemberByGender(ctx, familyID, childID, opposite, 0)
	if errors.Is(err, gorm.ErrRecordNotFound) || existingParent == nil {
		return nil, errLineageParentRequired
	}
	if err != nil {
		return nil, err
	}
	if normalizedStoredMemberType(existingParent.MemberType) != "LINEAGE_MEMBER" {
		return nil, errLineageParentRequired
	}
	return []relationshipPlan{{
		from: existingParent, to: member,
		relationshipType: string(relationshipenum.RelationshipTypeSpouse),
		input:            normalizedRelationshipInput{},
	}}, nil
}

// MapPlacementError preserves the public relationship error codes when the
// placement workflow is reused by another transactional service.
func MapPlacementError(err error) *apperrors.BusinessError {
	return mapRepositoryError(err)
}

func asPlacementBusinessError(err error) *apperrors.BusinessError {
	var businessErr *apperrors.BusinessError
	if errors.As(err, &businessErr) {
		return businessErr
	}
	return nil
}
