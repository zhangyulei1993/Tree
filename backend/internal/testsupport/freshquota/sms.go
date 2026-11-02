package freshquota

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/enums"
	"tree/backend/internal/common/security"
	usermodel "tree/backend/internal/user/model"
)

const bindPhoneScene = "BIND_PHONE"

// SeedBindPhoneCode inserts a pending verification code for real BindPhone business flow.
func SeedBindPhoneCode(t *testing.T, tx *gorm.DB, phone, code string) {
	t.Helper()
	codeHash, err := security.HashVerificationCode(phone, bindPhoneScene, code)
	if err != nil {
		t.Fatalf("HashVerificationCode: %v", err)
	}
	record := &usermodel.VerificationCode{
		PhoneHash: security.PhoneHash(phone),
		Scene:     bindPhoneScene,
		CodeHash:  codeHash,
		Status:    string(enums.StatusPending),
		ExpiredAt: time.Now().Add(10 * time.Minute),
	}
	if err := tx.WithContext(context.Background()).Create(record).Error; err != nil {
		t.Fatalf("seed verification code: %v", err)
	}
}
