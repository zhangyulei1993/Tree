package service_test

import (
	"context"
	"fmt"
	"testing"

	quotaenum "tree/backend/internal/accountquota/enum"
	authservice "tree/backend/internal/auth/service"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	coredto "tree/backend/internal/family/core/dto"
	familyservice "tree/backend/internal/family/core/service"
	invitedto "tree/backend/internal/family/invitation/dto"
	invitationservice "tree/backend/internal/family/invitation/service"
	memberdto "tree/backend/internal/family/member/dto"
	memberservice "tree/backend/internal/family/member/service"
	"tree/backend/internal/testsupport/freshauth"
	"tree/backend/internal/testsupport/freshfamily"
	"tree/backend/internal/testsupport/freshquota"
)

func runWechatOnlyTierBoundaries(t *testing.T) {
	ctx := context.Background()
	tx, runID := freshquota.TestDB(t)
	freshquota.DeferRestoreQuotaConfigs(t, tx, testRootAdminID, string(enums.AdminRoleRootAdmin))
	wechatClient := freshquota.NewFakeWechatMiniClient("fresh-quota-test-app")
	authSvc := freshauth.NewAuthService(t, tx, wechatClient)
	quotaSvc := freshquota.NewQuotaService(tx)
	familySvc := freshquota.NewFamilyService(tx, quotaSvc)
	memberSvc := freshquota.NewMemberService(tx, quotaSvc)
	inviteSvc := freshfamily.NewInvitationService(tx, quotaSvc)

	adminCfg := freshquota.ListTierConfigs(t, quotaSvc, string(enums.AdminRoleRootAdmin))
	wechatLimits := adminCfg[quotaenum.TrustTierWechatOnly]
	freshquota.AssertMigrationDefaults(t, quotaSvc)

	code := wechatClient.BindUniqueCode(runID, "wechat-only")
	login, err := authSvc.WechatMiniLogin(ctx, authservice.WechatMiniLoginInput{
		Code: code, ClientType: "WECHAT_MINI_PROGRAM", IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	userID := login.User.ID
	nickname := fmt.Sprintf("边界%s", runID[len(runID)-4:])
	_, err = authSvc.UpdateProfile(ctx, authservice.UpdateProfileInput{UserID: userID, Nickname: &nickname, IP: "127.0.0.1", UserAgent: "fresh-quota-test"})
	if err != nil {
		t.Fatalf("profile: %v", err)
	}

	caps, err := quotaSvc.GetCapabilities(ctx, userID)
	if err != nil {
		t.Fatalf("GetCapabilities: %v", err)
	}
	freshquota.AssertCapabilitiesLimits(t, caps, wechatLimits)

	gender := "MALE"
	owned, err := familySvc.Create(ctx, userID, coredto.CreateFamilyRequest{Surname: "张", FounderGender: &gender}, familyservice.AuditInput{})
	if err != nil {
		t.Fatalf("first family: %v", err)
	}
	if _, err := familySvc.Create(ctx, userID, coredto.CreateFamilyRequest{Surname: "李", FounderGender: &gender}, familyservice.AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaOwnedFamiliesExceeded {
		t.Fatalf("expected 41502 at owned limit %d, got %#v", wechatLimits.MaxOwnedFamilies, err)
	}

	if freshquota.CountActiveMembers(t, tx, owned.ID) != 1 {
		t.Fatalf("founder must count toward members, got %d", freshquota.CountActiveMembers(t, tx, owned.ID))
	}
	graphBefore := freshquota.GraphVersion(t, tx, owned.ID)
	for i := 0; i < wechatLimits.MaxMembersPerOwnedFamily-1; i++ {
		if _, err := memberSvc.Create(ctx, userID, owned.ID, memberdto.CreateMemberRequest{Name: fmt.Sprintf("成员%d", i)}, memberservice.AuditInput{}); err != nil {
			t.Fatalf("member %d: %v", i, err)
		}
	}
	if freshquota.CountActiveMembers(t, tx, owned.ID) != int64(wechatLimits.MaxMembersPerOwnedFamily) {
		t.Fatalf("expected %d members, got %d", wechatLimits.MaxMembersPerOwnedFamily, freshquota.CountActiveMembers(t, tx, owned.ID))
	}
	if _, err := memberSvc.Create(ctx, userID, owned.ID, memberdto.CreateMemberRequest{Name: "超额"}, memberservice.AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaMembersExceeded {
		t.Fatalf("expected 41504 at member limit %d, got %#v", wechatLimits.MaxMembersPerOwnedFamily, err)
	}
	if freshquota.CountActiveMembers(t, tx, owned.ID) != int64(wechatLimits.MaxMembersPerOwnedFamily) || freshquota.GraphVersion(t, tx, owned.ID) != graphBefore+wechatLimits.MaxMembersPerOwnedFamily-1 {
		t.Fatal("member overflow must not add data or change graph unexpectedly")
	}

	caps, err = quotaSvc.GetCapabilities(ctx, userID)
	if err != nil {
		t.Fatalf("GetCapabilities after members: %v", err)
	}
	if caps.Usage.JoinedFamilies != 0 {
		t.Fatalf("owned family must not count as joined: %#v", caps)
	}

	for i := 0; i < wechatLimits.MaxJoinedFamilies; i++ {
		host := freshquota.CreateActiveUser(t, tx, fmt.Sprintf("主持%d", i), false)
		familyID, memberID := freshquota.SeedHostFamilyWithGuest(t, tx, host, fmt.Sprintf("外%d", i))
		created, err := inviteSvc.Create(ctx, host.ID, familyID, memberID, invitedto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, invitationservice.AuditInput{})
		if err != nil {
			t.Fatalf("create invitation %d: %v", i, err)
		}
		if _, err := inviteSvc.Accept(ctx, userID, created.Invitation.InvitationID, invitationservice.AuditInput{}); err != nil {
			t.Fatalf("accept invitation %d: %v", i, err)
		}
	}
	caps, err = quotaSvc.GetCapabilities(ctx, userID)
	if err != nil {
		t.Fatalf("GetCapabilities after joins: %v", err)
	}
	if caps.Usage.JoinedFamilies != wechatLimits.MaxJoinedFamilies {
		t.Fatalf("expected joined=%d, got %#v", wechatLimits.MaxJoinedFamilies, caps)
	}

	host2 := freshquota.CreateActiveUser(t, tx, "主持超额", false)
	family2, member2 := freshquota.SeedHostFamilyWithGuest(t, tx, host2, "超额")
	created2, err := inviteSvc.Create(ctx, host2.ID, family2, member2, invitedto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, invitationservice.AuditInput{})
	if err != nil {
		t.Fatalf("second invitation: %v", err)
	}
	if _, err := inviteSvc.Accept(ctx, userID, created2.Invitation.InvitationID, invitationservice.AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaJoinedFamiliesExceeded {
		t.Fatalf("expected 41503 at joined limit %d, got %#v", wechatLimits.MaxJoinedFamilies, err)
	}
}

func TestPhoneVerifiedTierBoundariesAfterBindPhone(t *testing.T) {
	ctx := context.Background()
	tx, runID := freshquota.TestDB(t)
	freshquota.DeferRestoreQuotaConfigs(t, tx, testRootAdminID, string(enums.AdminRoleRootAdmin))
	wechatClient := freshquota.NewFakeWechatMiniClient("fresh-quota-test-app")
	authSvc := freshauth.NewAuthService(t, tx, wechatClient)
	quotaSvc := freshquota.NewQuotaService(tx)
	familySvc := freshquota.NewFamilyService(tx, quotaSvc)
	memberSvc := freshquota.NewMemberService(tx, quotaSvc)

	adminCfg := freshquota.ListTierConfigs(t, quotaSvc, string(enums.AdminRoleRootAdmin))
	wechatLimits := adminCfg[quotaenum.TrustTierWechatOnly]
	phoneLimits := adminCfg[quotaenum.TrustTierPhoneBound]

	code := wechatClient.BindUniqueCode(runID, "phone-verified")
	login, err := authSvc.WechatMiniLogin(ctx, authservice.WechatMiniLoginInput{
		Code: code, ClientType: "WECHAT_MINI_PROGRAM", IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	userID := login.User.ID
	nickname := fmt.Sprintf("升级%s", runID[len(runID)-4:])
	_, err = authSvc.UpdateProfile(ctx, authservice.UpdateProfileInput{UserID: userID, Nickname: &nickname, IP: "127.0.0.1", UserAgent: "fresh-quota-test"})
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	ownedID := freshquota.CreateFamily(t, familySvc, userID, "陈")
	for i := 0; i < wechatLimits.MaxMembersPerOwnedFamily-1; i++ {
		freshquota.AddMember(t, memberSvc, userID, ownedID, fmt.Sprintf("旧成员%d", i))
	}

	phone := freshquota.UniquePhone(runID, "upgrade")
	password := "bind-pass-" + runID[len(runID)-4:]
	if _, err := authSvc.BindPhoneCredential(ctx, authservice.BindPhoneCredentialInput{
		UserID: userID, Phone: phone, Password: password, ConfirmPassword: password,
		IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	}); err != nil {
		t.Fatalf("BindPhoneCredential: %v", err)
	}

	caps, err := quotaSvc.GetCapabilities(ctx, userID)
	if err != nil {
		t.Fatalf("GetCapabilities: %v", err)
	}
	freshquota.AssertCapabilitiesLimits(t, caps, phoneLimits)

	gender := "MALE"
	if _, err := familySvc.Create(ctx, userID, coredto.CreateFamilyRequest{Surname: "赵", FounderGender: &gender}, familyservice.AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaOwnedFamiliesExceeded {
		t.Fatalf("second owned family blocked at limit %d: %#v", phoneLimits.MaxOwnedFamilies, err)
	}

	graphBefore := freshquota.GraphVersion(t, tx, ownedID)
	extraMembers := phoneLimits.MaxMembersPerOwnedFamily - wechatLimits.MaxMembersPerOwnedFamily
	for i := 0; i < extraMembers; i++ {
		if _, err := memberSvc.Create(ctx, userID, ownedID, memberdto.CreateMemberRequest{Name: fmt.Sprintf("新成员%d", i)}, memberservice.AuditInput{}); err != nil {
			t.Fatalf("expanded member %d: %v", i, err)
		}
	}
	if freshquota.CountActiveMembers(t, tx, ownedID) != int64(phoneLimits.MaxMembersPerOwnedFamily) {
		t.Fatalf("expected %d members after upgrade, got %d", phoneLimits.MaxMembersPerOwnedFamily, freshquota.CountActiveMembers(t, tx, ownedID))
	}
	if _, err := memberSvc.Create(ctx, userID, ownedID, memberdto.CreateMemberRequest{Name: "超额成员"}, memberservice.AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaMembersExceeded {
		t.Fatalf("expected 41504 at member limit %d, got %#v", phoneLimits.MaxMembersPerOwnedFamily, err)
	}
	if freshquota.CountActiveMembers(t, tx, ownedID) != int64(phoneLimits.MaxMembersPerOwnedFamily) || freshquota.GraphVersion(t, tx, ownedID) != graphBefore+extraMembers {
		t.Fatal("member overflow must not persist or skip graph rules")
	}

	inviteSvc := freshfamily.NewInvitationService(tx, quotaSvc)
	for i := 0; i < phoneLimits.MaxJoinedFamilies; i++ {
		host := freshquota.CreateActiveUser(t, tx, fmt.Sprintf("主持%d", i), false)
		familyID, memberID := freshquota.SeedHostFamilyWithGuest(t, tx, host, fmt.Sprintf("外%d", i))
		created, err := inviteSvc.Create(ctx, host.ID, familyID, memberID, invitedto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, invitationservice.AuditInput{})
		if err != nil {
			t.Fatalf("invitation %d: %v", i, err)
		}
		if _, err := inviteSvc.Accept(ctx, userID, created.Invitation.InvitationID, invitationservice.AuditInput{}); err != nil {
			t.Fatalf("join %d: %v", i, err)
		}
	}
	caps, err = quotaSvc.GetCapabilities(ctx, userID)
	if err != nil {
		t.Fatalf("GetCapabilities after joins: %v", err)
	}
	if caps.Usage.JoinedFamilies != phoneLimits.MaxJoinedFamilies {
		t.Fatalf("expected joined=%d, got %#v", phoneLimits.MaxJoinedFamilies, caps)
	}
	hostExtra := freshquota.CreateActiveUser(t, tx, "主持超额", false)
	familyExtra, memberExtra := freshquota.SeedHostFamilyWithGuest(t, tx, hostExtra, "超额")
	createdExtra, err := inviteSvc.Create(ctx, hostExtra.ID, familyExtra, memberExtra, invitedto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, invitationservice.AuditInput{})
	if err != nil {
		t.Fatalf("extra invitation: %v", err)
	}
	if _, err := inviteSvc.Accept(ctx, userID, createdExtra.Invitation.InvitationID, invitationservice.AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaJoinedFamiliesExceeded {
		t.Fatalf("expected 41503 at joined limit %d, got %#v", phoneLimits.MaxJoinedFamilies, err)
	}
}

func TestWechatOnlyTierBoundariesFreshAccount(t *testing.T) {
	runWechatOnlyTierBoundaries(t)
}

func TestWechatOnlyTierBoundariesRepeated(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("iteration-%d", i+1), func(t *testing.T) {
			runWechatOnlyTierBoundaries(t)
		})
	}
}
