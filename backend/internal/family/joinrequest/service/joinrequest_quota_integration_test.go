package service

import (
	"context"
	"fmt"
	"testing"

	"gorm.io/gorm"

	authservice "tree/backend/internal/auth/service"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/family/joinrequest/dto"
	joinmodel "tree/backend/internal/family/joinrequest/model"
	joinrepo "tree/backend/internal/family/joinrequest/repository"
	rolemodel "tree/backend/internal/family/role/model"
	"tree/backend/internal/testsupport/freshauth"
	"tree/backend/internal/testsupport/freshquota"
)

func newFreshJoinRequestService(tx *gorm.DB) Service {
	quotaSvc := freshquota.NewQuotaService(tx)
	repo := joinrepo.NewRepository(tx)
	return NewService(repo, joinrepo.NewUnitOfWork(tx, repo), freshquota.AllowAllFamilyPerm{}, quotaSvc, contentsafety.AlwaysPass())
}

func TestJoinRequestApproveRespectsJoinedQuotaFreshAccount(t *testing.T) {
	ctx := context.Background()
	tx, runID := freshquota.TestDB(t)
	quotaSvc := freshquota.NewQuotaService(tx)
	joinSvc := newFreshJoinRequestService(tx)

	wechatClient := freshquota.NewFakeWechatMiniClient("fresh-quota-test-app")
	authSvc := freshauth.NewAuthService(t, tx, wechatClient)
	code := wechatClient.BindUniqueCode(runID, "applicant")
	login, err := authSvc.WechatMiniLogin(ctx, authservice.WechatMiniLoginInput{
		Code: code, ClientType: "WECHAT_MINI_PROGRAM", IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("applicant login: %v", err)
	}
	applicantID := login.User.ID
	nickname := fmt.Sprintf("申请%s", runID[len(runID)-4:])
	_, err = authSvc.UpdateProfile(ctx, authservice.UpdateProfileInput{UserID: applicantID, Nickname: &nickname, IP: "127.0.0.1", UserAgent: "fresh-quota-test"})
	if err != nil {
		t.Fatalf("applicant profile: %v", err)
	}
	freshquota.CreateFamily(t, freshquota.NewFamilyService(tx, quotaSvc), applicantID, "自")

	host := freshquota.CreateActiveUser(t, tx, "审批主持", false)
	familyID, memberID := freshquota.SeedHostFamilyWithGuest(t, tx, host, "申")
	graphBefore := freshquota.GraphVersion(t, tx, familyID)

	created, err := joinSvc.Create(ctx, applicantID, familyID, dto.CreateJoinRequest{ApplicantGender: stringPtr("MALE")}, AuditInput{})
	if err != nil {
		t.Fatalf("create join request: %v", err)
	}
	approved, err := joinSvc.Approve(ctx, host.ID, familyID, created.RequestID, dto.ApproveJoinRequest{
		ApproveMode: "BIND_EXISTING_MEMBER", MemberID: &memberID,
	}, AuditInput{})
	if err != nil {
		t.Fatalf("approve join request: %v", err)
	}
	if approved.RequestStatus != "APPROVED" {
		t.Fatalf("expected APPROVED, got %s", approved.RequestStatus)
	}
	if freshquota.GraphVersion(t, tx, familyID) != graphBefore {
		t.Fatal("BIND_EXISTING approve must not increment graph version")
	}
	var linkCount int64
	if err := tx.Model(&rolemodel.FamilyMemberUserLink{}).Where("user_id = ? AND family_id = ? AND link_status = 'ACTIVE'", applicantID, familyID).Count(&linkCount).Error; err != nil {
		t.Fatalf("count links: %v", err)
	}
	if linkCount != 1 {
		t.Fatalf("expected one link, got %d", linkCount)
	}
}

func TestJoinRequestApproveJoinedQuotaExceededKeepsPending(t *testing.T) {
	ctx := context.Background()
	tx, runID := freshquota.TestDB(t)
	quotaSvc := freshquota.NewQuotaService(tx)
	joinSvc := newFreshJoinRequestService(tx)

	wechatClient := freshquota.NewFakeWechatMiniClient("fresh-quota-test-app")
	authSvc := freshauth.NewAuthService(t, tx, wechatClient)
	code := wechatClient.BindUniqueCode(runID, "full-applicant")
	login, err := authSvc.WechatMiniLogin(ctx, authservice.WechatMiniLoginInput{
		Code: code, ClientType: "WECHAT_MINI_PROGRAM", IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("applicant login: %v", err)
	}
	applicantID := login.User.ID
	nickname := fmt.Sprintf("满申请%s", runID[len(runID)-4:])
	_, err = authSvc.UpdateProfile(ctx, authservice.UpdateProfileInput{UserID: applicantID, Nickname: &nickname, IP: "127.0.0.1", UserAgent: "fresh-quota-test"})
	if err != nil {
		t.Fatalf("applicant profile: %v", err)
	}
	freshquota.CreateFamily(t, freshquota.NewFamilyService(tx, quotaSvc), applicantID, "自")
	freshquota.SeedJoinedFamily(t, tx, freshquota.LoadUser(t, tx, applicantID), "已加入")

	host := freshquota.CreateActiveUser(t, tx, "超额审批", false)
	familyID, memberID := freshquota.SeedHostFamilyWithGuest(t, tx, host, "超额申")
	graphBefore := freshquota.GraphVersion(t, tx, familyID)

	created, err := joinSvc.Create(ctx, applicantID, familyID, dto.CreateJoinRequest{ApplicantGender: stringPtr("MALE")}, AuditInput{})
	if err != nil {
		t.Fatalf("create join request: %v", err)
	}
	requestID := created.RequestID
	logCountBefore := countJoinApproveLogs(t, tx, host.ID)

	if _, err := joinSvc.Approve(ctx, host.ID, familyID, requestID, dto.ApproveJoinRequest{
		ApproveMode: "BIND_EXISTING_MEMBER", MemberID: &memberID,
	}, AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaJoinedFamiliesExceeded {
		t.Fatalf("expected 41503, got %#v", err)
	}

	var request joinmodel.FamilyJoinRequest
	if err := tx.First(&request, requestID).Error; err != nil {
		t.Fatalf("reload request: %v", err)
	}
	if request.RequestStatus != "PENDING" {
		t.Fatalf("request must stay PENDING, got %s", request.RequestStatus)
	}
	var linkCount int64
	if err := tx.Model(&rolemodel.FamilyMemberUserLink{}).Where("user_id = ? AND family_id = ?", applicantID, familyID).Count(&linkCount).Error; err != nil {
		t.Fatalf("count links: %v", err)
	}
	if linkCount != 0 {
		t.Fatalf("must not create binding on quota failure, got %d", linkCount)
	}
	if freshquota.GraphVersion(t, tx, familyID) != graphBefore {
		t.Fatal("graph version must remain unchanged")
	}
	if countJoinApproveLogs(t, tx, host.ID) != logCountBefore {
		t.Fatal("must not write extra approve success logs on quota failure")
	}
}

func countJoinApproveLogs(t *testing.T, tx *gorm.DB, userID uint64) int64 {
	t.Helper()
	var count int64
	if err := tx.Table("operation_logs").Where("user_id = ? AND action = ?", userID, "APPROVE_JOIN_REQUEST").Count(&count).Error; err != nil {
		t.Fatalf("count logs: %v", err)
	}
	return count
}

func stringPtr(v string) *string { return &v }
