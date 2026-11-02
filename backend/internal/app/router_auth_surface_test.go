package app

import (
	"os"
	"strings"
	"testing"
)

func TestPersonalAuthRouteSurface(t *testing.T) {
	source, err := os.ReadFile("router.go")
	if err != nil {
		t.Fatalf("read router.go: %v", err)
	}
	body := string(source)
	allowed := []string{
		`auth.POST("/wechat-mini/login"`,
		`auth.POST("/wechat-mini/phone-login"`,
		`protected.POST("/logout"`,
		`protected.POST("/wechat-mini/bind-phone"`,
		`protected.POST("/cancel-account/wechat-reauth"`,
	}
	for _, snippet := range allowed {
		if !strings.Contains(body, snippet) {
			t.Fatalf("missing allowed auth route registration: %s", snippet)
		}
	}
	removed := []string{
		`auth.POST("/send-code"`,
		`auth.POST("/register-phone"`,
		`protected.POST("/change-phone"`,
		`protected.POST("/cancel-account"`,
		`protected.POST("/wechat-mini/bind-phone-credential"`,
		`protected.POST("/wechat-mini/change-phone-login-password"`,
	}
	for _, snippet := range removed {
		if strings.Contains(body, snippet) {
			t.Fatalf("legacy auth route must not be registered: %s", snippet)
		}
	}
}
