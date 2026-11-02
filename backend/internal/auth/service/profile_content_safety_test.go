package service

import (
	"context"
	"testing"

	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/security"
	usermodel "tree/backend/internal/user/model"
)

func TestUpdateProfileContentSafetyRejectedDoesNotPersist(t *testing.T) {
	const appID = "test-app-id"
	nickname := "违规昵称"
	users := newWechatIdentityUserRepoFake(&usermodel.User{
		ID: 601, Status: string(enums.StatusActive), PhoneVerified: true,
	})
	openID := "profile-test-openid"
	users.seedActiveIdentity(1, 601, providerWechatMini, appID, security.HashPlain(openID))
	for _, identity := range users.identityRecords {
		if identity.UserID == 601 {
			identity.OpenID = &openID
		}
	}

	client := contentsafety.NewFakeClient(contentsafety.SuggestReview)
	contentSafety := contentsafety.NewChecker(client, &bindUserOpenIDResolver{users: users, appID: appID}, nil)

	logs := &authLogCapture{}
	service := newBindServiceWithBlacklist(t, users, &codeRepoFake{}, logs, nil)
	service.contentSafety = contentSafety

	result, businessErr := service.UpdateProfile(context.Background(), UpdateProfileInput{
		UserID: 601, Nickname: &nickname, IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if businessErr == nil || businessErr.Code != apperrors.CodeContentSafetyRejected {
		t.Fatalf("expected content safety rejection, got result=%#v err=%#v", result, businessErr)
	}
	updated, err := users.FindByID(context.Background(), 601)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if updated.Nickname != nil {
		t.Fatalf("nickname should not be persisted, got %#v", updated.Nickname)
	}
}

type bindUserOpenIDResolver struct {
	users authOpenIDUserRepo
	appID string
}

type authOpenIDUserRepo interface {
	FindActiveWechatMiniOpenID(ctx context.Context, userID uint64, appID string) (string, error)
}

func (r *bindUserOpenIDResolver) ActiveWechatMiniOpenID(ctx context.Context, userID uint64) (string, error) {
	return r.users.FindActiveWechatMiniOpenID(ctx, userID, r.appID)
}
