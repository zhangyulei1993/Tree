package service

import (
	"context"
	"fmt"
	"testing"

	"gorm.io/gorm"

	authservice "tree/backend/internal/auth/service"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/family/invitation/dto"
	invitationmodel "tree/backend/internal/family/invitation/model"
	inviterepo "tree/backend/internal/family/invitation/repository"
	rolemodel "tree/backend/internal/family/role/model"
	"tree/backend/internal/testsupport/freshauth"
	"tree/backend/internal/testsupport/freshquota"
)

func newFreshInvitationService(tx *gorm.DB) Service {
	quotaSvc := freshquota.NewQuotaService(tx)
	repo := inviterepo.NewRepository(tx)
	return NewService(repo, inviterepo.NewUnitOfWork(tx, repo), freshquota.AllowAllFamilyPerm{}, quotaSvc)
}

func TestInvitationAcceptRespectsJoinedQuotaFreshAccount(t *testing.T) {
	ctx := context.Background()
	tx, runID := freshquota.TestDB(t)
	quotaSvc := freshquota.NewQuotaService(tx)
	inviteSvc := newFreshInvitationService(tx)

	wechatClient := freshquota.NewFakeWechatMiniClient("fresh-quota-test-app")
	authSvc := freshauth.NewAuthService(t, tx, wechatClient)
	code := wechatClient.BindUniqueCode(runID, "guest")
	login, err := authSvc.WechatMiniLogin(ctx, authservice.WechatMiniLoginInput{
		Code: code, ClientType: "WECHAT_MINI_PROGRAM", IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("guest login: %v", err)
	}
	guestID := login.User.ID
	nickname := fmt.Sprintf("访客%s", runID[len(runID)-4:])
	_, err = authSvc.UpdateProfile(ctx, authservice.UpdateProfileInput{UserID: guestID, Nickname: &nickname, IP: "127.0.0.1", UserAgent: "fresh-quota-test"})
	if err != nil {
		t.Fatalf("guest profile: %v", err)
	}
	freshquota.CreateFamily(t, freshquota.NewFamilyService(tx, quotaSvc), guestID, "自")

	host := freshquota.CreateActiveUser(t, tx, "邀请主持", false)
	familyID, memberID := freshquota.SeedHostFamilyWithGuest(t, tx, host, "邀")
	graphBefore := freshquota.GraphVersion(t, tx, familyID)
	created, err := inviteSvc.Create(ctx, host.ID, familyID, memberID, dto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, AuditInput{})
	if err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	result, err := inviteSvc.Accept(ctx, guestID, created.Invitation.InvitationID, AuditInput{})
	if err != nil {
		t.Fatalf("accept invitation: %v", err)
	}
	if result.Status != "ACCEPTED" {
		t.Fatalf("expected ACCEPTED, got %s", result.Status)
	}
	if freshquota.GraphVersion(t, tx, familyID) != graphBefore {
		t.Fatal("accept must not increment graph version")
	}
	var linkCount int64
	if err := tx.Model(&rolemodel.FamilyMemberUserLink{}).Where("user_id = ? AND family_id = ? AND link_status = 'ACTIVE'", guestID, familyID).Count(&linkCount).Error; err != nil {
		t.Fatalf("count links: %v", err)
	}
	if linkCount != 1 {
		t.Fatalf("expected one active link, got %d", linkCount)
	}
}

func TestInvitationAcceptJoinedQuotaExceededKeepsPending(t *testing.T) {
	ctx := context.Background()
	tx, runID := freshquota.TestDB(t)
	quotaSvc := freshquota.NewQuotaService(tx)
	inviteSvc := newFreshInvitationService(tx)

	wechatClient := freshquota.NewFakeWechatMiniClient("fresh-quota-test-app")
	authSvc := freshauth.NewAuthService(t, tx, wechatClient)
	code := wechatClient.BindUniqueCode(runID, "full-guest")
	login, err := authSvc.WechatMiniLogin(ctx, authservice.WechatMiniLoginInput{
		Code: code, ClientType: "WECHAT_MINI_PROGRAM", IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("guest login: %v", err)
	}
	guestID := login.User.ID
	nickname := fmt.Sprintf("满额%s", runID[len(runID)-4:])
	_, err = authSvc.UpdateProfile(ctx, authservice.UpdateProfileInput{UserID: guestID, Nickname: &nickname, IP: "127.0.0.1", UserAgent: "fresh-quota-test"})
	if err != nil {
		t.Fatalf("guest profile: %v", err)
	}
	freshquota.CreateFamily(t, freshquota.NewFamilyService(tx, quotaSvc), guestID, "自")
	freshquota.SeedJoinedFamily(t, tx, freshquota.LoadUser(t, tx, guestID), "已加入")

	host := freshquota.CreateActiveUser(t, tx, "超额主持", false)
	familyID, memberID := freshquota.SeedHostFamilyWithGuest(t, tx, host, "超额")
	graphBefore := freshquota.GraphVersion(t, tx, familyID)
	created, err := inviteSvc.Create(ctx, host.ID, familyID, memberID, dto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, AuditInput{})
	if err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	invitationID := created.Invitation.InvitationID
	logCountBefore := countInvitationAcceptLogs(t, tx, guestID)

	if _, err := inviteSvc.Accept(ctx, guestID, invitationID, AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaJoinedFamiliesExceeded {
		t.Fatalf("expected 41503, got %#v", err)
	}

	var invitation invitationmodel.FamilyInvitation
	if err := tx.First(&invitation, invitationID).Error; err != nil {
		t.Fatalf("reload invitation: %v", err)
	}
	if invitation.Status != "PENDING" {
		t.Fatalf("invitation must stay PENDING, got %s", invitation.Status)
	}
	var linkCount int64
	if err := tx.Model(&rolemodel.FamilyMemberUserLink{}).Where("user_id = ? AND family_id = ?", guestID, familyID).Count(&linkCount).Error; err != nil {
		t.Fatalf("count links: %v", err)
	}
	if linkCount != 0 {
		t.Fatalf("must not create binding link on quota failure, got %d", linkCount)
	}
	if freshquota.GraphVersion(t, tx, familyID) != graphBefore {
		t.Fatal("graph version must remain unchanged on quota failure")
	}
	if countInvitationAcceptLogs(t, tx, guestID) != logCountBefore {
		t.Fatalf("must not write extra accept success logs on quota failure")
	}
}

func countInvitationAcceptLogs(t *testing.T, tx *gorm.DB, userID uint64) int64 {
	t.Helper()
	var count int64
	if err := tx.Table("operation_logs").Where("user_id = ? AND action = ?", userID, "ACCEPT_INVITATION").Count(&count).Error; err != nil {
		t.Fatalf("count logs: %v", err)
	}
	return count
}
