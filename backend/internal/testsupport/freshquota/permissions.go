package freshquota

import (
	"context"
	"testing"

	quotaservice "tree/backend/internal/accountquota/service"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	rolemodel "tree/backend/internal/family/role/model"
)

type AllowAllFamilyPerm struct{}

func (AllowAllFamilyPerm) IsFamilyAdmin(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllFamilyPerm) IsFamilyMember(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllFamilyPerm) CanManageFamily(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllFamilyPerm) CanCreateMember(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllFamilyPerm) CanEditMember(context.Context, uint64, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllFamilyPerm) CanDeleteMember(context.Context, uint64, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllFamilyPerm) CanManageRelationships(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllFamilyPerm) GetActiveLink(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error) {
	return &rolemodel.FamilyMemberUserLink{FamilyRole: "FOUNDER"}, nil
}

type AllowAllMemberPerm struct{}

func (AllowAllMemberPerm) IsFamilyAdmin(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllMemberPerm) IsFamilyMember(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllMemberPerm) CanManageFamily(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllMemberPerm) CanCreateMember(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllMemberPerm) CanEditMember(context.Context, uint64, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllMemberPerm) CanDeleteMember(context.Context, uint64, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllMemberPerm) CanManageRelationships(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (AllowAllMemberPerm) GetActiveLink(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error) {
	return nil, nil
}

func AssertQuotaCode(t *testing.T, svc quotaservice.Service, err error, code apperrors.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected quota-related error %d, got nil", code)
	}
	businessErr := svc.MapQuotaError(err)
	if businessErr == nil || businessErr.Code != code {
		t.Fatalf("expected business error %d, got %#v from %v", code, businessErr, err)
	}
}

func AssertBusinessCode(t *testing.T, businessErr *apperrors.BusinessError, code apperrors.Code) {
	t.Helper()
	if businessErr == nil || businessErr.Code != code {
		t.Fatalf("expected business error %d, got %#v", code, businessErr)
	}
}

func AssertUserStatus(t *testing.T, status string, expected enums.RecordStatus) {
	t.Helper()
	if status != string(expected) {
		t.Fatalf("expected status %s, got %s", expected, status)
	}
}
