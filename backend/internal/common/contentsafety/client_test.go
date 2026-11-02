package contentsafety

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"tree/backend/internal/common/config"
)

type wechatHostRewriter struct {
	target *url.URL
	inner  http.RoundTripper
}

func (r *wechatHostRewriter) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Host, "api.weixin.qq.com") {
		req = req.Clone(req.Context())
		req.URL.Scheme = r.target.Scheme
		req.URL.Host = r.target.Host
	}
	if r.inner == nil {
		r.inner = http.DefaultTransport
	}
	return r.inner.RoundTrip(req)
}

func newTestWechatClient(t *testing.T, handler http.HandlerFunc) *WechatClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	httpClient := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &wechatHostRewriter{target: target},
	}
	client := NewWechatClientWithHTTP(config.WechatConfig{
		MiniAppID:     "test-app-id",
		MiniAppSecret: "test-app-secret",
	}, httpClient)
	client.SetAccessTokenForTest("cached-test-token", time.Now().Add(time.Hour))
	return client
}

func msgSecCheckHandler(t *testing.T, suggest string, extra func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "token") {
			_, _ = w.Write([]byte(`{"access_token":"token-from-server","expires_in":7200}`))
			return
		}
		if extra != nil {
			extra(w, r)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		for _, key := range []string{"version", "scene", "openid", "content"} {
			if _, ok := payload[key]; !ok {
				t.Fatalf("request missing %s: %s", key, string(body))
			}
		}
		if payload["version"].(float64) != float64(msgSecVersion) {
			t.Fatalf("unexpected version: %#v", payload["version"])
		}
		resp := map[string]any{
			"errcode":  0,
			"result":   map[string]any{"suggest": suggest, "label": 100},
			"trace_id": "trace-test-id",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func TestWechatClientMsgSecCheckPass(t *testing.T) {
	client := newTestWechatClient(t, msgSecCheckHandler(t, SuggestPass, nil))
	result, err := client.MsgSecCheck(context.Background(), "user-openid", 1, "正常昵称")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Suggest != SuggestPass || result.TraceID != "trace-test-id" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestWechatClientMsgSecCheckReviewAndRisky(t *testing.T) {
	for _, suggest := range []string{SuggestReview, SuggestRisky} {
		client := newTestWechatClient(t, msgSecCheckHandler(t, suggest, nil))
		result, err := client.MsgSecCheck(context.Background(), "user-openid", 2, "待审内容")
		if err != nil {
			t.Fatalf("%s: %v", suggest, err)
		}
		if result.Suggest != suggest {
			t.Fatalf("unexpected suggest: %#v", result)
		}
	}
}

func TestWechatClientMsgSecCheckErrcodeNonZero(t *testing.T) {
	client := newTestWechatClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "token") {
			_, _ = w.Write([]byte(`{"access_token":"token","expires_in":7200}`))
			return
		}
		_, _ = w.Write([]byte(`{"errcode":87014,"errmsg":"risky content"}`))
	})
	_, err := client.MsgSecCheck(context.Background(), "user-openid", 1, "违规")
	if err == nil || !strings.Contains(err.Error(), "87014") {
		t.Fatalf("expected errcode error, got %v", err)
	}
}

func TestWechatClientMsgSecCheckRefreshesExpiredAccessToken(t *testing.T) {
	var tokenRequests int
	var checkRequests int
	client := newTestWechatClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "token") {
			tokenRequests++
			_, _ = w.Write([]byte(`{"access_token":"fresh-token","expires_in":7200}`))
			return
		}
		checkRequests++
		if checkRequests == 1 {
			if got := r.URL.Query().Get("access_token"); got != "cached-test-token" {
				t.Fatalf("first check should use cached token, got %q", got)
			}
			_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid credential"}`))
			return
		}
		if got := r.URL.Query().Get("access_token"); got != "fresh-token" {
			t.Fatalf("retry should use refreshed token, got %q", got)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"result":{"suggest":"pass","label":100},"trace_id":"retry-trace"}`))
	})

	result, err := client.MsgSecCheck(context.Background(), "user-openid", 1, "正常昵称")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Suggest != SuggestPass || result.TraceID != "retry-trace" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if tokenRequests != 1 || checkRequests != 2 {
		t.Fatalf("expected one token refresh and two checks, got token=%d check=%d", tokenRequests, checkRequests)
	}
}

func TestWechatClientMsgSecCheckHTTPNon2xx(t *testing.T) {
	client := newTestWechatClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "token") {
			_, _ = w.Write([]byte(`{"access_token":"token","expires_in":7200}`))
			return
		}
		w.WriteHeader(http.StatusBadGateway)
	})
	_, err := client.MsgSecCheck(context.Background(), "user-openid", 1, "内容")
	if !errors.Is(err, ErrHTTPStatus) {
		t.Fatalf("expected ErrHTTPStatus, got %v", err)
	}
}

func TestWechatClientMsgSecCheckSuggestMissing(t *testing.T) {
	client := newTestWechatClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "token") {
			_, _ = w.Write([]byte(`{"access_token":"token","expires_in":7200}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"errcode":0,"result":{},"trace_id":"t1"}`))
	})
	_, err := client.MsgSecCheck(context.Background(), "user-openid", 1, "内容")
	if !errors.Is(err, ErrInvalidSuggest) {
		t.Fatalf("expected ErrInvalidSuggest, got %v", err)
	}
}

func TestWechatClientMsgSecCheckSuggestUnknown(t *testing.T) {
	client := newTestWechatClient(t, msgSecCheckHandler(t, "maybe-pass", nil))
	_, err := client.MsgSecCheck(context.Background(), "user-openid", 1, "内容")
	if !errors.Is(err, ErrInvalidSuggest) {
		t.Fatalf("expected ErrInvalidSuggest, got %v", err)
	}
}

func TestWechatClientMsgSecCheckInvalidJSON(t *testing.T) {
	client := newTestWechatClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "token") {
			_, _ = w.Write([]byte(`{"access_token":"token","expires_in":7200}`))
			return
		}
		_, _ = w.Write([]byte(`not-json`))
	})
	_, err := client.MsgSecCheck(context.Background(), "user-openid", 1, "内容")
	if err == nil {
		t.Fatal("expected json error")
	}
}

func TestWechatClientMsgSecCheckNetworkError(t *testing.T) {
	target, _ := url.Parse("http://127.0.0.1:1")
	httpClient := &http.Client{
		Timeout:   200 * time.Millisecond,
		Transport: &wechatHostRewriter{target: target},
	}
	client := NewWechatClientWithHTTP(config.WechatConfig{
		MiniAppID: "app", MiniAppSecret: "secret",
	}, httpClient)
	client.SetAccessTokenForTest("token", time.Now().Add(time.Hour))
	_, err := client.MsgSecCheck(context.Background(), "user-openid", 1, "内容")
	if err == nil {
		t.Fatal("expected network error")
	}
}

func TestWechatClientDoesNotExposeSensitiveValuesInErrors(t *testing.T) {
	const token = "super-secret-access-token"
	const openid = "wx-sensitive-openid"
	client := newTestWechatClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "access_token") && strings.Contains(r.URL.RawQuery, token) {
			t.Fatal("access token leaked in request url path")
		}
		if strings.Contains(r.URL.Path, "token") {
			_, _ = w.Write([]byte(`{"access_token":"` + token + `","expires_in":7200}`))
			return
		}
		_, _ = w.Write([]byte(`{"errcode":0,"result":{"suggest":"pass","label":1}}`))
	})
	result, err := client.MsgSecCheck(context.Background(), openid, 2, "检测文本")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Suggest != SuggestPass {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestWechatClientEmptyContentRejected(t *testing.T) {
	client := newTestWechatClient(t, msgSecCheckHandler(t, SuggestPass, nil))
	_, err := client.MsgSecCheck(context.Background(), "user-openid", 1, "   ")
	if !errors.Is(err, ErrEmptyContent) {
		t.Fatalf("expected ErrEmptyContent, got %v", err)
	}
}

func TestWechatClientTokenHTTPNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	client := NewWechatClientWithHTTP(config.WechatConfig{
		MiniAppID: "app", MiniAppSecret: "secret",
	}, &http.Client{Transport: &wechatHostRewriter{target: target}, Timeout: 3 * time.Second})
	_, err := client.MsgSecCheck(context.Background(), "user-openid", 1, "内容")
	if !errors.Is(err, ErrHTTPStatus) {
		t.Fatalf("expected ErrHTTPStatus from token fetch, got %v", err)
	}
}
