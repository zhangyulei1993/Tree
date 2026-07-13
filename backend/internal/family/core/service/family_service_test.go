package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/config"
	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/database"
	"tree/backend/internal/common/enums"
	"tree/backend/internal/common/permission"
	"tree/backend/internal/family/core/dto"
	familymodel "tree/backend/internal/family/core/model"
	familyrepo "tree/backend/internal/family/core/repository"
	membermodel "tree/backend/internal/family/member/model"
	rolemodel "tree/backend/internal/family/role/model"
	usermodel "tree/backend/internal/user/model"
)

func transactionalTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	db, err := database.Init(context.Background(), cfg.MySQL)
	if err != nil {
		t.Skipf("development MySQL unavailable: %v", err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin transaction: %v", tx.Error)
	}
	t.Cleanup(func() { _ = tx.Rollback().Error })
	return tx
}

func TestCreateFamilyWithoutPhoneVerified(t *testing.T) {
	ctx := context.Background()
	tx := transactionalTestDB(t)
	user := &usermodel.User{
		PhoneVerified: false, AccountOrigin: "P0_TEST", RegisterClient: "P0_TEST",
		Status: string(enums.StatusActive),
	}
	if err := tx.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatalf("create user fixture: %v", err)
	}

	service := NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), nil, nil, contentsafety.AlwaysPass())
	founderGender := string(enums.GenderMale)
	result, businessErr := service.Create(ctx, user.ID, dto.CreateFamilyRequest{
		Surname: "测", FounderGender: &founderGender,
	}, AuditInput{IP: "127.0.0.1", UserAgent: "p0-test"})
	if businessErr != nil {
		t.Fatalf("Create without phone verified: %v", businessErr)
	}
	if result.ID == 0 {
		t.Fatalf("expected created family, got %#v", result)
	}
}

func TestCreateFamilyInitializesFounderAndGraphVersion(t *testing.T) {
	ctx := context.Background()
	tx := transactionalTestDB(t)
	user := &usermodel.User{
		PhoneVerified: true, AccountOrigin: "P0_TEST", RegisterClient: "P0_TEST",
		Status: string(enums.StatusActive),
	}
	if err := tx.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatalf("create user fixture: %v", err)
	}

	service := NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), nil, nil, contentsafety.AlwaysPass())
	founderGender := string(enums.GenderMale)
	result, businessErr := service.Create(ctx, user.ID, dto.CreateFamilyRequest{
		Surname: "测", FounderGender: &founderGender,
	}, AuditInput{IP: "127.0.0.1", UserAgent: "p0-test"})
	if businessErr != nil {
		t.Fatalf("Create: %v", businessErr)
	}
	if result.GraphVersion != 1 || result.Role != string(enums.FamilyRoleFounder) ||
		result.CurrentFounderMemberID == nil {
		t.Fatalf("unexpected family result: %#v", result)
	}

	var family familymodel.Family
	if err := tx.WithContext(ctx).First(&family, result.ID).Error; err != nil {
		t.Fatalf("load family: %v", err)
	}
	if family.CurrentFounderMemberID == nil || *family.CurrentFounderMemberID != *result.CurrentFounderMemberID ||
		family.GraphVersion != 1 {
		t.Fatalf("founder/graph version mismatch: %#v", family)
	}

	var member membermodel.FamilyMember
	if err := tx.WithContext(ctx).First(&member, *family.CurrentFounderMemberID).Error; err != nil {
		t.Fatalf("load founder member: %v", err)
	}
	if member.FamilyID != family.ID || member.Status != string(enums.StatusActive) ||
		member.Gender != string(enums.GenderMale) {
		t.Fatalf("invalid founder member: %#v", member)
	}

	var links []rolemodel.FamilyMemberUserLink
	if err := tx.WithContext(ctx).
		Where("family_id = ? AND link_status = ? AND family_role = ?", family.ID, string(enums.StatusActive), string(enums.FamilyRoleFounder)).
		Find(&links).Error; err != nil {
		t.Fatalf("load founder links: %v", err)
	}
	if len(links) != 1 || links[0].MemberID != member.ID || links[0].UserID != user.ID {
		t.Fatalf("expected one initialized founder link, got %#v", links)
	}
}

type deniedFamilyPermission struct {
	permission.FamilyPermissionService
}

func (deniedFamilyPermission) CanManageFamily(context.Context, uint64, uint64) (bool, error) {
	return false, nil
}

type founderFamilyPermission struct {
	permission.FamilyPermissionService
	memberID uint64
}

func (p founderFamilyPermission) CanManageFamily(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}

func (p founderFamilyPermission) GetActiveLink(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error) {
	return &rolemodel.FamilyMemberUserLink{
		MemberID: p.memberID,
		FamilyRole: string(enums.FamilyRoleFounder),
		LinkStatus: string(enums.StatusActive),
	}, nil
}

func TestNonFounderCannotCreateDissolutionRequest(t *testing.T) {
	service := NewFamilyService(nil, nil, deniedFamilyPermission{}, nil, contentsafety.AlwaysPass())
	result, businessErr := service.CreateDissolutionRequest(
		context.Background(), 99, 88, dto.CreateDissolutionRequest{}, AuditInput{},
	)
	if result != nil || businessErr == nil || businessErr.Code != CodeFamilyDissolutionForbidden {
		t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
	}
}

func TestMemberCanLeaveWithoutDeletingTreeNodeOrChangingGraphVersion(t *testing.T) {
	ctx := context.Background()
	tx := transactionalTestDB(t)
	founder := &usermodel.User{PhoneVerified: true, AccountOrigin: "P0_TEST", RegisterClient: "P0_TEST", Status: string(enums.StatusActive)}
	memberUser := &usermodel.User{PhoneVerified: true, AccountOrigin: "P0_TEST", RegisterClient: "P0_TEST", Status: string(enums.StatusActive)}
	if err := tx.WithContext(ctx).Create(founder).Error; err != nil {
		t.Fatalf("create founder: %v", err)
	}
	if err := tx.WithContext(ctx).Create(memberUser).Error; err != nil {
		t.Fatalf("create member user: %v", err)
	}
	repo := familyrepo.NewFamilyRepository(tx)
	service := NewFamilyService(tx, repo, nil, nil, contentsafety.AlwaysPass())
	created, businessErr := service.Create(ctx, founder.ID, dto.CreateFamilyRequest{Surname: "退"}, AuditInput{})
	if businessErr != nil {
		t.Fatalf("create family: %v", businessErr)
	}
	member := &membermodel.FamilyMember{
		FamilyID: created.ID, MemberType: "LINEAGE_MEMBER", DisplayName: "保留节点", Gender: "FEMALE",
		UserBindingPolicy: "OPTIONAL", Status: string(enums.StatusActive),
	}
	if err := tx.WithContext(ctx).Create(member).Error; err != nil {
		t.Fatalf("create member: %v", err)
	}
	now := time.Now()
	link := &rolemodel.FamilyMemberUserLink{
		FamilyID: created.ID, MemberID: member.ID, UserID: memberUser.ID,
		LinkStatus: string(enums.StatusActive), LinkSource: "P0_TEST", FamilyRole: string(enums.FamilyRoleMember), RoleGrantedAt: &now,
	}
	if err := tx.WithContext(ctx).Create(link).Error; err != nil {
		t.Fatalf("create link: %v", err)
	}

	result, businessErr := service.Leave(ctx, memberUser.ID, created.ID, dto.LeaveFamilyRequest{}, AuditInput{})
	if businessErr != nil {
		t.Fatalf("Leave: %v", businessErr)
	}
	if result.MemberID != member.ID || result.Status != "LEFT" {
		t.Fatalf("unexpected leave result: %#v", result)
	}
	var savedLink rolemodel.FamilyMemberUserLink
	if err := tx.WithContext(ctx).First(&savedLink, link.ID).Error; err != nil {
		t.Fatalf("load link: %v", err)
	}
	if savedLink.LinkStatus != "INACTIVE" || savedLink.UnlinkedAt == nil {
		t.Fatalf("link was not deactivated: %#v", savedLink)
	}
	var savedMember membermodel.FamilyMember
	if err := tx.WithContext(ctx).First(&savedMember, member.ID).Error; err != nil || savedMember.Status != string(enums.StatusActive) {
		t.Fatalf("tree node must remain active: %#v %v", savedMember, err)
	}
	var savedFamily familymodel.Family
	if err := tx.WithContext(ctx).First(&savedFamily, created.ID).Error; err != nil || savedFamily.GraphVersion != created.GraphVersion {
		t.Fatalf("leave must not change graph version: %#v %v", savedFamily, err)
	}
}

func TestFounderMustTransferBeforeLeave(t *testing.T) {
	ctx := context.Background()
	tx := transactionalTestDB(t)
	user := &usermodel.User{PhoneVerified: true, AccountOrigin: "P0_TEST", RegisterClient: "P0_TEST", Status: string(enums.StatusActive)}
	if err := tx.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	service := NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), nil, nil, contentsafety.AlwaysPass())
	created, businessErr := service.Create(ctx, user.ID, dto.CreateFamilyRequest{Surname: "创"}, AuditInput{})
	if businessErr != nil {
		t.Fatalf("create family: %v", businessErr)
	}
	_, businessErr = service.Leave(ctx, user.ID, created.ID, dto.LeaveFamilyRequest{}, AuditInput{})
	if businessErr == nil || businessErr.Code != CodeFamilyLeaveForbidden {
		t.Fatalf("expected founder leave rejection, got %#v", businessErr)
	}
}

func TestFounderCanRestoreFamilyDuringCooldown(t *testing.T) {
	ctx := context.Background()
	tx := transactionalTestDB(t)
	user := &usermodel.User{PhoneVerified: true, AccountOrigin: "P0_TEST", RegisterClient: "P0_TEST", Status: string(enums.StatusActive)}
	if err := tx.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	baseService := NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), nil, nil, contentsafety.AlwaysPass())
	created, businessErr := baseService.Create(ctx, user.ID, dto.CreateFamilyRequest{Surname: "复"}, AuditInput{})
	if businessErr != nil {
		t.Fatalf("create family: %v", businessErr)
	}
	now := time.Now()
	if err := tx.WithContext(ctx).Model(&familymodel.Family{}).Where("id = ?", created.ID).Updates(map[string]any{
		"status":                     string(enums.StatusDissolutionCooldown),
		"dissolution_cooldown_until": now.AddDate(0, 0, 7),
		"dissolution_cooldown_days":  7,
		"dissolution_hidden_at":      now,
	}).Error; err != nil {
		t.Fatalf("seed cooldown status: %v", err)
	}
	service := NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), founderFamilyPermission{memberID: *created.CurrentFounderMemberID}, nil, contentsafety.AlwaysPass())
	result, businessErr := service.RestoreDissolution(ctx, user.ID, created.ID, AuditInput{})
	if businessErr != nil {
		t.Fatalf("RestoreDissolution: %v", businessErr)
	}
	if result.Status != string(enums.StatusNormal) || result.PublicDisplayStatus != string(enums.StatusPrivate) || !result.Searchable {
		t.Fatalf("unexpected restored family: %#v", result)
	}
	var family familymodel.Family
	if err := tx.WithContext(ctx).First(&family, created.ID).Error; err != nil {
		t.Fatalf("reload family: %v", err)
	}
	if family.Status != string(enums.StatusNormal) || family.DissolutionCooldownUntil != nil || family.DissolutionCompletedAt != nil {
		t.Fatalf("family not restored correctly: %#v", family)
	}
}
