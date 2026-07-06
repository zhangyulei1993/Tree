package contentsafety

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"tree/backend/internal/common/config"
)

const (
	msgSecCheckURL    = "https://api.weixin.qq.com/wxa/msg_sec_check"
	accessTokenURL    = "https://api.weixin.qq.com/cgi-bin/token"
	accessTokenSkew   = 2 * time.Minute
	defaultTokenTTL   = 2 * time.Hour
	msgSecVersion     = 2
)

var (
	ErrEmptyContent   = errors.New("msg_sec_check content is required")
	ErrInvalidSuggest = errors.New("msg_sec_check suggest invalid")
	ErrHTTPStatus     = errors.New("msg_sec_check http status error")
)

type Client interface {
	MsgSecCheck(ctx context.Context, openid string, scene int, content string) (Result, error)
}

type WechatClient struct {
	cfg          config.WechatConfig
	httpClient   *http.Client
	tokenMu      sync.Mutex
	cachedToken  string
	tokenExpires time.Time
}

func NewWechatClient(cfg config.WechatConfig) *WechatClient {
	return NewWechatClientWithHTTP(cfg, &http.Client{Timeout: 8 * time.Second})
}

func NewWechatClientWithHTTP(cfg config.WechatConfig, httpClient *http.Client) *WechatClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 8 * time.Second}
	}
	return &WechatClient{cfg: cfg, httpClient: httpClient}
}

func (c *WechatClient) SetAccessTokenForTest(token string, expiresAt time.Time) {
	c.tokenMu.Lock()
	c.cachedToken = token
	c.tokenExpires = expiresAt
	c.tokenMu.Unlock()
}

func (c *WechatClient) MsgSecCheck(ctx context.Context, openid string, scene int, content string) (Result, error) {
	if openid == "" {
		return Result{}, errors.New("openid is required")
	}
	if strings.TrimSpace(content) == "" {
		return Result{}, ErrEmptyContent
	}
	if c.cfg.MiniAppID == "" || c.cfg.MiniAppSecret == "" {
		return Result{}, errors.New("wechat mini app config is required")
	}

	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return Result{}, err
	}

	body, err := json.Marshal(map[string]any{
		"openid":  openid,
		"scene":   scene,
		"version": msgSecVersion,
		"content": content,
	})
	if err != nil {
		return Result{}, err
	}

	endpoint := msgSecCheckURL + "?access_token=" + url.QueryEscape(accessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Result{}, fmt.Errorf("%w: %d", ErrHTTPStatus, resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, err
	}

	var payload struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		Result  struct {
			Suggest string `json:"suggest"`
			Label   int    `json:"label"`
		} `json:"result"`
		TraceID string `json:"trace_id"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return Result{}, err
	}
	if payload.ErrCode != 0 {
		return Result{}, fmt.Errorf("wechat msg_sec_check failed: %d", payload.ErrCode)
	}

	suggest, err := normalizeSuggest(payload.Result.Suggest)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Suggest: suggest,
		Label:   payload.Result.Label,
		TraceID: payload.TraceID,
	}, nil
}

func normalizeSuggest(raw string) (string, error) {
	suggest := strings.TrimSpace(strings.ToLower(raw))
	switch suggest {
	case SuggestPass, SuggestReview, SuggestRisky:
		return suggest, nil
	default:
		return "", ErrInvalidSuggest
	}
}

func (c *WechatClient) getAccessToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	if c.cachedToken != "" && time.Now().Before(c.tokenExpires) {
		token := c.cachedToken
		c.tokenMu.Unlock()
		return token, nil
	}
	c.tokenMu.Unlock()

	values := url.Values{}
	values.Set("grant_type", "client_credential")
	values.Set("appid", c.cfg.MiniAppID)
	values.Set("secret", c.cfg.MiniAppSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, accessTokenURL+"?"+values.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("%w: %d", ErrHTTPStatus, resp.StatusCode)
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.ErrCode != 0 || payload.AccessToken == "" {
		return "", errors.New("wechat access token fetch failed")
	}

	ttl := time.Duration(payload.ExpiresIn) * time.Second
	if ttl <= 0 {
		ttl = defaultTokenTTL
	}
	expiresAt := time.Now().Add(ttl - accessTokenSkew)

	c.tokenMu.Lock()
	c.cachedToken = payload.AccessToken
	c.tokenExpires = expiresAt
	token := c.cachedToken
	c.tokenMu.Unlock()
	return token, nil
}
