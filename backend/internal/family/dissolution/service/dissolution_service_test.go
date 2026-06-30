package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	adminmodel "tree/backend/internal/admin/model"
	"tree/backend/internal/common/enums"
	familymodel "tree/backend/internal/family/core/model"
	dissolutiondto "tree/backend/internal/family/dissolution/dto"
	dissolutionenum "tree/backend/internal/family/dissolution/enum"
	dissolutionmodel "tree/backend/internal/family/dissolution/model"
	dissolutionrepo "tree/backend/internal/family/dissolution/repository"
	operationlog "tree/backend/internal/operationlog/service"
)

type fakeRepo struct {
	family   familymodel.Family
	requests map[uint64]*dissolutionmodel.FamilyDissolutionRequest
	logs     []operationlog.WriteInput
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		family: familymodel.Family{ID: 2, FamilyName: "Tree", Status: "DISSOLUTION_PENDING", PublicDisplayStatus: "APPROVED", Searchable: true, GraphVersion: 11},
		requests: map[uint64]*dissolutionmodel.FamilyDissolutionRequest{
			1: {ID: 1, FamilyID: 2, RequesterMemberID: 3, RequesterUserID: 8, RequestStatus: dissolutionenum.StatusPending},
		},
	}
}

func (r *fakeRepo) WithTx(*gorm.DB) dissolutionrepo.Repository { return r }
func (r *fakeRepo) FindAdmin(context.Context, uint64) (*adminmodel.AdminUser, error) {
	return &adminmodel.AdminUser{ID: 1, Role: string(enums.AdminRoleRootAdmin)}, nil
}
func (r *fakeRepo) FindFamily(context.Context, uint64, bool) (*familymodel.Family, error) {
	v := r.family
	return &v, nil
}
func (r *fakeRepo) FindByID(_ context.Context, id uint64, _ bool) (*dissolutionmodel.FamilyDissolutionRequest, error) {
	v, ok := r.requests[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *v
	return &copy, nil
}
func (r *fakeRepo) List(_ context.Context, _ dissolutionrepo.ListQuery) ([]dissolutionrepo.DissolutionRow, int64, error) {
	rows := make([]dissolutionrepo.DissolutionRow, 0)
	for _, v := range r.requests {
		rows = append(rows, dissolutionrepo.DissolutionRow{
			FamilyDissolutionRequest: *v,
			FamilyName:               r.family.FamilyName,
			FamilyStatus:             r.family.Status,
		})
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
	return nil
}
func (r *fakeRepo) UpdateFamily(_ context.Context, _ uint64, values map[string]any) error {
	if value, ok := values["status"].(string); ok {
		r.family.Status = value
	}
	if value, ok := values["public_display_status"].(string); ok {
		r.family.PublicDisplayStatus = value
	}
	if value, ok := values["searchable"].(bool); ok {
		r.family.Searchable = value
	}
	if value, ok := values["dissolved_at"].(time.Time); ok {
		r.family.DissolvedAt = &value
	}
	if value, ok := values["restored_at"].(time.Time); ok {
		r.family.RestoredAt = &value
	}
	return nil
}
func (r *fakeRepo) WriteLog(_ context.Context, input operationlog.WriteInput) error {
	r.logs = append(r.logs, input)
	return nil
}

type fakeUOW struct{ repo dissolutionrepo.Repository }

func (u fakeUOW) WithinTransaction(ctx context.Context, fn func(dissolutionrepo.Repository) error) error {
	return fn(u.repo)
}

func newTestService(repo *fakeRepo) *service {
	svc := NewService(repo, fakeUOW{repo: repo}).(*service)
	svc.now = func() time.Time { return time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC) }
	return svc
}

func TestDissolutionReviewRules(t *testing.T) {
	t.Run("platform admin denied", func(t *testing.T) {
		repo := newFakeRepo()
		if _, err := newTestService(repo).Approve(context.Background(), 1, string(enums.AdminRolePlatformAdmin), 1, dissolutiondto.ReviewDissolutionRequest{}, AuditInput{}); err == nil || err.Code != CodeDissolutionAdminDenied {
			t.Fatalf("unexpected %#v", err)
		}
		if _, err := newTestService(repo).Reject(context.Background(), 1, string(enums.AdminRolePlatformAdmin), 1, dissolutiondto.ReviewDissolutionRequest{}, AuditInput{}); err == nil || err.Code != CodeDissolutionAdminDenied {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("root approves dissolution", func(t *testing.T) {
		repo := newFakeRepo()
		before := repo.family.GraphVersion
		result, err := newTestService(repo).Approve(context.Background(), 1, string(enums.AdminRoleRootAdmin), 1, dissolutiondto.ReviewDissolutionRequest{}, AuditInput{})
		if err != nil || result.RequestStatus != dissolutionenum.StatusApproved {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		if repo.family.Status != string(enums.StatusDissolved) || repo.family.PublicDisplayStatus != string(enums.StatusPrivate) ||
			repo.family.Searchable || repo.family.DissolvedAt == nil || repo.family.GraphVersion != before || len(repo.logs) != 1 {
			t.Fatal("dissolution approval did not update family correctly")
		}
	})
	t.Run("super admin rejects dissolution", func(t *testing.T) {
		repo := newFakeRepo()
		result, err := newTestService(repo).Reject(context.Background(), 1, string(enums.AdminRoleSuperAdmin), 1, dissolutiondto.ReviewDissolutionRequest{}, AuditInput{})
		if err != nil || result.RequestStatus != dissolutionenum.StatusRejected || repo.family.Status != string(enums.StatusNormal) || len(repo.logs) != 1 {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
	})
}

func TestRestoreFamily(t *testing.T) {
	t.Run("restore dissolved family", func(t *testing.T) {
		repo := newFakeRepo()
		repo.family.Status = string(enums.StatusDissolved)
		repo.family.PublicDisplayStatus = string(enums.StatusTakenDown)
		repo.family.Searchable = false
		before := repo.family.GraphVersion
		result, err := newTestService(repo).Restore(context.Background(), 1, string(enums.AdminRoleRootAdmin), 2, dissolutiondto.RestoreFamilyRequest{}, AuditInput{})
		if err != nil || result.Status != string(enums.StatusNormal) || result.PublicDisplayStatus != string(enums.StatusPrivate) || !result.Searchable {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		if repo.family.GraphVersion != before || repo.family.RestoredAt == nil || len(repo.logs) != 1 {
			t.Fatal("restore updated graph version or failed to log")
		}
	})
	t.Run("platform cannot restore", func(t *testing.T) {
		repo := newFakeRepo()
		repo.family.Status = string(enums.StatusDissolved)
		if _, err := newTestService(repo).Restore(context.Background(), 1, string(enums.AdminRolePlatformAdmin), 2, dissolutiondto.RestoreFamilyRequest{}, AuditInput{}); err == nil || err.Code != CodeFamilyRestoreAdminDenied {
			t.Fatalf("unexpected %#v", err)
		}
	})
}
