package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	authrepo "tree/backend/internal/auth/repository"
	"tree/backend/internal/common/config"
	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	commonjwt "tree/backend/internal/common/jwt"
	"tree/backend/internal/common/security"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"
)

type userRepoFake struct {
	authrepo.UserRepository
	user             *usermodel.User
	recordLoginCalls int
}

func (r *userRepoFake) FindByPhoneHash(_ context.Context, phoneHash string) (*usermodel.User, error) {
	if r.user == nil || r.user.PhoneHash == nil || *r.user.PhoneHash != phoneHash {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *r.user
	return &copy, nil
}
func (r *userRepoFake) FindByID(_ context.Context, id uint64) (*usermodel.User, error) {
	if r.user == nil || r.user.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *r.user
	return &copy, nil
}
func (r *userRepoFake) Create(_ context.Context, user *usermodel.User) error {
	if user.ID == 0 {
		user.ID = 101
	}
	copy := *user
	r.user = &copy
	return nil
}
func (r *userRepoFake) RecordLoginSuccess(context.Context, uint64, string, string, time.Time) error {
	r.recordLoginCalls++
	return nil
}
func (r *userRepoFake) WithTx(*gorm.DB) authrepo.UserRepository { return r }

type codeRepoFake struct {
	authrepo.VerificationCodeRepository
	record *usermodel.VerificationCode
	used   bool
}

func (r *codeRepoFake) FindLatestPending(context.Context, string, string) (*usermodel.VerificationCode, error) {
	if r.record == nil {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *r.record
	return &copy, nil
}
func (r *codeRepoFake) MarkUsed(context.Context, uint64, time.Time) error {
	r.used = true
	return nil
}
func (r *codeRepoFake) WithTx(*gorm.DB) authrepo.VerificationCodeRepository { return r }

type authLogCapture struct {
	success []operationlog.WriteInput
	failed  []operationlog.WriteInput
}

func (l *authLogCapture) WriteSuccess(_ context.Context, input operationlog.WriteInput) error {
	l.success = append(l.success, input)
	return nil
}
func (l *authLogCapture) WriteFailed(_ context.Context, input operationlog.WriteInput) error {
	l.failed = append(l.failed, input)
	return nil
}

func authTestManager(t *testing.T) *commonjwt.Manager {
	t.Helper()
	manager, err := commonjwt.NewManager(config.JWTConfig{
		UserSecret: "auth-test-user-secret", AdminSecret: "auth-test-admin-secret",
		AccessTokenExpireMinutes: 5,
	}, "TreeTest")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return manager
}

func newPhoneService(t *testing.T, users *userRepoFake, codes *codeRepoFake, logs *authLogCapture) *PhoneAuthService {
	t.Helper()
	return NewPhoneAuthService(nil, users, codes, nil, authTestManager(t), nil, logs, contentsafety.AlwaysPass(), &config.Config{
		App:        config.AppConfig{Env: "test"},
		VerifyCode: config.VerifyCodeConfig{ExpireSeconds: 300, CooldownSeconds: 60},
	})
}

func TestPhoneRegistrationAndLoginFlow(t *testing.T) {
	const (
		phone    = "13800000001"
		code     = "654321"
		password = "unit-test-password"
	)
	codeHash, err := security.HashVerificationCode(phone, SceneRegister, code)
	if err != nil {
		t.Fatalf("HashVerificationCode: %v", err)
	}
	users := &userRepoFake{}
	codes := &codeRepoFake{record: &usermodel.VerificationCode{
		ID: 1, PhoneHash: security.PhoneHash(phone), Scene: SceneRegister,
		CodeHash: codeHash, Status: string(enums.StatusPending), ExpiredAt: time.Now().Add(time.Minute),
	}}
	logs := &authLogCapture{}
	service := newPhoneService(t, users, codes, logs)

	registered, businessErr := service.RegisterPhone(context.Background(), RegisterPhoneInput{
		Phone: phone, Code: code, Password: password, ClientType: ClientH5Web,
	})
	if businessErr != nil {
		t.Fatalf("RegisterPhone: %v", businessErr)
	}
	if !codes.used || users.user == nil || users.user.PasswordHash == nil ||
		!security.CheckPassword(password, *users.user.PasswordHash) {
		t.Fatal("registration did not consume the code or hash the password")
	}
	if _, err := service.jwtManager.Parse(registered.AccessToken, commonjwt.TokenTypeUser); err != nil {
		t.Fatalf("registered token: %v", err)
	}

	loggedIn, businessErr := service.LoginPhone(context.Background(), LoginPhoneInput{
		Phone: phone, Password: password, ClientType: ClientH5Web,
	})
	if businessErr != nil {
		t.Fatalf("LoginPhone: %v", businessErr)
	}
	if loggedIn.User.ID != users.user.ID || len(logs.success) != 2 {
		t.Fatalf("login response/log mismatch: %#v %#v", loggedIn, logs.success)
	}

	serialized, err := json.Marshal(struct {
		Response any
		Logs     any
	}{Response: loggedIn, Logs: logs})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	text := string(serialized)
	for _, forbidden := range []string{password, code, *users.user.PasswordHash, "password_hash", "verificationCode"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("response/logs exposed sensitive value %q", forbidden)
		}
	}
}

func TestPhoneRegistrationAndLoginFailures(t *testing.T) {
	const phone = "13800000002"

	t.Run("wrong verification code", func(t *testing.T) {
		hash, _ := security.HashVerificationCode(phone, SceneRegister, "correct-code")
		service := newPhoneService(t, &userRepoFake{}, &codeRepoFake{record: &usermodel.VerificationCode{
			ID: 2, PhoneHash: security.PhoneHash(phone), Scene: SceneRegister,
			CodeHash: hash, Status: string(enums.StatusPending), ExpiredAt: time.Now().Add(time.Minute),
		}}, &authLogCapture{})
		result, businessErr := service.RegisterPhone(context.Background(), RegisterPhoneInput{
			Phone: phone, Code: "wrong-code", Password: "unit-test-password", ClientType: ClientPCWeb,
		})
		if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeRegisterCodeInvalid {
			t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		hash, _ := security.HashPassword("correct-password")
		phoneHash := security.PhoneHash(phone)
		logs := &authLogCapture{}
		service := newPhoneService(t, &userRepoFake{user: &usermodel.User{
			ID: 9, Phone: stringPointer(phone), PhoneHash: &phoneHash, PasswordHash: &hash,
			PhoneVerified: true, Status: string(enums.StatusActive),
		}}, &codeRepoFake{}, logs)
		result, businessErr := service.LoginPhone(context.Background(), LoginPhoneInput{
			Phone: phone, Password: "wrong-password", ClientType: ClientPCWeb,
		})
		if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeLoginPhonePasswordInvalid ||
			len(logs.failed) != 1 {
			t.Fatalf("unexpected result/error/log: %#v %#v %#v", result, businessErr, logs.failed)
		}
	})

	for _, test := range []struct {
		status string
		code   apperrors.Code
	}{
		{status: string(enums.StatusDisabled), code: apperrors.CodeLoginDisabled},
		{status: string(enums.StatusCancelled), code: apperrors.CodeLoginCancelled},
	} {
		t.Run(test.status, func(t *testing.T) {
			hash, _ := security.HashPassword("correct-password")
			phoneHash := security.PhoneHash(phone)
			service := newPhoneService(t, &userRepoFake{user: &usermodel.User{
				ID: 10, Phone: stringPointer(phone), PhoneHash: &phoneHash, PasswordHash: &hash,
				PhoneVerified: true, Status: test.status,
			}}, &codeRepoFake{}, &authLogCapture{})
			result, businessErr := service.LoginPhone(context.Background(), LoginPhoneInput{
				Phone: phone, Password: "correct-password", ClientType: ClientPCWeb,
			})
			if result != nil || businessErr == nil || businessErr.Code != test.code {
				t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
			}
		})
	}
}

func TestMissingUserIsRejectedWithoutCredentialDetails(t *testing.T) {
	logs := &authLogCapture{}
	service := newPhoneService(t, &userRepoFake{}, &codeRepoFake{}, logs)
	result, businessErr := service.LoginPhone(context.Background(), LoginPhoneInput{
		Phone: "13800000003", Password: "not-recorded", ClientType: ClientPCWeb,
	})
	if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeLoginPhonePasswordInvalid {
		t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
	}
	if len(logs.failed) != 1 || logs.failed[0].DetailJSON != nil {
		t.Fatalf("unexpected failed log: %#v", logs.failed)
	}
}

func stringPointer(value string) *string { return &value }
