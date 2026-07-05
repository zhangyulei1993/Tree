package service

import (
	"testing"

	contentenum "tree/backend/internal/content/enum"
)

func stringPointer(value string) *string {
	return &value
}

func TestValidateInternalArticleContent(t *testing.T) {
	body, externalURL, err := validateArticleContent(contentenum.ArticleTypeInternal, "  正文  ", stringPointer("https://mp.weixin.qq.com/s/demo"))
	if err != nil {
		t.Fatalf("expected valid internal article, got %v", err)
	}
	if body != "正文" {
		t.Fatalf("expected trimmed body, got %q", body)
	}
	if externalURL != nil {
		t.Fatal("internal article must not retain external URL")
	}
}

func TestValidateWechatOfficialArticleContent(t *testing.T) {
	rawURL := "https://mp.weixin.qq.com/s/example"
	body, externalURL, err := validateArticleContent(contentenum.ArticleTypeWechatOfficial, "ignored", &rawURL)
	if err != nil {
		t.Fatalf("expected valid WeChat article, got %v", err)
	}
	if body != "" {
		t.Fatalf("WeChat article must not retain local body, got %q", body)
	}
	if externalURL == nil || *externalURL != rawURL {
		t.Fatalf("expected normalized external URL, got %#v", externalURL)
	}
}

func TestValidateWechatOfficialArticleRejectsUntrustedURL(t *testing.T) {
	cases := []string{
		"http://mp.weixin.qq.com/s/example",
		"https://example.com/s/example",
		"https://mp.weixin.qq.com.evil.example/s/example",
		"https://mp.weixin.qq.com:444/s/example",
		"https://mp.weixin.qq.com/",
	}
	for _, rawURL := range cases {
		if _, _, err := validateArticleContent(contentenum.ArticleTypeWechatOfficial, "", &rawURL); err == nil {
			t.Fatalf("expected URL to be rejected: %s", rawURL)
		}
	}
}
