package service

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"tree/backend/internal/common/config"
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

	service := NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), nil)
	result, businessErr := service.Create(ctx, user.ID, dto.CreateFamilyRequest{
		Surname: "测",
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
	if member.FamilyID != family.ID || member.Status != string(enums.StatusActive) {
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

func TestNonFounderCannotCreateDissolutionRequest(t *testing.T) {
	service := NewFamilyService(nil, nil, deniedFamilyPermission{})
	result, businessErr := service.CreateDissolutionRequest(
		context.Background(), 99, 88, dto.CreateDissolutionRequest{}, AuditInput{},
	)
	if result != nil || businessErr == nil || businessErr.Code != CodeFamilyDissolutionForbidden {
		t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
	}
}
