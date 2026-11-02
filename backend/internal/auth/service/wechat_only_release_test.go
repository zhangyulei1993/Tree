package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	accountmodel "tree/backend/internal/account/model"
	authrepo "tree/backend/internal/auth/repository"
	"tree/backend/internal/common/config"
	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/security"
	usermodel "tree/backend/internal/user/model"
)

type wechatIdentityUserRepoFake struct {
	bindUserRepoFake
	identityRecords []*usermodel.UserAuthIdentity
	nextIdentityID  uint64
}

func newWechatIdentityUserRepoFake(users ...*usermodel.User) *wechatIdentityUserRepoFake {
	return &wechatIdentityUserRepoFake{
		bindUserRepoFake: *newBindUserRepoFake(users...),
		identityRecords:  []*usermodel.UserAuthIdentity{},
		nextIdentityID:   1,
	}
}

func (r *wechatIdentityUserRepoFake) CreateIdentity(_ context.Context, identity *usermodel.UserAuthIdentity) error {
	copy := *identity
	if copy.ID == 0 {
		copy.ID = r.nextIdentityID
		r.nextIdentityID++
	}
	r.identityRecords = append(r.identityRecords, &copy)
	return nil
}

func (r *wechatIdentityUserRepoFake) FindIdentityByOpenIDHash(_ context.Context, provider, appID, openIDHash string) (*usermodel.UserAuthIdentity, error) {
	for _, identity := range r.identityRecords {
		if identity.Provider != provider ||
			derefString(identity.ProviderAppID) != appID ||
			identity.OpenIDHash == nil || *identity.OpenIDHash != openIDHash ||
			identity.IdentityStatus != string(enums.StatusActive) {
			continue
		}
		copy := *identity
		return &copy, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *wechatIdentityUserRepoFake) HasActiveWechatMiniIdentity(_ context.Context, userID uint64, appID string) (bool, error) {
	for _, identity := range r.identityRecords {
		if identity.UserID == userID &&
			identity.Provider == providerWechatMini &&
			derefString(identity.ProviderAppID) == appID &&
			identity.IdentityStatus == string(enums.StatusActive) {
			return true, nil
		}
	}
	return false, nil
}

func (r *wechatIdentityUserRepoFake) FindActiveWechatMiniOpenID(_ context.Context, userID uint64, appID string) (string, error) {
	for _, identity := range r.identityRecords {
		if identity.UserID == userID &&
			identity.Provider == providerWechatMini &&
			derefString(identity.ProviderAppID) == appID &&
			identity.IdentityStatus == string(enums.StatusActive) &&
			identity.OpenID != nil && strings.TrimSpace(*identity.OpenID) != "" {
			return strings.TrimSpace(*identity.OpenID), nil
		}
	}
	return "", gorm.ErrRecordNotFound
}

func (r *wechatIdentityUserRepoFake) WithTx(*gorm.DB) authrepo.UserRepository { return r }

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (r *wechatIdentityUserRepoFake) CancelActiveIdentities(_ context.Context, userID uint64) error {
	now := time.Now()
	for _, identity := range r.identityRecords {
		if identity.UserID == userID && identity.IdentityStatus == string(enums.StatusActive) {
			identity.IdentityStatus = string(enums.StatusCancelled)
			identity.UnboundAt = &now
		}
	}
	return nil
}

func (r *wechatIdentityUserRepoFake) countCancelledIdentitiesByOpenID(provider, appID, openIDHash string) int {
	count := 0
	for _, identity := range r.identityRecords {
		if identity.Provider == provider &&
			derefString(identity.ProviderAppID) == appID &&
			identity.OpenIDHash != nil && *identity.OpenIDHash == openIDHash &&
			identity.IdentityStatus == string(enums.StatusCancelled) {
			count++
		}
	}
	return count
}

func (r *wechatIdentityUserRepoFake) seedActiveIdentity(id, userID uint64, provider, appID, openIDHash string) {
	app := appID
	hash := openIDHash
	r.identityRecords = append(r.identityRecords, &usermodel.UserAuthIdentity{
		ID: id, UserID: userID, Provider: provider, ProviderAppID: &app,
		OpenIDHash: &hash, IdentityStatus: string(enums.StatusActive),
	})
	if id >= r.nextIdentityID {
		r.nextIdentityID = id + 1
	}
}

func (r *wechatIdentityUserRepoFake) UpdateUser(_ context.Context, userID uint64, values map[string]any) error {
	user, ok := r.users[userID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if status, ok := values["status"].(string); ok {
		user.Status = status
	}
	if nickname, ok := values["nickname"]; ok && nickname == nil {
		user.Nickname = nil
	}
	return r.bindUserRepoFake.UpdateUser(context.Background(), userID, values)
}

func TestWechatMiniLoginUpgradesPendingBindUser(t *testing.T) {
	openID := "mock_openid_upgrade"
	openIDHash := security.HashPlain(openID)
	appID := "test-app-id"
	users := newWechatIdentityUserRepoFake(&usermodel.User{
		ID: 501, Status: string(enums.StatusPendingBind), PhoneVerified: false,
		AccountOrigin: accountOriginWechatMini, RegisterClient: ClientWechatMini,
	})
	users.seedActiveIdentity(1, 501, providerWechatMini, appID, openIDHash)
	service := NewPhoneAuthService(nil, users, &codeRepoFake{}, &wechatClientFake{openID: openID}, authTestManager(t), nil, &authLogCapture{}, contentsafety.AlwaysPass(), &config.Config{
		App: config.AppConfig{Env: "test"}, Wechat: config.WechatConfig{MiniAppID: appID},
	})

	result, businessErr := service.WechatMiniLogin(context.Background(), WechatMiniLoginInput{
		Code: "wx-code", ClientType: ClientWechatMini, IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if businessErr != nil {
		t.Fatalf("WechatMiniLogin: %v", businessErr)
	}
	if result.User.Status != string(enums.StatusActive) {
		t.Fatalf("expected ACTIVE user, got %#v", result.User)
	}
	if users.users[501].Status != string(enums.StatusActive) {
		t.Fatalf("user not upgraded in repo: %#v", users.users[501])
	}
}

func TestWechatMiniLoginDoesNotRestoreDisabledOrCancelledUsers(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{name: "disabled", status: string(enums.StatusDisabled)},
		{name: "cancelled", status: string(enums.StatusCancelled)},
		{name: "deleted", status: string(enums.StatusDeleted)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			openID := "openid_" + test.name
			openIDHash := security.HashPlain(openID)
			appID := "test-app-id"
			users := newWechatIdentityUserRepoFake(&usermodel.User{
				ID: 601, Status: test.status, PhoneVerified: false,
			})
			users.seedActiveIdentity(2, 601, providerWechatMini, appID, openIDHash)
			service := NewPhoneAuthService(nil, users, &codeRepoFake{}, &wechatClientFake{openID: openID}, authTestManager(t), nil, &authLogCapture{}, contentsafety.AlwaysPass(), &config.Config{
				App: config.AppConfig{Env: "test"}, Wechat: config.WechatConfig{MiniAppID: appID},
			})
			result, businessErr := service.WechatMiniLogin(context.Background(), WechatMiniLoginInput{
				Code: "wx-code", ClientType: ClientWechatMini, IP: "127.0.0.1", UserAgent: "unit-test",
			})
			if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeWechatAccountInvalid {
				t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
			}
		})
	}
}

func TestCancelAccountByWechatReauthSuccess(t *testing.T) {
	openID := "cancel_openid"
	openIDHash := security.HashPlain(openID)
	appID := "test-app-id"
	users := newWechatIdentityUserRepoFake(&usermodel.User{
		ID: 701, Status: string(enums.StatusActive), PhoneVerified: false, PhoneLoginEnabled: true,
		Nickname: stringPointer("测试用户"),
	})
	users.seedActiveIdentity(3, 701, providerWechatMini, appID, openIDHash)
	logs := &authLogCapture{}
	service := NewPhoneAuthService(nil, users, &codeRepoFake{}, &wechatClientFake{openID: openID}, authTestManager(t), nil, logs, contentsafety.AlwaysPass(), &config.Config{
		App: config.AppConfig{Env: "test"}, Wechat: config.WechatConfig{MiniAppID: appID},
	})

	businessErr := service.CancelAccountByWechatReauth(context.Background(), CancelAccountByWechatInput{
		UserID: 701, Code: "wx-code", TokenID: "token-1", ExpiresAt: time.Now().Add(time.Hour).Unix(),
		IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if businessErr != nil {
		t.Fatalf("CancelAccountByWechatReauth: %v", businessErr)
	}
	if users.users[701].Status != string(enums.StatusCancelled) {
		t.Fatalf("expected cancelled user, got %#v", users.users[701])
	}
	if users.users[701].PhoneLoginEnabled {
		t.Fatal("cancel must clear phone_login_enabled")
	}
	if len(logs.success) != 1 || logs.success[0].Action != actionCancelAccountWechat {
		t.Fatalf("unexpected success log: %#v", logs.success)
	}
}

func TestCancelAccountByWechatReauthIdentityMismatch(t *testing.T) {
	openID := "owner_openid"
	otherOpenID := "other_openid"
	openIDHash := security.HashPlain(openID)
	appID := "test-app-id"
	users := newWechatIdentityUserRepoFake(&usermodel.User{ID: 801, Status: string(enums.StatusActive)})
	users.seedActiveIdentity(4, 802, providerWechatMini, appID, openIDHash)
	service := NewPhoneAuthService(nil, users, &codeRepoFake{}, &wechatClientFake{openID: otherOpenID}, authTestManager(t), nil, &authLogCapture{}, contentsafety.AlwaysPass(), &config.Config{
		App: config.AppConfig{Env: "test"}, Wechat: config.WechatConfig{MiniAppID: appID},
	})

	businessErr := service.CancelAccountByWechatReauth(context.Background(), CancelAccountByWechatInput{
		UserID: 801, Code: "wx-code", TokenID: "token-1", ExpiresAt: time.Now().Add(time.Hour).Unix(),
		IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if businessErr == nil || businessErr.Code != apperrors.CodeCancelWechatIdentityMismatch {
		t.Fatalf("expected identity mismatch, got %#v", businessErr)
	}
}

func TestCancelAccountByWechatReauthInvalidCode(t *testing.T) {
	users := newWechatIdentityUserRepoFake(&usermodel.User{ID: 901, Status: string(enums.StatusActive)})
	service := NewPhoneAuthService(nil, users, &codeRepoFake{}, &wechatClientFake{err: errors.New("wechat code2session failed")}, authTestManager(t), nil, &authLogCapture{}, contentsafety.AlwaysPass(), &config.Config{
		App: config.AppConfig{Env: "test"},
	})

	businessErr := service.CancelAccountByWechatReauth(context.Background(), CancelAccountByWechatInput{
		UserID: 901, Code: "bad-code", TokenID: "token-1", ExpiresAt: time.Now().Add(time.Hour).Unix(),
		IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if businessErr == nil || businessErr.Code != apperrors.CodeCancelWechatCodeInvalid {
		t.Fatalf("expected invalid code, got %#v", businessErr)
	}
}

func TestWechatAccountFullLifecycleLoginCancelReregister(t *testing.T) {
	const openID = "lifecycle_openid"
	openIDHash := security.HashPlain(openID)
	appID := "test-app-id"
	users := newWechatIdentityUserRepoFake()
	logs := &authLogCapture{}
	service := NewPhoneAuthService(nil, users, &codeRepoFake{}, &wechatClientFake{openID: openID}, authTestManager(t), nil, logs, contentsafety.AlwaysPass(), &config.Config{
		App: config.AppConfig{Env: "test"}, Wechat: config.WechatConfig{MiniAppID: appID},
	})
	ctx := context.Background()
	loginInput := WechatMiniLoginInput{
		Code: "wx-code", ClientType: ClientWechatMini, IP: "127.0.0.1", UserAgent: "unit-test",
	}
	cancelInput := func(userID uint64) CancelAccountByWechatInput {
		return CancelAccountByWechatInput{
			UserID: userID, Code: "wx-code", TokenID: "token-1", ExpiresAt: time.Now().Add(time.Hour).Unix(),
			IP: "127.0.0.1", UserAgent: "unit-test",
		}
	}

	firstLogin, businessErr := service.WechatMiniLogin(ctx, loginInput)
	if businessErr != nil {
		t.Fatalf("first WechatMiniLogin: %v", businessErr)
	}
	firstUserID := firstLogin.User.ID

	if businessErr = service.CancelAccountByWechatReauth(ctx, cancelInput(firstUserID)); businessErr != nil {
		t.Fatalf("first CancelAccountByWechatReauth: %v", businessErr)
	}

	secondLogin, businessErr := service.WechatMiniLogin(ctx, loginInput)
	if businessErr != nil {
		t.Fatalf("second WechatMiniLogin: %v", businessErr)
	}
	secondUserID := secondLogin.User.ID
	if secondUserID == 0 || secondUserID == firstUserID {
		t.Fatalf("expected new active user after re-register, first=%d second=%d users=%#v", firstUserID, secondUserID, users.users)
	}

	if businessErr = service.CancelAccountByWechatReauth(ctx, cancelInput(secondUserID)); businessErr != nil {
		t.Fatalf("second CancelAccountByWechatReauth: %v", businessErr)
	}

	if users.countCancelledIdentitiesByOpenID(providerWechatMini, appID, openIDHash) != 2 {
		t.Fatalf("expected two cancelled identities, got %#v", users.identityRecords)
	}
	cancelLogs := 0
	for _, log := range logs.success {
		if log.Action == actionCancelAccountWechat {
			cancelLogs++
		}
	}
	if cancelLogs != 2 {
		t.Fatalf("expected two cancel success logs, got %#v", logs.success)
	}
}

// Ensure wechatIdentityUserRepoFake satisfies repository methods used in cancel flow.
var _ authrepo.UserRepository = (*wechatIdentityUserRepoFake)(nil)

func (r *wechatIdentityUserRepoFake) CreatePhoneHistory(context.Context, *accountmodel.UserPhoneHistory) error {
	return nil
}

func (r *wechatIdentityUserRepoFake) MoveIdentities(context.Context, uint64, uint64) error {
	return nil
}
