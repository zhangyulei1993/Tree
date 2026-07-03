package service_test

import (
	"context"
	"fmt"
	"testing"

	quotadto "tree/backend/internal/accountquota/dto"
	quotaenum "tree/backend/internal/accountquota/enum"
	quotarepo "tree/backend/internal/accountquota/repository"
	quotaservice "tree/backend/internal/accountquota/service"
	authservice "tree/backend/internal/auth/service"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	coredto "tree/backend/internal/family/core/dto"
	familymodel "tree/backend/internal/family/core/model"
	familyservice "tree/backend/internal/family/core/service"
	invitedto "tree/backend/internal/family/invitation/dto"
	invitationservice "tree/backend/internal/family/invitation/service"
	memberdto "tree/backend/internal/family/member/dto"
	membermodel "tree/backend/internal/family/member/model"
	memberservice "tree/backend/internal/family/member/service"
	rolemodel "tree/backend/internal/family/role/model"
	"tree/backend/internal/testsupport/freshauth"
	"tree/backend/internal/testsupport/freshfamily"
	"tree/backend/internal/testsupport/freshquota"
)

const testRootAdminID uint64 = 1

func TestMigrationDefaultConfigsViaAdminList(t *testing.T) {
	ctx := context.Background()
	tx, _ := freshquota.TestDB(t)
	freshquota.DeferRestoreQuotaConfigs(t, tx, testRootAdminID, string(enums.AdminRoleRootAdmin))
	svc := freshquota.NewQuotaService(tx)
	freshquota.AssertMigrationDefaults(t, svc)

	wechatUser := freshquota.CreateActiveUser(t, tx, "默认微信", false)
	caps, err := svc.GetCapabilities(ctx, wechatUser.ID)
	if err != nil {
		t.Fatalf("GetCapabilities: %v", err)
	}
	freshquota.AssertCapabilitiesLimits(t, caps, freshquota.MigrationDefaults()[quotaenum.TrustTierWechatOnly])

	phoneUser := freshquota.CreateActiveUser(t, tx, "默认手机", true)
	phoneCaps, err := svc.GetCapabilities(ctx, phoneUser.ID)
	if err != nil {
		t.Fatalf("phone GetCapabilities: %v", err)
	}
	freshquota.AssertCapabilitiesLimits(t, phoneCaps, freshquota.MigrationDefaults()[quotaenum.TrustTierPhoneVerified])
}

func TestAdminDynamicConfigDrivesCapabilitiesAndBoundaries(t *testing.T) {
	ctx := context.Background()
	tx, runID := freshquota.TestDB(t)
	freshquota.DeferRestoreQuotaConfigs(t, tx, testRootAdminID, string(enums.AdminRoleRootAdmin))
	svc := freshquota.NewQuotaService(tx)
	familySvc := freshquota.NewFamilyService(tx, svc)
	memberSvc := freshquota.NewMemberService(tx, svc)
	inviteSvc := freshfamily.NewInvitationService(tx, svc)

	freshquota.ApplyDynamicTestConfigs(t, svc, testRootAdminID)
	dynamic := freshquota.ListTierConfigs(t, svc, string(enums.AdminRoleRootAdmin))
	wechatCfg := dynamic[quotaenum.TrustTierWechatOnly]
	phoneCfg := dynamic[quotaenum.TrustTierPhoneVerified]

	wechatClient := freshquota.NewFakeWechatMiniClient("fresh-quota-test-app")
	authSvc := freshauth.NewAuthService(t, tx, wechatClient)
	code := wechatClient.BindUniqueCode(runID, "dynamic-wechat")
	login, err := authSvc.WechatMiniLogin(ctx, authservice.WechatMiniLoginInput{
		Code: code, ClientType: "WECHAT_MINI_PROGRAM", IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	userID := login.User.ID
	nickname := fmt.Sprintf("动态%s", runID[len(runID)-4:])
	_, err = authSvc.UpdateProfile(ctx, authservice.UpdateProfileInput{UserID: userID, Nickname: &nickname, IP: "127.0.0.1", UserAgent: "fresh-quota-test"})
	if err != nil {
		t.Fatalf("profile: %v", err)
	}

	caps, err := svc.GetCapabilities(ctx, userID)
	if err != nil {
		t.Fatalf("wechat capabilities: %v", err)
	}
	freshquota.AssertCapabilitiesLimits(t, caps, wechatCfg)

	gender := "MALE"
	var ownedID uint64
	for i := 0; i < wechatCfg.MaxOwnedFamilies; i++ {
		created, err := familySvc.Create(ctx, userID, coredto.CreateFamilyRequest{Surname: fmt.Sprintf("张%d", i), FounderGender: &gender}, familyservice.AuditInput{})
		if err != nil {
			t.Fatalf("create family %d at limit: %v", i, err)
		}
		if i == 0 {
			ownedID = created.ID
		}
	}
	if _, err := familySvc.Create(ctx, userID, coredto.CreateFamilyRequest{Surname: "超额", FounderGender: &gender}, familyservice.AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaOwnedFamiliesExceeded {
		t.Fatalf("expected 41502 at owned=%d, got %#v", wechatCfg.MaxOwnedFamilies, err)
	}

	for i := 0; i < wechatCfg.MaxMembersPerOwnedFamily-1; i++ {
		if _, err := memberSvc.Create(ctx, userID, ownedID, memberdto.CreateMemberRequest{Name: fmt.Sprintf("成员%d", i)}, memberservice.AuditInput{}); err != nil {
			t.Fatalf("member %d: %v", i, err)
		}
	}
	if freshquota.CountActiveMembers(t, tx, ownedID) != int64(wechatCfg.MaxMembersPerOwnedFamily) {
		t.Fatalf("expected %d members at boundary", wechatCfg.MaxMembersPerOwnedFamily)
	}
	if _, err := memberSvc.Create(ctx, userID, ownedID, memberdto.CreateMemberRequest{Name: "超额成员"}, memberservice.AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaMembersExceeded {
		t.Fatalf("expected 41504 at members=%d, got %#v", wechatCfg.MaxMembersPerOwnedFamily, err)
	}

	for i := 0; i < wechatCfg.MaxJoinedFamilies; i++ {
		host := freshquota.CreateActiveUser(t, tx, fmt.Sprintf("主持%d", i), false)
		familyID, memberID := freshquota.SeedHostFamilyWithGuest(t, tx, host, fmt.Sprintf("外%d", i))
		created, err := inviteSvc.Create(ctx, host.ID, familyID, memberID, invitedto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, invitationservice.AuditInput{})
		if err != nil {
			t.Fatalf("invitation %d: %v", i, err)
		}
		if _, err := inviteSvc.Accept(ctx, userID, created.Invitation.InvitationID, invitationservice.AuditInput{}); err != nil {
			t.Fatalf("accept %d: %v", i, err)
		}
	}
	hostExtra := freshquota.CreateActiveUser(t, tx, "主持超额", false)
	familyExtra, memberExtra := freshquota.SeedHostFamilyWithGuest(t, tx, hostExtra, "超额")
	createdExtra, err := inviteSvc.Create(ctx, hostExtra.ID, familyExtra, memberExtra, invitedto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, invitationservice.AuditInput{})
	if err != nil {
		t.Fatalf("extra invitation: %v", err)
	}
	if _, err := inviteSvc.Accept(ctx, userID, createdExtra.Invitation.InvitationID, invitationservice.AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaJoinedFamiliesExceeded {
		t.Fatalf("expected 41503 at joined=%d, got %#v", wechatCfg.MaxJoinedFamilies, err)
	}

	phone := freshquota.UniquePhone(runID, "dynamic-bind")
	freshquota.SeedBindPhoneCode(t, tx, phone, "864200")
	if _, err := authSvc.BindPhone(ctx, authservice.BindPhoneInput{UserID: userID, Phone: phone, Code: "864200", IP: "127.0.0.1", UserAgent: "fresh-quota-test"}); err != nil {
		t.Fatalf("BindPhone: %v", err)
	}
	phoneCaps, err := svc.GetCapabilities(ctx, userID)
	if err != nil {
		t.Fatalf("phone capabilities: %v", err)
	}
	freshquota.AssertCapabilitiesLimits(t, phoneCaps, phoneCfg)
	if phoneCaps.TrustTier != quotaenum.TrustTierPhoneVerified {
		t.Fatalf("expected PHONE_VERIFIED tier, got %s", phoneCaps.TrustTier)
	}

	if _, err := familySvc.Create(ctx, userID, coredto.CreateFamilyRequest{Surname: "手机第三家", FounderGender: &gender}, familyservice.AuditInput{}); err != nil {
		t.Fatalf("third owned family at phone limit: %v", err)
	}
	if _, err := familySvc.Create(ctx, userID, coredto.CreateFamilyRequest{Surname: "手机超额家", FounderGender: &gender}, familyservice.AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaOwnedFamiliesExceeded {
		t.Fatalf("expected 41502 at phone owned=%d, got %#v", phoneCfg.MaxOwnedFamilies, err)
	}

	extraMembers := phoneCfg.MaxMembersPerOwnedFamily - wechatCfg.MaxMembersPerOwnedFamily
	for i := 0; i < extraMembers; i++ {
		if _, err := memberSvc.Create(ctx, userID, ownedID, memberdto.CreateMemberRequest{Name: fmt.Sprintf("手机扩员%d", i)}, memberservice.AuditInput{}); err != nil {
			t.Fatalf("phone expanded member %d: %v", i, err)
		}
	}
	if freshquota.CountActiveMembers(t, tx, ownedID) != int64(phoneCfg.MaxMembersPerOwnedFamily) {
		t.Fatalf("expected %d members at phone limit, got %d", phoneCfg.MaxMembersPerOwnedFamily, freshquota.CountActiveMembers(t, tx, ownedID))
	}
	if _, err := memberSvc.Create(ctx, userID, ownedID, memberdto.CreateMemberRequest{Name: "手机超额成员"}, memberservice.AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaMembersExceeded {
		t.Fatalf("expected 41504 at phone members=%d, got %#v", phoneCfg.MaxMembersPerOwnedFamily, err)
	}

	for i := 0; i < phoneCfg.MaxJoinedFamilies-wechatCfg.MaxJoinedFamilies; i++ {
		host := freshquota.CreateActiveUser(t, tx, fmt.Sprintf("手机主持%d", i), false)
		familyID, memberID := freshquota.SeedHostFamilyWithGuest(t, tx, host, fmt.Sprintf("手机外%d", i))
		created, err := inviteSvc.Create(ctx, host.ID, familyID, memberID, invitedto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, invitationservice.AuditInput{})
		if err != nil {
			t.Fatalf("phone invitation %d: %v", i, err)
		}
		if _, err := inviteSvc.Accept(ctx, userID, created.Invitation.InvitationID, invitationservice.AuditInput{}); err != nil {
			t.Fatalf("phone accept %d: %v", i, err)
		}
	}
	phoneCaps, err = svc.GetCapabilities(ctx, userID)
	if err != nil {
		t.Fatalf("phone capabilities after joins: %v", err)
	}
	if phoneCaps.Usage.JoinedFamilies != phoneCfg.MaxJoinedFamilies {
		t.Fatalf("expected phone joined=%d, got %#v", phoneCfg.MaxJoinedFamilies, phoneCaps.Usage)
	}
	hostPhoneExtra := freshquota.CreateActiveUser(t, tx, "手机主持超额", false)
	familyPhoneExtra, memberPhoneExtra := freshquota.SeedHostFamilyWithGuest(t, tx, hostPhoneExtra, "手机超额")
	createdPhoneExtra, err := inviteSvc.Create(ctx, hostPhoneExtra.ID, familyPhoneExtra, memberPhoneExtra, invitedto.CreateInvitationRequest{InviteChannel: "SHARE_LINK"}, invitationservice.AuditInput{})
	if err != nil {
		t.Fatalf("phone extra invitation: %v", err)
	}
	if _, err := inviteSvc.Accept(ctx, userID, createdPhoneExtra.Invitation.InvitationID, invitationservice.AuditInput{}); err == nil || err.Code != apperrors.CodeQuotaJoinedFamiliesExceeded {
		t.Fatalf("expected 41503 at phone joined=%d, got %#v", phoneCfg.MaxJoinedFamilies, err)
	}
}

func TestAdminUpdateConfigRoundTripViaListConfigs(t *testing.T) {
	tx, _ := freshquota.TestDB(t)
	freshquota.DeferRestoreQuotaConfigs(t, tx, testRootAdminID, string(enums.AdminRoleRootAdmin))
	svc := freshquota.NewQuotaService(tx)

	freshquota.ApplyDynamicTestConfigs(t, svc, testRootAdminID)
	target := freshquota.DynamicTestConfigs[quotaenum.TrustTierWechatOnly]
	got := freshquota.ListTierConfigs(t, svc, string(enums.AdminRoleSuperAdmin))
	if got[quotaenum.TrustTierWechatOnly] != target {
		t.Fatalf("ListConfigs after staged PUT mismatch: %#v", got[quotaenum.TrustTierWechatOnly])
	}
}

func TestAdminTierOrderReturns41507(t *testing.T) {
	tx, _ := freshquota.TestDB(t)
	freshquota.DeferRestoreQuotaConfigs(t, tx, testRootAdminID, string(enums.AdminRoleRootAdmin))
	svc := freshquota.NewQuotaService(tx)
	freshquota.ApplyDynamicTestConfigs(t, svc, testRootAdminID)

	_, err := svc.UpdateConfig(context.Background(), testRootAdminID, string(enums.AdminRoleRootAdmin), quotaenum.TrustTierPhoneVerified, quotadto.ConfigValues{
		MaxOwnedFamilies: 1, MaxMembersPerOwnedFamily: 5, MaxJoinedFamilies: 4,
	}, quotaservice.AuditInput{})
	if err == nil || err.Code != apperrors.CodeQuotaConfigTierOrder {
		t.Fatalf("expected 41507 when phone tier below wechat, got %#v", err)
	}
}

func TestAdminPreviewImpactExactAffectedCounts(t *testing.T) {
	ctx := context.Background()
	tx, _ := freshquota.TestDB(t)
	freshquota.DeferRestoreQuotaConfigs(t, tx, testRootAdminID, string(enums.AdminRoleRootAdmin))
	svc := freshquota.NewQuotaService(tx)
	repo := quotarepo.NewRepository(tx)
	baselineUsers, countErr := repo.CountUsersExceedingDistinct(ctx, quotaenum.TrustTierWechatOnly, 1, 0)
	if countErr != nil {
		t.Fatalf("baseline users: %v", countErr)
	}
	baselineFamilies, countErr := repo.CountFamiliesExceedingMembers(ctx, quotaenum.TrustTierWechatOnly, 6)
	if countErr != nil {
		t.Fatalf("baseline families: %v", countErr)
	}

	ownedUser := freshquota.CreateActiveUser(t, tx, "预览超额创始人", false)
	for i := 0; i < 2; i++ {
		family := &familymodel.Family{
			FamilyName: fmt.Sprintf("预览家%d", i), FamilySurname: fmt.Sprintf("预%d", i), Status: string(enums.StatusNormal),
			PublicDisplayStatus: string(enums.StatusPrivate), Searchable: true, GraphVersion: 1,
		}
		if err := tx.Create(family).Error; err != nil {
			t.Fatalf("seed family %d: %v", i, err)
		}
		member := &membermodel.FamilyMember{FamilyID: family.ID, DisplayName: "预览创始人", Status: string(enums.StatusActive)}
		if err := tx.Create(member).Error; err != nil {
			t.Fatalf("seed founder member %d: %v", i, err)
		}
		link := rolemodel.FamilyMemberUserLink{
			FamilyID: family.ID, MemberID: member.ID, UserID: ownedUser.ID,
			LinkStatus: string(enums.StatusActive), LinkSource: "FRESH_QUOTA_TEST", FamilyRole: string(enums.FamilyRoleFounder),
		}
		if err := tx.Create(&link).Error; err != nil {
			t.Fatalf("seed founder link %d: %v", i, err)
		}
	}

	memberUser := freshquota.CreateActiveUser(t, tx, "预览超额成员", false)
	memberFamily := &familymodel.Family{
		FamilyName: "预览成员家", FamilySurname: "预览", Status: string(enums.StatusNormal),
		PublicDisplayStatus: string(enums.StatusPrivate), Searchable: true, GraphVersion: 1,
	}
	if err := tx.Create(memberFamily).Error; err != nil {
		t.Fatalf("seed member family: %v", err)
	}
	founder := &membermodel.FamilyMember{FamilyID: memberFamily.ID, DisplayName: "预览成员创始人", Status: string(enums.StatusActive)}
	if err := tx.Create(founder).Error; err != nil {
		t.Fatalf("seed founder: %v", err)
	}
	founderLink := rolemodel.FamilyMemberUserLink{
		FamilyID: memberFamily.ID, MemberID: founder.ID, UserID: memberUser.ID,
		LinkStatus: string(enums.StatusActive), LinkSource: "FRESH_QUOTA_TEST", FamilyRole: string(enums.FamilyRoleFounder),
	}
	if err := tx.Create(&founderLink).Error; err != nil {
		t.Fatalf("seed founder link: %v", err)
	}
	for i := 0; i < 6; i++ {
		row := &membermodel.FamilyMember{FamilyID: memberFamily.ID, DisplayName: fmt.Sprintf("预览成员%d", i), Status: string(enums.StatusActive)}
		if err := tx.Create(row).Error; err != nil {
			t.Fatalf("seed member %d: %v", i, err)
		}
	}
	if freshquota.CountActiveMembers(t, tx, memberFamily.ID) != 7 {
		t.Fatalf("expected 7 active members before preview, got %d", freshquota.CountActiveMembers(t, tx, memberFamily.ID))
	}

	preview, previewErr := svc.PreviewImpact(ctx, string(enums.AdminRoleRootAdmin), quotaenum.TrustTierWechatOnly, quotadto.ConfigValues{
		MaxOwnedFamilies: 1, MaxMembersPerOwnedFamily: 6, MaxJoinedFamilies: 0,
	})
	if previewErr != nil {
		t.Fatalf("PreviewImpact: %v", previewErr)
	}
	if preview.AffectedUsers != baselineUsers+1 {
		t.Fatalf("expected %d affected users, got %d (%#v)", baselineUsers+1, preview.AffectedUsers, preview)
	}
	if preview.AffectedFamilies != baselineFamilies+1 {
		t.Fatalf("expected %d affected families, got %d (%#v)", baselineFamilies+1, preview.AffectedFamilies, preview)
	}
}

func TestAdminPlatformAdminCannotUpdateConfig(t *testing.T) {
	tx, _ := freshquota.TestDB(t)
	freshquota.DeferRestoreQuotaConfigs(t, tx, testRootAdminID, string(enums.AdminRoleRootAdmin))
	svc := freshquota.NewQuotaService(tx)

	_, err := svc.ListConfigs(context.Background(), string(enums.AdminRolePlatformAdmin))
	if err == nil || err.Code != apperrors.CodeQuotaConfigForbidden {
		t.Fatalf("PLATFORM_ADMIN should not list configs, got %#v", err)
	}
	_, err = svc.UpdateConfig(context.Background(), 2, string(enums.AdminRolePlatformAdmin), quotaenum.TrustTierWechatOnly, quotadto.ConfigValues{
		MaxOwnedFamilies: 2, MaxMembersPerOwnedFamily: 6, MaxJoinedFamilies: 2,
	}, quotaservice.AuditInput{})
	if err == nil || err.Code != apperrors.CodeQuotaConfigForbidden {
		t.Fatalf("PLATFORM_ADMIN should not update, got %#v", err)
	}
}

func TestAdminUpdateConfigWritesOperationLog(t *testing.T) {
	tx, _ := freshquota.TestDB(t)
	freshquota.DeferRestoreQuotaConfigs(t, tx, testRootAdminID, string(enums.AdminRoleRootAdmin))
	svc := freshquota.NewQuotaService(tx)

	freshquota.ApplyDynamicTestConfigs(t, svc, testRootAdminID)

	var count int64
	if err := tx.Table("operation_logs").Where("module = ? AND action = ?", "ACCOUNT_QUOTA", "UPDATE_ACCOUNT_QUOTA_CONFIG").Count(&count).Error; err != nil {
		t.Fatalf("count logs: %v", err)
	}
	if count < 1 {
		t.Fatal("expected UPDATE_ACCOUNT_QUOTA_CONFIG operation log")
	}
}
