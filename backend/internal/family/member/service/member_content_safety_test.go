package service

import (
	"context"
	"strings"
	"testing"

	"gorm.io/gorm"

	"tree/backend/internal/common/contentsafety"
	apperrors "tree/backend/internal/common/errors"
	coremodel "tree/backend/internal/family/core/model"
	"tree/backend/internal/family/member/dto"
	membermodel "tree/backend/internal/family/member/model"
	operationlog "tree/backend/internal/operationlog/service"
)

type memberCSOpenIDResolver struct{}

func (memberCSOpenIDResolver) ActiveWechatMiniOpenID(context.Context, uint64) (string, error) {
	return "member-test-openid", nil
}

type memberCSLogCapture struct {
	failed []operationlog.WriteInput
}

func (l *memberCSLogCapture) WriteSuccess(context.Context, operationlog.WriteInput) error { return nil }
func (l *memberCSLogCapture) WriteFailed(_ context.Context, input operationlog.WriteInput) error {
	l.failed = append(l.failed, input)
	return nil
}

func TestMemberCreateContentSafetyRejected(t *testing.T) {
	originalTx := runMemberTransaction
	runMemberTransaction = func(_ context.Context, _ *gorm.DB, fn func(tx *gorm.DB) error) error {
		return fn(nil)
	}
	t.Cleanup(func() { runMemberTransaction = originalTx })

	repo := &memberRepoFake{
		family: coremodel.Family{ID: 22, Status: "NORMAL", GraphVersion: 10},
	}
	beforeGV := repo.family.GraphVersion
	client := contentsafety.NewFakeClient(contentsafety.SuggestReview)
	logs := &memberCSLogCapture{}
	contentSafety := contentsafety.NewChecker(client, memberCSOpenIDResolver{}, logs)
	svc := NewMemberService(nil, repo, memberPermFake{canManage: true}, nil, contentSafety)

	_, businessErr := svc.Create(context.Background(), 8, 22, dto.CreateMemberRequest{Name: "违规成员名"}, AuditInput{})
	if businessErr == nil || businessErr.Code != apperrors.CodeContentSafetyRejected {
		t.Fatalf("expected 49007, got %#v", businessErr)
	}
	if repo.family.GraphVersion != beforeGV {
		t.Fatalf("graph version must not change: %d", repo.family.GraphVersion)
	}
	if len(logs.failed) != 1 {
		t.Fatalf("expected one content safety failed log, got %d", len(logs.failed))
	}
	detail := string(logs.failed[0].DetailJSON)
	if strings.Contains(detail, "违规成员名") {
		t.Fatalf("log leaked content: %s", detail)
	}
}

func TestMemberUpdateContentSafetyRejected(t *testing.T) {
	originalTx := runMemberTransaction
	runMemberTransaction = func(_ context.Context, _ *gorm.DB, fn func(tx *gorm.DB) error) error {
		return fn(nil)
	}
	t.Cleanup(func() { runMemberTransaction = originalTx })

	repo := &memberRepoFake{
		family: coremodel.Family{ID: 22, Status: "NORMAL", GraphVersion: 10},
		member: membermodel.FamilyMember{ID: 6, FamilyID: 22, DisplayName: "成员", Status: "ACTIVE"},
	}
	beforeGV := repo.family.GraphVersion
	client := contentsafety.NewFakeClient(contentsafety.SuggestRisky)
	logs := &memberCSLogCapture{}
	contentSafety := contentsafety.NewChecker(client, memberCSOpenIDResolver{}, logs)
	svc := NewMemberService(nil, repo, memberPermFake{canManage: true}, nil, contentSafety)
	badDesc := "违规描述"
	_, businessErr := svc.Update(context.Background(), 8, 22, 6, dto.UpdateMemberRequest{Description: &badDesc}, AuditInput{})
	if businessErr == nil || businessErr.Code != apperrors.CodeContentSafetyRejected {
		t.Fatalf("expected 49007, got %#v", businessErr)
	}
	if repo.family.GraphVersion != beforeGV {
		t.Fatalf("graph version must not change")
	}
	if len(logs.failed) != 1 {
		t.Fatalf("expected content safety failed log")
	}
}
