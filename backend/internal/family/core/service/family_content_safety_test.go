package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/family/core/dto"
	familymodel "tree/backend/internal/family/core/model"
	familyrepo "tree/backend/internal/family/core/repository"
	rolemodel "tree/backend/internal/family/role/model"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"
)

type stubOpenIDResolver struct{ openid string }

func (s stubOpenIDResolver) ActiveWechatMiniOpenID(context.Context, uint64) (string, error) {
	return s.openid, nil
}

type csLogCapture struct {
	success []operationlog.WriteInput
	failed  []operationlog.WriteInput
}

func (l *csLogCapture) WriteSuccess(_ context.Context, input operationlog.WriteInput) error {
	l.success = append(l.success, input)
	return nil
}

func (l *csLogCapture) WriteFailed(_ context.Context, input operationlog.WriteInput) error {
	l.failed = append(l.failed, input)
	return nil
}

func assertCSFailedLog(t *testing.T, logs *csLogCapture, forbidden ...string) {
	t.Helper()
	if len(logs.success) != 0 {
		t.Fatalf("expected no success logs, got %d", len(logs.success))
	}
	if len(logs.failed) != 1 {
		t.Fatalf("expected one failed log, got %d", len(logs.failed))
	}
	detail := string(logs.failed[0].DetailJSON)
	for _, word := range forbidden {
		if strings.Contains(detail, word) {
			t.Fatalf("log leaked %q: %s", word, detail)
		}
	}
	if !strings.Contains(detail, "suggest") || !strings.Contains(detail, "fieldLabels") {
		t.Fatalf("log missing metadata: %s", detail)
	}
}

type allowAllFamilyPerm struct{}

func (allowAllFamilyPerm) IsFamilyAdmin(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllFamilyPerm) IsFamilyMember(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllFamilyPerm) CanManageFamily(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllFamilyPerm) CanCreateMember(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllFamilyPerm) CanEditMember(context.Context, uint64, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllFamilyPerm) CanDeleteMember(context.Context, uint64, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllFamilyPerm) CanManageRelationships(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllFamilyPerm) GetActiveLink(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error) {
	return &rolemodel.FamilyMemberUserLink{FamilyRole: string(enums.FamilyRoleFounder)}, nil
}

func TestFamilyCreateContentSafetyRejected(t *testing.T) {
	tx := transactionalTestDB(t)
	ctx := context.Background()
	user := &usermodel.User{Status: string(enums.StatusActive), PhoneVerified: true}
	if err := tx.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	client := contentsafety.NewFakeClient(contentsafety.SuggestReview)
	logs := &csLogCapture{}
	contentSafety := contentsafety.NewChecker(client, stubOpenIDResolver{openid: "oid"}, logs)
	service := NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), nil, nil, contentSafety)
	gender := "MALE"
	desc := "违规家族描述"
	_, businessErr := service.Create(ctx, user.ID, dto.CreateFamilyRequest{
		Surname: "张", Description: &desc, FounderGender: &gender,
	}, AuditInput{})
	if businessErr == nil || businessErr.Code != apperrors.CodeContentSafetyRejected {
		t.Fatalf("expected 49007, got %#v", businessErr)
	}
	var count int64
	if err := tx.Model(&familymodel.Family{}).Count(&count).Error; err != nil {
		t.Fatalf("count families: %v", err)
	}
	if count != 0 {
		t.Fatalf("family should not be created, count=%d", count)
	}
	assertCSFailedLog(t, logs, "违规家族描述", "oid")
}

func TestFamilyUpdateContentSafetyRejected(t *testing.T) {
	tx := transactionalTestDB(t)
	ctx := context.Background()
	user := &usermodel.User{Status: string(enums.StatusActive), PhoneVerified: true}
	if err := tx.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	passSvc := NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), nil, nil, contentsafety.AlwaysPass())
	gender := "MALE"
	created, err := passSvc.Create(ctx, user.ID, dto.CreateFamilyRequest{Surname: "李", FounderGender: &gender}, AuditInput{})
	if err != nil {
		t.Fatalf("create family: %#v", err)
	}
	beforeGV := created.GraphVersion

	client := contentsafety.NewFakeClient(contentsafety.SuggestRisky)
	logs := &csLogCapture{}
	contentSafety := contentsafety.NewChecker(client, stubOpenIDResolver{openid: "oid"}, logs)
	service := NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), allowAllFamilyPerm{}, nil, contentSafety)
	badName := "违规家庭名"
	_, businessErr := service.Update(ctx, user.ID, created.ID, dto.UpdateFamilyRequest{FamilyName: &badName}, AuditInput{})
	if businessErr == nil || businessErr.Code != apperrors.CodeContentSafetyRejected {
		t.Fatalf("expected 49007, got %#v", businessErr)
	}
	family, dbErr := familyrepo.NewFamilyRepository(tx).FindFamilyByID(ctx, created.ID)
	if dbErr != nil {
		t.Fatalf("find family: %v", dbErr)
	}
	if family.FamilyName != created.FamilyName || family.GraphVersion != beforeGV {
		t.Fatalf("family should be unchanged: %#v", family)
	}
	assertCSFailedLog(t, logs, badName)
	serialized, _ := json.Marshal(logs.failed)
	if strings.Contains(string(serialized), "UPDATE_FAMILY") {
		t.Fatal("business success log should not be written")
	}
}

func TestDissolutionCreateContentSafetyRejected(t *testing.T) {
	tx := transactionalTestDB(t)
	ctx := context.Background()
	user := &usermodel.User{Status: string(enums.StatusActive), PhoneVerified: true}
	if err := tx.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	passSvc := NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), nil, nil, contentsafety.AlwaysPass())
	gender := "MALE"
	created, err := passSvc.Create(ctx, user.ID, dto.CreateFamilyRequest{Surname: "王", FounderGender: &gender}, AuditInput{})
	if err != nil {
		t.Fatalf("create family: %#v", err)
	}
	beforeGV := created.GraphVersion

	client := contentsafety.NewFakeClient(contentsafety.SuggestReview)
	logs := &csLogCapture{}
	contentSafety := contentsafety.NewChecker(client, stubOpenIDResolver{openid: "oid"}, logs)
	service := NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), allowAllFamilyPerm{}, nil, contentSafety)
	reason := "违规解散理由"
	_, businessErr := service.CreateDissolutionRequest(ctx, user.ID, created.ID, dto.CreateDissolutionRequest{
		RequestReason: &reason,
	}, AuditInput{})
	if businessErr == nil || businessErr.Code != apperrors.CodeContentSafetyRejected {
		t.Fatalf("expected 49007, got %#v", businessErr)
	}
	family, dbErr := familyrepo.NewFamilyRepository(tx).FindFamilyByID(ctx, created.ID)
	if dbErr != nil {
		t.Fatalf("find family: %v", dbErr)
	}
	if family.Status != familyStatusNormal || family.GraphVersion != beforeGV {
		t.Fatalf("dissolution should not proceed: status=%s gv=%d", family.Status, family.GraphVersion)
	}
	assertCSFailedLog(t, logs, reason)
}
