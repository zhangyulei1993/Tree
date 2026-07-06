package contentsafety

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	authrepo "tree/backend/internal/auth/repository"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"

	"gorm.io/gorm"
)

type openIDResolverFake struct {
	openid string
	err    error
}

func (f *openIDResolverFake) ActiveWechatMiniOpenID(context.Context, uint64) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if f.openid == "" {
		return "", gorm.ErrRecordNotFound
	}
	return f.openid, nil
}

type logCapture struct {
	success []operationlog.WriteInput
	failed  []operationlog.WriteInput
}

func (l *logCapture) WriteSuccess(_ context.Context, input operationlog.WriteInput) error {
	l.success = append(l.success, input)
	return nil
}

func (l *logCapture) WriteFailed(_ context.Context, input operationlog.WriteInput) error {
	l.failed = append(l.failed, input)
	return nil
}

func TestCheckTextsPassWritesAllowed(t *testing.T) {
	client := NewFakeClient(SuggestPass)
	logs := &logCapture{}
	service := NewChecker(client, &openIDResolverFake{openid: "test-openid"}, logs)
	err := service.CheckTexts(context.Background(), CheckInput{
		UserID: 1,
		Scene:  SceneProfile,
		Fields: StringField("nickname", "树友"),
	})
	if err != nil {
		t.Fatalf("expected pass, got %#v", err)
	}
	if len(logs.success) != 1 || len(logs.failed) != 0 {
		t.Fatalf("unexpected logs: success=%d failed=%d", len(logs.success), len(logs.failed))
	}
}

func TestCheckTextsReviewRejected(t *testing.T) {
	client := NewFakeClient(SuggestReview)
	service := NewChecker(client, &openIDResolverFake{openid: "test-openid"}, nil)
	err := service.CheckTexts(context.Background(), CheckInput{
		UserID: 1,
		Scene:  SceneProfile,
		Fields: StringField("nickname", "违规词"),
	})
	if err == nil || err.Code != apperrors.CodeContentSafetyRejected {
		t.Fatalf("expected rejected, got %#v", err)
	}
}

func TestCheckTextsRiskyRejected(t *testing.T) {
	client := NewFakeClient(SuggestRisky)
	service := NewChecker(client, &openIDResolverFake{openid: "test-openid"}, nil)
	err := service.CheckTexts(context.Background(), CheckInput{
		UserID: 1,
		Scene:  SceneSocial,
		Fields: StringField("invite_message", "风险内容"),
	})
	if err == nil || err.Code != apperrors.CodeContentSafetyRejected {
		t.Fatalf("expected rejected, got %#v", err)
	}
}

func TestCheckTextsWechatUnavailable(t *testing.T) {
	client := NewFakeClient(SuggestPass)
	client.Err = errors.New("wechat down")
	service := NewChecker(client, &openIDResolverFake{openid: "test-openid"}, nil)
	err := service.CheckTexts(context.Background(), CheckInput{
		UserID: 1,
		Scene:  SceneProfile,
		Fields: StringField("nickname", "树友"),
	})
	if err == nil || err.Code != apperrors.CodeContentSafetyUnavailable {
		t.Fatalf("expected unavailable, got %#v", err)
	}
}

func TestCheckTextsMissingWechatIdentity(t *testing.T) {
	client := NewFakeClient(SuggestPass)
	service := NewChecker(client, &openIDResolverFake{}, nil)
	err := service.CheckTexts(context.Background(), CheckInput{
		UserID: 1,
		Scene:  SceneProfile,
		Fields: StringField("nickname", "树友"),
	})
	if err == nil || err.Code != apperrors.CodeContentSafetyWechatRequired {
		t.Fatalf("expected wechat required, got %#v", err)
	}
}

func TestCheckTextsMergesMultipleFieldsSingleCall(t *testing.T) {
	client := NewFakeClient(SuggestPass)
	service := NewChecker(client, &openIDResolverFake{openid: "test-openid"}, nil)
	err := service.CheckTexts(context.Background(), CheckInput{
		UserID: 1,
		Scene:  SceneProfile,
		Fields: MergeFields(
			StringField("family_name", "张家"),
			StringField("description", "家族简介"),
		),
	})
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}
	if client.CallCount() != 1 {
		t.Fatalf("expected one call, got %d", client.CallCount())
	}
	call, ok := client.LastCall()
	if !ok || call.Content != "张家\n家族简介" {
		t.Fatalf("unexpected merged content: %#v", call)
	}
}

func TestCheckTextsCacheAvoidsDuplicateCalls(t *testing.T) {
	client := NewFakeClient(SuggestPass)
	service := NewChecker(client, &openIDResolverFake{openid: "test-openid"}, nil)
	input := CheckInput{
		UserID: 1,
		Scene:  SceneProfile,
		Fields: StringField("nickname", "树友"),
	}
	if err := service.CheckTexts(context.Background(), input); err != nil {
		t.Fatalf("first check: %#v", err)
	}
	if err := service.CheckTexts(context.Background(), input); err != nil {
		t.Fatalf("second check: %#v", err)
	}
	if client.CallCount() != 1 {
		t.Fatalf("expected cached single call, got %d", client.CallCount())
	}
}

func TestOperationLogDoesNotContainSensitiveValues(t *testing.T) {
	client := NewFakeClient(SuggestPass)
	logs := &logCapture{}
	service := NewChecker(client, &openIDResolverFake{openid: "secret_openid_value"}, logs)
	_ = service.CheckTexts(context.Background(), CheckInput{
		UserID: 1,
		Scene:  SceneProfile,
		Fields: StringField("nickname", "敏感昵称"),
		IP:     "127.0.0.1",
	})
	serialized, _ := json.Marshal(struct {
		Success []operationlog.WriteInput
		Failed  []operationlog.WriteInput
	}{Success: logs.success, Failed: logs.failed})
	text := string(serialized)
	if len(logs.success) != 1 {
		t.Fatalf("expected one success log, got %d", len(logs.success))
	}
	detailText := string(logs.success[0].DetailJSON)
	for _, forbidden := range []string{"敏感昵称", "secret_openid_value", "access_token", "accessToken"} {
		if strings.Contains(text, forbidden) || strings.Contains(detailText, forbidden) {
			t.Fatalf("operation log leaked %q", forbidden)
		}
	}
	if !strings.Contains(detailText, "traceId") {
		t.Fatalf("expected traceId in operation log detail, got %s", detailText)
	}
}

type userRepoOpenIDFake struct {
	authrepo.UserRepository
	openid string
}

func (f *userRepoOpenIDFake) FindActiveWechatMiniOpenID(context.Context, uint64, string) (string, error) {
	if f.openid == "" {
		return "", gorm.ErrRecordNotFound
	}
	return f.openid, nil
}

func TestNewServiceUsesRepositoryOpenID(t *testing.T) {
	client := NewFakeClient(SuggestPass)
	repo := &userRepoOpenIDFake{openid: "repo-openid"}
	service := NewService(client, repo, nil, "test-app")
	err := service.CheckTexts(context.Background(), CheckInput{
		UserID: 9,
		Scene:  SceneProfile,
		Fields: StringField("nickname", "树友"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}
	call, ok := client.LastCall()
	if !ok || call.OpenID != "repo-openid" {
		t.Fatalf("unexpected openid call: %#v", call)
	}
}

func TestRejectedMessage(t *testing.T) {
	if apperrors.Message(apperrors.CodeContentSafetyRejected) != "内容可能不符合平台规范，请修改后重试" {
		t.Fatal("unexpected rejected message")
	}
}

func TestAlwaysPassSkipsClient(t *testing.T) {
	client := NewFakeClient(SuggestRisky)
	service := AlwaysPass()
	err := service.CheckTexts(context.Background(), CheckInput{
		UserID: 1,
		Scene:  SceneProfile,
		Fields: StringField("nickname", "树友"),
	})
	if err != nil {
		t.Fatalf("always pass should not error: %#v", err)
	}
	if client.CallCount() != 0 {
		t.Fatal("always pass should not call client")
	}
}

func TestFailClosedReturnsUnavailable(t *testing.T) {
	err := FailClosed().CheckTexts(context.Background(), CheckInput{
		UserID: 1,
		Scene:  SceneProfile,
		Fields: StringField("nickname", "树友"),
	})
	if err == nil || err.Code != apperrors.CodeContentSafetyUnavailable {
		t.Fatalf("expected unavailable, got %#v", err)
	}
}

func TestNewCheckerNilClientFailsClosed(t *testing.T) {
	service := NewChecker(nil, &openIDResolverFake{openid: "oid"}, nil)
	err := service.CheckTexts(context.Background(), CheckInput{
		UserID: 1, Scene: SceneProfile, Fields: StringField("nickname", "树友"),
	})
	if err == nil || err.Code != apperrors.CodeContentSafetyUnavailable {
		t.Fatalf("expected unavailable, got %#v", err)
	}
}

func TestCacheIsolationDifferentUsers(t *testing.T) {
	client := NewFakeClient(SuggestPass)
	service := NewChecker(client, &openIDResolverFake{openid: "oid"}, nil)
	fields := StringField("nickname", "相同内容")
	if err := service.CheckTexts(context.Background(), CheckInput{UserID: 1, Scene: SceneProfile, Fields: fields}); err != nil {
		t.Fatalf("user1: %#v", err)
	}
	if err := service.CheckTexts(context.Background(), CheckInput{UserID: 2, Scene: SceneProfile, Fields: fields}); err != nil {
		t.Fatalf("user2: %#v", err)
	}
	if client.CallCount() != 2 {
		t.Fatalf("expected separate calls per user, got %d", client.CallCount())
	}
}

func TestCacheIsolationDifferentScenes(t *testing.T) {
	client := NewFakeClient(SuggestPass)
	service := NewChecker(client, &openIDResolverFake{openid: "oid"}, nil)
	fields := StringField("nickname", "相同内容")
	if err := service.CheckTexts(context.Background(), CheckInput{UserID: 1, Scene: SceneProfile, Fields: fields}); err != nil {
		t.Fatalf("scene1: %#v", err)
	}
	if err := service.CheckTexts(context.Background(), CheckInput{UserID: 1, Scene: SceneSocial, Fields: fields}); err != nil {
		t.Fatalf("scene2: %#v", err)
	}
	if client.CallCount() != 2 {
		t.Fatalf("expected separate calls per scene, got %d", client.CallCount())
	}
}

func TestCachedReviewStillRejected(t *testing.T) {
	client := NewFakeClient(SuggestReview)
	logs := &logCapture{}
	service := NewChecker(client, &openIDResolverFake{openid: "oid"}, logs)
	input := CheckInput{UserID: 1, Scene: SceneSocial, Fields: StringField("invite_message", "违规")}
	for i := 0; i < 2; i++ {
		err := service.CheckTexts(context.Background(), input)
		if err == nil || err.Code != apperrors.CodeContentSafetyRejected {
			t.Fatalf("attempt %d: expected rejected, got %#v", i+1, err)
		}
	}
	if client.CallCount() != 1 {
		t.Fatalf("expected cached single call, got %d", client.CallCount())
	}
	if len(logs.success) != 0 || len(logs.failed) != 2 {
		t.Fatalf("unexpected logs: success=%d failed=%d", len(logs.success), len(logs.failed))
	}
	detail := string(logs.failed[0].DetailJSON)
	for _, forbidden := range []string{"违规", "oid", "access_token"} {
		if strings.Contains(detail, forbidden) {
			t.Fatalf("failed log leaked %q: %s", forbidden, detail)
		}
	}
	if !strings.Contains(detail, "suggest") || !strings.Contains(detail, "fieldLabels") {
		t.Fatalf("failed log missing metadata: %s", detail)
	}
}

func TestInvalidSuggestFromClientUnavailable(t *testing.T) {
	client := NewFakeClient("unknown-verdict")
	service := NewChecker(client, &openIDResolverFake{openid: "oid"}, nil)
	err := service.CheckTexts(context.Background(), CheckInput{
		UserID: 1, Scene: SceneProfile, Fields: StringField("nickname", "树友"),
	})
	if err == nil || err.Code != apperrors.CodeContentSafetyUnavailable {
		t.Fatalf("expected unavailable, got %#v", err)
	}
}

func TestIdentityModelStoresOpenIDForResolver(t *testing.T) {
	openid := "wx_openid_abc"
	identity := usermodel.UserAuthIdentity{
		UserID: 1, Provider: "WECHAT_MINI", ProviderAppID: stringPtr("app"),
		OpenID: &openid, IdentityStatus: string(enums.StatusActive),
	}
	if identity.OpenID == nil || *identity.OpenID != openid {
		t.Fatalf("unexpected identity: %#v", identity)
	}
}

func stringPtr(value string) *string {
	return &value
}
