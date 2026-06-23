package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"tree/backend/internal/common/config"
	"tree/backend/internal/common/security"
)

const (
	code2SessionURL   = "https://api.weixin.qq.com/sns/jscode2session"
	accessTokenURL    = "https://api.weixin.qq.com/cgi-bin/token"
	getPhoneNumberURL = "https://api.weixin.qq.com/wxa/business/getuserphonenumber"
	accessTokenSkew   = 2 * time.Minute
	defaultTokenTTL   = 2 * time.Hour
)

type Code2SessionResult struct {
	OpenID     string
	UnionID    string
	SessionKey string
}

type Client interface {
	Code2Session(ctx context.Context, code string) (*Code2SessionResult, error)
	GetPhoneNumber(ctx context.Context, phoneCode string) (string, error)
}

type MiniProgramClient struct {
	cfg          config.WechatConfig
	httpClient   *http.Client
	tokenMu      sync.Mutex
	cachedToken  string
	tokenExpires time.Time
}

func NewMiniProgramClient(cfg config.WechatConfig) *MiniProgramClient {
	return &MiniProgramClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
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

func (c *MiniProgramClient) GetPhoneNumber(ctx context.Context, phoneCode string) (string, error) {
	if phoneCode == "" {
		return "", errors.New("wechat phone code is required")
	}
	if c.cfg.MockEnabled {
		return mockPhoneFromCode(phoneCode), nil
	}
	if c.cfg.MiniAppID == "" || c.cfg.MiniAppSecret == "" {
		return "", errors.New("wechat mini app config is required")
	}

	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return "", err
	}

	body, err := json.Marshal(map[string]string{"code": phoneCode})
	if err != nil {
		return "", err
	}
	endpoint := getPhoneNumberURL + "?access_token=" + url.QueryEscape(accessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload struct {
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
		PhoneInfo struct {
			PhoneNumber     string `json:"phoneNumber"`
			PurePhoneNumber string `json:"purePhoneNumber"`
			CountryCode     string `json:"countryCode"`
		} `json:"phone_info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.ErrCode != 0 {
		return "", fmt.Errorf("wechat get phone number failed: %d", payload.ErrCode)
	}

	phone := payload.PhoneInfo.PurePhoneNumber
	if phone == "" {
		phone = payload.PhoneInfo.PhoneNumber
	}
	if phone == "" {
		return "", errors.New("wechat phone number empty")
	}
	return phone, nil
}

func (c *MiniProgramClient) getAccessToken(ctx context.Context) (string, error) {
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

func mockPhoneFromCode(phoneCode string) string {
	hash := security.HashPlain(phoneCode)
	digits := ""
	for _, ch := range hash {
		if ch >= '0' && ch <= '9' {
			digits += string(ch)
		}
		if len(digits) >= 8 {
			break
		}
	}
	for len(digits) < 8 {
		digits += "0"
	}
	return "138" + digits[:8]
}
