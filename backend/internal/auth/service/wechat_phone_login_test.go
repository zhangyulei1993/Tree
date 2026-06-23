package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"tree/backend/internal/common/config"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/security"
	"tree/backend/internal/common/wechat"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"
)

type wechatClientFake struct {
	phone    string
	err      error
	callCode string
}

func (f *wechatClientFake) Code2Session(context.Context, string) (*wechat.Code2SessionResult, error) {
	return nil, errors.New("not implemented")
}

func (f *wechatClientFake) GetPhoneNumber(_ context.Context, phoneCode string) (string, error) {
	f.callCode = phoneCode
	if f.err != nil {
		return "", f.err
	}
	if f.phone != "" {
		return f.phone, nil
	}
	return "", errors.New("wechat phone fetch failed")
}

func newWechatPhoneService(t *testing.T, users *userRepoFake, wechatClient wechat.Client, logs *authLogCapture) *PhoneAuthService {
	t.Helper()
	return NewPhoneAuthService(nil, users, &codeRepoFake{}, wechatClient, authTestManager(t), nil, logs, &config.Config{
		App:        config.AppConfig{Env: "test"},
		VerifyCode: config.VerifyCodeConfig{ExpireSeconds: 300, CooldownSeconds: 60},
	})
}

func wechatPhoneLoginInput(phoneCode string) WechatMiniPhoneLoginInput {
	return WechatMiniPhoneLoginInput{
		PhoneCode:  phoneCode,
		ClientType: ClientWechatMini,
		IP:         "127.0.0.1",
		UserAgent:  "unit-test",
	}
}

func assertLogsSanitized(t *testing.T, logs *authLogCapture, phoneCode string, phone string) {
	t.Helper()
	serialized, err := json.Marshal(struct {
		Success []operationlog.WriteInput
		Failed  []operationlog.WriteInput
	}{Success: logs.success, Failed: logs.failed})
	if err != nil {
		t.Fatalf("Marshal logs: %v", err)
	}
	text := string(serialized)
	for _, forbidden := range []string{
		phoneCode,
		"access_token",
		"accessToken",
		"AppSecret",
		"session_key",
		"openid",
		"unionid",
		phone,
	} {
		if forbidden != "" && strings.Contains(text, forbidden) {
			t.Fatalf("operation log exposed sensitive value %q", forbidden)
		}
	}
}

func TestWechatMiniPhoneLoginFailures(t *testing.T) {
	t.Run("empty phone code", func(t *testing.T) {
		logs := &authLogCapture{}
		service := newWechatPhoneService(t, &userRepoFake{}, &wechatClientFake{}, logs)
		result, businessErr := service.WechatMiniPhoneLogin(context.Background(), wechatPhoneLoginInput(""))
		if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeWechatPhoneCodeInvalid {
			t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
		}
		if len(logs.failed) != 1 || logs.failed[0].Action != actionWechatPhoneLogin {
			t.Fatalf("unexpected failed log: %#v", logs.failed)
		}
	})

	t.Run("wechat phone fetch failed", func(t *testing.T) {
		logs := &authLogCapture{}
		client := &wechatClientFake{err: errors.New("wechat api error")}
		service := newWechatPhoneService(t, &userRepoFake{}, client, logs)
		phoneCode := "dynamic-phone-code-token"
		result, businessErr := service.WechatMiniPhoneLogin(context.Background(), wechatPhoneLoginInput(phoneCode))
		if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeWechatPhoneFetchFailed {
			t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
		}
		if client.callCode != phoneCode || len(logs.failed) != 1 {
			t.Fatalf("unexpected client/log state: call=%q logs=%#v", client.callCode, logs.failed)
		}
		assertLogsSanitized(t, logs, phoneCode, "")
	})
}

func TestWechatMiniPhoneLoginCreatesActiveUserWithoutPassword(t *testing.T) {
	const phone = "13812345678"
	phoneCode := "wx-phone-code-new-user"
	users := &userRepoFake{}
	logs := &authLogCapture{}
	service := newWechatPhoneService(t, users, &wechatClientFake{phone: phone}, logs)

	result, businessErr := service.WechatMiniPhoneLogin(context.Background(), wechatPhoneLoginInput(phoneCode))
	if businessErr != nil {
		t.Fatalf("WechatMiniPhoneLogin: %v", businessErr)
	}
	if users.user == nil {
		t.Fatal("expected created user")
	}
	if users.user.PasswordHash != nil {
		t.Fatalf("expected nil password hash, got %#v", users.user.PasswordHash)
	}
	if !users.user.PhoneVerified || users.user.Status != string(enums.StatusActive) ||
		users.user.AccountOrigin != accountOriginWechatMiniPhone ||
		users.user.RegisterClient != ClientWechatMini {
		t.Fatalf("unexpected created user: %#v", users.user)
	}
	if result.User.PasswordSet {
		t.Fatal("expected passwordSet=false for new user")
	}
	if result.User.Phone == nil || *result.User.Phone != phone || !result.User.PhoneVerified {
		t.Fatalf("unexpected login user info: %#v", result.User)
	}
	if len(logs.success) != 1 || logs.success[0].Action != actionWechatPhoneLogin {
		t.Fatalf("unexpected success log: %#v", logs.success)
	}
	assertLogsSanitized(t, logs, phoneCode, phone)
}

func TestWechatMiniPhoneLoginExistingActiveUser(t *testing.T) {
	const phone = "13898765432"
	phoneHash := security.PhoneHash(phone)
	passwordHash, err := security.HashPassword("existing-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	users := &userRepoFake{user: &usermodel.User{
		ID: 55, Phone: stringPointer(phone), PhoneHash: &phoneHash, PasswordHash: &passwordHash,
		PhoneVerified: true, Status: string(enums.StatusActive),
	}}
	logs := &authLogCapture{}
	service := newWechatPhoneService(t, users, &wechatClientFake{phone: phone}, logs)

	result, businessErr := service.WechatMiniPhoneLogin(context.Background(), wechatPhoneLoginInput("existing-user-code"))
	if businessErr != nil {
		t.Fatalf("WechatMiniPhoneLogin: %v", businessErr)
	}
	if result.User.ID != 55 || !result.User.PasswordSet {
		t.Fatalf("unexpected login response: %#v", result.User)
	}
	if users.recordLoginCalls != 1 || len(logs.success) != 1 {
		t.Fatalf("unexpected login side effects: calls=%d logs=%#v", users.recordLoginCalls, logs.success)
	}
}

func TestWechatMiniPhoneLoginRejectsInvalidAccountStatus(t *testing.T) {
	const phone = "13811112222"
	phoneHash := security.PhoneHash(phone)

	for _, test := range []struct {
		name   string
		status string
	}{
		{name: "disabled", status: string(enums.StatusDisabled)},
		{name: "cancelled", status: string(enums.StatusCancelled)},
		{name: "deleted", status: string(enums.StatusDeleted)},
	} {
		t.Run(test.name, func(t *testing.T) {
			users := &userRepoFake{user: &usermodel.User{
				ID: 66, Phone: stringPointer(phone), PhoneHash: &phoneHash,
				PhoneVerified: true, Status: test.status,
			}}
			logs := &authLogCapture{}
			service := newWechatPhoneService(t, users, &wechatClientFake{phone: phone}, logs)
			result, businessErr := service.WechatMiniPhoneLogin(context.Background(), wechatPhoneLoginInput("status-test-code"))
			if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeWechatPhoneLoginInvalid {
				t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
			}
			if len(logs.failed) != 1 || logs.failed[0].Action != actionWechatPhoneLogin {
				t.Fatalf("unexpected failed log: %#v", logs.failed)
			}
		})
	}
}

func TestWechatMiniPhoneLoginPasswordSetFlag(t *testing.T) {
	const phone = "13833334444"
	phoneHash := security.PhoneHash(phone)

	t.Run("without password", func(t *testing.T) {
		users := &userRepoFake{user: &usermodel.User{
			ID: 77, Phone: stringPointer(phone), PhoneHash: &phoneHash,
			PasswordHash: nil, PhoneVerified: true, Status: string(enums.StatusActive),
		}}
		service := newWechatPhoneService(t, users, &wechatClientFake{phone: phone}, &authLogCapture{})
		result, businessErr := service.WechatMiniPhoneLogin(context.Background(), wechatPhoneLoginInput("no-password-code"))
		if businessErr != nil {
			t.Fatalf("WechatMiniPhoneLogin: %v", businessErr)
		}
		if result.User.PasswordSet {
			t.Fatal("expected passwordSet=false")
		}
	})

	t.Run("with password", func(t *testing.T) {
		hash, _ := security.HashPassword("has-password")
		users := &userRepoFake{user: &usermodel.User{
			ID: 78, Phone: stringPointer(phone), PhoneHash: &phoneHash,
			PasswordHash: &hash, PhoneVerified: true, Status: string(enums.StatusActive),
		}}
		service := newWechatPhoneService(t, users, &wechatClientFake{phone: phone}, &authLogCapture{})
		result, businessErr := service.WechatMiniPhoneLogin(context.Background(), wechatPhoneLoginInput("has-password-code"))
		if businessErr != nil {
			t.Fatalf("WechatMiniPhoneLogin: %v", businessErr)
		}
		if !result.User.PasswordSet {
			t.Fatal("expected passwordSet=true")
		}
	})
}
