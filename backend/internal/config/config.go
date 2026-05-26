package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBCharset  string

	JWTSecret string
	UploadDir string
	StaticURL string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		AppPort: getenv("APP_PORT", "8080"),

		DBHost:     getenv("DB_HOST", "127.0.0.1"),
		DBPort:     getenv("DB_PORT", "3306"),
		DBUser:     getenv("DB_USER", "root"),
		DBPassword: getenv("DB_PASSWORD", ""),
		DBName:     getenv("DB_NAME", "tree_dev"),
		DBCharset:  getenv("DB_CHARSET", "utf8mb4"),

		JWTSecret: getenv("JWT_SECRET", "tree-admin-secret"),
		UploadDir: getenv("UPLOAD_DIR", "uploads"),
		StaticURL: getenv("STATIC_URL", "http://localhost:8080/uploads"),
	}
}

func (c *Config) Addr() string {
	if strings.HasPrefix(c.AppPort, ":") {
		return c.AppPort
	}
	return ":" + c.AppPort
}

func getenv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
