package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/enums"
	familymodel "tree/backend/internal/family/core/model"
	membermodel "tree/backend/internal/family/member/model"
	"tree/backend/internal/family/relationship/dto"
	relationshipmodel "tree/backend/internal/family/relationship/model"
	relationshiprepo "tree/backend/internal/family/relationship/repository"
	operationlog "tree/backend/internal/operationlog/service"
)

type fakePermission struct {
	allowed bool
}

func (p fakePermission) CanManageRelationships(context.Context, uint64, uint64) (bool, error) {
	return p.allowed, nil
}

type fakeUnitOfWork struct {
	repo relationshiprepo.Repository
}

func (u fakeUnitOfWork) WithinTransaction(ctx context.Context, fn func(relationshiprepo.Repository) error) error {
	return fn(u.repo)
}

type fakeRepository struct {
	family         familymodel.Family
	members        map[uint64]*membermodel.FamilyMember
	relationships  map[uint64]*relationshipmodel.FamilyRelationship
	nextMemberID   uint64
	nextRelationID uint64
	graphVersion   int64
	logActions     []string
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		family: familymodel.Family{ID: 2, Status: string(enums.StatusNormal), GraphVersion: 1},
		members: map[uint64]*membermodel.FamilyMember{
			3: {ID: 3, FamilyID: 2, DisplayName: "Base", Gender: string(enums.GenderMale), Status: string(enums.StatusActive)},
		},
		relationships:  make(map[uint64]*relationshipmodel.FamilyRelationship),
		nextMemberID:   10,
		nextRelationID: 20,
		graphVersion:   1,
	}
}

func (r *fakeRepository) WithTx(*gorm.DB) relationshiprepo.Repository { return r }

func (r *fakeRepository) LockFamily(context.Context, uint64) (*familymodel.Family, error) {
	return &r.family, nil
}

func (r *fakeRepository) FindMemberForUpdate(_ context.Context, familyID uint64, memberID uint64) (*membermodel.FamilyMember, error) {
	member, ok := r.members[memberID]
	if !ok || member.FamilyID != familyID || member.Status != string(enums.StatusActive) || member.DeletedAt != nil {
		return nil, gorm.ErrRecordNotFound
	}
	return member, nil
}

func (r *fakeRepository) CreateMember(_ context.Context, member *membermodel.FamilyMember) error {
	r.nextMemberID++
	member.ID = r.nextMemberID
	member.CreatedAt = time.Now()
	member.UpdatedAt = member.CreatedAt
	r.members[member.ID] = member
	return nil
}

func (r *fakeRepository) FindActiveRelationship(_ context.Context, familyID uint64, relationshipID uint64) (*relationshipmodel.FamilyRelationship, error) {
	relationship, ok := r.relationships[relationshipID]
	if !ok || relationship.FamilyID != familyID || relationship.Status != string(enums.StatusActive) || relationship.DeletedAt != nil {
		return nil, gorm.ErrRecordNotFound
	}
	return relationship, nil
}

func (r *fakeRepository) FindDuplicate(_ context.Context, familyID uint64, fromMemberID uint64, toMemberID uint64, relationshipType string) (*relationshipmodel.FamilyRelationship, error) {
	for _, relationship := range r.relationships {
		if relationship.FamilyID != familyID || relationship.RelationshipType != relationshipType ||
			relationship.Status != string(enums.StatusActive) || relationship.DeletedAt != nil {
			continue
		}
		direct := relationship.FromMemberID == fromMemberID && relationship.ToMemberID == toMemberID
		reverseSpouse := relationshipType == string(enums.RelationshipTypeSpouse) &&
			relationship.FromMemberID == toMemberID && relationship.ToMemberID == fromMemberID
		if direct || reverseSpouse {
			return relationship, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeRepository) ListActiveParents(_ context.Context, familyID uint64, childMemberID uint64) ([]relationshipmodel.FamilyRelationship, error) {
	result := make([]relationshipmodel.FamilyRelationship, 0)
	for _, relationship := range r.relationships {
		if relationship.FamilyID == familyID && relationship.ToMemberID == childMemberID &&
			relationship.RelationshipType == string(enums.RelationshipTypeParentChild) &&
			relationship.Status == string(enums.StatusActive) && relationship.DeletedAt == nil {
			result = append(result, *relationship)
		}
	}
	return result, nil
}

func (r *fakeRepository) FindPrimaryParentByGender(_ context.Context, familyID uint64, childMemberID uint64, gender string, excludeRelationshipID uint64) (*relationshipmodel.FamilyRelationship, error) {
	relationship, _, err := r.FindPrimaryParentWithMemberByGender(context.Background(), familyID, childMemberID, gender, excludeRelationshipID)
	return relationship, err
}

func (r *fakeRepository) FindPrimaryParentWithMemberByGender(_ context.Context, familyID uint64, childMemberID uint64, gender string, excludeRelationshipID uint64) (*relationshipmodel.FamilyRelationship, *membermodel.FamilyMember, error) {
	for _, relationship := range r.relationships {
		parent := r.members[relationship.FromMemberID]
		if relationship.ID != excludeRelationshipID &&
			relationship.FamilyID == familyID && relationship.ToMemberID == childMemberID &&
			relationship.RelationshipType == string(enums.RelationshipTypeParentChild) &&
			relationship.ParentLinkType != nil && *relationship.ParentLinkType == "PRIMARY" &&
			relationship.Status == string(enums.StatusActive) && relationship.DeletedAt == nil &&
			parent != nil && parent.Gender == gender {
			return relationship, parent, nil
		}
	}
	return nil, nil, gorm.ErrRecordNotFound
}

func (r *fakeRepository) ListActiveSpouseMembersByGender(_ context.Context, familyID uint64, baseMemberID uint64, gender string) ([]membermodel.FamilyMember, error) {
	rows, err := r.ListActiveSpouseRelationshipsByGender(context.Background(), familyID, baseMemberID, gender)
	if err != nil {
		return nil, err
	}
	result := make([]membermodel.FamilyMember, 0, len(rows))
	for i := range rows {
		result = append(result, rows[i].Spouse)
	}
	return result, nil
}

func (r *fakeRepository) ListActiveSpouseRelationshipsByGender(_ context.Context, familyID uint64, baseMemberID uint64, gender string) ([]relationshiprepo.SpouseRelationshipRow, error) {
	result := make([]relationshiprepo.SpouseRelationshipRow, 0)
	for _, relationship := range r.relationships {
		if relationship.FamilyID != familyID || relationship.RelationshipType != string(enums.RelationshipTypeSpouse) ||
			relationship.Status != string(enums.StatusActive) || relationship.DeletedAt != nil {
			continue
		}
		var spouseID uint64
		switch {
		case relationship.FromMemberID == baseMemberID:
			spouseID = relationship.ToMemberID
		case relationship.ToMemberID == baseMemberID:
			spouseID = relationship.FromMemberID
		default:
			continue
		}
		member := r.members[spouseID]
		if member != nil && member.Gender == gender && member.Status == string(enums.StatusActive) && member.DeletedAt == nil {
			result = append(result, relationshiprepo.SpouseRelationshipRow{Relationship: *relationship, Spouse: *member})
		}
	}
	return result, nil
}

func (r *fakeRepository) CreateRelationship(_ context.Context, relationship *relationshipmodel.FamilyRelationship) error {
	r.nextRelationID++
	relationship.ID = r.nextRelationID
	relationship.CreatedAt = time.Now()
	relationship.UpdatedAt = relationship.CreatedAt
	r.relationships[relationship.ID] = relationship
	return nil
}

func (r *fakeRepository) UpdateRelationship(_ context.Context, familyID uint64, relationshipID uint64, values map[string]any) error {
	relationship, err := r.FindActiveRelationship(context.Background(), familyID, relationshipID)
	if err != nil {
		return err
	}
	applyRelationshipValues(relationship, values)
	return nil
}

func (r *fakeRepository) SoftDeleteRelationship(_ context.Context, familyID uint64, relationshipID uint64, actorID uint64, reason *string, now time.Time) error {
	relationship, err := r.FindActiveRelationship(context.Background(), familyID, relationshipID)
	if err != nil {
		return err
	}
	relationship.Status = "DELETED"
	relationship.DeletedAt = &now
	relationship.DeletedByUserID = &actorID
	relationship.DeleteReason = reason
	return nil
}

func (r *fakeRepository) IncrementGraphVersion(context.Context, uint64) (int64, error) {
	r.graphVersion++
	r.family.GraphVersion = r.graphVersion
	return r.graphVersion, nil
}

func (r *fakeRepository) WriteOperationLog(_ context.Context, input operationlog.WriteInput) error {
	r.logActions = append(r.logActions, input.Action)
	return nil
}

func testService(repo *fakeRepository, allowed bool) RelationshipService {
	return NewRelationshipService(repo, fakeUnitOfWork{repo: repo}, fakePermission{allowed: allowed})
}

func createRequest(addType string, name string) dto.CreateRelationshipRequest {
	gender := string(enums.GenderMale)
	if addType == "ADD_MOTHER" || addType == "ADD_SPOUSE" {
		gender = string(enums.GenderFemale)
	}
	return createRequestWithGender(addType, name, gender)
}

func createRequestWithGender(addType string, name string, gender string) dto.CreateRelationshipRequest {
	return dto.CreateRelationshipRequest{
		BaseMemberID: 3,
		AddType:      addType,
		NewMember:    dto.NewMemberInput{Name: name, Gender: &gender},
	}
}

func TestAddChildCreatesParentChildInCorrectDirection(t *testing.T) {
	repo := newFakeRepository()
	result, businessErr := testService(repo, true).Create(context.Background(), 8, 2, createRequest("ADD_CHILD", "Child"), AuditInput{})
	if businessErr != nil {
		t.Fatalf("Create returned error: %v", businessErr)
	}
	relationship := result.Relationships[0]
	if relationship.RelationshipType != "PARENT_CHILD" || relationship.FromMemberID != 3 ||
		relationship.ToMemberID != result.CreatedMember.MemberID {
		t.Fatalf("unexpected child direction: %#v", relationship)
	}
	if result.GraphVersion != 2 {
		t.Fatalf("expected graph version 2, got %d", result.GraphVersion)
	}
}

func TestAddFatherAndMotherDirections(t *testing.T) {
	tests := []struct {
		addType string
		gender  string
	}{
		{addType: "ADD_FATHER", gender: "MALE"},
		{addType: "ADD_MOTHER", gender: "FEMALE"},
	}
	for _, tt := range tests {
		t.Run(tt.addType, func(t *testing.T) {
			repo := newFakeRepository()
			result, businessErr := testService(repo, true).Create(context.Background(), 8, 2, createRequest(tt.addType, tt.addType), AuditInput{})
			if businessErr != nil {
				t.Fatalf("Create returned error: %v", businessErr)
			}
			relationship := result.Relationships[0]
			if relationship.FromMemberID != result.CreatedMember.MemberID || relationship.ToMemberID != 3 ||
				result.CreatedMember.Gender != tt.gender {
				t.Fatalf("unexpected parent direction: %#v", result)
			}
		})
	}
}

func TestAddParentRejectsMismatchedGender(t *testing.T) {
	tests := []struct {
		name    string
		addType string
		gender  string
		message string
	}{
		{name: "female cannot be father", addType: "ADD_FATHER", gender: string(enums.GenderFemale), message: "父亲成员性别必须为男"},
		{name: "male cannot be mother", addType: "ADD_MOTHER", gender: string(enums.GenderMale), message: "母亲成员性别必须为女"},
		{name: "unknown cannot be father", addType: "ADD_FATHER", gender: string(enums.GenderUnknown), message: "父亲成员性别必须为男"},
		{name: "unknown cannot be mother", addType: "ADD_MOTHER", gender: string(enums.GenderUnknown), message: "母亲成员性别必须为女"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			_, businessErr := testService(repo, true).Create(context.Background(), 8, 2, createRequestWithGender(tt.addType, "Wrong Gender", tt.gender), AuditInput{})
			if businessErr == nil || businessErr.Code != CodeRelationshipMember || businessErr.Message != tt.message {
				t.Fatalf("expected %s, got %#v", tt.message, businessErr)
			}
			if len(repo.members) != 1 {
				t.Fatalf("mismatched parent gender must not create member, got %d members", len(repo.members))
			}
		})
	}
}

func TestAddSpouseRejectsMismatchedOrUnknownGender(t *testing.T) {
	tests := []struct {
		name   string
		gender string
	}{
		{name: "male cannot be wife of male base", gender: string(enums.GenderMale)},
		{name: "unknown cannot be spouse", gender: string(enums.GenderUnknown)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			_, businessErr := testService(repo, true).Create(context.Background(), 8, 2, createRequestWithGender("ADD_SPOUSE", "Wrong Spouse", tt.gender), AuditInput{})
			if businessErr == nil || businessErr.Code != CodeRelationshipMember || businessErr.Message != "配偶性别必须与基准成员相对" {
				t.Fatalf("expected spouse gender mismatch, got %#v", businessErr)
			}
			if len(repo.members) != 1 {
				t.Fatalf("mismatched spouse gender must not create member, got %d members", len(repo.members))
			}
		})
	}
}

func TestDuplicatePrimaryParentsReturnSpecificErrors(t *testing.T) {
	tests := []struct {
		addType string
		gender  string
		code    int
	}{
		{addType: "ADD_FATHER", gender: "MALE", code: int(CodePrimaryFatherExists)},
		{addType: "ADD_MOTHER", gender: "FEMALE", code: int(CodePrimaryMotherExists)},
	}
	for _, tt := range tests {
		t.Run(tt.addType, func(t *testing.T) {
			repo := newFakeRepository()
			parentID := uint64(9)
			repo.members[parentID] = &membermodel.FamilyMember{
				ID: parentID, FamilyID: 2, DisplayName: "Existing", Gender: tt.gender, Status: string(enums.StatusActive),
			}
			primary := "PRIMARY"
			repo.relationships[19] = &relationshipmodel.FamilyRelationship{
				ID: 19, FamilyID: 2, FromMemberID: parentID, ToMemberID: 3,
				RelationshipType: "PARENT_CHILD", ParentLinkType: &primary, Status: string(enums.StatusActive),
			}
			_, businessErr := testService(repo, true).Create(context.Background(), 8, 2, createRequest(tt.addType, "Duplicate"), AuditInput{})
			if businessErr == nil || int(businessErr.Code) != tt.code {
				t.Fatalf("expected code %d, got %#v", tt.code, businessErr)
			}
		})
	}
}

func TestSpouseLikeExistingParentAllowsStepParent(t *testing.T) {
	tests := []struct {
		name     string
		addType  string
		gender   string
		noteType string
	}{
		{name: "step father", addType: "ADD_FATHER", gender: "MALE", noteType: "STEP_FATHER"},
		{name: "step mother", addType: "ADD_MOTHER", gender: "FEMALE", noteType: "STEP_MOTHER"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			parentID := uint64(9)
			repo.members[parentID] = &membermodel.FamilyMember{
				ID: parentID, FamilyID: 2, DisplayName: "Existing spouse-like parent",
				Gender: tt.gender, MemberType: "SPOUSE", Status: string(enums.StatusActive),
			}
			primary := "PRIMARY"
			repo.relationships[19] = &relationshipmodel.FamilyRelationship{
				ID: 19, FamilyID: 2, FromMemberID: parentID, ToMemberID: 3,
				RelationshipType: "PARENT_CHILD", ParentLinkType: &primary, Status: string(enums.StatusActive),
			}

			result, businessErr := testService(repo, true).Create(context.Background(), 8, 2, createRequest(tt.addType, "Step Parent"), AuditInput{})
			if businessErr != nil {
				t.Fatalf("Create returned error: %v", businessErr)
			}
			relationship := result.Relationships[0]
			if relationship.ParentLinkType == nil || *relationship.ParentLinkType != "STEP" {
				t.Fatalf("expected STEP parent link, got %#v", relationship.ParentLinkType)
			}
			if relationship.RelationNoteType == nil || *relationship.RelationNoteType != tt.noteType {
				t.Fatalf("expected note type %s, got %#v", tt.noteType, relationship.RelationNoteType)
			}
		})
	}
}

func TestNonFamilySurnameExistingParentAllowsStepParent(t *testing.T) {
	repo := newFakeRepository()
	repo.family.FamilySurname = "张"
	parentID := uint64(9)
	repo.members[parentID] = &membermodel.FamilyMember{
		ID: parentID, FamilyID: 2, DisplayName: "李某", Gender: string(enums.GenderFemale),
		MemberType: "LINEAGE_MEMBER", Status: string(enums.StatusActive),
	}
	primary := "PRIMARY"
	repo.relationships[19] = &relationshipmodel.FamilyRelationship{
		ID: 19, FamilyID: 2, FromMemberID: parentID, ToMemberID: 3,
		RelationshipType: "PARENT_CHILD", ParentLinkType: &primary, Status: string(enums.StatusActive),
	}

	result, businessErr := testService(repo, true).Create(context.Background(), 8, 2, createRequest("ADD_MOTHER", "Step Mother"), AuditInput{})
	if businessErr != nil {
		t.Fatalf("Create returned error: %v", businessErr)
	}
	relationship := result.Relationships[0]
	if relationship.ParentLinkType == nil || *relationship.ParentLinkType != "STEP" {
		t.Fatalf("expected STEP parent link, got %#v", relationship.ParentLinkType)
	}
	if relationship.RelationNoteType == nil || *relationship.RelationNoteType != "STEP_MOTHER" {
		t.Fatalf("expected STEP_MOTHER note, got %#v", relationship.RelationNoteType)
	}
}

func TestAddSpouseAllowsMultipleSpouses(t *testing.T) {
	repo := newFakeRepository()
	service := testService(repo, true)
	for _, name := range []string{"Spouse One", "Spouse Two"} {
		result, businessErr := service.Create(context.Background(), 8, 2, createRequest("ADD_SPOUSE", name), AuditInput{})
		if businessErr != nil {
			t.Fatalf("Create %s returned error: %v", name, businessErr)
		}
		if result.Relationships[0].RelationshipType != "SPOUSE" {
			t.Fatalf("expected SPOUSE, got %#v", result.Relationships[0])
		}
	}
	if len(repo.relationships) != 2 {
		t.Fatalf("expected two spouse relationships, got %d", len(repo.relationships))
	}
}

func TestLineageExistingSpouseBlocksDuplicateSpouse(t *testing.T) {
	repo := newFakeRepository()
	wifeID := uint64(9)
	repo.members[wifeID] = &membermodel.FamilyMember{
		ID: wifeID, FamilyID: 2, DisplayName: "Lineage Wife",
		Gender: string(enums.GenderFemale), MemberType: "LINEAGE_MEMBER", Status: string(enums.StatusActive),
	}
	repo.relationships[19] = &relationshipmodel.FamilyRelationship{
		ID: 19, FamilyID: 2, FromMemberID: 3, ToMemberID: wifeID,
		RelationshipType: "SPOUSE", Status: string(enums.StatusActive),
	}

	_, businessErr := testService(repo, true).Create(context.Background(), 8, 2, createRequest("ADD_SPOUSE", "New Wife"), AuditInput{})
	if businessErr == nil || businessErr.Code != CodeRelationshipType {
		t.Fatalf("expected duplicate lineage spouse rejection, got %#v", businessErr)
	}
}

func TestSpouseLikeExistingSpouseAllowsSecondSpouseNote(t *testing.T) {
	repo := newFakeRepository()
	wifeID := uint64(9)
	repo.members[wifeID] = &membermodel.FamilyMember{
		ID: wifeID, FamilyID: 2, DisplayName: "Spouse One",
		Gender: string(enums.GenderFemale), MemberType: "SPOUSE", Status: string(enums.StatusActive),
	}
	repo.relationships[19] = &relationshipmodel.FamilyRelationship{
		ID: 19, FamilyID: 2, FromMemberID: 3, ToMemberID: wifeID,
		RelationshipType: "SPOUSE", Status: string(enums.StatusActive),
	}

	result, businessErr := testService(repo, true).Create(context.Background(), 8, 2, createRequest("ADD_SPOUSE", "Spouse Two"), AuditInput{})
	if businessErr != nil {
		t.Fatalf("Create returned error: %v", businessErr)
	}
	relationship := result.Relationships[0]
	if relationship.RelationNoteType == nil || *relationship.RelationNoteType != "SECOND_WIFE" {
		t.Fatalf("expected SECOND_WIFE note, got %#v", relationship.RelationNoteType)
	}
	existing := repo.relationships[19]
	if existing.RelationNoteType == nil || *existing.RelationNoteType != "EX_WIFE" {
		t.Fatalf("expected existing wife marked EX_WIFE, got %#v", existing.RelationNoteType)
	}
}

func TestNonFamilySurnameExistingSpouseAllowsSecondSpouseAndMarksEx(t *testing.T) {
	repo := newFakeRepository()
	repo.family.FamilySurname = "张"
	wifeID := uint64(9)
	repo.members[wifeID] = &membermodel.FamilyMember{
		ID: wifeID, FamilyID: 2, DisplayName: "李某",
		Gender: string(enums.GenderFemale), MemberType: "LINEAGE_MEMBER", Status: string(enums.StatusActive),
	}
	repo.relationships[19] = &relationshipmodel.FamilyRelationship{
		ID: 19, FamilyID: 2, FromMemberID: 3, ToMemberID: wifeID,
		RelationshipType: "SPOUSE", Status: string(enums.StatusActive),
	}

	result, businessErr := testService(repo, true).Create(context.Background(), 8, 2, createRequest("ADD_SPOUSE", "New Wife"), AuditInput{})
	if businessErr != nil {
		t.Fatalf("Create returned error: %v", businessErr)
	}
	if result.Relationships[0].RelationNoteType == nil || *result.Relationships[0].RelationNoteType != "SECOND_WIFE" {
		t.Fatalf("expected new wife SECOND_WIFE, got %#v", result.Relationships[0].RelationNoteType)
	}
	if repo.relationships[19].RelationNoteType == nil || *repo.relationships[19].RelationNoteType != "EX_WIFE" {
		t.Fatalf("expected previous wife EX_WIFE, got %#v", repo.relationships[19].RelationNoteType)
	}
}

func TestAddSiblingRequiresAndReusesParents(t *testing.T) {
	t.Run("without parents", func(t *testing.T) {
		repo := newFakeRepository()
		_, businessErr := testService(repo, true).Create(context.Background(), 8, 2, createRequest("ADD_SIBLING", "Sibling"), AuditInput{})
		if businessErr == nil || businessErr.Code != CodeSiblingParentRequired {
			t.Fatalf("expected 43301, got %#v", businessErr)
		}
	})

	t.Run("with parents", func(t *testing.T) {
		repo := newFakeRepository()
		primary := "PRIMARY"
		for id, gender := range map[uint64]string{4: "MALE", 5: "FEMALE"} {
			repo.members[id] = &membermodel.FamilyMember{
				ID: id, FamilyID: 2, DisplayName: "Parent", Gender: gender, Status: string(enums.StatusActive),
			}
			repo.nextRelationID++
			repo.relationships[repo.nextRelationID] = &relationshipmodel.FamilyRelationship{
				ID: repo.nextRelationID, FamilyID: 2, FromMemberID: id, ToMemberID: 3,
				RelationshipType: "PARENT_CHILD", ParentLinkType: &primary, Status: string(enums.StatusActive),
			}
		}
		result, businessErr := testService(repo, true).Create(context.Background(), 8, 2, createRequest("ADD_SIBLING", "Sibling"), AuditInput{})
		if businessErr != nil {
			t.Fatalf("Create returned error: %v", businessErr)
		}
		if len(result.Relationships) != 2 {
			t.Fatalf("expected two shared-parent relationships, got %d", len(result.Relationships))
		}
		for _, relationship := range result.Relationships {
			if relationship.RelationshipType != "PARENT_CHILD" || relationship.ToMemberID != result.CreatedMember.MemberID {
				t.Fatalf("unexpected sibling relationship: %#v", relationship)
			}
		}
		if result.GraphVersion != 2 {
			t.Fatalf("ADD_SIBLING must increment once, got %d", result.GraphVersion)
		}
	})
}

func TestUpdateRejectsStructuralFieldsAndIncrementsGraphVersion(t *testing.T) {
	repo := newFakeRepository()
	relationship := seedRelationship(repo)
	newTarget := uint64(99)
	_, businessErr := testService(repo, true).Update(context.Background(), 8, 2, relationship.ID, dto.UpdateRelationshipRequest{
		ToMemberID: &newTarget,
	}, AuditInput{})
	if businessErr == nil || businessErr.Code != CodeRelationshipType {
		t.Fatalf("expected structural update rejection, got %#v", businessErr)
	}
	if repo.graphVersion != 1 {
		t.Fatalf("rejected update changed graph version")
	}

	note := "updated"
	result, businessErr := testService(repo, true).Update(context.Background(), 8, 2, relationship.ID, dto.UpdateRelationshipRequest{
		RelationNote: &note,
	}, AuditInput{})
	if businessErr != nil {
		t.Fatalf("Update returned error: %v", businessErr)
	}
	if result.GraphVersion != 2 || result.Relationships[0].RelationNote == nil || *result.Relationships[0].RelationNote != note {
		t.Fatalf("unexpected update result: %#v", result)
	}
}

func TestUpdateToPrimaryRejectsDuplicateFather(t *testing.T) {
	repo := newFakeRepository()
	step := "STEP"
	relationship := seedRelationship(repo)
	relationship.ParentLinkType = &step

	existingFatherID := uint64(4)
	repo.members[existingFatherID] = &membermodel.FamilyMember{
		ID: existingFatherID, FamilyID: 2, DisplayName: "Existing Father",
		Gender: string(enums.GenderMale), Status: string(enums.StatusActive),
	}
	primary := "PRIMARY"
	repo.relationships[21] = &relationshipmodel.FamilyRelationship{
		ID: 21, FamilyID: 2, FromMemberID: existingFatherID, ToMemberID: relationship.ToMemberID,
		RelationshipType: "PARENT_CHILD", ParentLinkType: &primary, Status: string(enums.StatusActive),
	}

	_, businessErr := testService(repo, true).Update(context.Background(), 8, 2, relationship.ID, dto.UpdateRelationshipRequest{
		ParentLinkType: &primary,
	}, AuditInput{})
	if businessErr == nil || businessErr.Code != CodePrimaryFatherExists {
		t.Fatalf("expected primary father conflict, got %#v", businessErr)
	}
	if repo.graphVersion != 1 {
		t.Fatalf("rejected primary update changed graph version")
	}
}

func TestDeleteSoftDeletesAndIncrementsGraphVersion(t *testing.T) {
	repo := newFakeRepository()
	relationship := seedRelationship(repo)
	result, businessErr := testService(repo, true).Delete(context.Background(), 8, 2, relationship.ID, dto.DeleteRelationshipRequest{}, AuditInput{})
	if businessErr != nil {
		t.Fatalf("Delete returned error: %v", businessErr)
	}
	deleted := repo.relationships[relationship.ID]
	if deleted.Status != "DELETED" || deleted.DeletedAt == nil {
		t.Fatalf("relationship was not soft deleted: %#v", deleted)
	}
	if result.GraphVersion != 2 {
		t.Fatalf("expected graph version 2, got %d", result.GraphVersion)
	}
}

func TestPermissionDenied(t *testing.T) {
	repo := newFakeRepository()
	_, businessErr := testService(repo, false).Create(context.Background(), 9, 2, createRequest("ADD_CHILD", "Child"), AuditInput{})
	if businessErr == nil || businessErr.Code != CodeRelationshipForbidden {
		t.Fatalf("expected permission error, got %#v", businessErr)
	}
	if len(repo.members) != 1 || repo.graphVersion != 1 {
		t.Fatalf("permission denial changed data")
	}
}

func seedRelationship(repo *fakeRepository) *relationshipmodel.FamilyRelationship {
	primary := "PRIMARY"
	relationship := &relationshipmodel.FamilyRelationship{
		ID: 20, FamilyID: 2, FromMemberID: 3, ToMemberID: 10,
		RelationshipType: "PARENT_CHILD", ParentLinkType: &primary, Status: string(enums.StatusActive),
	}
	repo.members[10] = &membermodel.FamilyMember{
		ID: 10, FamilyID: 2, DisplayName: "Child", Gender: "UNKNOWN", Status: string(enums.StatusActive),
	}
	repo.relationships[relationship.ID] = relationship
	return relationship
}
