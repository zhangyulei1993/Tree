package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"tree/backend/internal/admin/model"
	"tree/backend/internal/common/config"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	commonjwt "tree/backend/internal/common/jwt"
	"tree/backend/internal/common/security"
	operationlog "tree/backend/internal/operationlog/service"
)

type adminRepoFake struct {
	admin        *model.AdminUser
	findErr      error
	successCalls int
	failureCalls int
}

func (r *adminRepoFake) FindByUsername(context.Context, string) (*model.AdminUser, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	copy := *r.admin
	return &copy, nil
}
func (r *adminRepoFake) FindByID(context.Context, uint64) (*model.AdminUser, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	copy := *r.admin
	return &copy, nil
}
func (r *adminRepoFake) RecordLoginSuccess(context.Context, uint64, string, time.Time) error {
	r.successCalls++
	return nil
}
func (r *adminRepoFake) RecordLoginFailure(context.Context, uint64, int, *time.Time) error {
	r.failureCalls++
	return nil
}
func (r *adminRepoFake) UpdatePassword(context.Context, uint64, string, time.Time) error {
	return nil
}

type logCapture struct {
	success []operationlog.WriteInput
	failed  []operationlog.WriteInput
}

func (l *logCapture) WriteSuccess(_ context.Context, input operationlog.WriteInput) error {
	l.success = append(l.success, input)
	return nil
}
func (l *logCapture) WriteFailed(_ context.Context, input operationlog.WriteInput) error {
	l.failed = append(l.failed, input)
	return nil
}

func adminTestManager(t *testing.T) *commonjwt.Manager {
	t.Helper()
	manager, err := commonjwt.NewManager(config.JWTConfig{
		UserSecret: "admin-test-user-secret", AdminSecret: "admin-test-admin-secret",
		AccessTokenExpireMinutes: 5,
	}, "TreeTest")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return manager
}

func activeAdmin(t *testing.T, role string) *model.AdminUser {
	t.Helper()
	hash, err := security.HashPassword("unit-test-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	return &model.AdminUser{
		ID: 7, Username: "unit_admin", PasswordHash: hash,
		Role: role, Status: string(enums.StatusActive),
	}
}

func TestAdminLoginSuccessAndRoleClaims(t *testing.T) {
	for _, role := range []string{
		string(enums.AdminRoleRootAdmin),
		string(enums.AdminRolePlatformAdmin),
	} {
		t.Run(role, func(t *testing.T) {
			repo := &adminRepoFake{admin: activeAdmin(t, role)}
			logs := &logCapture{}
			service := NewAdminAuthService(repo, adminTestManager(t), nil, logs, config.AdminSecurityConfig{})

			result, businessErr := service.Login(context.Background(), LoginInput{
				Username: "unit_admin", Password: "unit-test-password",
			})
			if businessErr != nil {
				t.Fatalf("Login: %v", businessErr)
			}
			if result.Admin.Role != role || repo.successCalls != 1 || len(logs.success) != 1 {
				t.Fatalf("role/login/log mismatch: %#v", result)
			}
			claims, err := service.jwtManager.Parse(result.AccessToken, commonjwt.TokenTypeAdmin)
			if err != nil || claims.Role != role {
				t.Fatalf("admin token claims = %#v, %v", claims, err)
			}
		})
	}
}

func TestAdminLoginRejectsWrongPasswordDisabledAndLocked(t *testing.T) {
	t.Run("wrong password", func(t *testing.T) {
		repo := &adminRepoFake{admin: activeAdmin(t, string(enums.AdminRolePlatformAdmin))}
		logs := &logCapture{}
		service := NewAdminAuthService(repo, adminTestManager(t), nil, logs, config.AdminSecurityConfig{})
		result, businessErr := service.Login(context.Background(), LoginInput{
			Username: "unit_admin", Password: "wrong-password",
		})
		if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeAdminUsernameOrPasswordInvalid {
			t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
		}
		if repo.failureCalls != 1 || len(logs.failed) != 1 {
			t.Fatal("failed login was not recorded")
		}
	})

	t.Run("disabled", func(t *testing.T) {
		admin := activeAdmin(t, string(enums.AdminRoleRootAdmin))
		admin.Status = string(enums.StatusDisabled)
		logs := &logCapture{}
		result, businessErr := NewAdminAuthService(
			&adminRepoFake{admin: admin}, adminTestManager(t), nil, logs, config.AdminSecurityConfig{},
		).Login(context.Background(), LoginInput{Username: admin.Username, Password: "unit-test-password"})
		if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeAdminDisabled || len(logs.failed) != 1 {
			t.Fatalf("unexpected result/error/log: %#v %#v %#v", result, businessErr, logs.failed)
		}
	})

	t.Run("locked", func(t *testing.T) {
		admin := activeAdmin(t, string(enums.AdminRolePlatformAdmin))
		lockedUntil := time.Now().Add(time.Hour)
		admin.LockedUntil = &lockedUntil
		result, businessErr := NewAdminAuthService(
			&adminRepoFake{admin: admin}, adminTestManager(t), nil, &logCapture{}, config.AdminSecurityConfig{},
		).Login(context.Background(), LoginInput{Username: admin.Username, Password: "unit-test-password"})
		if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeAdminLocked {
			t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
		}
	})

	t.Run("unknown username", func(t *testing.T) {
		result, businessErr := NewAdminAuthService(
			&adminRepoFake{findErr: errors.New("not found")}, adminTestManager(t), nil, &logCapture{}, config.AdminSecurityConfig{},
		).Login(context.Background(), LoginInput{Username: "missing", Password: "irrelevant"})
		if result != nil || businessErr == nil || businessErr.Code != apperrors.CodeAdminUsernameOrPasswordInvalid {
			t.Fatalf("unexpected result/error: %#v %#v", result, businessErr)
		}
	})
}
