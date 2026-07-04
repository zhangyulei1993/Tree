package app

import (
	"os"
	"strings"
	"testing"
)

func TestPublicFamilyShowcaseRouteIsRegistered(t *testing.T) {
	source, err := os.ReadFile("router.go")
	if err != nil {
		t.Fatalf("read router.go: %v", err)
	}
	body := string(source)
	if !strings.Contains(body, `api.GET("/public/families", handler.ListPublicFamilyShowcase)`) {
		t.Fatal("public family showcase route must be registered")
	}
	if strings.Contains(body, `handler.PublicFamiliesSearchOffline`) {
		t.Fatal("offline search handler must not be registered")
	}
	required := []string{
		`api.GET("/families/:familyId/public", handler.PublicDetail)`,
		`api.GET("/public/families/:familyId/tree", treeHandler.PublicTree)`,
	}
	for _, snippet := range required {
		if !strings.Contains(body, snippet) {
			t.Fatalf("missing required public family route: %s", snippet)
		}
	}
}
