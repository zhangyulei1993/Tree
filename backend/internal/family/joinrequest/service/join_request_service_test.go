package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/contentsafety"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/family/core/model"
	"tree/backend/internal/family/joinrequest/dto"
	joinmodel "tree/backend/internal/family/joinrequest/model"
	joinrepo "tree/backend/internal/family/joinrequest/repository"
	membermodel "tree/backend/internal/family/member/model"
	relationshipdto "tree/backend/internal/family/relationship/dto"
	relationshipmodel "tree/backend/internal/family/relationship/model"
	relationshiprepo "tree/backend/internal/family/relationship/repository"
	rolemodel "tree/backend/internal/family/role/model"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"
)

func createJoinRequestInput() dto.CreateJoinRequest {
	gender := "MALE"
	return dto.CreateJoinRequest{ApplicantGender: &gender}
}

type fakePermission struct{ allowed bool }

func (p fakePermission) IsFamilyAdmin(context.Context, uint64, uint64) (bool, error) {
	return p.allowed, nil
}
func (p fakePermission) IsFamilyMember(context.Context, uint64, uint64) (bool, error) {
	return p.allowed, nil
}
func (p fakePermission) CanManageFamily(context.Context, uint64, uint64) (bool, error) {
	return p.allowed, nil
}
func (p fakePermission) CanCreateMember(context.Context, uint64, uint64) (bool, error) {
	return p.allowed, nil
}
func (p fakePermission) CanEditMember(context.Context, uint64, uint64, uint64) (bool, error) {
	return p.allowed, nil
}
func (p fakePermission) CanDeleteMember(context.Context, uint64, uint64, uint64) (bool, error) {
	return p.allowed, nil
}
func (p fakePermission) CanManageRelationships(context.Context, uint64, uint64) (bool, error) {
	return p.allowed, nil
}
func (p fakePermission) GetActiveLink(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error) {
	return nil, gorm.ErrRecordNotFound
}

type fakeRepo struct {
	family                                    model.Family
	users                                     map[uint64]usermodel.User
	members                                   map[uint64]*membermodel.FamilyMember
	requests                                  map[uint64]*joinmodel.FamilyJoinRequest
	links                                     []rolemodel.FamilyMemberUserLink
	relationships                             map[uint64]*relationshipmodel.FamilyRelationship
	logs                                      []operationlog.WriteInput
	nextRequest, nextMember, nextRelationship uint64
	pendingInvitations                        int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		family:   model.Family{ID: 2, FamilyName: "Tree", Status: "NORMAL", GraphVersion: 10},
		users:    map[uint64]usermodel.User{8: {ID: 8, Status: "ACTIVE", PhoneVerified: true}, 9: {ID: 9, Status: "ACTIVE", PhoneVerified: true}},
		members:  map[uint64]*membermodel.FamilyMember{6: {ID: 6, FamilyID: 2, DisplayName: "Existing", Status: "ACTIVE", UserBindingPolicy: "OPTIONAL"}},
		requests: map[uint64]*joinmodel.FamilyJoinRequest{}, relationships: map[uint64]*relationshipmodel.FamilyRelationship{},
		nextRequest: 1, nextMember: 20, nextRelationship: 30,
	}
}
func (r *fakeRepo) WithTx(*gorm.DB) joinrepo.Repository { return r }
func (r *fakeRepo) DB() *gorm.DB                        { return nil }
func (r *fakeRepo) FindFamily(context.Context, uint64, bool) (*model.Family, error) {
	v := r.family
	return &v, nil
}
func (r *fakeRepo) FindUser(_ context.Context, id uint64, _ bool) (*usermodel.User, error) {
	v, ok := r.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &v, nil
}
func (r *fakeRepo) FindMember(_ context.Context, familyID, id uint64, _ bool) (*membermodel.FamilyMember, error) {
	v, ok := r.members[id]
	if !ok || v.FamilyID != familyID {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *v
	return &copy, nil
}
func (r *fakeRepo) FindActiveLinkByUser(_ context.Context, familyID, userID uint64) (*rolemodel.FamilyMemberUserLink, error) {
	for i := range r.links {
		if r.links[i].FamilyID == familyID && r.links[i].UserID == userID && r.links[i].LinkStatus == "ACTIVE" {
			return &r.links[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeRepo) FindActiveLinkByMember(_ context.Context, familyID, memberID uint64) (*rolemodel.FamilyMemberUserLink, error) {
	for i := range r.links {
		if r.links[i].FamilyID == familyID && r.links[i].MemberID == memberID && r.links[i].LinkStatus == "ACTIVE" {
			return &r.links[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeRepo) FindPending(_ context.Context, familyID, userID uint64) (*joinmodel.FamilyJoinRequest, error) {
	for _, v := range r.requests {
		if v.FamilyID == familyID && v.ApplicantUserID == userID && v.RequestStatus == "PENDING" {
			return v, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeRepo) Create(_ context.Context, value *joinmodel.FamilyJoinRequest) error {
	value.ID = r.nextRequest
	r.nextRequest++
	value.CreatedAt = time.Now()
	value.UpdatedAt = value.CreatedAt
	copy := *value
	r.requests[value.ID] = &copy
	return nil
}
func (r *fakeRepo) FindByID(_ context.Context, familyID, id uint64, _ bool) (*joinmodel.FamilyJoinRequest, error) {
	v, ok := r.requests[id]
	if !ok || v.FamilyID != familyID {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *v
	return &copy, nil
}
func (r *fakeRepo) ListMine(_ context.Context, userID uint64) ([]joinrepo.JoinRequestRow, error) {
	var rows []joinrepo.JoinRequestRow
	for _, v := range r.requests {
		if v.ApplicantUserID == userID {
			rows = append(rows, joinrepo.JoinRequestRow{FamilyJoinRequest: *v})
		}
	}
	return rows, nil
}
func (r *fakeRepo) ListFamily(_ context.Context, familyID uint64) ([]joinrepo.JoinRequestRow, error) {
	var rows []joinrepo.JoinRequestRow
	for _, v := range r.requests {
		if v.FamilyID == familyID {
			rows = append(rows, joinrepo.JoinRequestRow{FamilyJoinRequest: *v})
		}
	}
	return rows, nil
}
func (r *fakeRepo) UpdateStatus(_ context.Context, id uint64, current string, values map[string]any) error {
	v, ok := r.requests[id]
	if !ok || v.RequestStatus != current {
		return gorm.ErrRecordNotFound
	}
	if value, ok := values["request_status"].(string); ok {
		v.RequestStatus = value
	}
	if value, ok := values["approve_mode"].(string); ok {
		v.ApproveMode = &value
	}
	if value, ok := values["bound_member_id"].(uint64); ok {
		v.BoundMemberID = &value
	}
	if value, ok := values["created_member_id"].(uint64); ok {
		v.CreatedMemberID = &value
	}
	if value, ok := values["handled_at"].(time.Time); ok {
		v.HandledAt = &value
	}
	if value, ok := values["cancelled_at"].(time.Time); ok {
		v.CancelledAt = &value
	}
	return nil
}
func (r *fakeRepo) CreateMember(_ context.Context, value *membermodel.FamilyMember) error {
	value.ID = r.nextMember
	r.nextMember++
	copy := *value
	r.members[value.ID] = &copy
	return nil
}
func (r *fakeRepo) UpdateMember(_ context.Context, familyID, memberID uint64, values map[string]any) error {
	member, ok := r.members[memberID]
	if !ok || member.FamilyID != familyID {
		return gorm.ErrRecordNotFound
	}
	if value, ok := values["member_type"].(string); ok {
		member.MemberType = value
	}
	return nil
}
func (r *fakeRepo) FindMemberForUpdate(ctx context.Context, familyID, memberID uint64) (*membermodel.FamilyMember, error) {
	return r.FindMember(ctx, familyID, memberID, true)
}
func (r *fakeRepo) FindDuplicate(_ context.Context, familyID, fromMemberID, toMemberID uint64, relationshipType string) (*relationshipmodel.FamilyRelationship, error) {
	for _, value := range r.relationships {
		if value.FamilyID != familyID || value.RelationshipType != relationshipType || value.Status != "ACTIVE" || value.DeletedAt != nil {
			continue
		}
		direct := value.FromMemberID == fromMemberID && value.ToMemberID == toMemberID
		reverseSpouse := relationshipType == "SPOUSE" && value.FromMemberID == toMemberID && value.ToMemberID == fromMemberID
		if direct || reverseSpouse {
			copy := *value
			return &copy, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeRepo) ListActiveParents(_ context.Context, familyID, childMemberID uint64) ([]relationshipmodel.FamilyRelationship, error) {
	var result []relationshipmodel.FamilyRelationship
	for _, value := range r.relationships {
		if value.FamilyID == familyID && value.ToMemberID == childMemberID && value.RelationshipType == "PARENT_CHILD" && value.Status == "ACTIVE" && value.DeletedAt == nil {
			result = append(result, *value)
		}
	}
	return result, nil
}
func (r *fakeRepo) FindPrimaryParentByGender(ctx context.Context, familyID, childMemberID uint64, gender string, excludeRelationshipID uint64) (*relationshipmodel.FamilyRelationship, error) {
	relationship, _, err := r.FindPrimaryParentWithMemberByGender(ctx, familyID, childMemberID, gender, excludeRelationshipID)
	return relationship, err
}
func (r *fakeRepo) FindPrimaryParentWithMemberByGender(_ context.Context, familyID, childMemberID uint64, gender string, excludeRelationshipID uint64) (*relationshipmodel.FamilyRelationship, *membermodel.FamilyMember, error) {
	for _, value := range r.relationships {
		parent := r.members[value.FromMemberID]
		if value.ID != excludeRelationshipID && value.FamilyID == familyID && value.ToMemberID == childMemberID && value.RelationshipType == "PARENT_CHILD" && value.ParentLinkType != nil && *value.ParentLinkType == "PRIMARY" && value.Status == "ACTIVE" && parent != nil && parent.Gender == gender {
			relationshipCopy, parentCopy := *value, *parent
			return &relationshipCopy, &parentCopy, nil
		}
	}
	return nil, nil, gorm.ErrRecordNotFound
}
func (r *fakeRepo) ListActiveSpouseRelationshipsByGender(_ context.Context, familyID, baseMemberID uint64, gender string) ([]relationshiprepo.SpouseRelationshipRow, error) {
	var result []relationshiprepo.SpouseRelationshipRow
	for _, value := range r.relationships {
		if value.FamilyID != familyID || value.RelationshipType != "SPOUSE" || value.Status != "ACTIVE" {
			continue
		}
		spouseID := uint64(0)
		if value.FromMemberID == baseMemberID {
			spouseID = value.ToMemberID
		} else if value.ToMemberID == baseMemberID {
			spouseID = value.FromMemberID
		}
		spouse := r.members[spouseID]
		if spouse != nil && spouse.Gender == gender {
			result = append(result, relationshiprepo.SpouseRelationshipRow{Relationship: *value, Spouse: *spouse})
		}
	}
	return result, nil
}
func (r *fakeRepo) CreateRelationship(_ context.Context, value *relationshipmodel.FamilyRelationship) error {
	value.ID = r.nextRelationship
	r.nextRelationship++
	copy := *value
	r.relationships[value.ID] = &copy
	return nil
}
func (r *fakeRepo) UpdateRelationship(_ context.Context, familyID, relationshipID uint64, values map[string]any) error {
	value := r.relationships[relationshipID]
	if value == nil || value.FamilyID != familyID {
		return gorm.ErrRecordNotFound
	}
	if noteType, ok := values["relation_note_type"].(*string); ok {
		value.RelationNoteType = noteType
	}
	if note, ok := values["relation_note"].(*string); ok {
		value.RelationNote = note
	}
	return nil
}
func (r *fakeRepo) CreateLink(_ context.Context, value *rolemodel.FamilyMemberUserLink) error {
	r.links = append(r.links, *value)
	return nil
}
func (r *fakeRepo) CancelPendingInvitations(context.Context, uint64, uint64, uint64, time.Time) (int64, error) {
	count := r.pendingInvitations
	r.pendingInvitations = 0
	return count, nil
}
func (r *fakeRepo) IncrementGraphVersion(context.Context, uint64) (int64, error) {
	r.family.GraphVersion++
	return r.family.GraphVersion, nil
}
func (r *fakeRepo) WriteLog(_ context.Context, input operationlog.WriteInput) error {
	r.logs = append(r.logs, input)
	return nil
}

type fakeUOW struct{ repo joinrepo.Repository }

func (u fakeUOW) WithinTransaction(ctx context.Context, fn func(joinrepo.Repository) error) error {
	repo, ok := u.repo.(*fakeRepo)
	if !ok {
		return fn(u.repo)
	}
	snapshot := repo.clone()
	if err := fn(u.repo); err != nil {
		*repo = *snapshot
		return err
	}
	return nil
}

func (r *fakeRepo) clone() *fakeRepo {
	copyRepo := *r
	copyRepo.users = make(map[uint64]usermodel.User, len(r.users))
	for key, value := range r.users {
		copyRepo.users[key] = value
	}
	copyRepo.members = make(map[uint64]*membermodel.FamilyMember, len(r.members))
	for key, value := range r.members {
		copyValue := *value
		copyRepo.members[key] = &copyValue
	}
	copyRepo.requests = make(map[uint64]*joinmodel.FamilyJoinRequest, len(r.requests))
	for key, value := range r.requests {
		copyValue := *value
		copyRepo.requests[key] = &copyValue
	}
	copyRepo.relationships = make(map[uint64]*relationshipmodel.FamilyRelationship, len(r.relationships))
	for key, value := range r.relationships {
		copyValue := *value
		copyRepo.relationships[key] = &copyValue
	}
	copyRepo.links = append([]rolemodel.FamilyMemberUserLink(nil), r.links...)
	copyRepo.logs = append([]operationlog.WriteInput(nil), r.logs...)
	return &copyRepo
}
func testService(repo *fakeRepo, allowed bool) *service {
	s := NewService(repo, fakeUOW{repo}, fakePermission{allowed}, nil, contentsafety.AlwaysPass()).(*service)
	s.now = func() time.Time { return time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC) }
	return s
}
func addPending(repo *fakeRepo, applicant uint64) {
	repo.requests[1] = &joinmodel.FamilyJoinRequest{ID: 1, FamilyID: 2, ApplicantUserID: applicant, RequestStatus: "PENDING"}
}

func TestJoinRequestCreationRules(t *testing.T) {
	t.Run("create and log", func(t *testing.T) {
		repo := newFakeRepo()
		result, err := testService(repo, true).Create(context.Background(), 9, 2, createJoinRequestInput(), AuditInput{})
		if err != nil || result.RequestStatus != "PENDING" || result.ApplicantGender == nil || *result.ApplicantGender != "MALE" || len(repo.logs) != 1 {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		data, _ := json.Marshal(repo.logs)
		if containsSensitive(string(data)) {
			t.Fatalf("sensitive log: %s", data)
		}
	})
	t.Run("gender required", func(t *testing.T) {
		repo := newFakeRepo()
		if _, err := testService(repo, true).Create(context.Background(), 9, 2, dto.CreateJoinRequest{}, AuditInput{}); err == nil || err.Code != CodeApproveModeInvalid {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("gender invalid", func(t *testing.T) {
		repo := newFakeRepo()
		gender := "UNKNOWN"
		if _, err := testService(repo, true).Create(context.Background(), 9, 2, dto.CreateJoinRequest{ApplicantGender: &gender}, AuditInput{}); err == nil || err.Code != CodeApproveModeInvalid {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("duplicate rejected", func(t *testing.T) {
		repo := newFakeRepo()
		addPending(repo, 9)
		if _, err := testService(repo, true).Create(context.Background(), 9, 2, createJoinRequestInput(), AuditInput{}); err == nil || err.Code != CodeJoinRequestDuplicate {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("already linked rejected", func(t *testing.T) {
		repo := newFakeRepo()
		repo.links = append(repo.links, rolemodel.FamilyMemberUserLink{FamilyID: 2, MemberID: 6, UserID: 9, LinkStatus: "ACTIVE"})
		if _, err := testService(repo, true).Create(context.Background(), 9, 2, createJoinRequestInput(), AuditInput{}); err == nil || err.Code != CodeApplicantUnavailable {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("active user without phone verified allowed", func(t *testing.T) {
		repo := newFakeRepo()
		repo.users[9] = usermodel.User{ID: 9, Status: "ACTIVE", PhoneVerified: false}
		result, err := testService(repo, true).Create(context.Background(), 9, 2, createJoinRequestInput(), AuditInput{})
		if err != nil || result.RequestStatus != "PENDING" {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
	})
}

func TestJoinRequestApprovalModes(t *testing.T) {
	t.Run("non-admin denied", func(t *testing.T) {
		repo := newFakeRepo()
		addPending(repo, 9)
		memberID := uint64(6)
		if _, err := testService(repo, false).Approve(context.Background(), 8, 2, 1, dto.ApproveJoinRequest{ApproveMode: "BIND_EXISTING_MEMBER", MemberID: &memberID}, AuditInput{}); err == nil || err.Code != CodeJoinRequestForbidden {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("bind existing does not increment graph version", func(t *testing.T) {
		repo := newFakeRepo()
		repo.pendingInvitations = 2
		gender := "MALE"
		repo.requests[1] = &joinmodel.FamilyJoinRequest{
			ID: 1, FamilyID: 2, ApplicantUserID: 9, ApplicantGender: &gender, RequestStatus: "PENDING",
		}
		repo.members[6].Gender = gender
		before := repo.family.GraphVersion
		memberID := uint64(6)
		result, err := testService(repo, true).Approve(context.Background(), 8, 2, 1, dto.ApproveJoinRequest{ApproveMode: "BIND_EXISTING_MEMBER", MemberID: &memberID}, AuditInput{})
		if err != nil || result.RequestStatus != "APPROVED" || len(repo.links) != 1 {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		if repo.links[0].LinkSource != "JOIN_REQUEST_APPROVED" || repo.family.GraphVersion != before {
			t.Fatal("bad existing-member approval")
		}
		if repo.pendingInvitations != 0 {
			t.Fatal("approval must resolve conflicting pending invitations")
		}
		if len(repo.logs) != 2 || repo.logs[0].Action != "AUTO_CANCEL_INVITATIONS" || repo.logs[1].Action != "APPROVE_JOIN_REQUEST" {
			t.Fatalf("unexpected logs %#v", repo.logs)
		}
	})
	t.Run("bind existing requires matching applicant gender", func(t *testing.T) {
		repo := newFakeRepo()
		applicantGender := "MALE"
		repo.requests[1] = &joinmodel.FamilyJoinRequest{
			ID: 1, FamilyID: 2, ApplicantUserID: 9, ApplicantGender: &applicantGender, RequestStatus: "PENDING",
		}
		repo.members[6].Gender = "FEMALE"
		memberID := uint64(6)
		_, businessErr := testService(repo, true).Approve(context.Background(), 8, 2, 1, dto.ApproveJoinRequest{
			ApproveMode: "BIND_EXISTING_MEMBER", MemberID: &memberID,
		}, AuditInput{})
		if businessErr == nil || businessErr.Code != CodeApproveModeInvalid || len(repo.links) != 0 {
			t.Fatalf("unexpected result err=%#v links=%d", businessErr, len(repo.links))
		}
	})
	t.Run("create new increments graph version", func(t *testing.T) {
		repo := newFakeRepo()
		addPending(repo, 9)
		before := repo.family.GraphVersion
		result, err := testService(repo, true).Approve(context.Background(), 8, 2, 1, dto.ApproveJoinRequest{
			ApproveMode: "CREATE_NEW_MEMBER", NewMember: &dto.NewMemberInput{Name: "New Member"},
		}, AuditInput{})
		if err != nil || result.CreatedMemberID == nil || result.GraphVersion == nil {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		if repo.family.GraphVersion != before+1 || len(repo.links) != 1 {
			t.Fatal("new member approval must increment once and link")
		}
	})
	t.Run("create and locate commits member relationship link and logs together", func(t *testing.T) {
		repo := newFakeRepo()
		repo.family.FamilySurname = "测"
		repo.members[6].Gender = "MALE"
		gender := "FEMALE"
		repo.requests[1] = &joinmodel.FamilyJoinRequest{
			ID: 1, FamilyID: 2, ApplicantUserID: 9, ApplicantGender: &gender, RequestStatus: "PENDING",
		}
		beforeVersion := repo.family.GraphVersion
		result, businessErr := testService(repo, true).Approve(context.Background(), 8, 2, 1, dto.ApproveJoinRequest{
			ApproveMode: "CREATE_NEW_MEMBER",
			NewMember:   &dto.NewMemberInput{Name: "Located Member", Gender: &gender},
			Location: &dto.MemberLocation{
				BaseMemberID: 6,
				AddType:      "ADD_CHILD",
				Relationship: relationshipdto.RelationshipInput{},
			},
		}, AuditInput{})
		if businessErr != nil || result.RequestStatus != "APPROVED" || result.CreatedMemberID == nil {
			t.Fatalf("unexpected result=%#v err=%#v", result, businessErr)
		}
		createdID := *result.CreatedMemberID
		if repo.family.GraphVersion != beforeVersion+1 || len(repo.relationships) != 1 || len(repo.links) != 1 {
			t.Fatalf("transaction result relationships=%d links=%d version=%d", len(repo.relationships), len(repo.links), repo.family.GraphVersion)
		}
		for _, relationship := range repo.relationships {
			if relationship.FromMemberID != 6 || relationship.ToMemberID != createdID || relationship.RelationshipType != "PARENT_CHILD" {
				t.Fatalf("unexpected relationship %#v", relationship)
			}
		}
		if len(repo.logs) != 3 || repo.logs[0].Action != "CREATE_MEMBER" || repo.logs[1].Action != "CREATE_RELATIONSHIP" || repo.logs[2].Action != "APPROVE_JOIN_REQUEST" {
			t.Fatalf("unexpected logs %#v", repo.logs)
		}
	})
	t.Run("placement failure rolls back newly created member", func(t *testing.T) {
		repo := newFakeRepo()
		repo.family.FamilySurname = "测"
		repo.members[6].Gender = "MALE"
		repo.members[7] = &membermodel.FamilyMember{ID: 7, FamilyID: 2, DisplayName: "测父", Gender: "MALE", MemberType: "LINEAGE_MEMBER", Status: "ACTIVE"}
		primary := "PRIMARY"
		repo.relationships[29] = &relationshipmodel.FamilyRelationship{
			ID: 29, FamilyID: 2, FromMemberID: 7, ToMemberID: 6,
			RelationshipType: "PARENT_CHILD", ParentLinkType: &primary, Status: "ACTIVE",
		}
		gender := "MALE"
		memberType := "LINEAGE_MEMBER"
		repo.requests[1] = &joinmodel.FamilyJoinRequest{
			ID: 1, FamilyID: 2, ApplicantUserID: 9, ApplicantGender: &gender, RequestStatus: "PENDING",
		}
		beforeMembers := len(repo.members)
		_, businessErr := testService(repo, true).Approve(context.Background(), 8, 2, 1, dto.ApproveJoinRequest{
			ApproveMode: "CREATE_NEW_MEMBER",
			NewMember:   &dto.NewMemberInput{Name: "Duplicate Father", Gender: &gender},
			Location: &dto.MemberLocation{
				BaseMemberID: 6,
				AddType:      "ADD_FATHER",
				MemberType:   &memberType,
				Relationship: relationshipdto.RelationshipInput{},
			},
		}, AuditInput{})
		if businessErr == nil || businessErr.Code != 43304 {
			t.Fatalf("unexpected error %#v", businessErr)
		}
		if len(repo.members) != beforeMembers || len(repo.links) != 0 || len(repo.logs) != 0 || repo.requests[1].RequestStatus != "PENDING" {
			t.Fatalf("failed approval left mutations members=%d links=%d logs=%d request=%s", len(repo.members), len(repo.links), len(repo.logs), repo.requests[1].RequestStatus)
		}
	})
	t.Run("new member gender must match applicant gender", func(t *testing.T) {
		repo := newFakeRepo()
		applicantGender := "MALE"
		memberGender := "FEMALE"
		repo.requests[1] = &joinmodel.FamilyJoinRequest{
			ID: 1, FamilyID: 2, ApplicantUserID: 9, ApplicantGender: &applicantGender, RequestStatus: "PENDING",
		}
		_, businessErr := testService(repo, true).Approve(context.Background(), 8, 2, 1, dto.ApproveJoinRequest{
			ApproveMode: "CREATE_NEW_MEMBER",
			NewMember:   &dto.NewMemberInput{Name: "Wrong Gender", Gender: &memberGender},
		}, AuditInput{})
		if businessErr == nil || businessErr.Code != CodeApproveModeInvalid || len(repo.members) != 1 || len(repo.links) != 0 {
			t.Fatalf("unexpected result err=%#v members=%d links=%d", businessErr, len(repo.members), len(repo.links))
		}
	})
}

func TestJoinRequestRejectAndCancel(t *testing.T) {
	t.Run("reject writes log", func(t *testing.T) {
		repo := newFakeRepo()
		addPending(repo, 9)
		result, err := testService(repo, true).Reject(context.Background(), 8, 2, 1, dto.RejectJoinRequest{}, AuditInput{})
		if err != nil || result.RequestStatus != "REJECTED" || len(repo.logs) != 1 {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
	})
	t.Run("applicant cancels and other user cannot", func(t *testing.T) {
		repo := newFakeRepo()
		addPending(repo, 9)
		if _, err := testService(repo, true).Cancel(context.Background(), 8, 2, 1, dto.CancelJoinRequest{}, AuditInput{}); err == nil || err.Code != CodeJoinRequestForbidden {
			t.Fatalf("unexpected %#v", err)
		}
		result, err := testService(repo, true).Cancel(context.Background(), 9, 2, 1, dto.CancelJoinRequest{}, AuditInput{})
		if err != nil || result.RequestStatus != "CANCELLED" {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
	})
}

func containsSensitive(value string) bool {
	for _, needle := range []string{"password", "token", "phone", "openid", "unionid", "verification"} {
		for i := 0; i+len(needle) <= len(value); i++ {
			if value[i:i+len(needle)] == needle {
				return true
			}
		}
	}
	return false
}

type joinCSOpenID struct{}

func (joinCSOpenID) ActiveWechatMiniOpenID(context.Context, uint64) (string, error) {
	return "join-test-openid", nil
}

func TestJoinRequestCreateContentSafetyRejected(t *testing.T) {
	repo := newFakeRepo()
	before := len(repo.requests)
	client := contentsafety.NewFakeClient(contentsafety.SuggestRisky)
	contentSafety := contentsafety.NewChecker(client, joinCSOpenID{}, nil)
	svc := NewService(repo, fakeUOW{repo}, fakePermission{allowed: true}, nil, contentSafety).(*service)
	svc.now = func() time.Time { return time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC) }
	gender := "MALE"
	msg := "违规申请留言"
	_, businessErr := svc.Create(context.Background(), 9, 2, dto.CreateJoinRequest{
		ApplicantGender: &gender, ApplicantMessage: &msg,
	}, AuditInput{})
	if businessErr == nil || businessErr.Code != apperrors.CodeContentSafetyRejected {
		t.Fatalf("expected 49007, got %#v", businessErr)
	}
	if len(repo.requests) != before {
		t.Fatal("join request must not be created")
	}
}

func TestJoinRequestApproveContentSafetyRejected(t *testing.T) {
	repo := newFakeRepo()
	addPending(repo, 9)
	beforeLinks := len(repo.links)
	client := contentsafety.NewFakeClient(contentsafety.SuggestReview)
	contentSafety := contentsafety.NewChecker(client, joinCSOpenID{}, nil)
	svc := NewService(repo, fakeUOW{repo}, fakePermission{allowed: true}, nil, contentSafety).(*service)
	svc.now = func() time.Time { return time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC) }
	memberID := uint64(6)
	comment := "违规审批意见"
	_, businessErr := svc.Approve(context.Background(), 8, 2, 1, dto.ApproveJoinRequest{
		ApproveMode: "BIND_EXISTING_MEMBER", MemberID: &memberID, HandleComment: &comment,
	}, AuditInput{})
	if businessErr == nil || businessErr.Code != apperrors.CodeContentSafetyRejected {
		t.Fatalf("expected 49007, got %#v", businessErr)
	}
	if len(repo.links) != beforeLinks {
		t.Fatal("approve must not create link")
	}
}
