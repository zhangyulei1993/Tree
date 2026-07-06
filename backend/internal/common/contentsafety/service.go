package contentsafety

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	authrepo "tree/backend/internal/auth/repository"
	"tree/backend/internal/common/config"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	operationlog "tree/backend/internal/operationlog/service"

	"gorm.io/gorm"
)

const (
	defaultCacheTTL = 5 * time.Minute
	moduleName      = "CONTENT_SAFETY"
	defaultAction   = "MSG_SEC_CHECK"
)

type OpenIDResolver interface {
	ActiveWechatMiniOpenID(ctx context.Context, userID uint64) (string, error)
}

type Service interface {
	CheckTexts(ctx context.Context, input CheckInput) *apperrors.BusinessError
}

type noopService struct{}

func (noopService) CheckTexts(context.Context, CheckInput) *apperrors.BusinessError {
	return nil
}

type failClosedService struct{}

func (failClosedService) CheckTexts(context.Context, CheckInput) *apperrors.BusinessError {
	return apperrors.New(apperrors.CodeContentSafetyUnavailable)
}

func AlwaysPass() Service {
	return noopService{}
}

func FailClosed() Service {
	return failClosedService{}
}

func NewService(client Client, userRepo authrepo.UserRepository, opLog operationlog.Service, appID string) Service {
	if client == nil {
		return FailClosed()
	}
	return NewChecker(client, &repoOpenIDResolver{repo: userRepo, appID: appID}, opLog)
}

func NewClient(cfg config.WechatConfig) Client {
	if cfg.MockEnabled {
		return NewFakeClient(SuggestPass)
	}
	return NewWechatClient(cfg)
}

type repoOpenIDResolver struct {
	repo  authrepo.UserRepository
	appID string
}

func (r *repoOpenIDResolver) ActiveWechatMiniOpenID(ctx context.Context, userID uint64) (string, error) {
	return r.repo.FindActiveWechatMiniOpenID(ctx, userID, r.appID)
}

type checker struct {
	client   Client
	openid   OpenIDResolver
	opLog    operationlog.Service
	cache    verdictCache
	now      func() time.Time
	disabled bool
}

type CheckerOption func(*checker)

func WithDisabled(disabled bool) CheckerOption {
	return func(c *checker) {
		c.disabled = disabled
	}
}

func NewChecker(client Client, openid OpenIDResolver, opLog operationlog.Service, opts ...CheckerOption) Service {
	c := &checker{
		client: client,
		openid: openid,
		opLog:  opLog,
		cache:  newMemoryVerdictCache(defaultCacheTTL),
		now:    time.Now,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.client == nil {
		return FailClosed()
	}
	return c
}

func (c *checker) CheckTexts(ctx context.Context, input CheckInput) *apperrors.BusinessError {
	if c == nil || c.disabled {
		return nil
	}
	fields := normalizeFields(input.Fields)
	if len(fields) == 0 {
		return nil
	}

	openid, err := c.openid.ActiveWechatMiniOpenID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.writeLog(ctx, input, nil, false, "wechat identity missing")
			return apperrors.New(apperrors.CodeContentSafetyWechatRequired)
		}
		c.writeLog(ctx, input, nil, false, "openid lookup failed")
		return apperrors.New(apperrors.CodeContentSafetyUnavailable)
	}
	if openid == "" {
		c.writeLog(ctx, input, nil, false, "wechat identity missing")
		return apperrors.New(apperrors.CodeContentSafetyWechatRequired)
	}

	content := mergeFieldContent(fields)
	cacheKey := contentCacheKey(input.UserID, input.Scene, content)
	if cached, ok := c.cache.Get(cacheKey); ok {
		return c.handleResult(ctx, input, fields, &cached)
	}

	result, err := c.client.MsgSecCheck(ctx, openid, int(input.Scene), content)
	if err != nil {
		c.writeLog(ctx, input, nil, false, "wechat api unavailable")
		return apperrors.New(apperrors.CodeContentSafetyUnavailable)
	}
	c.cache.Set(cacheKey, result)
	return c.handleResult(ctx, input, fields, &result)
}

func (c *checker) handleResult(ctx context.Context, input CheckInput, fields []Field, result *Result) *apperrors.BusinessError {
	switch result.Suggest {
	case SuggestPass:
		c.writeLog(ctx, input, result, true, "")
		return nil
	case SuggestReview, SuggestRisky:
		c.writeLog(ctx, input, result, false, result.Suggest)
		return apperrors.New(apperrors.CodeContentSafetyRejected)
	default:
		c.writeLog(ctx, input, result, false, "unknown suggest")
		return apperrors.New(apperrors.CodeContentSafetyUnavailable)
	}
}

func normalizeFields(fields []Field) []Field {
	if len(fields) == 0 {
		return nil
	}
	out := make([]Field, 0, len(fields))
	for _, field := range fields {
		label := strings.TrimSpace(field.Label)
		value := strings.TrimSpace(field.Value)
		if label == "" || value == "" {
			continue
		}
		out = append(out, Field{Label: label, Value: value})
	}
	return out
}

func mergeFieldContent(fields []Field) string {
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		parts = append(parts, field.Value)
	}
	return strings.Join(parts, "\n")
}

func fieldLabels(fields []Field) []string {
	labels := make([]string, 0, len(fields))
	for _, field := range fields {
		labels = append(labels, field.Label)
	}
	return labels
}

func contentCacheKey(userID uint64, scene Scene, content string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d\n%d\n%s", userID, scene, content)))
	return hex.EncodeToString(sum[:])
}

func (c *checker) writeLog(ctx context.Context, input CheckInput, result *Result, success bool, reason string) {
	if c.opLog == nil {
		return
	}
	action := strings.TrimSpace(input.LogAction)
	if action == "" {
		action = defaultAction
	}
	detail := map[string]any{
		"scene":       int(input.Scene),
		"fieldLabels": fieldLabels(normalizeFields(input.Fields)),
	}
	if result != nil {
		detail["suggest"] = result.Suggest
		detail["label"] = result.Label
		if result.TraceID != "" {
			detail["traceId"] = result.TraceID
		}
	}
	if reason != "" {
		detail["reason"] = reason
	}
	detailJSON, err := json.Marshal(detail)
	if err != nil {
		return
	}
	userID := input.UserID
	logInput := operationlog.WriteInput{
		OperatorType:   string(enums.OperatorTypeUser),
		OperatorUserID: &userID,
		Module:         moduleName,
		Action:         action,
		FamilyID:       input.FamilyID,
		DetailJSON:     detailJSON,
		IP:             optionalString(input.IP),
		UserAgent:      optionalString(input.UserAgent),
	}
	if success {
		_ = c.opLog.WriteSuccess(ctx, logInput)
		return
	}
	msg := reason
	if msg == "" {
		msg = "content safety rejected"
	}
	logInput.ErrorMessage = &msg
	_ = c.opLog.WriteFailed(ctx, logInput)
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
