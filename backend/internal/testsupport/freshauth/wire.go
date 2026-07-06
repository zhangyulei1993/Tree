package freshauth

import (
	"testing"

	"gorm.io/gorm"

	authrepo "tree/backend/internal/auth/repository"
	authservice "tree/backend/internal/auth/service"
	"tree/backend/internal/common/config"
	"tree/backend/internal/common/contentsafety"
	commonjwt "tree/backend/internal/common/jwt"
	"tree/backend/internal/testsupport/freshquota"
)

func testJWTManager(t *testing.T) *commonjwt.Manager {
	t.Helper()
	manager, err := commonjwt.NewManager(config.JWTConfig{
		UserSecret: "fresh-quota-user-secret", AdminSecret: "fresh-quota-admin-secret",
		AccessTokenExpireMinutes: 30,
	}, "TreeFreshQuotaTest")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return manager
}

func testConfig() *config.Config {
	return &config.Config{
		App:        config.AppConfig{Env: "test"},
		Wechat:     config.WechatConfig{MiniAppID: "fresh-quota-test-app"},
		VerifyCode: config.VerifyCodeConfig{ExpireSeconds: 300, CooldownSeconds: 60},
	}
}

// NewAuthService wires real repositories with FakeWechatMiniClient (no external WeChat API).
func NewAuthService(t *testing.T, tx *gorm.DB, wechatClient *freshquota.FakeWechatMiniClient) *authservice.PhoneAuthService {
	t.Helper()
	return authservice.NewPhoneAuthService(
		tx,
		authrepo.NewGormUserRepository(tx),
		authrepo.NewGormVerificationCodeRepository(tx),
		wechatClient,
		testJWTManager(t),
		nil,
		nil,
		contentsafety.AlwaysPass(),
		testConfig(),
	)
}

// NewAuthServiceWithJWT returns auth service and JWT manager for token refresh assertions.
func NewAuthServiceWithJWT(t *testing.T, tx *gorm.DB, wechatClient *freshquota.FakeWechatMiniClient) (*authservice.PhoneAuthService, *commonjwt.Manager) {
	t.Helper()
	manager := testJWTManager(t)
	return authservice.NewPhoneAuthService(
		tx,
		authrepo.NewGormUserRepository(tx),
		authrepo.NewGormVerificationCodeRepository(tx),
		wechatClient,
		manager,
		nil,
		nil,
		contentsafety.AlwaysPass(),
		testConfig(),
	), manager
}
