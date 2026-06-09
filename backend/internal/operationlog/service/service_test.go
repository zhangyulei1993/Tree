package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/config"
	"tree/backend/internal/common/database"
	operationmodel "tree/backend/internal/operationlog/model"
)

func operationLogTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	db, err := database.Init(context.Background(), cfg.MySQL)
	if err != nil {
		t.Skipf("development MySQL unavailable: %v", err)
	}
	return db
}

func TestWriteSuccessAndFailedResults(t *testing.T) {
	ctx := context.Background()
	tx := operationLogTestDB(t).Begin()
	if tx.Error != nil {
		t.Fatalf("begin transaction: %v", tx.Error)
	}
	t.Cleanup(func() { _ = tx.Rollback().Error })

	service := NewGormService(tx)
	if err := service.WriteSuccess(ctx, WriteInput{
		OperatorType: "SYSTEM", Module: "P0_TEST", Action: "SUCCESS_CASE",
	}); err != nil {
		t.Fatalf("WriteSuccess: %v", err)
	}
	failure := "expected test failure"
	if err := service.WriteFailed(ctx, WriteInput{
		OperatorType: "SYSTEM", Module: "P0_TEST", Action: "FAILED_CASE", ErrorMessage: &failure,
	}); err != nil {
		t.Fatalf("WriteFailed: %v", err)
	}

	var records []operationmodel.OperationLog
	if err := tx.WithContext(ctx).Where("module = ?", "P0_TEST").Order("id").Find(&records).Error; err != nil {
		t.Fatalf("load logs: %v", err)
	}
	if len(records) != 2 || records[0].Result != ResultSuccess || records[1].Result != ResultFailed {
		t.Fatalf("unexpected operation log results: %#v", records)
	}
}

func TestSensitiveDetailJSONIsSanitized(t *testing.T) {
	ctx := context.Background()
	tx := operationLogTestDB(t).Begin()
	if tx.Error != nil {
		t.Fatalf("begin transaction: %v", tx.Error)
	}
	t.Cleanup(func() { _ = tx.Rollback().Error })

	detail := []byte(`{
		"password":"sensitive-value",
		"password_hash":"sensitive-hash",
		"accessToken":"sensitive-access",
		"refreshToken":"sensitive-refresh",
		"verificationCode":"sensitive-code",
		"openid":"sensitive-openid",
		"unionid":"sensitive-unionid",
		"AppSecret":"sensitive-secret",
		"safeField":"retained"
	}`)
	if err := NewGormService(tx).WriteSuccess(ctx, WriteInput{
		OperatorType: "SYSTEM", Module: "P0_TEST", Action: "SANITIZE_CASE", DetailJSON: detail,
	}); err != nil {
		t.Fatalf("WriteSuccess: %v", err)
	}

	var record operationmodel.OperationLog
	if err := tx.WithContext(ctx).Where("module = ? AND action = ?", "P0_TEST", "SANITIZE_CASE").Last(&record).Error; err != nil {
		t.Fatalf("load log: %v", err)
	}
	stored := strings.ToLower(string(record.DetailJSON))
	for _, forbidden := range []string{
		"sensitive-value", "sensitive-hash", "sensitive-access",
		"sensitive-refresh", "sensitive-code", "sensitive-openid",
		"sensitive-unionid", "sensitive-secret",
	} {
		if strings.Contains(stored, strings.ToLower(forbidden)) {
			t.Fatalf("detail_json retained sensitive value %q: %s", forbidden, stored)
		}
	}
	if !strings.Contains(stored, "safefield") {
		t.Fatalf("sanitization removed non-sensitive content: %s", stored)
	}
	if count := strings.Count(string(record.DetailJSON), redactedValue); count != 8 {
		t.Fatalf("redacted value count = %d, want 8: %s", count, record.DetailJSON)
	}
}

func TestSensitiveDetailJSONIsSanitizedRecursively(t *testing.T) {
	type nestedDetail struct {
		DisplayName string `json:"displayName"`
		Password    string `json:"password"`
		Metadata    struct {
			AccessToken string `json:"accessToken"`
		} `json:"metadata"`
	}
	detail := nestedDetail{DisplayName: "retained", Password: "struct-secret"}
	detail.Metadata.AccessToken = "nested-secret"
	raw, err := json.Marshal(map[string]any{
		"struct": detail,
		"items": []any{
			map[string]any{"verification_code": "array-secret", "safe": "retained"},
		},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}

	cleaned, err := sanitizeDetailJSON(raw)
	if err != nil {
		t.Fatalf("sanitizeDetailJSON: %v", err)
	}
	text := string(cleaned)
	for _, secret := range []string{"struct-secret", "nested-secret", "array-secret"} {
		if strings.Contains(text, secret) {
			t.Fatalf("nested sensitive value retained: %s", text)
		}
	}
	if strings.Count(text, redactedValue) != 3 || !strings.Contains(text, "retained") {
		t.Fatalf("unexpected recursive sanitization result: %s", text)
	}
}

func TestRolledBackTransactionLeavesNoSuccessLog(t *testing.T) {
	ctx := context.Background()
	db := operationLogTestDB(t)
	action := "ROLLBACK_" + time.Now().UTC().Format("20060102150405.000000000")
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin transaction: %v", tx.Error)
	}
	if err := NewGormService(tx).WriteSuccess(ctx, WriteInput{
		OperatorType: "SYSTEM", Module: "P0_TEST", Action: action,
	}); err != nil {
		_ = tx.Rollback().Error
		t.Fatalf("WriteSuccess: %v", err)
	}
	if err := tx.Rollback().Error; err != nil {
		t.Fatalf("rollback: %v", err)
	}

	var count int64
	if err := db.WithContext(ctx).Model(&operationmodel.OperationLog{}).
		Where("module = ? AND action = ?", "P0_TEST", action).Count(&count).Error; err != nil {
		t.Fatalf("count logs: %v", err)
	}
	if count != 0 {
		t.Fatalf("rolled-back SUCCESS log persisted: count=%d", count)
	}
}
