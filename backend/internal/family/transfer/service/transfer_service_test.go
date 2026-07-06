package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	adminmodel "tree/backend/internal/admin/model"
	"tree/backend/internal/common/enums"
	"tree/backend/internal/common/contentsafety"
	apperrors "tree/backend/internal/common/errors"
	familymodel "tree/backend/internal/family/core/model"
	membermodel "tree/backend/internal/family/member/model"
	rolemodel "tree/backend/internal/family/role/model"
	transferdto "tree/backend/internal/family/transfer/dto"
	transferenum "tree/backend/internal/family/transfer/enum"
	transfermodel "tree/backend/internal/family/transfer/model"
	transferrepo "tree/backend/internal/family/transfer/repository"
	operationlog "tree/backend/internal/operationlog/service"
)

type fakeRepo struct {
	family   familymodel.Family
	members  map[uint64]*membermodel.FamilyMember
	links    map[uint64]*rolemodel.FamilyMemberUserLink
	requests map[uint64]*transfermodel.FamilyFounderTransferRequest
	logs     []operationlog.WriteInput
	nextID   uint64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		family: familymodel.Family{ID: 2, FamilyName: "Tree", Status: string(enums.StatusNormal), CurrentFounderMemberID: ptr(3), GraphVersion: 20},
		members: map[uint64]*membermodel.FamilyMember{
			3: {ID: 3, FamilyID: 2, DisplayName: "Founder", Status: string(enums.StatusActive), UserBindingPolicy: string(enums.UserBindingOptional)},
			4: {ID: 4, FamilyID: 2, DisplayName: "Next", Status: string(enums.StatusActive), UserBindingPolicy: string(enums.UserBindingOptional)},
		},
		links: map[uint64]*rolemodel.FamilyMemberUserLink{
			1: {ID: 1, FamilyID: 2, MemberID: 3, UserID: 8, LinkStatus: string(enums.StatusActive), FamilyRole: string(enums.FamilyRoleFounder)},
			2: {ID: 2, FamilyID: 2, MemberID: 4, UserID: 9, LinkStatus: string(enums.StatusActive), FamilyRole: string(enums.FamilyRoleMember)},
		},
		requests: map[uint64]*transfermodel.FamilyFounderTransferRequest{},
		nextID:   1,
	}
}

func ptr(v uint64) *uint64 { return &v }

func (r *fakeRepo) WithTx(*gorm.DB) transferrepo.Repository { return r }
func (r *fakeRepo) DB() *gorm.DB                            { return nil }
func (r *fakeRepo) FindAdmin(context.Context, uint64) (*adminmodel.AdminUser, error) {
	return &adminmodel.AdminUser{ID: 1, Role: string(enums.AdminRoleRootAdmin)}, nil
}
func (r *fakeRepo) FindFamily(context.Context, uint64, bool) (*familymodel.Family, error) {
	v := r.family
	return &v, nil
}
func (r *fakeRepo) FindMember(_ context.Context, familyID, memberID uint64, _ bool) (*membermodel.FamilyMember, error) {
	v, ok := r.members[memberID]
	if !ok || v.FamilyID != familyID {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *v
	return &copy, nil
}
func (r *fakeRepo) FindActiveLinkByUser(_ context.Context, familyID, userID uint64, _ bool) (*rolemodel.FamilyMemberUserLink, error) {
	for _, v := range r.links {
		if v.FamilyID == familyID && v.UserID == userID && v.LinkStatus == string(enums.StatusActive) {
			copy := *v
			return &copy, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeRepo) FindActiveLinkByMember(_ context.Context, familyID, memberID uint64, _ bool) (*rolemodel.FamilyMemberUserLink, error) {
	for _, v := range r.links {
		if v.FamilyID == familyID && v.MemberID == memberID && v.LinkStatus == string(enums.StatusActive) {
			copy := *v
			return &copy, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeRepo) FindPendingByFamily(_ context.Context, familyID uint64) (*transfermodel.FamilyFounderTransferRequest, error) {
	for _, v := range r.requests {
		if v.FamilyID == familyID && v.RequestStatus == transferenum.StatusPending {
			return v, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeRepo) Create(_ context.Context, request *transfermodel.FamilyFounderTransferRequest) error {
	request.ID = r.nextID
	r.nextID++
	request.CreatedAt = time.Now()
	request.UpdatedAt = request.CreatedAt
	copy := *request
	r.requests[request.ID] = &copy
	return nil
}
func (r *fakeRepo) FindByID(_ context.Context, id uint64, _ bool) (*transfermodel.FamilyFounderTransferRequest, error) {
	v, ok := r.requests[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *v
	return &copy, nil
}
func (r *fakeRepo) List(_ context.Context, _ transferrepo.ListQuery) ([]transferrepo.TransferRow, int64, error) {
	rows := make([]transferrepo.TransferRow, 0)
	for _, v := range r.requests {
		rows = append(rows, transferrepo.TransferRow{FamilyFounderTransferRequest: *v, FamilyName: r.family.FamilyName})
	}
	return rows, int64(len(rows)), nil
}
func (r *fakeRepo) UpdateRequest(_ context.Context, id uint64, current string, values map[string]any) error {
	v, ok := r.requests[id]
	if !ok || v.RequestStatus != current {
		return gorm.ErrRecordNotFound
	}
	if value, ok := values["request_status"].(string); ok {
		v.RequestStatus = value
	}
	if value, ok := values["review_result"].(string); ok {
		v.ReviewResult = &value
	}
	if value, ok := values["reviewed_by_admin_id"].(uint64); ok {
		v.ReviewedByAdminID = &value
	}
	if value, ok := values["reviewed_at"].(time.Time); ok {
		v.ReviewedAt = &value
	}
	if value, ok := values["completed_at"].(time.Time); ok {
		v.CompletedAt = &value
	}
	if value, ok := values["cancelled_at"].(time.Time); ok {
		v.CancelledAt = &value
	}
	return nil
}
func (r *fakeRepo) UpdateLinkRole(_ context.Context, linkID uint64, values map[string]any) error {
	link := r.links[linkID]
	if value, ok := values["family_role"].(string); ok {
		link.FamilyRole = value
	}
	return nil
}
func (r *fakeRepo) UpdateFamily(_ context.Context, _ uint64, values map[string]any) error {
	if value, ok := values["current_founder_member_id"].(uint64); ok {
		r.family.CurrentFounderMemberID = &value
	}
	return nil
}
func (r *fakeRepo) CountActiveFounders(context.Context, uint64) (int64, error) {
	var count int64
	for _, v := range r.links {
		if v.LinkStatus == string(enums.StatusActive) && v.FamilyRole == string(enums.FamilyRoleFounder) {
			count++
		}
	}
	return count, nil
}
func (r *fakeRepo) WriteLog(_ context.Context, input operationlog.WriteInput) error {
	r.logs = append(r.logs, input)
	return nil
}

type fakeUOW struct{ repo transferrepo.Repository }

func (u fakeUOW) WithinTransaction(ctx context.Context, fn func(transferrepo.Repository) error) error {
	return fn(u.repo)
}

func newTestService(repo *fakeRepo) *service {
	svc := NewService(repo, fakeUOW{repo: repo}, nil, contentsafety.AlwaysPass()).(*service)
	svc.now = func() time.Time { return time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC) }
	return svc
}

func TestTransferCreateRules(t *testing.T) {
	t.Run("founder creates request", func(t *testing.T) {
		repo := newFakeRepo()
		result, err := newTestService(repo).Create(context.Background(), 8, 2, transferdto.CreateTransferRequest{ToMemberID: 4}, AuditInput{})
		if err != nil || result.RequestStatus != transferenum.StatusPending || len(repo.logs) != 1 {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
	})
	t.Run("non-founder denied", func(t *testing.T) {
		repo := newFakeRepo()
		if _, err := newTestService(repo).Create(context.Background(), 9, 2, transferdto.CreateTransferRequest{ToMemberID: 4}, AuditInput{}); err == nil || err.Code != CodeTransferForbidden {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("target unavailable rejected", func(t *testing.T) {
		repo := newFakeRepo()
		delete(repo.links, 2)
		if _, err := newTestService(repo).Create(context.Background(), 8, 2, transferdto.CreateTransferRequest{ToMemberID: 4}, AuditInput{}); err == nil || err.Code != CodeTransferTargetInvalid {
			t.Fatalf("unexpected %#v", err)
		}
		repo = newFakeRepo()
		repo.members[4].UserBindingPolicy = string(enums.UserBindingNotRequired)
		if _, err := newTestService(repo).Create(context.Background(), 8, 2, transferdto.CreateTransferRequest{ToMemberID: 4}, AuditInput{}); err == nil || err.Code != CodeTransferTargetInvalid {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("duplicate pending rejected", func(t *testing.T) {
		repo := newFakeRepo()
		_, _ = newTestService(repo).Create(context.Background(), 8, 2, transferdto.CreateTransferRequest{ToMemberID: 4}, AuditInput{})
		if _, err := newTestService(repo).Create(context.Background(), 8, 2, transferdto.CreateTransferRequest{ToMemberID: 4}, AuditInput{}); err == nil || err.Code != CodeTransferDuplicate {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("active family member reads current request", func(t *testing.T) {
		repo := newFakeRepo()
		created, _ := newTestService(repo).Create(context.Background(), 8, 2, transferdto.CreateTransferRequest{ToMemberID: 4}, AuditInput{})
		result, err := newTestService(repo).Current(context.Background(), 9, 2)
		if err != nil || result.RequestID != created.RequestID || result.RequestStatus != transferenum.StatusPending {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
	})
	t.Run("outsider cannot read current request", func(t *testing.T) {
		repo := newFakeRepo()
		if _, err := newTestService(repo).Current(context.Background(), 99, 2); err == nil || err.Code != CodeTransferForbidden {
			t.Fatalf("unexpected %#v", err)
		}
	})
}

func TestTransferReviewRules(t *testing.T) {
	t.Run("cancel request", func(t *testing.T) {
		repo := newFakeRepo()
		created, _ := newTestService(repo).Create(context.Background(), 8, 2, transferdto.CreateTransferRequest{ToMemberID: 4}, AuditInput{})
		result, err := newTestService(repo).Cancel(context.Background(), 8, 2, created.RequestID, transferdto.CancelTransferRequest{}, AuditInput{})
		if err != nil || result.RequestStatus != transferenum.StatusCancelled {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
	})
	t.Run("platform admin denied", func(t *testing.T) {
		repo := newFakeRepo()
		created, _ := newTestService(repo).Create(context.Background(), 8, 2, transferdto.CreateTransferRequest{ToMemberID: 4}, AuditInput{})
		if _, err := newTestService(repo).Approve(context.Background(), 1, string(enums.AdminRolePlatformAdmin), created.RequestID, transferdto.ReviewTransferRequest{}, AuditInput{}); err == nil || err.Code != CodeTransferAdminDenied {
			t.Fatalf("unexpected %#v", err)
		}
		if _, err := newTestService(repo).Reject(context.Background(), 1, string(enums.AdminRolePlatformAdmin), created.RequestID, transferdto.ReviewTransferRequest{}, AuditInput{}); err == nil || err.Code != CodeTransferAdminDenied {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("root approves and keeps one founder", func(t *testing.T) {
		repo := newFakeRepo()
		before := repo.family.GraphVersion
		created, _ := newTestService(repo).Create(context.Background(), 8, 2, transferdto.CreateTransferRequest{ToMemberID: 4}, AuditInput{})
		result, err := newTestService(repo).Approve(context.Background(), 1, string(enums.AdminRoleRootAdmin), created.RequestID, transferdto.ReviewTransferRequest{}, AuditInput{})
		if err != nil || result.RequestStatus != transferenum.StatusApproved {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		count, _ := repo.CountActiveFounders(context.Background(), 2)
		if count != 1 || repo.links[2].FamilyRole != string(enums.FamilyRoleFounder) || *repo.family.CurrentFounderMemberID != 4 || repo.family.GraphVersion != before {
			t.Fatal("founder transfer did not complete correctly")
		}
	})
	t.Run("reject does not change founder", func(t *testing.T) {
		repo := newFakeRepo()
		created, _ := newTestService(repo).Create(context.Background(), 8, 2, transferdto.CreateTransferRequest{ToMemberID: 4}, AuditInput{})
		result, err := newTestService(repo).Reject(context.Background(), 1, string(enums.AdminRoleSuperAdmin), created.RequestID, transferdto.ReviewTransferRequest{}, AuditInput{})
		if err != nil || result.RequestStatus != transferenum.StatusRejected || repo.links[1].FamilyRole != string(enums.FamilyRoleFounder) || repo.links[2].FamilyRole == string(enums.FamilyRoleFounder) {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
	})
}

type transferCSOpenID struct{}

func (transferCSOpenID) ActiveWechatMiniOpenID(context.Context, uint64) (string, error) {
	return "transfer-test-openid", nil
}

func TestTransferCreateContentSafetyRejected(t *testing.T) {
	repo := newFakeRepo()
	before := len(repo.requests)
	beforeGV := repo.family.GraphVersion
	client := contentsafety.NewFakeClient(contentsafety.SuggestRisky)
	contentSafety := contentsafety.NewChecker(client, transferCSOpenID{}, nil)
	svc := NewService(repo, fakeUOW{repo: repo}, nil, contentSafety).(*service)
	svc.now = func() time.Time { return time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC) }
	reason := "违规转让理由"
	_, businessErr := svc.Create(context.Background(), 8, 2, transferdto.CreateTransferRequest{
		ToMemberID: 4, RequestReason: &reason,
	}, AuditInput{})
	if businessErr == nil || businessErr.Code != apperrors.CodeContentSafetyRejected {
		t.Fatalf("expected 49007, got %#v", businessErr)
	}
	if len(repo.requests) != before || repo.family.GraphVersion != beforeGV {
		t.Fatal("transfer request must not be created")
	}
}
