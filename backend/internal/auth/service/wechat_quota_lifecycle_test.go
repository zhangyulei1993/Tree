package service

import (
	"context"
	"fmt"
	"testing"

	"gorm.io/gorm"

	quotaenum "tree/backend/internal/accountquota/enum"
	authrepo "tree/backend/internal/auth/repository"
	"tree/backend/internal/common/config"
	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	commonjwt "tree/backend/internal/common/jwt"
	"tree/backend/internal/common/security"
	coredto "tree/backend/internal/family/core/dto"
	familyservice "tree/backend/internal/family/core/service"
	"tree/backend/internal/testsupport/freshquota"
	usermodel "tree/backend/internal/user/model"
)

func newFreshQuotaAuthService(t *testing.T, tx *gorm.DB, wechatClient *freshquota.FakeWechatMiniClient) *PhoneAuthService {
	t.Helper()
	manager, err := commonjwt.NewManager(config.JWTConfig{
		UserSecret: "fresh-quota-user-secret", AdminSecret: "fresh-quota-admin-secret",
		AccessTokenExpireMinutes: 30,
	}, "TreeFreshQuotaTest")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return NewPhoneAuthService(
		tx,
		authrepo.NewGormUserRepository(tx),
		authrepo.NewGormVerificationCodeRepository(tx),
		wechatClient,
		manager,
		nil,
		nil,
		contentsafety.AlwaysPass(),
		&config.Config{
			App:        config.AppConfig{Env: "test"},
			Wechat:     config.WechatConfig{MiniAppID: "fresh-quota-test-app"},
			VerifyCode: config.VerifyCodeConfig{ExpireSeconds: 300, CooldownSeconds: 60},
		},
	)
}

func runWechatQuotaLifecycle(t *testing.T, tx *gorm.DB, runID string) {
	t.Helper()
	ctx := context.Background()
	wechatClient := freshquota.NewFakeWechatMiniClient("fresh-quota-test-app")
	authSvc := newFreshQuotaAuthService(t, tx, wechatClient)
	quotaSvc := freshquota.NewQuotaService(tx)
	familySvc := freshquota.NewFamilyService(tx, quotaSvc)
	adminCfg := freshquota.ListTierConfigs(t, quotaSvc, string(enums.AdminRoleRootAdmin))

	codeA := wechatClient.BindUniqueCode(runID, "login-a")
	loginA, err := authSvc.WechatMiniLogin(ctx, WechatMiniLoginInput{
		Code: codeA, ClientType: ClientWechatMini, IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("first WechatMiniLogin: %v", err)
	}
	userA := loginA.User
	freshquota.AssertUserStatus(t, userA.Status, enums.StatusActive)
	if userA.PhoneVerified || userA.Phone != nil {
		t.Fatalf("expected no phone on new wechat user: %#v", userA)
	}
	if userA.Nickname != nil {
		t.Fatalf("expected nil nickname: %#v", userA.Nickname)
	}

	dbUser := freshquota.LoadUser(t, tx, userA.ID)
	if dbUser.PasswordHash != nil {
		t.Fatal("expected nil password hash")
	}
	caps, capErr := quotaSvc.GetCapabilities(ctx, userA.ID)
	if capErr != nil || caps.TrustTier != quotaenum.TrustTierWechatOnly {
		t.Fatalf("expected WECHAT_ONLY tier: %#v %v", caps, capErr)
	}
	freshquota.AssertCapabilitiesLimits(t, caps, adminCfg[quotaenum.TrustTierWechatOnly])

	var identity usermodel.UserAuthIdentity
	if err := tx.Where("user_id = ? AND provider = ?", userA.ID, providerWechatMini).First(&identity).Error; err != nil {
		t.Fatalf("load identity: %v", err)
	}
	if identity.OpenID == nil || identity.OpenIDHash == nil {
		t.Fatal("expected openid stored")
	}

	loginA2, err := authSvc.WechatMiniLogin(ctx, WechatMiniLoginInput{
		Code: codeA, ClientType: ClientWechatMini, IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("second WechatMiniLogin same openid: %v", err)
	}
	if loginA2.User.ID != userA.ID {
		t.Fatalf("same openid should return same user: %d vs %d", loginA2.User.ID, userA.ID)
	}

	codeB := wechatClient.BindUniqueCode(runID, "login-b")
	loginB, err := authSvc.WechatMiniLogin(ctx, WechatMiniLoginInput{
		Code: codeB, ClientType: ClientWechatMini, IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("different openid login: %v", err)
	}
	if loginB.User.ID == userA.ID {
		t.Fatalf("different openid should create different user")
	}

	gender := "MALE"
	_, profileErr := familySvc.Create(ctx, userA.ID, coredto.CreateFamilyRequest{Surname: "测", FounderGender: &gender}, familyservice.AuditInput{})
	freshquota.AssertBusinessCode(t, profileErr, apperrors.CodeProfileIncomplete)

	nickname := fmt.Sprintf("用户%s", runID[len(runID)-6:])
	_, err = authSvc.UpdateProfile(ctx, UpdateProfileInput{
		UserID: userA.ID, Nickname: &nickname, IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	created, err := familySvc.Create(ctx, userA.ID, coredto.CreateFamilyRequest{Surname: "张", FounderGender: &gender}, familyservice.AuditInput{})
	if err != nil {
		t.Fatalf("create family after profile: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected created family")
	}
}

func TestWechatQuotaLifecycle(t *testing.T) {
	tx, runID := freshquota.TestDB(t)
	runWechatQuotaLifecycle(t, tx, runID)
}

func TestWechatQuotaLifecycleRepeated(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("iteration-%d", i+1), func(t *testing.T) {
			tx, runID := freshquota.TestDB(t)
			runWechatQuotaLifecycle(t, tx, runID)
		})
	}
}

func TestWechatLoginBindPhoneUpgradePreservesIdentity(t *testing.T) {
	ctx := context.Background()
	tx, runID := freshquota.TestDB(t)
	wechatClient := freshquota.NewFakeWechatMiniClient("fresh-quota-test-app")
	authSvc := newFreshQuotaAuthService(t, tx, wechatClient)
	quotaSvc := freshquota.NewQuotaService(tx)
	familySvc := freshquota.NewFamilyService(tx, quotaSvc)
	memberSvc := freshquota.NewMemberService(tx, quotaSvc)

	code := wechatClient.BindUniqueCode(runID, "upgrade")
	login, err := authSvc.WechatMiniLogin(ctx, WechatMiniLoginInput{
		Code: code, ClientType: ClientWechatMini, IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("WechatMiniLogin: %v", err)
	}
	userID := login.User.ID
	nickname := fmt.Sprintf("升级%s", runID[len(runID)-4:])
	_, err = authSvc.UpdateProfile(ctx, UpdateProfileInput{UserID: userID, Nickname: &nickname, IP: "127.0.0.1", UserAgent: "fresh-quota-test"})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	ownedFamilyID := freshquota.CreateFamily(t, familySvc, userID, "李")
	for i := 0; i < 9; i++ {
		freshquota.AddMember(t, memberSvc, userID, ownedFamilyID, fmt.Sprintf("成员%d", i))
	}

	var identityBefore usermodel.UserAuthIdentity
	if err := tx.Where("user_id = ?", userID).First(&identityBefore).Error; err != nil {
		t.Fatalf("identity before bind: %v", err)
	}
	graphBefore := freshquota.GraphVersion(t, tx, ownedFamilyID)
	memberCountBefore := freshquota.CountActiveMembers(t, tx, ownedFamilyID)

	phone := freshquota.UniquePhone(runID, "bind")
	password := "bind-pass-" + runID[len(runID)-4:]
	bindResult, bindErr := authSvc.BindPhoneCredential(ctx, BindPhoneCredentialInput{
		UserID: userID, Phone: phone, Password: password, ConfirmPassword: password,
		IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if bindErr != nil {
		t.Fatalf("BindPhoneCredential: %v", bindErr)
	}
	if bindResult.User.ID != userID || !bindResult.User.PhoneLoginEnabled {
		t.Fatalf("unexpected bind result: %#v", bindResult.User)
	}

	var identityAfter usermodel.UserAuthIdentity
	if err := tx.Where("user_id = ?", userID).First(&identityAfter).Error; err != nil {
		t.Fatalf("identity after bind: %v", err)
	}
	if identityAfter.ID != identityBefore.ID || derefString(identityAfter.OpenIDHash) != derefString(identityBefore.OpenIDHash) {
		t.Fatal("wechat identity must remain unchanged after bind phone")
	}
	if freshquota.GraphVersion(t, tx, ownedFamilyID) != graphBefore {
		t.Fatal("bind phone must not change graph version")
	}
	if freshquota.CountActiveMembers(t, tx, ownedFamilyID) != memberCountBefore {
		t.Fatal("bind phone must not change member count")
	}

	caps, capErr := quotaSvc.GetCapabilities(ctx, userID)
	if capErr != nil || caps.TrustTier != quotaenum.TrustTierPhoneBound {
		t.Fatalf("unexpected PHONE_BOUND capabilities: %#v %v", caps, capErr)
	}
	phoneCfg := freshquota.ListTierConfigs(t, quotaSvc, "ROOT_ADMIN")[quotaenum.TrustTierPhoneBound]
	freshquota.AssertCapabilitiesLimits(t, caps, phoneCfg)

	relogin, err := authSvc.WechatMiniLogin(ctx, WechatMiniLoginInput{
		Code: code, ClientType: ClientWechatMini, IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("relogin after bind: %v", err)
	}
	if relogin.User.ID != userID || !relogin.User.PhoneLoginEnabled {
		t.Fatalf("relogin should keep same upgraded user: %#v", relogin.User)
	}
	caps2, _ := quotaSvc.GetCapabilities(ctx, userID)
	if caps2.TrustTier != quotaenum.TrustTierPhoneBound {
		t.Fatalf("token refresh path capabilities wrong: %#v", caps2)
	}
}

func TestWechatIdentityLookupUsesHashNotPlainLog(t *testing.T) {
	ctx := context.Background()
	tx, runID := freshquota.TestDB(t)
	wechatClient := freshquota.NewFakeWechatMiniClient("fresh-quota-test-app")
	openID := "secret_openid_" + runID
	wechatClient.BindCode(runID+"-code", openID, "secret_union_"+runID)
	authSvc := newFreshQuotaAuthService(t, tx, wechatClient)

	_, err := authSvc.WechatMiniLogin(ctx, WechatMiniLoginInput{
		Code: runID + "-code", ClientType: ClientWechatMini, IP: "127.0.0.1", UserAgent: "fresh-quota-test",
	})
	if err != nil {
		t.Fatalf("WechatMiniLogin: %v", err)
	}

	repo := authrepo.NewGormUserRepository(tx)
	identity, findErr := repo.FindIdentityByOpenIDHash(ctx, providerWechatMini, "fresh-quota-test-app", security.HashPlain(openID))
	if findErr != nil {
		t.Fatalf("FindIdentityByOpenIDHash: %v", findErr)
	}
	if identity.OpenID == nil || *identity.OpenID != openID {
		t.Fatal("identity should store openid for business use")
	}
}
