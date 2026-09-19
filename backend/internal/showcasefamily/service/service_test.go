package service

import (
	"testing"

	showcasemodel "tree/backend/internal/showcasefamily/model"
)

func TestValidateSpouseLink(t *testing.T) {
	male := showcasemodel.ShowcaseFamilyMember{ID: 1, Gender: "MALE"}
	female := showcasemodel.ShowcaseFamilyMember{ID: 2, Gender: "FEMALE"}
	secondFemale := showcasemodel.ShowcaseFamilyMember{ID: 3, Gender: "FEMALE"}
	secondMale := showcasemodel.ShowcaseFamilyMember{ID: 4, Gender: "MALE"}

	rows := []showcasemodel.ShowcaseFamilyRelationship{{FromMemberID: male.ID, ToMemberID: female.ID, RelationshipType: "SPOUSE"}}

	if businessErr := validateSpouseLink(rows, male, secondFemale); businessErr != nil {
		t.Fatalf("male member should be able to add another spouse: %v", businessErr)
	}
	if businessErr := validateSpouseLink(rows, female, secondMale); businessErr == nil {
		t.Fatal("female member with an existing spouse must not add another spouse")
	}
	if businessErr := validateSpouseLink(nil, female, secondMale); businessErr != nil {
		t.Fatalf("female member without a spouse should be able to add a first spouse: %v", businessErr)
	}
	if businessErr := validateSpouseLink(rows, female, male); businessErr == nil {
		t.Fatal("existing spouse relationship must be rejected regardless of direction")
	}
}

func TestValidateParentLink(t *testing.T) {
	child := showcasemodel.ShowcaseFamilyMember{ID: 10}
	thirdParent := showcasemodel.ShowcaseFamilyMember{ID: 13}
	rows := []showcasemodel.ShowcaseFamilyRelationship{
		{FromMemberID: 11, ToMemberID: child.ID, RelationshipType: "PARENT_CHILD"},
		{FromMemberID: 12, ToMemberID: child.ID, RelationshipType: "PARENT_CHILD"},
	}

	if businessErr := validateParentLink(rows, thirdParent, child); businessErr == nil {
		t.Fatal("a child must not accept a third parent")
	}
}

func TestMemberRelationshipCount(t *testing.T) {
	rows := []showcasemodel.ShowcaseFamilyRelationship{
		{FromMemberID: 1, ToMemberID: 2, RelationshipType: "SPOUSE"},
		{FromMemberID: 1, ToMemberID: 3, RelationshipType: "PARENT_CHILD"},
		{FromMemberID: 2, ToMemberID: 4, RelationshipType: "PARENT_CHILD"},
	}

	if got := memberRelationshipCount(rows, 1); got != 2 {
		t.Fatalf("member 1 relationship count = %d, want 2", got)
	}
	if got := memberRelationshipCount(rows, 5); got != 0 {
		t.Fatalf("unrelated member relationship count = %d, want 0", got)
	}
}

func TestDescendantCount(t *testing.T) {
	rows := []showcasemodel.ShowcaseFamilyRelationship{
		{FromMemberID: 1, ToMemberID: 2, RelationshipType: "PARENT_CHILD"},
		{FromMemberID: 1, ToMemberID: 3, RelationshipType: "PARENT_CHILD"},
		{FromMemberID: 2, ToMemberID: 4, RelationshipType: "PARENT_CHILD"},
		{FromMemberID: 1, ToMemberID: 5, RelationshipType: "SPOUSE"},
	}

	if got := descendantCount(rows, 1); got != 2 {
		t.Fatalf("member 1 descendant count = %d, want 2", got)
	}
	if got := descendantCount(rows, 5); got != 0 {
		t.Fatalf("spouse member descendant count = %d, want 0", got)
	}
}

func TestHasChildWithoutOtherLineageParent(t *testing.T) {
	rows := []showcasemodel.ShowcaseFamilyRelationship{
		{FromMemberID: 2, ToMemberID: 10, RelationshipType: "PARENT_CHILD"},
		{FromMemberID: 1, ToMemberID: 10, RelationshipType: "PARENT_CHILD"},
	}
	members := map[uint64]showcasemodel.ShowcaseFamilyMember{
		1: {ID: 1, MemberType: "LINEAGE_MEMBER"},
		2: {ID: 2, MemberType: "SPOUSE"},
	}
	if hasChildWithoutOtherLineageParent(rows, members, 2) {
		t.Fatal("external spouse child with a lineage co-parent should remain deletable")
	}
	if !hasChildWithoutOtherLineageParent(rows[:1], members, 2) {
		t.Fatal("external spouse child without a lineage co-parent must block deletion")
	}
}
