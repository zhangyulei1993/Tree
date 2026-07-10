package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	familymodel "tree/backend/internal/family/core/model"
	publicdto "tree/backend/internal/family/publicdisplay/dto"
	publicenum "tree/backend/internal/family/publicdisplay/enum"
	publicmodel "tree/backend/internal/family/publicdisplay/model"
	publicrepo "tree/backend/internal/family/publicdisplay/repository"
	"tree/backend/internal/family/publicdisplay/vo"
	rolemodel "tree/backend/internal/family/role/model"
	operationlog "tree/backend/internal/operationlog/service"
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
	family       familymodel.Family
	applications map[uint64]*publicmodel.FamilyPublicApplication
	logs         []operationlog.WriteInput
	nextID       uint64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		family: familymodel.Family{
			ID: 2, FamilyName: "Tree", Status: string(enums.StatusNormal),
			PublicDisplayStatus: publicenum.PublicPrivate, GraphVersion: 12,
		},
		applications: map[uint64]*publicmodel.FamilyPublicApplication{},
		nextID:       1,
	}
}

func (r *fakeRepo) WithTx(*gorm.DB) publicrepo.Repository { return r }
func (r *fakeRepo) FindFamily(context.Context, uint64, bool) (*familymodel.Family, error) {
	family := r.family
	return &family, nil
}
func (r *fakeRepo) UpdateFamilyPublicStatus(_ context.Context, _ uint64, values map[string]any) error {
	if value, ok := values["public_display_status"].(string); ok {
		r.family.PublicDisplayStatus = value
	}
	if value, ok := values["public_display_enabled"].(bool); ok {
		r.family.PublicDisplayEnabled = value
	}
	if value, ok := values["public_applied_at"].(time.Time); ok {
		r.family.PublicAppliedAt = &value
	}
	if value, ok := values["public_approved_at"].(time.Time); ok {
		r.family.PublicApprovedAt = &value
	}
	if value, ok := values["public_enabled_at"].(time.Time); ok {
		r.family.PublicEnabledAt = &value
	}
	if value, ok := values["public_taken_down_at"].(time.Time); ok {
		r.family.PublicTakenDownAt = &value
	}
	return nil
}
func (r *fakeRepo) FindPendingByFamily(_ context.Context, familyID uint64) (*publicmodel.FamilyPublicApplication, error) {
	for _, app := range r.applications {
		if app.FamilyID == familyID && app.ApplicationStatus == publicenum.StatusPending {
			return app, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeRepo) Create(_ context.Context, app *publicmodel.FamilyPublicApplication) error {
	app.ID = r.nextID
	r.nextID++
	app.CreatedAt = time.Now()
	app.UpdatedAt = app.CreatedAt
	copy := *app
	r.applications[app.ID] = &copy
	return nil
}
func (r *fakeRepo) FindByID(_ context.Context, id uint64, _ bool) (*publicmodel.FamilyPublicApplication, error) {
	app, ok := r.applications[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *app
	return &copy, nil
}
func (r *fakeRepo) List(_ context.Context, q publicrepo.ListQuery) ([]publicrepo.ApplicationRow, int64, error) {
	rows := make([]publicrepo.ApplicationRow, 0)
	for _, app := range r.applications {
		if q.FamilyID != nil && app.FamilyID != *q.FamilyID {
			continue
		}
		if q.Status != "" && app.ApplicationStatus != q.Status {
			continue
		}
		rows = append(rows, publicrepo.ApplicationRow{
			FamilyPublicApplication:   *app,
			FamilyName:                r.family.FamilyName,
			FamilyPublicDisplayStatus: r.family.PublicDisplayStatus,
		})
	}
	return rows, int64(len(rows)), nil
}
func (r *fakeRepo) UpdateApplication(_ context.Context, id uint64, current string, values map[string]any) error {
	app, ok := r.applications[id]
	if !ok || app.ApplicationStatus != current {
		return gorm.ErrRecordNotFound
	}
	if value, ok := values["application_status"].(string); ok {
		app.ApplicationStatus = value
	}
	if value, ok := values["review_result"].(string); ok {
		app.ReviewResult = &value
	}
	if value, ok := values["reviewed_by_admin_id"].(uint64); ok {
		app.ReviewedByAdminID = &value
	}
	if value, ok := values["reviewed_at"].(time.Time); ok {
		app.ReviewedAt = &value
	}
	if value, ok := values["review_comment"].(*string); ok {
		app.ReviewComment = value
	}
	if value, ok := values["cancelled_at"].(time.Time); ok {
		app.CancelledAt = &value
	}
	if value, ok := values["cancel_reason"].(*string); ok {
		app.CancelReason = value
	}
	return nil
}
func (r *fakeRepo) WriteLog(_ context.Context, input operationlog.WriteInput) error {
	r.logs = append(r.logs, input)
	return nil
}

type fakeUOW struct{ repo publicrepo.Repository }

func (u fakeUOW) WithinTransaction(ctx context.Context, fn func(publicrepo.Repository) error) error {
	return fn(u.repo)
}

func newTestService(repo *fakeRepo, allowed bool, now time.Time) *service {
	svc := NewService(repo, fakeUOW{repo: repo}, fakePermission{allowed: allowed}, contentsafety.AlwaysPass()).(*service)
	svc.now = func() time.Time { return now }
	return svc
}

func TestPublicApplicationFlow(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	t.Run("founder submits application and does not change graph version", func(t *testing.T) {
		repo := newFakeRepo()
		before := repo.family.GraphVersion
		result, err := newTestService(repo, true, now).Submit(context.Background(), 8, 2, publicdto.CreatePublicApplicationRequest{}, AuditInput{})
		if err != nil || result.Status != publicenum.StatusPending {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		if repo.family.PublicDisplayStatus != publicenum.PublicPending || repo.family.GraphVersion != before || len(repo.logs) != 1 {
			t.Fatal("submit did not update public status, preserve graph version, and log")
		}
	})
	t.Run("non-admin cannot submit", func(t *testing.T) {
		if _, err := newTestService(newFakeRepo(), false, now).Submit(context.Background(), 8, 2, publicdto.CreatePublicApplicationRequest{}, AuditInput{}); err == nil || err.Code != CodePublicApplicationForbidden {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("non-normal family cannot submit", func(t *testing.T) {
		repo := newFakeRepo()
		repo.family.Status = "DISSOLVED"
		if _, err := newTestService(repo, true, now).Submit(context.Background(), 8, 2, publicdto.CreatePublicApplicationRequest{}, AuditInput{}); err == nil || err.Code != CodePublicFamilyUnavailable {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("duplicate pending rejected", func(t *testing.T) {
		repo := newFakeRepo()
		svc := newTestService(repo, true, now)
		_, _ = svc.Submit(context.Background(), 8, 2, publicdto.CreatePublicApplicationRequest{}, AuditInput{})
		if _, err := svc.Submit(context.Background(), 8, 2, publicdto.CreatePublicApplicationRequest{}, AuditInput{}); err == nil || err.Code != CodePublicApplicationDuplicate {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("cancel pending by applicant resets private", func(t *testing.T) {
		repo := newFakeRepo()
		svc := newTestService(repo, false, now)
		actor := uint64(8)
		repo.applications[1] = &publicmodel.FamilyPublicApplication{ID: 1, FamilyID: 2, ApplicantUserID: &actor, ApplicationStatus: publicenum.StatusPending}
		result, err := svc.Cancel(context.Background(), actor, 2, 1, publicdto.CancelPublicApplicationRequest{}, AuditInput{})
		if err != nil || result.Status != publicenum.StatusCancelled || repo.family.PublicDisplayStatus != publicenum.PublicPrivate {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
	})
	t.Run("cancel denied for non-applicant non-admin", func(t *testing.T) {
		repo := newFakeRepo()
		actor := uint64(8)
		repo.applications[1] = &publicmodel.FamilyPublicApplication{ID: 1, FamilyID: 2, ApplicantUserID: &actor, ApplicationStatus: publicenum.StatusPending}
		if _, err := newTestService(repo, false, now).Cancel(context.Background(), 9, 2, 1, publicdto.CancelPublicApplicationRequest{}, AuditInput{}); err == nil || err.Code != CodePublicApplicationForbidden {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("family manager disables approved public display without revoking permission", func(t *testing.T) {
		repo := newFakeRepo()
		repo.family.PublicDisplayStatus = publicenum.PublicApproved
		repo.family.PublicDisplayEnabled = true
		before := repo.family.GraphVersion
		result, err := newTestService(repo, true, now).TakeDownUser(
			context.Background(), 8, 2, publicdto.TakeDownPublicFamilyRequest{}, AuditInput{},
		)
		if err != nil || result.PublicDisplayStatus != publicenum.PublicApproved || result.PublicDisplayEnabled {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		if repo.family.GraphVersion != before || len(repo.logs) != 1 ||
			repo.logs[0].OperatorUserID == nil || repo.logs[0].Action != "DISABLE_PUBLIC_DISPLAY" {
			t.Fatal("user disable must preserve graph version and write user operation log")
		}
	})
	t.Run("non-manager cannot close public display", func(t *testing.T) {
		repo := newFakeRepo()
		repo.family.PublicDisplayStatus = publicenum.PublicApproved
		if _, err := newTestService(repo, false, now).TakeDownUser(
			context.Background(), 8, 2, publicdto.TakeDownPublicFamilyRequest{}, AuditInput{},
		); err == nil || err.Code != CodePublicApplicationForbidden {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("private family cannot be closed again", func(t *testing.T) {
		repo := newFakeRepo()
		if _, err := newTestService(repo, true, now).TakeDownUser(
			context.Background(), 8, 2, publicdto.TakeDownPublicFamilyRequest{}, AuditInput{},
		); err == nil || err.Code != CodePublicFamilyTakeDownDenied {
			t.Fatalf("unexpected %#v", err)
		}
	})
}

func TestPublicApplicationAdminReviewAndTakeDown(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	t.Run("approve and reject update statuses without graph version change", func(t *testing.T) {
		for _, tc := range []struct {
			name         string
			review       func(Service, context.Context, uint64, string, uint64, publicdto.ReviewPublicApplicationRequest, AuditInput) (*vo.PublicApplication, *apperrors.BusinessError)
			appStatus    string
			familyStatus string
		}{
			{name: "approve", review: Service.Approve, appStatus: publicenum.StatusApproved, familyStatus: publicenum.PublicApproved},
			{name: "reject", review: Service.Reject, appStatus: publicenum.StatusRejected, familyStatus: publicenum.PublicRejected},
		} {
			t.Run(tc.name, func(t *testing.T) {
				repo := newFakeRepo()
				before := repo.family.GraphVersion
				repo.applications[1] = &publicmodel.FamilyPublicApplication{ID: 1, FamilyID: 2, ApplicationStatus: publicenum.StatusPending}
				result, err := tc.review(newTestService(repo, true, now), context.Background(), 1, string(enums.AdminRolePlatformAdmin), 1, publicdto.ReviewPublicApplicationRequest{}, AuditInput{})
				if err != nil || result.Status != tc.appStatus || repo.family.PublicDisplayStatus != tc.familyStatus || repo.family.GraphVersion != before || len(repo.logs) != 1 {
					t.Fatalf("unexpected %#v %#v", result, err)
				}
			})
		}
	})
	t.Run("take down approved family", func(t *testing.T) {
		repo := newFakeRepo()
		repo.family.PublicDisplayStatus = publicenum.PublicApproved
		before := repo.family.GraphVersion
		result, err := newTestService(repo, true, now).TakeDown(context.Background(), 1, string(enums.AdminRoleRootAdmin), 2, publicdto.TakeDownPublicFamilyRequest{}, AuditInput{})
		if err != nil || result.PublicDisplayStatus != publicenum.PublicTakenDown || repo.family.GraphVersion != before || len(repo.logs) != 1 {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
	})
}

type publicCSOpenID struct{}

func (publicCSOpenID) ActiveWechatMiniOpenID(context.Context, uint64) (string, error) {
	return "public-test-openid", nil
}

func TestPublicApplicationSubmitContentSafetyRejected(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	repo := newFakeRepo()
	beforeGV := repo.family.GraphVersion
	beforeApps := len(repo.applications)
	client := contentsafety.NewFakeClient(contentsafety.SuggestReview)
	contentSafety := contentsafety.NewChecker(client, publicCSOpenID{}, nil)
	svc := NewService(repo, fakeUOW{repo: repo}, fakePermission{allowed: true}, contentSafety).(*service)
	svc.now = func() time.Time { return now }
	reason := "违规公开申请理由"
	_, businessErr := svc.Submit(context.Background(), 8, 2, publicdto.CreatePublicApplicationRequest{
		ApplicationReason: &reason,
	}, AuditInput{})
	if businessErr == nil || businessErr.Code != apperrors.CodeContentSafetyRejected {
		t.Fatalf("expected 49007, got %#v", businessErr)
	}
	if repo.family.GraphVersion != beforeGV || len(repo.applications) != beforeApps {
		t.Fatal("submit must not persist application or change graph version")
	}
}
