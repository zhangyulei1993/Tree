package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/enums"
	familymodel "tree/backend/internal/family/core/model"
	membermodel "tree/backend/internal/family/member/model"
	roledto "tree/backend/internal/family/role/dto"
	rolemodel "tree/backend/internal/family/role/model"
	rolerepo "tree/backend/internal/family/role/repository"
	operationlog "tree/backend/internal/operationlog/service"
)

type fakeRepo struct {
	family  familymodel.Family
	members map[uint64]*membermodel.FamilyMember
	links   map[uint64]*rolemodel.FamilyMemberUserLink
	logs    []operationlog.WriteInput
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		family: familymodel.Family{ID: 2, Status: string(enums.StatusNormal), GraphVersion: 9},
		members: map[uint64]*membermodel.FamilyMember{
			3: {ID: 3, FamilyID: 2, DisplayName: "Founder", Status: string(enums.StatusActive), UserBindingPolicy: string(enums.UserBindingOptional)},
			4: {ID: 4, FamilyID: 2, DisplayName: "Target", Status: string(enums.StatusActive), UserBindingPolicy: string(enums.UserBindingOptional)},
		},
		links: map[uint64]*rolemodel.FamilyMemberUserLink{
			1: {ID: 1, FamilyID: 2, MemberID: 3, UserID: 8, LinkStatus: string(enums.StatusActive), FamilyRole: string(enums.FamilyRoleFounder)},
			2: {ID: 2, FamilyID: 2, MemberID: 4, UserID: 9, LinkStatus: string(enums.StatusActive), FamilyRole: string(enums.FamilyRoleMember)},
		},
	}
}

func (r *fakeRepo) WithTx(*gorm.DB) rolerepo.Repository { return r }
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
func (r *fakeRepo) UpdateLinkRole(_ context.Context, linkID uint64, values map[string]any) error {
	link, ok := r.links[linkID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, ok := values["family_role"].(string); ok {
		link.FamilyRole = v
	}
	if v, ok := values["role_granted_at"].(time.Time); ok {
		link.RoleGrantedAt = &v
	}
	if v, ok := values["role_granted_by_user_id"].(uint64); ok {
		link.RoleGrantedByUserID = &v
	}
	return nil
}
func (r *fakeRepo) WriteLog(_ context.Context, input operationlog.WriteInput) error {
	r.logs = append(r.logs, input)
	return nil
}

type fakeUOW struct{ repo rolerepo.Repository }

func (u fakeUOW) WithinTransaction(ctx context.Context, fn func(rolerepo.Repository) error) error {
	return fn(u.repo)
}

func newTestService(repo *fakeRepo) *service {
	svc := NewService(repo, fakeUOW{repo: repo}).(*service)
	svc.now = func() time.Time { return time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC) }
	return svc
}

func TestSetAdminRules(t *testing.T) {
	t.Run("founder sets family admin", func(t *testing.T) {
		repo := newFakeRepo()
		before := repo.family.GraphVersion
		result, err := newTestService(repo).SetAdmin(context.Background(), 8, 2, 4, roledto.RoleChangeRequest{}, AuditInput{})
		if err != nil || result.FamilyRole != string(enums.FamilyRoleFamilyAdmin) {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		if repo.links[2].FamilyRole != string(enums.FamilyRoleFamilyAdmin) || repo.links[2].RoleGrantedByUserID == nil || repo.family.GraphVersion != before || len(repo.logs) != 1 {
			t.Fatal("role update, graph version, or operation log mismatch")
		}
	})
	t.Run("non-founder denied", func(t *testing.T) {
		repo := newFakeRepo()
		if _, err := newTestService(repo).SetAdmin(context.Background(), 9, 2, 4, roledto.RoleChangeRequest{}, AuditInput{}); err == nil || err.Code != CodeRoleForbidden {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("not-required member rejected", func(t *testing.T) {
		repo := newFakeRepo()
		repo.members[4].UserBindingPolicy = string(enums.UserBindingNotRequired)
		if _, err := newTestService(repo).SetAdmin(context.Background(), 8, 2, 4, roledto.RoleChangeRequest{}, AuditInput{}); err == nil || err.Code != CodeRoleTargetInvalid {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("member without active link rejected", func(t *testing.T) {
		repo := newFakeRepo()
		delete(repo.links, 2)
		if _, err := newTestService(repo).SetAdmin(context.Background(), 8, 2, 4, roledto.RoleChangeRequest{}, AuditInput{}); err == nil || err.Code != CodeRoleNoActiveLink {
			t.Fatalf("unexpected %#v", err)
		}
	})
}

func TestUnsetAdminRules(t *testing.T) {
	t.Run("founder cannot be unset", func(t *testing.T) {
		repo := newFakeRepo()
		if _, err := newTestService(repo).UnsetAdmin(context.Background(), 8, 2, 3, roledto.RoleChangeRequest{}, AuditInput{}); err == nil || err.Code != CodeRoleFounderGuard {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("unset family admin", func(t *testing.T) {
		repo := newFakeRepo()
		repo.links[2].FamilyRole = string(enums.FamilyRoleFamilyAdmin)
		result, err := newTestService(repo).UnsetAdmin(context.Background(), 8, 2, 4, roledto.RoleChangeRequest{}, AuditInput{})
		if err != nil || result.FamilyRole != string(enums.FamilyRoleMember) || repo.links[2].FamilyRole != string(enums.FamilyRoleMember) || len(repo.logs) != 1 {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
	})
}
