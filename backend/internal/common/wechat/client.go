package wechat

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"tree/backend/internal/common/config"
	"tree/backend/internal/common/security"
)

const code2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"

type Code2SessionResult struct {
	OpenID     string
	UnionID    string
	SessionKey string
}

type Client interface {
	Code2Session(ctx context.Context, code string) (*Code2SessionResult, error)
}

type MiniProgramClient struct {
	cfg        config.WechatConfig
	httpClient *http.Client
}

func NewMiniProgramClient(cfg config.WechatConfig) *MiniProgramClient {
	return &MiniProgramClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *MiniProgramClient) Code2Session(ctx context.Context, code string) (*Code2SessionResult, error) {
	if code == "" {
		return nil, errors.New("wechat code is required")
	}
	if c.cfg.MockEnabled {
		hash := security.HashPlain(code)
		return &Code2SessionResult{
			OpenID:     "mock_openid_" + hash[:24],
			UnionID:    "mock_unionid_" + hash[24:48],
			SessionKey: "mock_session_" + hash[48:],
		}, nil
	}
	if c.cfg.MiniAppID == "" || c.cfg.MiniAppSecret == "" {
		return nil, errors.New("wechat mini app config is required")
	}

	values := url.Values{}
	values.Set("appid", c.cfg.MiniAppID)
	values.Set("secret", c.cfg.MiniAppSecret)
	values.Set("js_code", code)
	values.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, code2SessionURL+"?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		OpenID     string `json:"openid"`
		UnionID    string `json:"unionid"`
		SessionKey string `json:"session_key"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.ErrCode != 0 || payload.OpenID == "" {
		return nil, errors.New("wechat code2session failed")
	}

	return &Code2SessionResult{
		OpenID:     payload.OpenID,
		UnionID:    payload.UnionID,
		SessionKey: payload.SessionKey,
	}, nil
}
