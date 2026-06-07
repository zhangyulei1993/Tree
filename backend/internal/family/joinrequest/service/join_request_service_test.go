package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/family/core/model"
	"tree/backend/internal/family/joinrequest/dto"
	joinmodel "tree/backend/internal/family/joinrequest/model"
	joinrepo "tree/backend/internal/family/joinrequest/repository"
	membermodel "tree/backend/internal/family/member/model"
	rolemodel "tree/backend/internal/family/role/model"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"
)

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
	family                  model.Family
	users                   map[uint64]usermodel.User
	members                 map[uint64]*membermodel.FamilyMember
	requests                map[uint64]*joinmodel.FamilyJoinRequest
	links                   []rolemodel.FamilyMemberUserLink
	logs                    []operationlog.WriteInput
	nextRequest, nextMember uint64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		family:   model.Family{ID: 2, FamilyName: "Tree", Status: "NORMAL", GraphVersion: 10},
		users:    map[uint64]usermodel.User{8: {ID: 8, Status: "ACTIVE", PhoneVerified: true}, 9: {ID: 9, Status: "ACTIVE", PhoneVerified: true}},
		members:  map[uint64]*membermodel.FamilyMember{6: {ID: 6, FamilyID: 2, DisplayName: "Existing", Status: "ACTIVE", UserBindingPolicy: "OPTIONAL"}},
		requests: map[uint64]*joinmodel.FamilyJoinRequest{}, nextRequest: 1, nextMember: 20,
	}
}
func (r *fakeRepo) WithTx(*gorm.DB) joinrepo.Repository { return r }
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
func (r *fakeRepo) CreateLink(_ context.Context, value *rolemodel.FamilyMemberUserLink) error {
	r.links = append(r.links, *value)
	return nil
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
	return fn(u.repo)
}
func testService(repo *fakeRepo, allowed bool) *service {
	s := NewService(repo, fakeUOW{repo}, fakePermission{allowed}).(*service)
	s.now = func() time.Time { return time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC) }
	return s
}
func addPending(repo *fakeRepo, applicant uint64) {
	repo.requests[1] = &joinmodel.FamilyJoinRequest{ID: 1, FamilyID: 2, ApplicantUserID: applicant, RequestStatus: "PENDING"}
}

func TestJoinRequestCreationRules(t *testing.T) {
	t.Run("create and log", func(t *testing.T) {
		repo := newFakeRepo()
		result, err := testService(repo, true).Create(context.Background(), 9, 2, dto.CreateJoinRequest{}, AuditInput{})
		if err != nil || result.RequestStatus != "PENDING" || len(repo.logs) != 1 {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		data, _ := json.Marshal(repo.logs)
		if containsSensitive(string(data)) {
			t.Fatalf("sensitive log: %s", data)
		}
	})
	t.Run("duplicate rejected", func(t *testing.T) {
		repo := newFakeRepo()
		addPending(repo, 9)
		if _, err := testService(repo, true).Create(context.Background(), 9, 2, dto.CreateJoinRequest{}, AuditInput{}); err == nil || err.Code != CodeJoinRequestDuplicate {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("already linked rejected", func(t *testing.T) {
		repo := newFakeRepo()
		repo.links = append(repo.links, rolemodel.FamilyMemberUserLink{FamilyID: 2, MemberID: 6, UserID: 9, LinkStatus: "ACTIVE"})
		if _, err := testService(repo, true).Create(context.Background(), 9, 2, dto.CreateJoinRequest{}, AuditInput{}); err == nil || err.Code != CodeApplicantUnavailable {
			t.Fatalf("unexpected %#v", err)
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
		addPending(repo, 9)
		before := repo.family.GraphVersion
		memberID := uint64(6)
		result, err := testService(repo, true).Approve(context.Background(), 8, 2, 1, dto.ApproveJoinRequest{ApproveMode: "BIND_EXISTING_MEMBER", MemberID: &memberID}, AuditInput{})
		if err != nil || result.RequestStatus != "APPROVED" || len(repo.links) != 1 {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		if repo.links[0].LinkSource != "JOIN_REQUEST_APPROVED" || repo.family.GraphVersion != before {
			t.Fatal("bad existing-member approval")
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
