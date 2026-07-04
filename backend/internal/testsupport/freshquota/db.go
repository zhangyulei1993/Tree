package freshquota

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"

	quotamodel "tree/backend/internal/accountquota/model"
	"tree/backend/internal/common/config"
	"tree/backend/internal/common/database"
	"tree/backend/internal/common/enums"
	"tree/backend/internal/common/security"
	usermodel "tree/backend/internal/user/model"
)

// RunID returns a unique prefix for the current test run (never reuses familyId=20 or fixed IDs).
func RunID(t *testing.T) string {
	t.Helper()
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return fmt.Sprintf("fq-%d-%s", time.Now().UnixNano(), hex.EncodeToString(buf[:]))
}

// TestDB starts a MySQL transaction and rolls back on cleanup.
func TestDB(t *testing.T) (*gorm.DB, string) {
	t.Helper()
	ensureQuotaConfigLock(t)
	runID := RunID(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	db, err := database.Init(context.Background(), cfg.MySQL)
	if err != nil {
		t.Skipf("mysql unavailable: %v", err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin tx: %v", tx.Error)
	}
	t.Cleanup(func() { _ = tx.Rollback().Error })
	SeedQuotaConfigs(t, tx)
	return tx, runID
}

func SeedQuotaConfigs(t *testing.T, tx *gorm.DB) {
	t.Helper()
	if err := tx.Where("1 = 1").Delete(&quotamodel.AccountQuotaConfig{}).Error; err != nil {
		t.Fatalf("clear quota configs: %v", err)
	}
	rows := migrationSeedRows()
	if err := tx.Create(&rows).Error; err != nil {
		t.Fatalf("seed quota configs: %v", err)
	}
}

func UniquePhone(runID string, suffix string) string {
	digits := ""
	for _, ch := range runID + suffix {
		if ch >= '0' && ch <= '9' {
			digits += string(ch)
		}
	}
	for len(digits) < 8 {
		digits += "7"
	}
	return "166" + digits[len(digits)-8:]
}

func LoadUser(t *testing.T, tx *gorm.DB, userID uint64) *usermodel.User {
	t.Helper()
	var user usermodel.User
	if err := tx.First(&user, userID).Error; err != nil {
		t.Fatalf("load user %d: %v", userID, err)
	}
	return &user
}

func CreateActiveUser(t *testing.T, tx *gorm.DB, nickname string, phoneLoginEnabled bool) *usermodel.User {
	t.Helper()
	var nick *string
	if nickname != "" {
		n := nickname
		nick = &n
	}
	user := &usermodel.User{
		AccountOrigin: "FRESH_QUOTA_TEST", RegisterClient: "WECHAT_MINI_PROGRAM",
		Status: string(enums.StatusActive), Nickname: nick,
	}
	if err := tx.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if phoneLoginEnabled {
		phone := fmt.Sprintf("166%08d", (time.Now().UnixNano()+int64(user.ID))%100000000)
		phoneHash := security.PhoneHash(phone)
		passwordHash, err := security.HashPassword("fresh-quota-test-password")
		if err != nil {
			t.Fatalf("HashPassword: %v", err)
		}
		if err := tx.Model(user).Updates(map[string]any{
			"phone": phone, "phone_hash": phoneHash, "password_hash": passwordHash,
			"phone_login_enabled": true,
		}).Error; err != nil {
			t.Fatalf("enable phone login: %v", err)
		}
		user.Phone = &phone
		user.PhoneHash = &phoneHash
		user.PasswordHash = &passwordHash
		user.PhoneLoginEnabled = true
	}
	return user
}

func CountActiveMembers(t *testing.T, tx *gorm.DB, familyID uint64) int64 {
	t.Helper()
	var count int64
	if err := tx.Table("family_members").
		Where("family_id = ? AND status = 'ACTIVE' AND deleted_at IS NULL", familyID).
		Count(&count).Error; err != nil {
		t.Fatalf("count members: %v", err)
	}
	return count
}

func GraphVersion(t *testing.T, tx *gorm.DB, familyID uint64) int {
	t.Helper()
	var version int
	if err := tx.Table("families").Select("graph_version").Where("id = ?", familyID).Scan(&version).Error; err != nil {
		t.Fatalf("graph version: %v", err)
	}
	return version
}
