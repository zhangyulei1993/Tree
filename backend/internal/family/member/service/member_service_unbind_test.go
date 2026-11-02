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
	family            coremodel.Family
	member            membermodel.FamilyMember
	user              *usermodel.User
	links             []rolemodel.FamilyMemberUserLink
	blockingChildCnt  int64
	unsafeSpouseChild bool
	relationsDeleted  bool
	restoreConflict   bool
	restored          bool
	recoveryDiscarded bool
	lastUpdate        map[string]any
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

func (r *memberRepoFake) Update(_ context.Context, _ uint64, _ uint64, values map[string]any) error {
	r.lastUpdate = make(map[string]any, len(values))
	for key, value := range values {
		r.lastUpdate[key] = value
	}
	return nil
}

func (r *memberRepoFake) SoftDelete(context.Context, uint64, uint64, uint64, *string, time.Time) error {
	r.member.Status = "DELETED"
	return nil
}

func (r *memberRepoFake) CountBlockingRelationships(context.Context, uint64, uint64) (int64, error) {
	return r.blockingChildCnt, nil
}

func (r *memberRepoFake) HasChildWithoutOtherLineageParent(context.Context, uint64, uint64) (bool, error) {
	return r.unsafeSpouseChild, nil
}

func (r *memberRepoFake) SoftDeleteDeletableRelationships(context.Context, uint64, uint64, uint64, *string, time.Time) (int64, error) {
	r.relationsDeleted = true
	return 1, nil
}

func (r *memberRepoFake) ListDeletedSince(context.Context, uint64, time.Time) ([]membermodel.FamilyMember, error) {
	return nil, nil
}

func (r *memberRepoFake) FindDeletedForUpdate(context.Context, uint64, uint64, time.Time) (*membermodel.FamilyMember, error) {
	if r.member.Status != "DELETED" {
		return nil, gorm.ErrRecordNotFound
	}
	value := r.member
	return &value, nil
}

func (r *memberRepoFake) HasRestorePositionConflict(context.Context, *membermodel.FamilyMember, time.Time) (bool, error) {
	return r.restoreConflict, nil
}

func (r *memberRepoFake) RestoreDeleted(context.Context, uint64, uint64, time.Time) error {
	r.restored = true
	return nil
}

func (r *memberRepoFake) DiscardRecovery(context.Context, uint64, uint64, uint64, time.Time) error {
	r.recoveryDiscarded = true
	return nil
}

func (r *memberRepoFake) RestoreRelationshipsDeletedAt(context.Context, uint64, uint64, time.Time) error {
	return nil
}

func (r *memberRepoFake) FindUserForUpdate(_ context.Context, userID uint64) (*usermodel.User, error) {
	if r.user == nil || r.user.ID != userID {
		return nil, gorm.ErrRecordNotFound
	}
	value := *r.user
	return &value, nil
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

type memberTreeCache struct {
	deleted []string
}

func (c *memberTreeCache) Get(context.Context, string) ([]byte, error) {
	return nil, gorm.ErrRecordNotFound
}
func (c *memberTreeCache) Set(context.Context, string, []byte, time.Duration) error {
	return nil
}
func (c *memberTreeCache) Delete(_ context.Context, key string) error {
	c.deleted = append(c.deleted, key)
	return nil
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
			family:           coremodel.Family{ID: 22, Status: "NORMAL", GraphVersion: 10},
			member:           membermodel.FamilyMember{ID: 6, FamilyID: 22, Status: "ACTIVE"},
			blockingChildCnt: 2,
			links: []rolemodel.FamilyMemberUserLink{
				{ID: 1, FamilyID: 22, MemberID: 6, UserID: 9, LinkStatus: "ACTIVE", FamilyRole: "MEMBER"},
			},
		}
		beforeGV := repo.family.GraphVersion
		cache := &memberTreeCache{}
		service := NewMemberService(nil, repo, memberPermFake{canManage: true}, nil, contentsafety.AlwaysPass(), cache)
		result, err := service.UnbindUser(context.Background(), 8, 22, 6, dto.UnbindUserRequest{}, AuditInput{})
		if err != nil || result == nil {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		if repo.links[0].LinkStatus != "UNBOUND" {
			t.Fatal("link should be unbound")
		}
		if repo.member.Status != "ACTIVE" {
			t.Fatal("member node must remain active")
		}
		if repo.blockingChildCnt != 2 {
			t.Fatal("relationship count must remain unchanged")
		}
		if repo.family.GraphVersion != beforeGV {
			t.Fatal("unbind must not increment graph version")
		}
		if len(cache.deleted) != 1 || cache.deleted[0] != "family:22:tree:v10" {
			t.Fatalf("unbind must invalidate current tree cache, got %#v", cache.deleted)
		}
	})
}

func TestBindUserInvalidatesTreeCacheWithoutGraphVersionChange(t *testing.T) {
	originalTx := runMemberTransaction
	runMemberTransaction = func(_ context.Context, _ *gorm.DB, fn func(tx *gorm.DB) error) error {
		return fn(nil)
	}
	t.Cleanup(func() { runMemberTransaction = originalTx })

	repo := &memberRepoFake{
		family: coremodel.Family{ID: 22, Status: "NORMAL", GraphVersion: 10},
		member: membermodel.FamilyMember{ID: 6, FamilyID: 22, Status: "ACTIVE", UserBindingPolicy: "OPTIONAL"},
		user:   &usermodel.User{ID: 9, Status: "ACTIVE"},
	}
	cache := &memberTreeCache{}
	service := NewMemberService(nil, repo, memberPermFake{canManage: true}, nil, contentsafety.AlwaysPass(), cache)
	result, err := service.BindUser(context.Background(), 8, 22, 6, dto.BindUserRequest{UserID: 9}, AuditInput{})
	if err != nil || result == nil {
		t.Fatalf("unexpected %#v %#v", result, err)
	}
	if repo.family.GraphVersion != 10 {
		t.Fatal("manual binding must not increment graph version")
	}
	if len(cache.deleted) != 1 || cache.deleted[0] != "family:22:tree:v10" {
		t.Fatalf("bind must invalidate current tree cache, got %#v", cache.deleted)
	}
}

func TestUpdateProfileVerificationFollowsEditorIdentity(t *testing.T) {
	originalTx := runMemberTransaction
	runMemberTransaction = func(_ context.Context, _ *gorm.DB, fn func(tx *gorm.DB) error) error {
		return fn(nil)
	}
	t.Cleanup(func() { runMemberTransaction = originalTx })

	name := "新名字"
	t.Run("bound member saving their own profile confirms it automatically", func(t *testing.T) {
		repo := &memberRepoFake{
			family: coremodel.Family{ID: 22, Status: "NORMAL", GraphVersion: 10},
			member: membermodel.FamilyMember{ID: 6, FamilyID: 22, Status: "ACTIVE", Gender: "MALE", Revision: 1},
			links:  []rolemodel.FamilyMemberUserLink{{ID: 1, FamilyID: 22, MemberID: 6, UserID: 9, LinkStatus: "ACTIVE"}},
		}
		result, err := testMemberService(repo, true).Update(context.Background(), 9, 22, 6, dto.UpdateMemberRequest{Name: &name}, AuditInput{})
		if err != nil || result == nil {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		if got := repo.lastUpdate["profile_verification_status"]; got != "SELF_CONFIRMED" {
			t.Fatalf("self edit should confirm the profile, got %#v", got)
		}
		if got := repo.lastUpdate["profile_confirmed_by_user_id"]; got != uint64(9) {
			t.Fatalf("self edit should record the editor, got %#v", got)
		}
		if _, ok := repo.lastUpdate["profile_confirmed_at"].(time.Time); !ok {
			t.Fatalf("self edit should store a confirmation time, got %#v", repo.lastUpdate["profile_confirmed_at"])
		}
	})

	t.Run("manager editing another member still leaves the profile pending", func(t *testing.T) {
		repo := &memberRepoFake{
			family: coremodel.Family{ID: 22, Status: "NORMAL", GraphVersion: 10},
			member: membermodel.FamilyMember{ID: 6, FamilyID: 22, Status: "ACTIVE", Gender: "MALE", Revision: 1},
			links:  []rolemodel.FamilyMemberUserLink{{ID: 1, FamilyID: 22, MemberID: 6, UserID: 9, LinkStatus: "ACTIVE"}},
		}
		result, err := testMemberService(repo, true).Update(context.Background(), 8, 22, 6, dto.UpdateMemberRequest{Name: &name}, AuditInput{})
		if err != nil || result == nil {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		if got := repo.lastUpdate["profile_verification_status"]; got != "PENDING" {
			t.Fatalf("manager edit should keep the profile pending, got %#v", got)
		}
		if repo.lastUpdate["profile_confirmed_at"] != nil || repo.lastUpdate["profile_confirmed_by_user_id"] != nil {
			t.Fatalf("manager edit should clear prior confirmation, got %#v", repo.lastUpdate)
		}
	})
}

func TestRestoreMemberRejectsReplacementAtSamePosition(t *testing.T) {
	originalTx := runMemberTransaction
	runMemberTransaction = func(_ context.Context, _ *gorm.DB, fn func(tx *gorm.DB) error) error {
		return fn(nil)
	}
	t.Cleanup(func() { runMemberTransaction = originalTx })

	creatorID := uint64(8)
	deletedAt := time.Now().Add(-time.Hour)
	repo := &memberRepoFake{
		family: coremodel.Family{ID: 22, Status: "NORMAL", CreatorUserID: &creatorID, GraphVersion: 10},
		member: membermodel.FamilyMember{
			ID: 22, FamilyID: 22, DisplayName: "樊一瑾", Gender: "FEMALE", Status: "DELETED", DeletedAt: &deletedAt,
		},
		restoreConflict: true,
	}

	cache := &memberTreeCache{}
	service := NewMemberService(nil, repo, memberPermFake{canManage: true}, nil, contentsafety.AlwaysPass(), cache)
	err := service.Restore(context.Background(), creatorID, 22, 22, AuditInput{})
	if err == nil || err.Code != CodeMemberDataConflict {
		t.Fatalf("expected restore conflict, got %#v", err)
	}
	if repo.restored {
		t.Fatal("conflicted restore must not reactivate the deleted member")
	}
	if repo.family.GraphVersion != 10 {
		t.Fatal("conflicted restore must not increment graph version")
	}

	t.Run("restores when the original position is still vacant", func(t *testing.T) {
		repo.restoreConflict = false
		repo.restored = false
		err := service.Restore(context.Background(), creatorID, 22, 22, AuditInput{})
		if err != nil {
			t.Fatalf("expected restore to succeed, got %#v", err)
		}
		if !repo.restored {
			t.Fatal("vacant restore should reactivate the deleted member")
		}
		if repo.family.GraphVersion != 11 {
			t.Fatalf("successful restore must increment graph version once, got %d", repo.family.GraphVersion)
		}
		if len(cache.deleted) != 1 || cache.deleted[0] != "family:22:tree:v10" {
			t.Fatalf("restore must invalidate the prior tree cache, got %#v", cache.deleted)
		}
	})
}

func TestDiscardRecoveryRequiresCreatorAndKeepsGraphVersion(t *testing.T) {
	originalTx := runMemberTransaction
	runMemberTransaction = func(_ context.Context, _ *gorm.DB, fn func(tx *gorm.DB) error) error {
		return fn(nil)
	}
	t.Cleanup(func() { runMemberTransaction = originalTx })

	creatorID := uint64(8)
	deletedAt := time.Now().Add(-time.Hour)
	repo := &memberRepoFake{
		family: coremodel.Family{ID: 22, Status: "NORMAL", CreatorUserID: &creatorID, GraphVersion: 10},
		member: membermodel.FamilyMember{ID: 22, FamilyID: 22, Status: "DELETED", DeletedAt: &deletedAt},
	}
	if err := testMemberService(repo, true).DiscardRecovery(context.Background(), creatorID, 22, 22, AuditInput{}); err != nil {
		t.Fatalf("expected discard to succeed, got %#v", err)
	}
	if !repo.recoveryDiscarded {
		t.Fatal("discard should remove recovery eligibility")
	}
	if repo.family.GraphVersion != 10 {
		t.Fatal("discarding recovery eligibility must not change the tree graph")
	}
}

func TestDeleteMemberRelationshipRules(t *testing.T) {
	originalTx := runMemberTransaction
	runMemberTransaction = func(_ context.Context, _ *gorm.DB, fn func(tx *gorm.DB) error) error {
		return fn(nil)
	}
	t.Cleanup(func() { runMemberTransaction = originalTx })

	t.Run("member with only incoming parent links can be deleted", func(t *testing.T) {
		repo := &memberRepoFake{
			family: coremodel.Family{ID: 22, Status: "NORMAL", GraphVersion: 10},
			member: membermodel.FamilyMember{ID: 6, FamilyID: 22, Status: "ACTIVE"},
		}
		err := testMemberService(repo, true).Delete(context.Background(), 8, 22, 6, dto.DeleteMemberRequest{}, AuditInput{})
		if err != nil {
			t.Fatalf("unexpected %#v", err)
		}
		if repo.member.Status != "DELETED" {
			t.Fatal("member should be deleted")
		}
		if !repo.relationsDeleted {
			t.Fatal("deletable relationships should be soft deleted with the member")
		}
		if repo.family.GraphVersion != 11 {
			t.Fatal("delete must increment graph version once")
		}
	})

	t.Run("member with children is rejected", func(t *testing.T) {
		repo := &memberRepoFake{
			family:           coremodel.Family{ID: 22, Status: "NORMAL", GraphVersion: 10},
			member:           membermodel.FamilyMember{ID: 6, FamilyID: 22, Status: "ACTIVE"},
			blockingChildCnt: 1,
		}
		err := testMemberService(repo, true).Delete(context.Background(), 8, 22, 6, dto.DeleteMemberRequest{}, AuditInput{})
		if err == nil || err.Code != CodeMemberHasRelationships {
			t.Fatalf("unexpected %#v", err)
		}
		if repo.member.Status != "ACTIVE" {
			t.Fatal("blocked delete must keep member active")
		}
		if repo.relationsDeleted {
			t.Fatal("blocked delete must not delete relationships")
		}
		if repo.family.GraphVersion != 10 {
			t.Fatal("blocked delete must not increment graph version")
		}
	})

	t.Run("external spouse with lineage co-parented children can be deleted", func(t *testing.T) {
		repo := &memberRepoFake{
			family:           coremodel.Family{ID: 22, Status: "NORMAL", GraphVersion: 10},
			member:           membermodel.FamilyMember{ID: 6, FamilyID: 22, MemberType: "SPOUSE", Status: "ACTIVE"},
			blockingChildCnt: 2,
		}
		err := testMemberService(repo, true).Delete(context.Background(), 8, 22, 6, dto.DeleteMemberRequest{}, AuditInput{})
		if err != nil {
			t.Fatalf("unexpected %#v", err)
		}
		if repo.member.Status != "DELETED" || !repo.relationsDeleted {
			t.Fatal("safe external spouse deletion must remove the node and related edges")
		}
	})

	t.Run("external spouse with an unpaired child is rejected", func(t *testing.T) {
		repo := &memberRepoFake{
			family:            coremodel.Family{ID: 22, Status: "NORMAL", GraphVersion: 10},
			member:            membermodel.FamilyMember{ID: 6, FamilyID: 22, MemberType: "SPOUSE", Status: "ACTIVE"},
			blockingChildCnt:  1,
			unsafeSpouseChild: true,
		}
		err := testMemberService(repo, true).Delete(context.Background(), 8, 22, 6, dto.DeleteMemberRequest{}, AuditInput{})
		if err == nil || err.Code != CodeMemberHasRelationships || repo.member.Status != "ACTIVE" {
			t.Fatalf("unexpected %#v", err)
		}
	})
}
