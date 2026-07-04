package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	accountmodel "tree/backend/internal/account/model"
	authrepo "tree/backend/internal/auth/repository"
	"tree/backend/internal/common/config"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	commonjwt "tree/backend/internal/common/jwt"
	commonredis "tree/backend/internal/common/redis"
	"tree/backend/internal/common/security"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"
)

type bindUserRepoFake struct {
	users      map[uint64]*usermodel.User
	phoneUsers map[string]*usermodel.User
	nextID     uint64
}

func newBindUserRepoFake(users ...*usermodel.User) *bindUserRepoFake {
	repo := &bindUserRepoFake{
		users:      make(map[uint64]*usermodel.User),
		phoneUsers: make(map[string]*usermodel.User),
		nextID:     100,
	}
	for _, user := range users {
		copy := *user
		if copy.ID == 0 {
			repo.nextID++
			copy.ID = repo.nextID
		}
		repo.users[copy.ID] = &copy
		if copy.PhoneHash != nil {
			repo.phoneUsers[*copy.PhoneHash] = &copy
		}
	}
	return repo
}

func (r *bindUserRepoFake) clone(user *usermodel.User) *usermodel.User {
	copy := *user
	return &copy
}

func (r *bindUserRepoFake) FindByPhoneHash(_ context.Context, phoneHash string) (*usermodel.User, error) {
	user, ok := r.phoneUsers[phoneHash]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return r.clone(user), nil
}

func (r *bindUserRepoFake) FindByID(_ context.Context, id uint64) (*usermodel.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return r.clone(user), nil
}

func (r *bindUserRepoFake) LockByID(ctx context.Context, id uint64) (*usermodel.User, error) {
	return r.FindByID(ctx, id)
}

func (r *bindUserRepoFake) Create(_ context.Context, user *usermodel.User) error {
	if user.ID == 0 {
		r.nextID++
		user.ID = r.nextID
	}
	copy := *user
	r.users[user.ID] = &copy
	return nil
}

func (r *bindUserRepoFake) RecordLoginSuccess(context.Context, uint64, string, string, time.Time) error {
	return nil
}

func (r *bindUserRepoFake) UpdateUser(_ context.Context, userID uint64, values map[string]any) error {
	user, ok := r.users[userID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if phone, ok := values["phone"].(string); ok {
		user.Phone = &phone
	}
	if phoneHash, ok := values["phone_hash"].(string); ok {
		user.PhoneHash = &phoneHash
	}
	if verified, ok := values["phone_verified"].(bool); ok {
		user.PhoneVerified = verified
	}
	if enabled, ok := values["phone_login_enabled"].(bool); ok {
		user.PhoneLoginEnabled = enabled
	}
	if raw, ok := values["phone_hash"]; ok && raw == nil {
		user.PhoneHash = nil
	}
	if _, ok := values["phone"]; ok && values["phone"] == nil {
		user.Phone = nil
	}
	if _, ok := values["password_hash"]; ok && values["password_hash"] == nil {
		user.PasswordHash = nil
	}
	if status, ok := values["status"].(string); ok {
		user.Status = status
		if status == string(enums.StatusMerged) {
			user.Phone = nil
			user.PhoneHash = nil
			user.PhoneVerified = false
			user.PhoneLoginEnabled = false
			user.PasswordHash = nil
		}
	}
	if nickname, ok := values["nickname"].(*string); ok {
		user.Nickname = nickname
	} else if nickname, ok := values["nickname"].(string); ok {
		user.Nickname = &nickname
	}
	if avatarURL, ok := values["avatar_url"].(string); ok {
		user.AvatarURL = &avatarURL
	}
	if hash, ok := values["password_hash"].(string); ok {
		user.PasswordHash = &hash
	}
	if mergedTo, ok := values["merged_to_user_id"].(uint64); ok {
		user.MergedToUserID = &mergedTo
	}
	if user.PhoneHash != nil {
		r.phoneUsers[*user.PhoneHash] = user
	}
	return nil
}

func (r *bindUserRepoFake) CreateIdentity(context.Context, *usermodel.UserAuthIdentity) error {
	return nil
}

func (r *bindUserRepoFake) FindIdentityByOpenIDHash(context.Context, string, string, string) (*usermodel.UserAuthIdentity, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *bindUserRepoFake) HasActiveWechatMiniIdentity(_ context.Context, _ uint64, _ string) (bool, error) {
	return false, nil
}

func (r *bindUserRepoFake) UpdateIdentity(context.Context, uint64, map[string]any) error { return nil }
func (r *bindUserRepoFake) CancelActiveIdentities(context.Context, uint64) error         { return nil }
func (r *bindUserRepoFake) MoveIdentities(context.Context, uint64, uint64) error         { return nil }
func (r *bindUserRepoFake) HasMemberBindingConflict(context.Context, uint64, uint64) (bool, error) {
	return false, nil
}
func (r *bindUserRepoFake) MoveFamilyLinks(context.Context, uint64, uint64) error { return nil }
func (r *bindUserRepoFake) CountActiveFamilyLinks(context.Context, uint64) (int64, error) {
	return 0, nil
}
func (r *bindUserRepoFake) CountFounderLinks(context.Context, uint64) (int64, error) { return 0, nil }
func (r *bindUserRepoFake) CountPendingFounderTransfer(context.Context, uint64) (int64, error) {
	return 0, nil
}
func (r *bindUserRepoFake) CountPendingDissolution(context.Context, uint64) (int64, error) {
	return 0, nil
}
func (r *bindUserRepoFake) CreatePhoneHistory(context.Context, *accountmodel.UserPhoneHistory) error {
	return nil
}
func (r *bindUserRepoFake) CreateMergeLog(context.Context, *accountmodel.UserAccountMergeLog) error {
	return nil
}
func (r *bindUserRepoFake) CreateClaimLog(context.Context, *accountmodel.UserAccountClaimLog) error {
	return nil
}
func (r *bindUserRepoFake) WithTx(*gorm.DB) authrepo.UserRepository { return r }

func bindCodeRepo(t *testing.T, phone, scene, code string) *codeRepoFake {
	t.Helper()
	codeHash, err := security.HashVerificationCode(phone, scene, code)
	if err != nil {
		t.Fatalf("HashVerificationCode: %v", err)
	}
	return &codeRepoFake{record: &usermodel.VerificationCode{
		ID: 1, PhoneHash: security.PhoneHash(phone), Scene: scene,
		CodeHash: codeHash, Status: string(enums.StatusPending), ExpiredAt: time.Now().Add(time.Minute),
	}}
}

func newBindService(t *testing.T, users *bindUserRepoFake, codes *codeRepoFake, logs *authLogCapture) *PhoneAuthService {
	t.Helper()
	return newBindServiceWithBlacklist(t, users, codes, logs, nil)
}

func newBindServiceWithBlacklist(t *testing.T, users authrepo.UserRepository, codes *codeRepoFake, logs *authLogCapture, blacklist commonredis.TokenBlacklist) *PhoneAuthService {
	t.Helper()
	return NewPhoneAuthService(nil, users, codes, &wechatClientFake{}, authTestManager(t), blacklist, logs, &config.Config{
		App:        config.AppConfig{Env: "test"},
		Wechat:     config.WechatConfig{MiniAppID: "test-app-id"},
		VerifyCode: config.VerifyCodeConfig{ExpireSeconds: 300, CooldownSeconds: 60},
	})
}

func bindPhoneInput(userID uint64, phone, code, password string) BindPhoneInput {
	pwd := password
	return BindPhoneInput{
		UserID: userID, Phone: phone, Code: code, Password: &pwd,
		IP: "127.0.0.1", UserAgent: "unit-test",
	}
}

func assertBindLogsSanitized(t *testing.T, logs *authLogCapture, phone, password, code string) {
	t.Helper()
	serialized, err := json.Marshal(struct {
		Success []operationlog.WriteInput
		Failed  []operationlog.WriteInput
	}{Success: logs.success, Failed: logs.failed})
	if err != nil {
		t.Fatalf("Marshal logs: %v", err)
	}
	text := string(serialized)
	for _, forbidden := range []string{password, code, phone, "openid", "unionid", "session_key", "AppSecret", "accessToken"} {
		if forbidden != "" && strings.Contains(text, forbidden) {
			t.Fatalf("operation log exposed sensitive value %q", forbidden)
		}
	}
}

func TestPrepareBindPhonePassword(t *testing.T) {
	service := newBindService(t, newBindUserRepoFake(), &codeRepoFake{}, &authLogCapture{})
	weak := "12345"
	hash, businessErr := service.prepareBindPhonePassword(&weak)
	if hash != nil || businessErr == nil || businessErr.Code != apperrors.CodeRegisterPasswordWeak {
		t.Fatalf("unexpected weak password result: %#v %#v", hash, businessErr)
	}
}

func TestApplyPasswordIfUnset(t *testing.T) {
	hash := "hashed-password"
	user := &usermodel.User{PasswordHash: stringPointer("existing")}
	updates := map[string]any{"status": string(enums.StatusActive)}
	applyPasswordIfUnset(updates, user, &hash)
	if _, ok := updates["password_hash"]; ok {
		t.Fatal("should not overwrite existing password")
	}

	emptyUser := &usermodel.User{}
	applyPasswordIfUnset(updates, emptyUser, &hash)
	if updates["password_hash"] != hash {
		t.Fatalf("expected password hash in updates, got %#v", updates)
	}
}

func TestWechatMiniLoginCreatesActiveUserWithoutPhone(t *testing.T) {
	const code = "wx-login-code"
	users := newBindUserRepoFake()
	logs := &authLogCapture{}
	wechatClient := &wechatClientFake{
		openID:  "mock_openid_abc",
		unionID: "mock_unionid_xyz",
	}
	service := NewPhoneAuthService(nil, users, &codeRepoFake{}, wechatClient, authTestManager(t), nil, logs, &config.Config{
		App: config.AppConfig{Env: "test"}, VerifyCode: config.VerifyCodeConfig{ExpireSeconds: 300},
	})

	result, businessErr := service.WechatMiniLogin(context.Background(), WechatMiniLoginInput{
		Code: code, ClientType: ClientWechatMini, IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if businessErr != nil {
		t.Fatalf("WechatMiniLogin: %v", businessErr)
	}
	if result.User.Status != string(enums.StatusActive) || result.User.PhoneVerified || result.User.Phone != nil {
		t.Fatalf("unexpected login user: %#v", result.User)
	}
	if len(users.users) != 1 {
		t.Fatalf("expected one created user, got %d", len(users.users))
	}
	if len(logs.success) != 1 || logs.success[0].Action != actionWechatLogin {
		t.Fatalf("unexpected success log: %#v", logs.success)
	}
}

func TestBindPhonePendingBindUserSetsPassword(t *testing.T) {
	const (
		phone    = "13800001001"
		code     = "654321"
		password = "bind-pass-01"
	)
	users := newBindUserRepoFake(&usermodel.User{
		ID: 201, Status: string(enums.StatusPendingBind), PhoneVerified: false,
		AccountOrigin: accountOriginWechatMini, RegisterClient: ClientWechatMini,
	})
	logs := &authLogCapture{}
	service := newBindService(t, users, bindCodeRepo(t, phone, SceneBindPhone, code), logs)

	result, businessErr := service.BindPhone(context.Background(), bindPhoneInput(201, phone, code, password))
	if businessErr != nil {
		t.Fatalf("BindPhone: %v", businessErr)
	}
	user := users.users[201]
	if user.PasswordHash == nil || !security.CheckPassword(password, *user.PasswordHash) {
		t.Fatal("password hash missing or invalid")
	}
	if user.Status != string(enums.StatusActive) || !user.PhoneVerified || user.Phone == nil || *user.Phone != phone {
		t.Fatalf("unexpected bound user: %#v", user)
	}
	if !result.User.PasswordSet {
		t.Fatalf("expected passwordSet=true, got %#v", result.User)
	}
	if len(logs.success) != 1 || logs.success[0].Action != actionBindPhone {
		t.Fatalf("unexpected success log: %#v", logs.success)
	}
	assertBindLogsSanitized(t, logs, phone, password, code)
}

func TestBindPhonePasswordTooShort(t *testing.T) {
	users := newBindUserRepoFake(&usermodel.User{ID: 202, Status: string(enums.StatusPendingBind)})
	logs := &authLogCapture{}
	service := newBindService(t, users, bindCodeRepo(t, "13800001002", SceneBindPhone, "654321"), logs)
	result, businessErr := service.BindPhone(context.Background(), bindPhoneInput(202, "13800001002", "654321", "12345"))
	if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeRegisterPasswordWeak {
		t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
	}
	if len(logs.failed) != 1 {
		t.Fatalf("expected failed log, got %#v", logs.failed)
	}
}

func TestBindPhoneMergeDoesNotOverwriteExistingPassword(t *testing.T) {
	const phone = "13800001003"
	phoneHash := security.PhoneHash(phone)
	existingHash, _ := security.HashPassword("existing-password")
	newPassword := "new-password-03"
	users := newBindUserRepoFake(
		&usermodel.User{ID: 301, Status: string(enums.StatusPendingBind), PhoneVerified: false},
		&usermodel.User{
			ID: 302, Phone: stringPointer(phone), PhoneHash: &phoneHash, PasswordHash: &existingHash,
			PhoneVerified: true, Status: string(enums.StatusActive),
		},
	)
	service := newBindService(t, users, bindCodeRepo(t, phone, SceneBindPhone, "111222"), &authLogCapture{})
	result, businessErr := service.BindPhone(context.Background(), bindPhoneInput(301, phone, "111222", newPassword))
	if businessErr != nil {
		t.Fatalf("BindPhone: %v", businessErr)
	}
	target := users.users[302]
	if target.PasswordHash == nil || !security.CheckPassword("existing-password", *target.PasswordHash) {
		t.Fatal("existing password was overwritten")
	}
	if !result.User.PasswordSet {
		t.Fatal("expected passwordSet=true on merged target")
	}
}

func TestBindPhoneMergeSetsPasswordWhenTargetEmpty(t *testing.T) {
	const phone = "13800001004"
	phoneHash := security.PhoneHash(phone)
	newPassword := "new-password-04"
	users := newBindUserRepoFake(
		&usermodel.User{ID: 401, Status: string(enums.StatusPendingBind), PhoneVerified: false},
		&usermodel.User{
			ID: 402, Phone: stringPointer(phone), PhoneHash: &phoneHash, PasswordHash: nil,
			PhoneVerified: true, Status: string(enums.StatusActive),
		},
	)
	service := newBindService(t, users, bindCodeRepo(t, phone, SceneBindPhone, "333444"), &authLogCapture{})
	result, businessErr := service.BindPhone(context.Background(), bindPhoneInput(401, phone, "333444", newPassword))
	if businessErr != nil {
		t.Fatalf("BindPhone: %v", businessErr)
	}
	target := users.users[402]
	if target.PasswordHash == nil || !security.CheckPassword(newPassword, *target.PasswordHash) {
		t.Fatal("expected password to be set on merged target")
	}
	if !result.User.PasswordSet {
		t.Fatal("expected passwordSet=true")
	}
}

func bindCredentialInput(userID uint64, phone, password string) BindPhoneCredentialInput {
	return BindPhoneCredentialInput{
		UserID: userID, Phone: phone, Password: password, ConfirmPassword: password,
		IP: "127.0.0.1", UserAgent: "unit-test",
	}
}

func TestBindPhoneCredentialKeepsSameUserAndUpgradesTier(t *testing.T) {
	const phone = "13800002001"
	const password = "secret-password-01"
	nickname := "微信用户"
	users := newBindUserRepoFake(&usermodel.User{
		ID: 501, Status: string(enums.StatusActive), Nickname: &nickname,
	})
	logs := &authLogCapture{}
	service := newBindService(t, users, &codeRepoFake{}, logs)
	result, businessErr := service.BindPhoneCredential(context.Background(), bindCredentialInput(501, phone, password))
	if businessErr != nil {
		t.Fatalf("BindPhoneCredential: %v", businessErr)
	}
	if result.User.ID != 501 || !result.User.PhoneLoginEnabled {
		t.Fatalf("unexpected bind result: %#v", result.User)
	}
	user := users.users[501]
	if user.Phone == nil || *user.Phone != phone || !user.PhoneLoginEnabled {
		t.Fatalf("unexpected stored user: %#v", user)
	}
	if user.PasswordHash == nil || security.CheckPassword(password, *user.PasswordHash) == false {
		t.Fatal("password must be stored as hash")
	}
	if user.PasswordHash != nil && *user.PasswordHash == password {
		t.Fatal("password must not be stored in plain text")
	}
	login, loginErr := service.LoginPhone(context.Background(), LoginPhoneInput{
		Phone: phone, Password: password, ClientType: "WECHAT_MINI", IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if loginErr != nil || login.User.ID != 501 {
		t.Fatalf("LoginPhone after bind: %#v %v", login, loginErr)
	}
	var bindLog bool
	for _, entry := range logs.success {
		if entry.Action == actionBindPhoneCredential {
			bindLog = true
		}
	}
	if !bindLog {
		t.Fatalf("expected BIND_PHONE_CREDENTIAL log: %#v", logs.success)
	}
	assertCredentialLogsSanitized(t, logs, phone, password)
}

func TestBindPhoneCredentialRejectsDuplicatePhone(t *testing.T) {
	const phone = "13800002002"
	phoneHash := security.PhoneHash(phone)
	existingHash, _ := security.HashPassword("existing-password")
	users := newBindUserRepoFake(
		&usermodel.User{ID: 601, Status: string(enums.StatusActive), Nickname: stringPointer("新用户")},
		&usermodel.User{
			ID: 602, Phone: stringPointer(phone), PhoneHash: &phoneHash, PasswordHash: &existingHash,
			PhoneLoginEnabled: true, Status: string(enums.StatusActive),
		},
	)
	service := newBindService(t, users, &codeRepoFake{}, &authLogCapture{})
	result, businessErr := service.BindPhoneCredential(context.Background(), bindCredentialInput(601, phone, "new-password-02"))
	if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeBindPhoneExists {
		t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
	}
}

func TestLoginPhoneRequiresPhoneLoginEnabled(t *testing.T) {
	const phone = "13800002003"
	phoneHash := security.PhoneHash(phone)
	passwordHash, _ := security.HashPassword("password-03")
	users := newBindUserRepoFake(&usermodel.User{
		ID: 701, Phone: stringPointer(phone), PhoneHash: &phoneHash, PasswordHash: &passwordHash,
		PhoneLoginEnabled: false, Status: string(enums.StatusActive),
	})
	service := newBindService(t, users, &codeRepoFake{}, &authLogCapture{})
	result, businessErr := service.LoginPhone(context.Background(), LoginPhoneInput{
		Phone: phone, Password: "password-03", ClientType: "WECHAT_MINI", IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeLoginPhonePasswordInvalid {
		t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
	}
}

func TestAdminUnbindPhoneLoginClearsCredential(t *testing.T) {
	const phone = "13800002004"
	phoneHash := security.PhoneHash(phone)
	passwordHash, _ := security.HashPassword("password-04")
	users := newWechatIdentityUserRepoFake(&usermodel.User{
		ID: 801, Phone: stringPointer(phone), PhoneHash: &phoneHash, PasswordHash: &passwordHash,
		PhoneLoginEnabled: true, Status: string(enums.StatusActive),
	})
	users.seedActiveIdentity(1, 801, providerWechatMini, "test-app-id", "openid-hash-801")
	logs := &authLogCapture{}
	service := newBindServiceWithBlacklist(t, users, &codeRepoFake{}, logs, nil)
	if err := service.AdminUnbindPhoneLogin(context.Background(), AdminUnbindPhoneLoginInput{
		AdminID: 1, AdminRole: string(enums.AdminRoleRootAdmin), TargetUserID: 801, Confirm: true,
		IP: "127.0.0.1", UserAgent: "unit-test",
	}); err != nil {
		t.Fatalf("AdminUnbindPhoneLogin: %v", err)
	}
	user := users.users[801]
	if user.Phone != nil || user.PhoneHash != nil || user.PasswordHash != nil || user.PhoneLoginEnabled {
		t.Fatalf("credential fields should be cleared: %#v", user)
	}
	login, loginErr := service.LoginPhone(context.Background(), LoginPhoneInput{
		Phone: phone, Password: "password-04", ClientType: "WECHAT_MINI", IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if login != nil || loginErr == nil {
		t.Fatalf("phone login should fail after unbind: %#v %v", login, loginErr)
	}
}

func TestAdminUnbindPhoneLoginRejectsPhoneOnlyAccount(t *testing.T) {
	const phone = "13800002005"
	phoneHash := security.PhoneHash(phone)
	passwordHash, _ := security.HashPassword("password-05")
	users := newBindUserRepoFake(&usermodel.User{
		ID: 802, Phone: stringPointer(phone), PhoneHash: &phoneHash, PasswordHash: &passwordHash,
		PhoneLoginEnabled: true, Status: string(enums.StatusActive),
	})
	service := newBindService(t, users, &codeRepoFake{}, &authLogCapture{})
	err := service.AdminUnbindPhoneLogin(context.Background(), AdminUnbindPhoneLoginInput{
		AdminID: 1, AdminRole: string(enums.AdminRoleRootAdmin), TargetUserID: 802, Confirm: true,
		IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if err == nil || err.Code != apperrors.CodeAdminUnbindPhoneLoginNoWechat {
		t.Fatalf("expected no-wechat error, got %#v", err)
	}
}

func TestAdminUnbindPhoneLoginRejectsOtherAppIDIdentity(t *testing.T) {
	const phone = "13800002007"
	phoneHash := security.PhoneHash(phone)
	passwordHash, _ := security.HashPassword("password-07")
	users := newWechatIdentityUserRepoFake(&usermodel.User{
		ID: 804, Phone: stringPointer(phone), PhoneHash: &phoneHash, PasswordHash: &passwordHash,
		PhoneLoginEnabled: true, Status: string(enums.StatusActive),
	})
	users.seedActiveIdentity(2, 804, providerWechatMini, "other-app-id", "openid-hash-804")
	service := newBindServiceWithBlacklist(t, users, &codeRepoFake{}, &authLogCapture{}, nil)
	err := service.AdminUnbindPhoneLogin(context.Background(), AdminUnbindPhoneLoginInput{
		AdminID: 1, AdminRole: string(enums.AdminRoleRootAdmin), TargetUserID: 804, Confirm: true,
		IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if err == nil || err.Code != apperrors.CodeAdminUnbindPhoneLoginNoWechat {
		t.Fatalf("expected no-wechat error for other app id, got %#v", err)
	}
}

func TestChangePhoneLoginPasswordRotatesToken(t *testing.T) {
	const phone = "13800002006"
	phoneHash := security.PhoneHash(phone)
	oldPassword := "password-old-06"
	newPassword := "password-new-06"
	passwordHash, _ := security.HashPassword(oldPassword)
	users := newBindUserRepoFake(&usermodel.User{
		ID: 803, Phone: stringPointer(phone), PhoneHash: &phoneHash, PasswordHash: &passwordHash,
		PhoneLoginEnabled: true, Status: string(enums.StatusActive),
	})
	blacklist := &tokenBlacklistFake{userRevokedAt: make(map[uint64]int64)}
	service := newBindServiceWithBlacklist(t, users, &codeRepoFake{}, &authLogCapture{}, blacklist)
	user := users.users[803]
	token1, err1 := service.issueUserToken(user)
	token2, err2 := service.issueUserToken(user)
	if err1 != nil || err2 != nil || token1 == nil || token2 == nil {
		t.Fatalf("issue old tokens: %#v %#v %v %v", token1, token2, err1, err2)
	}
	manager := authTestManager(t)
	claims1, parseErr1 := manager.Parse(token1.AccessToken, commonjwt.TokenTypeUser)
	claims2, parseErr2 := manager.Parse(token2.AccessToken, commonjwt.TokenTypeUser)
	if parseErr1 != nil || parseErr2 != nil {
		t.Fatalf("parse old tokens: %v %v", parseErr1, parseErr2)
	}

	result, businessErr := service.ChangePhoneLoginPassword(context.Background(), ChangePhoneLoginPasswordInput{
		UserID: 803, CurrentPassword: oldPassword, NewPassword: newPassword,
		IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if businessErr != nil || result == nil || result.AccessToken == "" {
		t.Fatalf("ChangePhoneLoginPassword: %#v %v", result, businessErr)
	}
	revokedAt, ok := blacklist.userRevokedAt[803]
	if !ok {
		t.Fatal("expected user-wide token revocation")
	}
	if claims1.IssuedAt > revokedAt || claims2.IssuedAt > revokedAt {
		t.Fatalf("old tokens should be revoked: iat1=%d iat2=%d revokedAt=%d", claims1.IssuedAt, claims2.IssuedAt, revokedAt)
	}
	newClaims, parseErr := manager.Parse(result.AccessToken, commonjwt.TokenTypeUser)
	if parseErr != nil {
		t.Fatalf("parse new token: %v", parseErr)
	}
	if newClaims.IssuedAt <= revokedAt {
		t.Fatalf("new token iat must exceed revokedAt: iat=%d revokedAt=%d", newClaims.IssuedAt, revokedAt)
	}
	if revoked, _ := blacklist.IsUserTokenRevoked(context.Background(), 803, newClaims.IssuedAt); revoked {
		t.Fatalf("new token must remain valid: iat=%d revokedAt=%d", newClaims.IssuedAt, revokedAt)
	}
	if revoked, _ := blacklist.IsUserTokenRevoked(context.Background(), 803, claims1.IssuedAt); !revoked {
		t.Fatal("token1 should be revoked")
	}
	if revoked, _ := blacklist.IsUserTokenRevoked(context.Background(), 803, claims2.IssuedAt); !revoked {
		t.Fatal("token2 should be revoked")
	}
	user = users.users[803]
	if user.PasswordHash == nil || !security.CheckPassword(newPassword, *user.PasswordHash) {
		t.Fatal("password hash not updated")
	}
}

type tokenBlacklistFake struct {
	revoked       []string
	userRevokedAt map[uint64]int64
}

func (f *tokenBlacklistFake) Revoke(_ context.Context, tokenID string, _ time.Duration) error {
	f.revoked = append(f.revoked, tokenID)
	return nil
}

func (f *tokenBlacklistFake) IsRevoked(context.Context, string) (bool, error) { return false, nil }

func (f *tokenBlacklistFake) RevokeUserTokens(_ context.Context, userID uint64, revokedAt int64, _ time.Duration) error {
	if f.userRevokedAt == nil {
		f.userRevokedAt = make(map[uint64]int64)
	}
	f.userRevokedAt[userID] = revokedAt
	return nil
}

func (f *tokenBlacklistFake) IsUserTokenRevoked(_ context.Context, userID uint64, issuedAt int64) (bool, error) {
	revokedAt, ok := f.userRevokedAt[userID]
	if !ok {
		return false, nil
	}
	return issuedAt <= revokedAt, nil
}

func TestAdminUnbindPhoneLoginForbiddenForPlatformAdmin(t *testing.T) {
	users := newBindUserRepoFake(&usermodel.User{
		ID: 901, PhoneLoginEnabled: true, Status: string(enums.StatusActive),
	})
	service := newBindService(t, users, &codeRepoFake{}, &authLogCapture{})
	err := service.AdminUnbindPhoneLogin(context.Background(), AdminUnbindPhoneLoginInput{
		AdminID: 2, AdminRole: string(enums.AdminRolePlatformAdmin), TargetUserID: 901, Confirm: true,
		IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if err == nil || err.Code != apperrors.CodeForbidden {
		t.Fatalf("expected forbidden, got %#v", err)
	}
}

func assertCredentialLogsSanitized(t *testing.T, logs *authLogCapture, phone, password string) {
	t.Helper()
	serialized, err := json.Marshal(struct {
		Success []operationlog.WriteInput
		Failed  []operationlog.WriteInput
	}{Success: logs.success, Failed: logs.failed})
	if err != nil {
		t.Fatalf("marshal logs: %v", err)
	}
	body := string(serialized)
	if strings.Contains(body, phone) || strings.Contains(body, password) {
		t.Fatalf("logs must not contain phone or password: %s", body)
	}
}
