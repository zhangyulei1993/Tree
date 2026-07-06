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
