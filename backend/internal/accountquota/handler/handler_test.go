package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	quotadto "tree/backend/internal/accountquota/dto"
	quotahandler "tree/backend/internal/accountquota/handler"
	quotaservice "tree/backend/internal/accountquota/service"
	quotavo "tree/backend/internal/accountquota/vo"
	apperrors "tree/backend/internal/common/errors"
)

type mockQuotaService struct {
	updateValues   quotadto.ConfigValues
	previewValues  quotadto.ConfigValues
	overridePhones []string
	deletedID      uint64
}

func (m *mockQuotaService) UpdateConfig(_ context.Context, _ uint64, _ string, _ string, values quotadto.ConfigValues, _ quotaservice.AuditInput) (*quotavo.ConfigItem, *apperrors.BusinessError) {
	m.updateValues = values
	return &quotavo.ConfigItem{TrustTier: "WECHAT_ONLY", MaxOwnedFamilies: values.MaxOwnedFamilies, SupportsFeaturePreview: values.SupportsFeaturePreview}, nil
}

func (m *mockQuotaService) PreviewImpact(_ context.Context, _ string, _ string, values quotadto.ConfigValues) (*quotavo.ImpactPreview, *apperrors.BusinessError) {
	m.previewValues = values
	return &quotavo.ImpactPreview{AffectedUsers: 0, AffectedFamilies: 0}, nil
}

func TestUpdateConfigAcceptsExplicitZeroValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockQuotaService{}
	handler := quotahandler.NewHandler(mock)

	body := `{"maxOwnedFamilies":0,"maxMembersPerOwnedFamily":1,"maxJoinedFamilies":0,"supportsFeaturePreview":false}`
	recorder := httptest.NewRecorder()
	ctx, engine := gin.CreateTestContext(recorder)
	engine.POST("/api/admin/account-quota-configs/:tier", func(c *gin.Context) {
		c.Set("currentAdminID", uint64(1))
		c.Set("currentAdminRole", "ROOT_ADMIN")
		handler.UpdateConfig(c)
	})
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/admin/account-quota-configs/WECHAT_ONLY", bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, ctx.Request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if mock.updateValues.MaxOwnedFamilies != 0 || mock.updateValues.MaxMembersPerOwnedFamily != 1 || mock.updateValues.MaxJoinedFamilies != 0 || mock.updateValues.SupportsFeaturePreview {
		t.Fatalf("unexpected parsed values: %#v", mock.updateValues)
	}
}

func TestUpdateConfigRejectsMissingField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := quotahandler.NewHandler(&mockQuotaService{})

	body := `{"maxOwnedFamilies":1,"maxJoinedFamilies":1,"supportsFeaturePreview":false}`
	recorder := httptest.NewRecorder()
	ctx, engine := gin.CreateTestContext(recorder)
	engine.POST("/api/admin/account-quota-configs/:tier", func(c *gin.Context) {
		c.Set("currentAdminID", uint64(1))
		c.Set("currentAdminRole", "ROOT_ADMIN")
		handler.UpdateConfig(c)
	})
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/admin/account-quota-configs/WECHAT_ONLY", bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, ctx.Request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestImpactPreviewRequestParseExplicitZero(t *testing.T) {
	var req quotadto.ImpactPreviewRequest
	if err := json.Unmarshal([]byte(`{"maxOwnedFamilies":0,"maxMembersPerOwnedFamily":10,"maxJoinedFamilies":0,"supportsFeaturePreview":false}`), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	values, err := req.Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if values.MaxOwnedFamilies != 0 || values.MaxJoinedFamilies != 0 {
		t.Fatalf("expected explicit zero values, got %#v", values)
	}
}

func TestUpdateConfigRequestParseMissingField(t *testing.T) {
	var req quotadto.UpdateConfigRequest
	if err := json.Unmarshal([]byte(`{"maxOwnedFamilies":1}`), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	_, err := req.Parse()
	if !errors.Is(err, quotadto.ErrMissingConfigField) {
		t.Fatalf("expected ErrMissingConfigField, got %v", err)
	}
}

func TestDeleteFeatureOverrideParsesID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockQuotaService{}
	handler := quotahandler.NewHandler(mock)

	recorder := httptest.NewRecorder()
	ctx, engine := gin.CreateTestContext(recorder)
	engine.DELETE("/api/admin/account-feature-overrides/:featureKey/:overrideId", func(c *gin.Context) {
		c.Set("currentAdminID", uint64(1))
		c.Set("currentAdminRole", "ROOT_ADMIN")
		handler.DeleteFeatureOverride(c)
	})
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/api/admin/account-feature-overrides/FEATURE_PREVIEW/42", nil)
	engine.ServeHTTP(recorder, ctx.Request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if mock.deletedID != 42 {
		t.Fatalf("expected deleted id 42, got %d", mock.deletedID)
	}
}

// satisfy interface for unused methods
func (m *mockQuotaService) GetCapabilities(context.Context, uint64) (*quotavo.Capabilities, *apperrors.BusinessError) {
	return nil, nil
}
func (m *mockQuotaService) ListConfigs(context.Context, string) ([]quotavo.ConfigItem, *apperrors.BusinessError) {
	return nil, nil
}
func (m *mockQuotaService) ListFeatureOverrides(context.Context, string, string) ([]quotavo.FeatureOverrideItem, *apperrors.BusinessError) {
	return nil, nil
}
func (m *mockQuotaService) UpdateFeatureOverrides(_ context.Context, _ uint64, _ string, _ string, phones []string, _ quotaservice.AuditInput) ([]quotavo.FeatureOverrideItem, *apperrors.BusinessError) {
	m.overridePhones = phones
	return []quotavo.FeatureOverrideItem{{ID: 42, FeatureKey: quotaservice.FeaturePreview, PhoneMask: "138****5678"}}, nil
}
func (m *mockQuotaService) DeleteFeatureOverride(_ context.Context, _ uint64, _ string, _ string, overrideID uint64, _ quotaservice.AuditInput) ([]quotavo.FeatureOverrideItem, *apperrors.BusinessError) {
	m.deletedID = overrideID
	return []quotavo.FeatureOverrideItem{}, nil
}
func (m *mockQuotaService) AssertProfileComplete(context.Context, *gorm.DB, uint64) error { return nil }
func (m *mockQuotaService) AssertCanCreateFamily(context.Context, *gorm.DB, uint64) error { return nil }
func (m *mockQuotaService) AssertCanJoinFamily(context.Context, *gorm.DB, uint64) error   { return nil }
func (m *mockQuotaService) AssertCanAddMember(context.Context, *gorm.DB, uint64, int) error {
	return nil
}
func (m *mockQuotaService) AssertCanRestoreFamily(context.Context, *gorm.DB, uint64) error {
	return nil
}
func (m *mockQuotaService) AssertFounderTransferReceiver(context.Context, *gorm.DB, uint64, uint64) error {
	return nil
}
func (m *mockQuotaService) AssertPostFounderTransfer(context.Context, *gorm.DB, uint64, uint64, uint64) error {
	return nil
}
func (m *mockQuotaService) MapQuotaError(error) *apperrors.BusinessError { return nil }
