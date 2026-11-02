package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App           AppConfig
	MySQL         MySQLConfig
	Redis         RedisConfig
	JWT           JWTConfig
	Log           LogConfig
	AdminSecurity AdminSecurityConfig
	VerifyCode    VerifyCodeConfig
	Wechat        WechatConfig
	WechatPay     WechatPayConfig
	Family        FamilyConfig
	Capacity      CapacityConfig
}

type AppConfig struct {
	Env  string
	Name string
	Port int
}

type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Charset  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	UserSecret               string
	AdminSecret              string
	AccessTokenExpireMinutes int
	RefreshTokenExpireDays   int
}

type LogConfig struct {
	Level string
}

type AdminSecurityConfig struct {
	LockMaxFailures int
	LockMinutes     int
}

type VerifyCodeConfig struct {
	ExpireSeconds   int
	CooldownSeconds int
}

type WechatConfig struct {
	MiniAppID     string
	MiniAppSecret string
	MockEnabled   bool
	Subscribe     WechatSubscribeConfig
}

type WechatSubscribeConfig struct {
	Enabled                bool
	AuditTemplateID        string
	PaymentTemplateID      string
	RefundTemplateID       string
	MemberTemplateID       string
	MemberChangeTemplateID string
	MiniProgramState       string
}

// WechatPayConfig contains only server-side payment credentials. Never expose
// these values to the mini program or write them to operation logs.
type WechatPayConfig struct {
	Enabled                     bool
	MerchantID                  string
	MerchantCertificateSerialNo string
	MerchantPrivateKeyPath      string
	APIv3Key                    string
	PublicKeyID                 string
	PublicKeyPath               string
	NotifyURL                   string
	RefundNotifyURL             string
}

type FamilyConfig struct {
	DissolutionCooldownDays int
}

type CapacityConfig struct {
	OrderTimeoutMinutes int
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	cfg := &Config{
		App: AppConfig{
			Env:  v.GetString("app.env"),
			Name: v.GetString("app.name"),
			Port: v.GetInt("app.port"),
		},
		MySQL: MySQLConfig{
			Host:     v.GetString("mysql.host"),
			Port:     v.GetInt("mysql.port"),
			User:     v.GetString("mysql.user"),
			Password: v.GetString("mysql.password"),
			Database: v.GetString("mysql.database"),
			Charset:  v.GetString("mysql.charset"),
		},
		Redis: RedisConfig{
			Host:     v.GetString("redis.host"),
			Port:     v.GetInt("redis.port"),
			Password: v.GetString("redis.password"),
			DB:       v.GetInt("redis.db"),
		},
		JWT: JWTConfig{
			UserSecret:               v.GetString("jwt.user_secret"),
			AdminSecret:              v.GetString("jwt.admin_secret"),
			AccessTokenExpireMinutes: v.GetInt("jwt.access_token_expire_minutes"),
			RefreshTokenExpireDays:   v.GetInt("jwt.refresh_token_expire_days"),
		},
		Log: LogConfig{
			Level: v.GetString("log.level"),
		},
		AdminSecurity: AdminSecurityConfig{
			LockMaxFailures: v.GetInt("admin_security.lock_max_failures"),
			LockMinutes:     v.GetInt("admin_security.lock_minutes"),
		},
		VerifyCode: VerifyCodeConfig{
			ExpireSeconds:   v.GetInt("verify_code.expire_seconds"),
			CooldownSeconds: v.GetInt("verify_code.cooldown_seconds"),
		},
		Wechat: WechatConfig{
			MiniAppID:     v.GetString("wechat.mini_app_id"),
			MiniAppSecret: v.GetString("wechat.mini_app_secret"),
			MockEnabled:   v.GetBool("wechat.mock_enabled"),
			Subscribe: WechatSubscribeConfig{
				Enabled:                v.GetBool("wechat.subscribe.enabled"),
				AuditTemplateID:        v.GetString("wechat.subscribe.audit_template_id"),
				PaymentTemplateID:      v.GetString("wechat.subscribe.payment_template_id"),
				RefundTemplateID:       v.GetString("wechat.subscribe.refund_template_id"),
				MemberTemplateID:       v.GetString("wechat.subscribe.member_template_id"),
				MemberChangeTemplateID: v.GetString("wechat.subscribe.member_change_template_id"),
				MiniProgramState:       v.GetString("wechat.subscribe.miniprogram_state"),
			},
		},
		WechatPay: WechatPayConfig{
			Enabled:                     v.GetBool("wechat_pay.enabled"),
			MerchantID:                  v.GetString("wechat_pay.merchant_id"),
			MerchantCertificateSerialNo: v.GetString("wechat_pay.merchant_certificate_serial_no"),
			MerchantPrivateKeyPath:      v.GetString("wechat_pay.merchant_private_key_path"),
			APIv3Key:                    v.GetString("wechat_pay.api_v3_key"),
			PublicKeyID:                 v.GetString("wechat_pay.public_key_id"),
			PublicKeyPath:               v.GetString("wechat_pay.public_key_path"),
			NotifyURL:                   v.GetString("wechat_pay.notify_url"),
			RefundNotifyURL:             v.GetString("wechat_pay.refund_notify_url"),
		},
		Family: FamilyConfig{
			DissolutionCooldownDays: v.GetInt("family.dissolution_cooldown_days"),
		},
		Capacity: CapacityConfig{OrderTimeoutMinutes: v.GetInt("capacity.order_timeout_minutes")},
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	env := strings.ToLower(strings.TrimSpace(c.App.Env))
	if c.WechatPay.Enabled {
		if strings.TrimSpace(c.Wechat.MiniAppID) == "" ||
			strings.TrimSpace(c.WechatPay.MerchantID) == "" ||
			strings.TrimSpace(c.WechatPay.MerchantCertificateSerialNo) == "" ||
			strings.TrimSpace(c.WechatPay.MerchantPrivateKeyPath) == "" ||
			len(strings.TrimSpace(c.WechatPay.APIv3Key)) != 32 ||
			strings.TrimSpace(c.WechatPay.NotifyURL) == "" ||
			strings.TrimSpace(c.WechatPay.RefundNotifyURL) == "" {
			return fmt.Errorf("wechat_pay enabled but required payment configuration is incomplete")
		}
	}
	if env == "staging" || env == "production" || env == "prod" {
		if c.Wechat.MockEnabled {
			return fmt.Errorf("wechat mock_enabled must not be enabled when app.env is %s", c.App.Env)
		}
		if isDefaultSecret(c.JWT.UserSecret, "change_me_user_secret") || isDefaultSecret(c.JWT.AdminSecret, "change_me_admin_secret") {
			return fmt.Errorf("production jwt secrets must be explicitly configured")
		}
		if strings.TrimSpace(c.Wechat.MiniAppID) == "" || strings.TrimSpace(c.Wechat.MiniAppSecret) == "" {
			return fmt.Errorf("production wechat mini program credentials must be explicitly configured")
		}
		if c.Capacity.OrderTimeoutMinutes <= 0 {
			return fmt.Errorf("production capacity order timeout must be positive")
		}
	}
	return nil
}

func isDefaultSecret(value, fallback string) bool {
	value = strings.TrimSpace(value)
	return value == "" || value == fallback
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", "dev")
	v.SetDefault("app.name", "Tree")
	v.SetDefault("app.port", 8080)
	v.SetDefault("mysql.host", "mysql")
	v.SetDefault("mysql.port", 3306)
	v.SetDefault("mysql.user", "tree_user")
	v.SetDefault("mysql.password", "tree_pass")
	v.SetDefault("mysql.database", "tree_platform")
	v.SetDefault("mysql.charset", "utf8mb4")
	v.SetDefault("redis.host", "redis")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("jwt.user_secret", "change_me_user_secret")
	v.SetDefault("jwt.admin_secret", "change_me_admin_secret")
	v.SetDefault("jwt.access_token_expire_minutes", 120)
	v.SetDefault("jwt.refresh_token_expire_days", 7)
	v.SetDefault("admin_security.lock_max_failures", 5)
	v.SetDefault("admin_security.lock_minutes", 30)
	v.SetDefault("verify_code.expire_seconds", 300)
	v.SetDefault("verify_code.cooldown_seconds", 60)
	v.SetDefault("wechat.mini_app_id", "")
	v.SetDefault("wechat.mini_app_secret", "")
	v.SetDefault("wechat.mock_enabled", false)
	v.SetDefault("wechat.subscribe.enabled", false)
	v.SetDefault("wechat.subscribe.miniprogram_state", "formal")
	v.SetDefault("family.dissolution_cooldown_days", 7)
	v.SetDefault("capacity.order_timeout_minutes", 30)
	v.SetDefault("log.level", "debug")
}
