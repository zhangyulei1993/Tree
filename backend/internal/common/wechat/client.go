package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"tree/backend/internal/common/config"
	"tree/backend/internal/common/security"
)

const (
	code2SessionURL   = "https://api.weixin.qq.com/sns/jscode2session"
	accessTokenURL    = "https://api.weixin.qq.com/cgi-bin/token"
	getPhoneNumberURL = "https://api.weixin.qq.com/wxa/business/getuserphonenumber"
	subscribeSendURL  = "https://api.weixin.qq.com/cgi-bin/message/subscribe/send"
	templateListURL   = "https://api.weixin.qq.com/wxaapi/newtmpl/gettemplate"
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
	templateMu   sync.Mutex
	templates    []SubscribeTemplate
	templateAt   time.Time
}

type SubscribeTemplate struct {
	ID      string `json:"priTmplId"`
	Title   string `json:"title"`
	Content string `json:"content"`
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

// SendSubscribeMessage resolves the template's actual DATA keys from WeChat's
// current template definition, then maps values to those keys in template order.
// This keeps field identifiers such as thing1/time2 out of business code.
func (c *MiniProgramClient) SendSubscribeMessage(ctx context.Context, openID, templateID, page, miniprogramState string, values []string) error {
	if strings.TrimSpace(openID) == "" || strings.TrimSpace(templateID) == "" {
		return errors.New("wechat subscription recipient and template are required")
	}
	if c.cfg.MockEnabled {
		return nil
	}
	template, err := c.findSubscribeTemplate(ctx, templateID)
	if err != nil {
		return err
	}
	keys := subscribeDataKeys(template.Content)
	if len(keys) == 0 || len(keys) != len(values) {
		return fmt.Errorf("wechat subscription template fields mismatch: want %d values, got %d", len(keys), len(values))
	}
	data := make(map[string]map[string]string, len(keys))
	for i, key := range keys {
		data[key] = map[string]string{"value": normalizeSubscribeValue(key, values[i])}
	}
	body := map[string]any{
		"touser":      openID,
		"template_id": templateID,
		"data":        data,
	}
	if strings.TrimSpace(page) != "" {
		body["page"] = page
	}
	if state := strings.TrimSpace(miniprogramState); state != "" {
		body["miniprogram_state"] = state
	}
	return c.postSubscribeMessage(ctx, body)
}

func (c *MiniProgramClient) postSubscribeMessage(ctx context.Context, body map[string]any) error {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, subscribeSendURL+"?access_token="+url.QueryEscape(token), bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var payload struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}
	if payload.ErrCode != 0 {
		return fmt.Errorf("wechat subscription send failed: %d", payload.ErrCode)
	}
	return nil
}

func (c *MiniProgramClient) findSubscribeTemplate(ctx context.Context, templateID string) (*SubscribeTemplate, error) {
	c.templateMu.Lock()
	if time.Since(c.templateAt) < 10*time.Minute {
		for i := range c.templates {
			if c.templates[i].ID == templateID {
				result := c.templates[i]
				c.templateMu.Unlock()
				return &result, nil
			}
		}
	}
	c.templateMu.Unlock()

	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, templateListURL+"?access_token="+url.QueryEscape(token), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var payload struct {
		ErrCode int                 `json:"errcode"`
		ErrMsg  string              `json:"errmsg"`
		Data    []SubscribeTemplate `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.ErrCode != 0 {
		return nil, fmt.Errorf("wechat subscription template query failed: %d", payload.ErrCode)
	}
	c.templateMu.Lock()
	c.templates = payload.Data
	c.templateAt = time.Now()
	c.templateMu.Unlock()
	for i := range payload.Data {
		if payload.Data[i].ID == templateID {
			return &payload.Data[i], nil
		}
	}
	return nil, errors.New("wechat subscription template is not configured")
}

var subscribeKeyPattern = regexp.MustCompile(`\{\{([A-Za-z0-9_]+)\.DATA\}\}`)

func subscribeDataKeys(content string) []string {
	matches := subscribeKeyPattern.FindAllStringSubmatch(content, -1)
	keys := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			if _, ok := seen[match[1]]; !ok {
				seen[match[1]] = struct{}{}
				keys = append(keys, match[1])
			}
		}
	}
	return keys
}

// normalizeSubscribeValue keeps dynamic business text within WeChat template
// field constraints while preserving the canonical template field order.
func normalizeSubscribeValue(key, value string) string {
	value = strings.TrimSpace(value)
	switch {
	case strings.HasPrefix(key, "date"):
		if parsed, ok := parseSubscribeTime(value); ok {
			return parsed.Format("2006-01-02")
		}
	case strings.HasPrefix(key, "time"):
		if parsed, ok := parseSubscribeTime(value); ok {
			return parsed.Format("15:04")
		}
	case strings.HasPrefix(key, "phrase"):
		return truncateSubscribeValue(value, 5)
	case strings.HasPrefix(key, "thing"):
		return truncateSubscribeValue(value, 20)
	case strings.HasPrefix(key, "name"):
		return truncateSubscribeValue(value, 10)
	case strings.HasPrefix(key, "character_string"):
		return truncateSubscribeValue(value, 32)
	}
	return value
}

func parseSubscribeTime(value string) (time.Time, bool) {
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04", "2006/1/2 15:04:05", "2006/1/2 15:04"} {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func truncateSubscribeValue(value string, limit int) string {
	chars := []rune(value)
	if len(chars) <= limit {
		return value
	}
	return string(chars[:limit])
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
