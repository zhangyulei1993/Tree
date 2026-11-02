package app

import (
	"os"
	"strings"
	"testing"
)

func TestVisitorMessageRoutesAreNotRegistered(t *testing.T) {
	source, err := os.ReadFile("router.go")
	if err != nil {
		t.Fatalf("read router.go: %v", err)
	}
	body := string(source)
	removed := []string{
		`api.POST("/public/families/:familyId/visitor-messages"`,
		`api.GET("/public/families/:familyId/visitor-messages"`,
		`admin.GET("/visitor-messages"`,
		`admin.POST("/visitor-messages/:messageId/approve"`,
		`admin.POST("/visitor-messages/:messageId/reject"`,
		`admin.DELETE("/visitor-messages/:messageId"`,
	}
	for _, snippet := range removed {
		if strings.Contains(body, snippet) {
			t.Fatalf("visitor message route must not be registered: %s", snippet)
		}
	}
	if strings.Contains(body, "messagehandler.NewHandler") {
		t.Fatal("buildFamilyCore must not create visitor message handler")
	}
}
