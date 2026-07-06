package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/contentsafety"
	coremodel "tree/backend/internal/family/core/model"
	"tree/backend/internal/family/member/dto"
	membermodel "tree/backend/internal/family/member/model"
	memberrepo "tree/backend/internal/family/member/repository"
	rolemodel "tree/backend/internal/family/role/model"
	usermodel "tree/backend/internal/user/model"
)

type memberPermFake struct {
	canManage bool
}

func (p memberPermFake) IsFamilyAdmin(context.Context, uint64, uint64) (bool, error) {
	return p.canManage, nil
}
func (p memberPermFake) IsFamilyMember(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (p memberPermFake) CanManageFamily(context.Context, uint64, uint64) (bool, error) {
	return p.canManage, nil
}
func (p memberPermFake) CanCreateMember(context.Context, uint64, uint64) (bool, error) {
	return p.canManage, nil
}
func (p memberPermFake) CanEditMember(context.Context, uint64, uint64, uint64) (bool, error) {
	return p.canManage, nil
}
func (p memberPermFake) CanDeleteMember(context.Context, uint64, uint64, uint64) (bool, error) {
	return p.canManage, nil
}
func (p memberPermFake) CanManageRelationships(context.Context, uint64, uint64) (bool, error) {
	return p.canManage, nil
}
func (p memberPermFake) GetActiveLink(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error) {
	return nil, gorm.ErrRecordNotFound
}

type memberRepoFake struct {
	family          coremodel.Family
	member          membermodel.FamilyMember
	links           []rolemodel.FamilyMemberUserLink
	relationshipCnt int64
}

func (r *memberRepoFake) WithTx(*gorm.DB) memberrepo.MemberRepository { return r }

func (r *memberRepoFake) LockFamily(context.Context, uint64) (*coremodel.Family, error) {
	value := r.family
	return &value, nil
}

func (r *memberRepoFake) IncrementGraphVersion(context.Context, uint64) error {
	r.family.GraphVersion++
	return nil
}

func (r *memberRepoFake) Create(context.Context, *membermodel.FamilyMember) error { return nil }

func (r *memberRepoFake) List(context.Context, uint64) ([]memberrepo.MemberRow, error) {
	return nil, nil
}

func (r *memberRepoFake) Find(_ context.Context, familyID, memberID uint64) (*memberrepo.MemberRow, error) {
	if r.member.ID != memberID || r.member.FamilyID != familyID {
		return nil, gorm.ErrRecordNotFound
	}
	return &memberrepo.MemberRow{FamilyMember: r.member}, nil
}

func (r *memberRepoFake) FindForUpdate(_ context.Context, familyID, memberID uint64) (*membermodel.FamilyMember, error) {
	if r.member.ID != memberID || r.member.FamilyID != familyID {
		return nil, gorm.ErrRecordNotFound
	}
	value := r.member
	return &value, nil
}

func (r *memberRepoFake) Update(context.Context, uint64, uint64, map[string]any) error { return nil }

func (r *memberRepoFake) SoftDelete(context.Context, uint64, uint64, uint64, *string, time.Time) error {
	return nil
}

func (r *memberRepoFake) CountActiveRelationships(context.Context, uint64, uint64) (int64, error) {
	return r.relationshipCnt, nil
}

func (r *memberRepoFake) FindUserForUpdate(context.Context, uint64) (*usermodel.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *memberRepoFake) FindActiveLinkByMember(_ context.Context, familyID, memberID uint64) (*rolemodel.FamilyMemberUserLink, error) {
	for i := range r.links {
		if r.links[i].FamilyID == familyID && r.links[i].MemberID == memberID && r.links[i].LinkStatus == "ACTIVE" {
			value := r.links[i]
			return &value, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *memberRepoFake) FindActiveLinkByUser(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *memberRepoFake) CreateLink(_ context.Context, value *rolemodel.FamilyMemberUserLink) error {
	r.links = append(r.links, *value)
	return nil
}

func (r *memberRepoFake) UnbindLink(_ context.Context, linkID, _ uint64, _ *string, _ time.Time) error {
	for i := range r.links {
		if r.links[i].ID == linkID {
			r.links[i].LinkStatus = "UNBOUND"
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func testMemberService(repo *memberRepoFake, canManage bool) MemberService {
	return NewMemberService(nil, repo, memberPermFake{canManage: canManage}, nil, contentsafety.AlwaysPass())
}

func TestUnbindUserRules(t *testing.T) {
	originalTx := runMemberTransaction
	runMemberTransaction = func(_ context.Context, _ *gorm.DB, fn func(tx *gorm.DB) error) error {
		return fn(nil)
	}
	t.Cleanup(func() { runMemberTransaction = originalTx })

	t.Run("forbidden without manage permission", func(t *testing.T) {
		repo := &memberRepoFake{
			family: coremodel.Family{ID: 22, Status: "NORMAL", GraphVersion: 10},
			member: membermodel.FamilyMember{ID: 6, FamilyID: 22, Status: "ACTIVE"},
			links: []rolemodel.FamilyMemberUserLink{
				{ID: 1, FamilyID: 22, MemberID: 6, UserID: 9, LinkStatus: "ACTIVE", FamilyRole: "MEMBER"},
			},
		}
		_, err := testMemberService(repo, false).UnbindUser(context.Background(), 8, 22, 6, dto.UnbindUserRequest{}, AuditInput{})
		if err == nil || err.Code != CodeMemberEditForbidden {
			t.Fatalf("unexpected %#v", err)
		}
	})

	t.Run("protected founder link rejected", func(t *testing.T) {
		repo := &memberRepoFake{
			family: coremodel.Family{ID: 22, Status: "NORMAL", GraphVersion: 10},
			member: membermodel.FamilyMember{ID: 6, FamilyID: 22, Status: "ACTIVE"},
			links: []rolemodel.FamilyMemberUserLink{
				{ID: 1, FamilyID: 22, MemberID: 6, UserID: 9, LinkStatus: "ACTIVE", FamilyRole: "FOUNDER"},
			},
		}
		beforeGV := repo.family.GraphVersion
		_, err := testMemberService(repo, true).UnbindUser(context.Background(), 8, 22, 6, dto.UnbindUserRequest{}, AuditInput{})
		if err == nil || err.Code != CodeMemberUnbindProtected {
			t.Fatalf("unexpected %#v", err)
		}
		if repo.links[0].LinkStatus != "ACTIVE" {
			t.Fatal("protected unbind must not change link")
		}
		if repo.family.GraphVersion != beforeGV {
			t.Fatal("failed unbind must not increment graph version")
		}
	})

	t.Run("member unbind keeps node and does not bump graph version", func(t *testing.T) {
		repo := &memberRepoFake{
			family:          coremodel.Family{ID: 22, Status: "NORMAL", GraphVersion: 10},
			member:          membermodel.FamilyMember{ID: 6, FamilyID: 22, Status: "ACTIVE"},
			relationshipCnt: 2,
			links: []rolemodel.FamilyMemberUserLink{
				{ID: 1, FamilyID: 22, MemberID: 6, UserID: 9, LinkStatus: "ACTIVE", FamilyRole: "MEMBER"},
			},
		}
		beforeGV := repo.family.GraphVersion
		result, err := testMemberService(repo, true).UnbindUser(context.Background(), 8, 22, 6, dto.UnbindUserRequest{}, AuditInput{})
		if err != nil || result == nil {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		if repo.links[0].LinkStatus != "UNBOUND" {
			t.Fatal("link should be unbound")
		}
		if repo.member.Status != "ACTIVE" {
			t.Fatal("member node must remain active")
		}
		if repo.relationshipCnt != 2 {
			t.Fatal("relationship count must remain unchanged")
		}
		if repo.family.GraphVersion != beforeGV {
			t.Fatal("unbind must not increment graph version")
		}
	})
}
