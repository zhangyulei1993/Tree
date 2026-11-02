package freshquota

import (
	"context"
	"errors"
	"fmt"

	"tree/backend/internal/common/security"
	"tree/backend/internal/common/wechat"
)

// Session holds deterministic test WeChat credentials (never log full values).
type Session struct {
	OpenID     string
	UnionID    string
	SessionKey string
}

// FakeWechatMiniClient simulates WeChat jscode2session for integration tests.
type FakeWechatMiniClient struct {
	appID    string
	sessions map[string]Session
}

func NewFakeWechatMiniClient(appID string) *FakeWechatMiniClient {
	return &FakeWechatMiniClient{appID: appID, sessions: make(map[string]Session)}
}

// BindCode maps a login code to the session returned by ExchangeCode / Code2Session.
func (c *FakeWechatMiniClient) BindCode(code string, openID, unionID string) {
	hash := security.HashPlain(code)
	c.sessions[code] = Session{
		OpenID:     openID,
		UnionID:    unionID,
		SessionKey: "sess_" + hash[:16],
	}
}

// BindUniqueCode derives openid/unionid from runID and code suffix.
func (c *FakeWechatMiniClient) BindUniqueCode(runID, codeSuffix string) string {
	code := fmt.Sprintf("%s-%s", runID, codeSuffix)
	hash := security.HashPlain(code)
	c.BindCode(code, "oid_"+hash[:20], "uid_"+hash[20:40])
	return code
}

// ExchangeCode is the test-facing alias for Code2Session.
func (c *FakeWechatMiniClient) ExchangeCode(ctx context.Context, code string) (*wechat.Code2SessionResult, error) {
	return c.Code2Session(ctx, code)
}

func (c *FakeWechatMiniClient) Code2Session(_ context.Context, code string) (*wechat.Code2SessionResult, error) {
	session, ok := c.sessions[code]
	if !ok {
		return nil, errors.New("wechat code2session failed")
	}
	return &wechat.Code2SessionResult{
		OpenID:     session.OpenID,
		UnionID:    session.UnionID,
		SessionKey: session.SessionKey,
	}, nil
}

func (c *FakeWechatMiniClient) GetPhoneNumber(_ context.Context, phoneCode string) (string, error) {
	if phoneCode == "" {
		return "", errors.New("wechat phone code is required")
	}
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
		digits += "3"
	}
	return "138" + digits[:8], nil
}
