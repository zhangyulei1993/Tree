package service

import (
	"context"
	"errors"

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
	return placeExistingMember(ctx, repo, familyID, familySurname, actorID, baseMemberID, member, addType, input)
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
	input normalizedRelationshipInput,
) ([]*relationshipmodel.FamilyRelationship, error) {
	baseMember, err := repo.FindMemberForUpdate(ctx, familyID, baseMemberID)
	if err != nil {
		return nil, errMemberUnavailable
	}
	if err := validatePlacementMembers(baseMember, member, addType); err != nil {
		return nil, err
	}

	var plans []relationshipPlan
	switch addType {
	case relationshipenum.AddTypeFather:
		input = applyParentPlacement(ctx, repo, familyID, familySurname, baseMember.ID, member.Gender, input, "STEP_FATHER", "继父")
		plans = []relationshipPlan{{from: member, to: baseMember, relationshipType: string(relationshipenum.RelationshipTypeParentChild), input: input}}
		if err := ensurePrimaryParentAvailable(ctx, repo, familyID, baseMember.ID, member.Gender, input.parentLinkType, 0); err != nil {
			return nil, err
		}
	case relationshipenum.AddTypeMother:
		input = applyParentPlacement(ctx, repo, familyID, familySurname, baseMember.ID, member.Gender, input, "STEP_MOTHER", "继母")
		plans = []relationshipPlan{{from: member, to: baseMember, relationshipType: string(relationshipenum.RelationshipTypeParentChild), input: input}}
		if err := ensurePrimaryParentAvailable(ctx, repo, familyID, baseMember.ID, member.Gender, input.parentLinkType, 0); err != nil {
			return nil, err
		}
	case relationshipenum.AddTypeChild:
		plans = []relationshipPlan{{from: baseMember, to: member, relationshipType: string(relationshipenum.RelationshipTypeParentChild), input: input}}
	case relationshipenum.AddTypeSpouse:
		member.MemberType = "SPOUSE"
		input, err = applySpousePlacement(ctx, repo, familyID, familySurname, baseMember.ID, member.Gender, input)
		if err != nil {
			return nil, err
		}
		plans = []relationshipPlan{{from: baseMember, to: member, relationshipType: string(relationshipenum.RelationshipTypeSpouse), input: input}}
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
				from: parent, to: member, relationshipType: string(relationshipenum.RelationshipTypeParentChild), input: parentInput,
			})
		}
	}

	created := make([]*relationshipmodel.FamilyRelationship, 0, len(plans))
	for i := range plans {
		if plans[i].from.ID == plans[i].to.ID {
			return nil, errSelfRelationship
		}
		if _, err := repo.FindDuplicate(ctx, familyID, plans[i].from.ID, plans[i].to.ID, plans[i].relationshipType); err == nil {
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
	return created, nil
}

func validatePlacementMembers(baseMember, member *membermodel.FamilyMember, addType relationshipenum.AddType) error {
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
