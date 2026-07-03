package freshquota

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"tree/backend/internal/common/config"
	"tree/backend/internal/common/database"
)

const quotaConfigLockName = "tree_account_quota_config_test"

type quotaLockState struct {
	mu   sync.Mutex
	conn *sql.Conn
	held bool
}

var quotaLocks sync.Map

// ensureQuotaConfigLock serializes account_quota_configs mutations across test packages and processes.
func ensureQuotaConfigLock(t *testing.T) {
	t.Helper()
	stateAny, _ := quotaLocks.LoadOrStore(t.Name(), &quotaLockState{})
	state := stateAny.(*quotaLockState)
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.held {
		return
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	db, err := database.Init(context.Background(), cfg.MySQL)
	if err != nil {
		t.Skipf("mysql unavailable: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql DB: %v", err)
	}
	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		t.Fatalf("sql conn: %v", err)
	}

	lockCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var locked sql.NullInt64
	if err := conn.QueryRowContext(lockCtx, "SELECT GET_LOCK(?, 120)", quotaConfigLockName).Scan(&locked); err != nil {
		_ = conn.Close()
		t.Fatalf("GET_LOCK: %v", err)
	}
	if !locked.Valid || locked.Int64 != 1 {
		_ = conn.Close()
		t.Fatalf("GET_LOCK timeout for %s", quotaConfigLockName)
	}

	state.conn = conn
	state.held = true
	t.Cleanup(func() {
		state.mu.Lock()
		defer state.mu.Unlock()
		if !state.held {
			return
		}
		releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer releaseCancel()
		_, _ = state.conn.ExecContext(releaseCtx, "SELECT RELEASE_LOCK(?)", quotaConfigLockName)
		_ = state.conn.Close()
		state.held = false
		quotaLocks.Delete(t.Name())
	})
}
