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
	Family        FamilyConfig
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
}

type FamilyConfig struct {
	DissolutionCooldownDays int
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
		},
		Family: FamilyConfig{
			DissolutionCooldownDays: v.GetInt("family.dissolution_cooldown_days"),
		},
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
	if !c.Wechat.MockEnabled {
		return nil
	}
	switch env {
	case "staging", "production", "prod":
		return fmt.Errorf("wechat mock_enabled must not be enabled when app.env is %s", c.App.Env)
	default:
		return nil
	}
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
	v.SetDefault("family.dissolution_cooldown_days", 7)
	v.SetDefault("log.level", "debug")
}
