package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/family/core/model"
	"tree/backend/internal/family/invitation/dto"
	invitationmodel "tree/backend/internal/family/invitation/model"
	inviterepo "tree/backend/internal/family/invitation/repository"
	membermodel "tree/backend/internal/family/member/model"
	rolemodel "tree/backend/internal/family/role/model"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"
)

type fakePermission struct {
	allowed bool
	link    rolemodel.FamilyMemberUserLink
}

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
	value := p.link
	return &value, nil
}

type fakeRepo struct {
	family              model.Family
	member              membermodel.FamilyMember
	users               map[uint64]usermodel.User
	invitations         map[uint64]*invitationmodel.FamilyInvitation
	links               []rolemodel.FamilyMemberUserLink
	logs                []operationlog.WriteInput
	nextID              uint64
	pendingJoinRequests int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		family: model.Family{ID: 2, FamilyName: "Tree", Status: "NORMAL", GraphVersion: 14},
		member: membermodel.FamilyMember{ID: 6, FamilyID: 2, DisplayName: "Member", Status: "ACTIVE", UserBindingPolicy: "OPTIONAL"},
		users: map[uint64]usermodel.User{
			8: {ID: 8, Status: "ACTIVE", PhoneVerified: true},
			9: {ID: 9, Status: "ACTIVE", PhoneVerified: true},
		},
		invitations: map[uint64]*invitationmodel.FamilyInvitation{}, nextID: 1,
	}
}
func (r *fakeRepo) WithTx(*gorm.DB) inviterepo.Repository { return r }
func (r *fakeRepo) FindFamily(context.Context, uint64, bool) (*model.Family, error) {
	v := r.family
	return &v, nil
}
func (r *fakeRepo) FindMember(context.Context, uint64, uint64, bool) (*membermodel.FamilyMember, error) {
	v := r.member
	return &v, nil
}
func (r *fakeRepo) FindUser(_ context.Context, id uint64, _ bool) (*usermodel.User, error) {
	v, ok := r.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &v, nil
}
func (r *fakeRepo) FindActiveLinkByMember(_ context.Context, familyID, memberID uint64) (*rolemodel.FamilyMemberUserLink, error) {
	for i := range r.links {
		if r.links[i].FamilyID == familyID && r.links[i].MemberID == memberID && r.links[i].LinkStatus == "ACTIVE" {
			return &r.links[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeRepo) FindActiveLinkByUser(_ context.Context, familyID, userID uint64) (*rolemodel.FamilyMemberUserLink, error) {
	for i := range r.links {
		if r.links[i].FamilyID == familyID && r.links[i].UserID == userID && r.links[i].LinkStatus == "ACTIVE" {
			return &r.links[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeRepo) FindPendingByMember(_ context.Context, familyID, memberID uint64, now time.Time) (*invitationmodel.FamilyInvitation, error) {
	for _, v := range r.invitations {
		if v.FamilyID == familyID && v.TargetMemberID == memberID && v.Status == "PENDING" && v.ExpiredAt.After(now) {
			return v, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeRepo) Create(_ context.Context, value *invitationmodel.FamilyInvitation) error {
	value.ID = r.nextID
	r.nextID++
	value.CreatedAt = time.Now()
	copy := *value
	r.invitations[value.ID] = &copy
	return nil
}
func (r *fakeRepo) FindByID(_ context.Context, id uint64, _ bool) (*invitationmodel.FamilyInvitation, error) {
	v, ok := r.invitations[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *v
	return &copy, nil
}
func (r *fakeRepo) FindRowByID(_ context.Context, id uint64) (*inviterepo.InvitationRow, error) {
	v, ok := r.invitations[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &inviterepo.InvitationRow{
		FamilyInvitation: *v, FamilyName: "Tree", TargetMemberName: "Member",
		InviterDisplayName: "邀请人",
	}, nil
}
func (r *fakeRepo) FindByTokenHash(_ context.Context, hash string) (*inviterepo.InvitationRow, error) {
	for _, v := range r.invitations {
		if v.InviteToken != nil && *v.InviteToken == hash {
			return &inviterepo.InvitationRow{
				FamilyInvitation: *v, FamilyName: "Tree", TargetMemberName: "Member",
				InviterDisplayName: "邀请人",
			}, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeRepo) ListForUser(_ context.Context, userID uint64) ([]inviterepo.InvitationRow, error) {
	var rows []inviterepo.InvitationRow
	for _, v := range r.invitations {
		if v.TargetUserID != nil && *v.TargetUserID == userID {
			rows = append(rows, inviterepo.InvitationRow{
				FamilyInvitation: *v, InviterDisplayName: "邀请人",
			})
		}
	}
	return rows, nil
}
func (r *fakeRepo) ListForFamily(_ context.Context, familyID uint64) ([]inviterepo.InvitationRow, error) {
	var rows []inviterepo.InvitationRow
	for _, v := range r.invitations {
		if v.FamilyID == familyID {
			rows = append(rows, inviterepo.InvitationRow{
				FamilyInvitation: *v, FamilyName: "Tree", TargetMemberName: "Member",
				InviterDisplayName: "邀请人",
			})
		}
	}
	return rows, nil
}
func (r *fakeRepo) CancelPendingJoinRequestsForUser(context.Context, uint64, uint64, time.Time) (int64, error) {
	count := r.pendingJoinRequests
	r.pendingJoinRequests = 0
	return count, nil
}
func (r *fakeRepo) UpdateStatus(_ context.Context, id uint64, current string, values map[string]any) error {
	v, ok := r.invitations[id]
	if !ok || v.Status != current {
		return gorm.ErrRecordNotFound
	}
	if status, ok := values["status"].(string); ok {
		v.Status = status
	}
	if value, ok := values["accepted_by_user_id"].(uint64); ok {
		v.AcceptedByUserID = &value
	}
	if value, ok := values["accepted_at"].(time.Time); ok {
		v.AcceptedAt = &value
	}
	if value, ok := values["rejected_at"].(time.Time); ok {
		v.RejectedAt = &value
	}
	if value, ok := values["cancelled_at"].(time.Time); ok {
		v.CancelledAt = &value
	}
	return nil
}
func (r *fakeRepo) CreateLink(_ context.Context, value *rolemodel.FamilyMemberUserLink) error {
	r.links = append(r.links, *value)
	return nil
}
func (r *fakeRepo) WriteLog(_ context.Context, input operationlog.WriteInput) error {
	r.logs = append(r.logs, input)
	return nil
}

type fakeUOW struct{ repo inviterepo.Repository }

func (u fakeUOW) WithinTransaction(ctx context.Context, fn func(inviterepo.Repository) error) error {
	return fn(u.repo)
}

func testService(repo *fakeRepo, allowed bool, now time.Time) *service {
	s := NewService(repo, fakeUOW{repo}, fakePermission{allowed: allowed, link: rolemodel.FamilyMemberUserLink{FamilyRole: "FOUNDER"}}).(*service)
	s.now = func() time.Time { return now }
	s.token = func() (string, error) { return "raw-secret-invite-token", nil }
	return s
}

func TestInvitationCreationRules(t *testing.T) {
	now := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC)
	t.Run("share invite hashes token and expires in seven days", func(t *testing.T) {
		repo := newFakeRepo()
		svc := testService(repo, true, now)
		result, err := svc.Create(context.Background(), 8, 2, 6, dto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, AuditInput{})
		if err != nil || result.InviteToken != "raw-secret-invite-token" {
			t.Fatalf("unexpected result: %#v %#v", result, err)
		}
		stored := repo.invitations[result.Invitation.InvitationID]
		if stored.InviteToken == nil || *stored.InviteToken == result.InviteToken {
			t.Fatal("raw invite token must not be stored")
		}
		if !stored.ExpiredAt.Equal(now.Add(7 * 24 * time.Hour)) {
			t.Fatalf("wrong expiry: %v", stored.ExpiredAt)
		}
		logJSON, _ := json.Marshal(repo.logs)
		if string(logJSON) == "" || contains(string(logJSON), result.InviteToken) {
			t.Fatal("operation log contains raw token")
		}
	})
	t.Run("in-app invite validates target", func(t *testing.T) {
		repo := newFakeRepo()
		svc := testService(repo, true, now)
		target := uint64(9)
		result, err := svc.Create(context.Background(), 8, 2, 6, dto.CreateInvitationRequest{InviteChannel: "IN_APP", TargetUserID: &target}, AuditInput{})
		if err != nil || result.Invitation.InviteChannel != "IN_APP" {
			t.Fatalf("unexpected: %#v %#v", result, err)
		}
	})
	t.Run("duplicate pending rejected", func(t *testing.T) {
		repo := newFakeRepo()
		svc := testService(repo, true, now)
		_, _ = svc.Create(context.Background(), 8, 2, 6, dto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, AuditInput{})
		_, err := svc.Create(context.Background(), 8, 2, 6, dto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, AuditInput{})
		if err == nil || err.Code != CodeInvitationDuplicate {
			t.Fatalf("expected duplicate, got %#v", err)
		}
	})
	t.Run("expired pending does not block replacement", func(t *testing.T) {
		repo := newFakeRepo()
		repo.invitations[1] = &invitationmodel.FamilyInvitation{
			ID: 1, FamilyID: 2, TargetMemberID: 6, Status: "PENDING", ExpiredAt: now.Add(-time.Second),
		}
		repo.nextID = 2
		result, err := testService(repo, true, now).Create(context.Background(), 8, 2, 6, dto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, AuditInput{})
		if err != nil || result.Invitation.InvitationID != 2 {
			t.Fatalf("expired invitation blocked replacement: %#v %#v", result, err)
		}
	})
	t.Run("not-required and linked member rejected", func(t *testing.T) {
		repo := newFakeRepo()
		repo.member.UserBindingPolicy = "NOT_REQUIRED"
		svc := testService(repo, true, now)
		if _, err := svc.Create(context.Background(), 8, 2, 6, dto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, AuditInput{}); err == nil || err.Code != CodeMemberNotInvitable {
			t.Fatalf("unexpected %#v", err)
		}
		repo = newFakeRepo()
		repo.links = append(repo.links, rolemodel.FamilyMemberUserLink{FamilyID: 2, MemberID: 6, UserID: 9, LinkStatus: "ACTIVE"})
		svc = testService(repo, true, now)
		if _, err := svc.Create(context.Background(), 8, 2, 6, dto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, AuditInput{}); err == nil || err.Code != CodeMemberAlreadyLinked {
			t.Fatalf("unexpected %#v", err)
		}
	})
}

func TestInvitationMutations(t *testing.T) {
	now := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC)
	makeInvite := func(repo *fakeRepo, channel, status string, target *uint64, expiry time.Time) {
		repo.invitations[1] = &invitationmodel.FamilyInvitation{ID: 1, FamilyID: 2, TargetMemberID: 6, InviterUserID: uint64Ptr(8), TargetUserID: target, InviteChannel: channel, FamilyRoleAfterAccept: "MEMBER", Status: status, ExpiredAt: expiry}
	}
	t.Run("wrong in-app target rejected", func(t *testing.T) {
		repo := newFakeRepo()
		target := uint64(9)
		makeInvite(repo, "IN_APP", "PENDING", &target, now.Add(time.Hour))
		_, err := testService(repo, true, now).Accept(context.Background(), 8, 1, AuditInput{})
		if err == nil || err.Code != CodeInvitationTargetMismatch {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("accept creates link without graph change", func(t *testing.T) {
		repo := newFakeRepo()
		repo.pendingJoinRequests = 1
		makeInvite(repo, "SHARE_LINK", "PENDING", nil, now.Add(time.Hour))
		before := repo.family.GraphVersion
		result, err := testService(repo, true, now).Accept(context.Background(), 9, 1, AuditInput{})
		if err != nil || result.Status != "ACCEPTED" || len(repo.links) != 1 {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		link := repo.links[0]
		if link.LinkStatus != "ACTIVE" || link.LinkSource != "INVITATION_ACCEPTED" || link.FamilyRole != "MEMBER" {
			t.Fatalf("bad link %#v", link)
		}
		if repo.family.GraphVersion != before {
			t.Fatal("accept must not increment graph version")
		}
		if repo.pendingJoinRequests != 0 {
			t.Fatal("accept must resolve the user's pending join request")
		}
		if len(repo.logs) != 2 || repo.logs[0].Action != "AUTO_CANCEL_JOIN_REQUESTS" || repo.logs[1].Action != "ACCEPT_INVITATION" {
			t.Fatalf("unexpected logs %#v", repo.logs)
		}
	})
	t.Run("expired and non-pending rejected", func(t *testing.T) {
		repo := newFakeRepo()
		makeInvite(repo, "SHARE_LINK", "PENDING", nil, now.Add(-time.Second))
		if _, err := testService(repo, true, now).Accept(context.Background(), 9, 1, AuditInput{}); err == nil || err.Code != CodeInvitationExpired {
			t.Fatalf("unexpected %#v", err)
		}
		repo = newFakeRepo()
		makeInvite(repo, "SHARE_LINK", "REJECTED", nil, now.Add(time.Hour))
		if _, err := testService(repo, true, now).Accept(context.Background(), 9, 1, AuditInput{}); err == nil || err.Code != CodeInvitationInvalidStatus {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("reject and cancel", func(t *testing.T) {
		repo := newFakeRepo()
		makeInvite(repo, "SHARE_LINK", "PENDING", nil, now.Add(time.Hour))
		result, err := testService(repo, true, now).Reject(context.Background(), 9, 1, dto.RejectInvitationRequest{}, AuditInput{})
		if err != nil || result.Status != "REJECTED" {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		repo = newFakeRepo()
		makeInvite(repo, "SHARE_LINK", "PENDING", nil, now.Add(time.Hour))
		result, err = testService(repo, true, now).Cancel(context.Background(), 8, 1, dto.CancelInvitationRequest{}, AuditInput{})
		if err != nil || result.Status != "CANCELLED" {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		repo = newFakeRepo()
		makeInvite(repo, "SHARE_LINK", "PENDING", nil, now.Add(time.Hour))
		if _, err = testService(repo, false, now).Cancel(context.Background(), 9, 1, dto.CancelInvitationRequest{}, AuditInput{}); err == nil || err.Code != CodeInvitationForbidden {
			t.Fatalf("unexpected %#v", err)
		}
	})
}

func TestInvitationSenderManagement(t *testing.T) {
	now := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC)
	repo := newFakeRepo()
	repo.invitations[1] = &invitationmodel.FamilyInvitation{
		ID: 1, FamilyID: 2, TargetMemberID: 6, InviterUserID: uint64Ptr(8),
		InviteType: "CLAIM_EXISTING_MEMBER", InviteChannel: "SHARE_LINK",
		InviteActorType: "FAMILY_FOUNDER", FamilyRoleAfterAccept: "MEMBER",
		Status: "PENDING", ExpiredAt: now.Add(time.Hour),
	}
	repo.nextID = 2
	svc := testService(repo, true, now)
	svc.token = func() (string, error) { return "replacement-token", nil }

	items, listErr := svc.ListFamily(context.Background(), 8, 2)
	if listErr != nil || len(items) != 1 || items[0].Status != "PENDING" ||
		items[0].InviterDisplayName != "邀请人" || items[0].InviterRole != "FAMILY_FOUNDER" {
		t.Fatalf("unexpected list %#v %#v", items, listErr)
	}
	result, regenerateErr := svc.Regenerate(context.Background(), 8, 1, AuditInput{})
	if regenerateErr != nil || result.InviteToken != "replacement-token" || result.Invitation.InvitationID != 2 {
		t.Fatalf("unexpected regenerate %#v %#v", result, regenerateErr)
	}
	if repo.invitations[1].Status != "CANCELLED" || repo.invitations[2].Status != "PENDING" {
		t.Fatalf("unexpected statuses old=%s new=%s", repo.invitations[1].Status, repo.invitations[2].Status)
	}
	if len(repo.logs) != 1 || repo.logs[0].Action != "REGENERATE_INVITATION" {
		t.Fatalf("unexpected logs %#v", repo.logs)
	}
}

func contains(value, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
func uint64Ptr(value uint64) *uint64 { return &value }
