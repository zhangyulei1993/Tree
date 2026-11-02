package config

import (
	"strings"
	"testing"
)

func loadWechatMockForTest(t *testing.T, envVars map[string]string) (*Config, error) {
	t.Helper()
	for key, value := range envVars {
		t.Setenv(key, value)
	}
	return Load()
}

func TestWechatMockEnabledDefaultFalse(t *testing.T) {
	cfg, err := loadWechatMockForTest(t, map[string]string{
		"WECHAT_MOCK_ENABLED": "",
		"APP_ENV":             "test",
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Wechat.MockEnabled {
		t.Fatal("expected MockEnabled=false when WECHAT_MOCK_ENABLED is unset")
	}
}

func TestWechatMockEnabledExplicitTrue(t *testing.T) {
	cfg, err := loadWechatMockForTest(t, map[string]string{
		"WECHAT_MOCK_ENABLED": "true",
		"APP_ENV":             "dev",
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Wechat.MockEnabled {
		t.Fatal("expected MockEnabled=true when WECHAT_MOCK_ENABLED=true")
	}
}

func TestWechatMockEnabledExplicitFalse(t *testing.T) {
	cfg, err := loadWechatMockForTest(t, map[string]string{
		"WECHAT_MOCK_ENABLED": "false",
		"APP_ENV":             "test",
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Wechat.MockEnabled {
		t.Fatal("expected MockEnabled=false when WECHAT_MOCK_ENABLED=false")
	}
}

func TestWechatMockEnabledRejectedInStagingAndProduction(t *testing.T) {
	for _, env := range []string{"staging", "production", "prod"} {
		t.Run(env, func(t *testing.T) {
			_, err := loadWechatMockForTest(t, map[string]string{
				"WECHAT_MOCK_ENABLED": "true",
				"APP_ENV":             env,
			})
			if err == nil {
				t.Fatalf("expected Load to fail for app.env=%s with mock enabled", env)
			}
			if !strings.Contains(err.Error(), "mock_enabled") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestWechatMockEnabledAllowedInDevAndTest(t *testing.T) {
	for _, env := range []string{"dev", "test"} {
		t.Run(env, func(t *testing.T) {
			cfg, err := loadWechatMockForTest(t, map[string]string{
				"WECHAT_MOCK_ENABLED": "true",
				"APP_ENV":             env,
			})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if !cfg.Wechat.MockEnabled {
				t.Fatal("expected MockEnabled=true")
			}
		})
	}
}

func TestProductionRejectsDefaultJWTSecrets(t *testing.T) {
	for _, env := range []string{"staging", "production", "prod"} {
		t.Run(env, func(t *testing.T) {
			t.Setenv("APP_ENV", env)
			t.Setenv("JWT_USER_SECRET", "change_me_user_secret")
			t.Setenv("JWT_ADMIN_SECRET", "change_me_admin_secret")
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), "jwt secrets") {
				t.Fatalf("expected default JWT secrets to be rejected, got %v", err)
			}
		})
	}
}

func TestProductionAcceptsExplicitJWTSecrets(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_USER_SECRET", "user-secret-for-production")
	t.Setenv("JWT_ADMIN_SECRET", "admin-secret-for-production")
	t.Setenv("WECHAT_MINI_APP_ID", "wx-production-test")
	t.Setenv("WECHAT_MINI_APP_SECRET", "production-test-secret")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.JWT.UserSecret == "" || cfg.JWT.AdminSecret == "" {
		t.Fatal("expected explicit JWT secrets")
	}
}
