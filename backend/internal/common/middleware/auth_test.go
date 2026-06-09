package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"tree/backend/internal/common/config"
	commonjwt "tree/backend/internal/common/jwt"
)

func middlewareTestManager(t *testing.T) *commonjwt.Manager {
	t.Helper()
	manager, err := commonjwt.NewManager(config.JWTConfig{
		UserSecret: "middleware-user-secret", AdminSecret: "middleware-admin-secret",
		AccessTokenExpireMinutes: 5,
	}, "TreeTest")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return manager
}

func TestAuthMiddlewareRejectsMissingAndWrongTokenTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := middlewareTestManager(t)
	userToken, _ := manager.GenerateAccessToken(10, commonjwt.TokenTypeUser, "", "user-jti")
	adminToken, _ := manager.GenerateAccessToken(20, commonjwt.TokenTypeAdmin, "ROOT_ADMIN", "admin-jti")

	tests := []struct {
		name       string
		middleware gin.HandlerFunc
		path       string
		token      string
	}{
		{name: "admin route without token", middleware: AdminAuth(manager, nil), path: "/api/admin/me"},
		{name: "user route without token", middleware: UserAuth(manager, nil), path: "/api/families/1"},
		{name: "user cannot access admin", middleware: AdminAuth(manager, nil), path: "/api/admin/me", token: userToken},
		{name: "admin cannot access user write", middleware: UserAuth(manager, nil), path: "/api/families/1/members", token: adminToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, tt.path, nil)
			if tt.token != "" {
				ctx.Request.Header.Set("Authorization", "Bearer "+tt.token)
			}
			tt.middleware(ctx)
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestAuthMiddlewareInjectsOnlyExpectedIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := middlewareTestManager(t)
	userToken, _ := manager.GenerateAccessToken(10, commonjwt.TokenTypeUser, "", "user-jti")
	adminToken, _ := manager.GenerateAccessToken(20, commonjwt.TokenTypeAdmin, "PLATFORM_ADMIN", "admin-jti")

	t.Run("user context", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/families", nil)
		ctx.Request.Header.Set("Authorization", "Bearer "+userToken)
		UserAuth(manager, nil)(ctx)
		if id, err := CurrentUserID(ctx); err != nil || id != 10 {
			t.Fatalf("CurrentUserID = %d, %v", id, err)
		}
		if _, err := CurrentAdminID(ctx); err == nil {
			t.Fatal("USER token injected admin identity")
		}
	})

	t.Run("admin context", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/admin/me", nil)
		ctx.Request.Header.Set("Authorization", "Bearer "+adminToken)
		AdminAuth(manager, nil)(ctx)
		if id, err := CurrentAdminID(ctx); err != nil || id != 20 {
			t.Fatalf("CurrentAdminID = %d, %v", id, err)
		}
		if role, err := CurrentAdminRole(ctx); err != nil || role != "PLATFORM_ADMIN" {
			t.Fatalf("CurrentAdminRole = %q, %v", role, err)
		}
		if _, err := CurrentUserID(ctx); err == nil {
			t.Fatal("ADMIN token injected user identity")
		}
	})
}
