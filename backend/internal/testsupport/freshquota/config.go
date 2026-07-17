package freshquota

import (
	"context"
	"testing"

	"gorm.io/gorm"

	quotadto "tree/backend/internal/accountquota/dto"
	quotaenum "tree/backend/internal/accountquota/enum"
	quotamodel "tree/backend/internal/accountquota/model"
	quotarepo "tree/backend/internal/accountquota/repository"
	quotaservice "tree/backend/internal/accountquota/service"
	quotavo "tree/backend/internal/accountquota/vo"
	"tree/backend/internal/common/enums"
)

// TierConfig mirrors admin PUT payload fields.
type TierConfig struct {
	MaxOwnedFamilies         int
	MaxMembersPerOwnedFamily int
	MaxJoinedFamilies        int
	SupportsFeaturePreview   bool
}

// ConfigMap is keyed by trust tier.
type ConfigMap map[string]TierConfig

// MigrationDefaults are seeded by migration 000018 (verified by tests, not used as runtime limits).
func MigrationDefaults() ConfigMap {
	return ConfigMap{
		quotaenum.TrustTierWechatOnly: {
			MaxOwnedFamilies: 1, MaxMembersPerOwnedFamily: 10, MaxJoinedFamilies: 1, SupportsFeaturePreview: false,
		},
		quotaenum.TrustTierPhoneBound: {
			MaxOwnedFamilies: 1, MaxMembersPerOwnedFamily: 20, MaxJoinedFamilies: 5, SupportsFeaturePreview: true,
		},
	}
}

func migrationSeedRows() []quotamodel.AccountQuotaConfig {
	defaults := MigrationDefaults()
	return []quotamodel.AccountQuotaConfig{
		{
			TrustTier:                quotaenum.TrustTierWechatOnly,
			MaxOwnedFamilies:         defaults[quotaenum.TrustTierWechatOnly].MaxOwnedFamilies,
			MaxMembersPerOwnedFamily: defaults[quotaenum.TrustTierWechatOnly].MaxMembersPerOwnedFamily,
			MaxJoinedFamilies:        defaults[quotaenum.TrustTierWechatOnly].MaxJoinedFamilies,
			SupportsFeaturePreview:   defaults[quotaenum.TrustTierWechatOnly].SupportsFeaturePreview,
		},
		{
			TrustTier:                quotaenum.TrustTierPhoneBound,
			MaxOwnedFamilies:         defaults[quotaenum.TrustTierPhoneBound].MaxOwnedFamilies,
			MaxMembersPerOwnedFamily: defaults[quotaenum.TrustTierPhoneBound].MaxMembersPerOwnedFamily,
			MaxJoinedFamilies:        defaults[quotaenum.TrustTierPhoneBound].MaxJoinedFamilies,
			SupportsFeaturePreview:   defaults[quotaenum.TrustTierPhoneBound].SupportsFeaturePreview,
		},
	}
}

func SnapshotConfigs(t *testing.T, db *gorm.DB) ConfigMap {
	t.Helper()
	repo := quotarepo.NewRepository(db)
	rows, err := repo.ListConfigs(context.Background())
	if err != nil {
		t.Fatalf("snapshot ListConfigs: %v", err)
	}
	snapshot := make(ConfigMap, len(rows))
	for _, row := range rows {
		snapshot[row.TrustTier] = TierConfig{
			MaxOwnedFamilies: row.MaxOwnedFamilies, MaxMembersPerOwnedFamily: row.MaxMembersPerOwnedFamily,
			MaxJoinedFamilies: row.MaxJoinedFamilies, SupportsFeaturePreview: row.SupportsFeaturePreview,
		}
	}
	return snapshot
}

// DeferRestoreQuotaConfigs saves current configs and restores them on cleanup (even when tests fail).
func DeferRestoreQuotaConfigs(t *testing.T, db *gorm.DB, adminID uint64, role string) {
	t.Helper()
	ensureQuotaConfigLock(t)
	snapshot := SnapshotConfigs(t, db)
	svc := quotaservice.NewService(db, quotarepo.NewRepository(db))
	t.Cleanup(func() {
		if err := restoreConfigsStaged(svc, adminID, role, snapshot); err != nil {
			t.Errorf("restore quota configs: %v", err)
		}
	})
}

func restoreConfigsStaged(svc quotaservice.Service, adminID uint64, role string, want ConfigMap) error {
	wechatWant, okW := want[quotaenum.TrustTierWechatOnly]
	phoneWant, okP := want[quotaenum.TrustTierPhoneBound]
	if !okW || !okP {
		return nil
	}
	ctx := context.Background()
	audit := quotaservice.AuditInput{IP: "127.0.0.1", UserAgent: "fresh-quota-restore"}
	current, businessErr := svc.ListConfigs(ctx, role)
	if businessErr != nil {
		return businessErr
	}
	currentMap := make(ConfigMap, len(current))
	for _, item := range current {
		currentMap[item.TrustTier] = TierConfig{
			MaxOwnedFamilies: item.MaxOwnedFamilies, MaxMembersPerOwnedFamily: item.MaxMembersPerOwnedFamily,
			MaxJoinedFamilies: item.MaxJoinedFamilies, SupportsFeaturePreview: item.SupportsFeaturePreview,
		}
	}
	phoneNow := currentMap[quotaenum.TrustTierPhoneBound]
	interimMembers := wechatWant.MaxMembersPerOwnedFamily
	if phoneNow.MaxMembersPerOwnedFamily < interimMembers {
		interimMembers = phoneNow.MaxMembersPerOwnedFamily
	}
	interim := TierConfig{
		MaxOwnedFamilies:         minInt(wechatWant.MaxOwnedFamilies, phoneWant.MaxOwnedFamilies, currentMap[quotaenum.TrustTierWechatOnly].MaxOwnedFamilies),
		MaxMembersPerOwnedFamily: interimMembers,
		MaxJoinedFamilies:        minInt(wechatWant.MaxJoinedFamilies, phoneWant.MaxJoinedFamilies, currentMap[quotaenum.TrustTierWechatOnly].MaxJoinedFamilies),
		SupportsFeaturePreview:   false,
	}
	if _, businessErr = svc.UpdateConfig(ctx, adminID, role, quotaenum.TrustTierWechatOnly, toDTOValues(interim), audit); businessErr != nil {
		return businessErr
	}
	if _, businessErr = svc.UpdateConfig(ctx, adminID, role, quotaenum.TrustTierPhoneBound, toDTOValues(phoneWant), audit); businessErr != nil {
		return businessErr
	}
	if _, businessErr = svc.UpdateConfig(ctx, adminID, role, quotaenum.TrustTierWechatOnly, toDTOValues(wechatWant), audit); businessErr != nil {
		return businessErr
	}
	return nil
}

func minInt(values ...int) int {
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func ListTierConfigs(t *testing.T, svc quotaservice.Service, adminRole string) ConfigMap {
	t.Helper()
	items, err := svc.ListConfigs(context.Background(), adminRole)
	if err != nil {
		t.Fatalf("ListConfigs: %v", err)
	}
	result := make(ConfigMap, len(items))
	for _, item := range items {
		result[item.TrustTier] = TierConfig{
			MaxOwnedFamilies: item.MaxOwnedFamilies, MaxMembersPerOwnedFamily: item.MaxMembersPerOwnedFamily,
			MaxJoinedFamilies: item.MaxJoinedFamilies, SupportsFeaturePreview: item.SupportsFeaturePreview,
		}
	}
	return result
}

func UpdateTierConfig(t *testing.T, svc quotaservice.Service, adminID uint64, role, tier string, cfg TierConfig) quotavo.ConfigItem {
	t.Helper()
	ensureQuotaConfigLock(t)
	item, err := svc.UpdateConfig(context.Background(), adminID, role, tier, toDTOValues(cfg), quotaservice.AuditInput{IP: "127.0.0.1", UserAgent: "fresh-quota-test"})
	if err != nil {
		t.Fatalf("UpdateConfig %s: %v", tier, err)
	}
	return *item
}

func AssertMigrationDefaults(t *testing.T, svc quotaservice.Service) {
	t.Helper()
	got := ListTierConfigs(t, svc, string(enums.AdminRoleRootAdmin))
	want := MigrationDefaults()
	for tier, expected := range want {
		actual, ok := got[tier]
		if !ok {
			t.Fatalf("missing tier %s in ListConfigs", tier)
		}
		if actual != expected {
			t.Fatalf("tier %s: expected %#v from migration seed, got %#v", tier, expected, actual)
		}
	}
}

func AssertCapabilitiesLimits(t *testing.T, caps *quotavo.Capabilities, expected TierConfig) {
	t.Helper()
	if caps == nil {
		t.Fatal("capabilities is nil")
	}
	if caps.Limits.MaxOwnedFamilies != expected.MaxOwnedFamilies ||
		caps.Limits.MaxMembersPerOwnedFamily != expected.MaxMembersPerOwnedFamily ||
		caps.Limits.MaxJoinedFamilies != expected.MaxJoinedFamilies ||
		caps.Limits.SupportsFeaturePreview != expected.SupportsFeaturePreview {
		t.Fatalf("capabilities limits %#v != admin config %#v", caps.Limits, expected)
	}
}

func toDTOValues(cfg TierConfig) quotadto.ConfigValues {
	return quotadto.ConfigValues{
		MaxOwnedFamilies: cfg.MaxOwnedFamilies, MaxMembersPerOwnedFamily: cfg.MaxMembersPerOwnedFamily,
		MaxJoinedFamilies: cfg.MaxJoinedFamilies, SupportsFeaturePreview: cfg.SupportsFeaturePreview,
	}
}

// DynamicTestConfigs are non-default values used to prove limits are not hard-coded.
var DynamicTestConfigs = ConfigMap{
	quotaenum.TrustTierWechatOnly: {
		MaxOwnedFamilies: 2, MaxMembersPerOwnedFamily: 6, MaxJoinedFamilies: 2, SupportsFeaturePreview: false,
	},
	quotaenum.TrustTierPhoneBound: {
		MaxOwnedFamilies: 3, MaxMembersPerOwnedFamily: 8, MaxJoinedFamilies: 4, SupportsFeaturePreview: true,
	},
}

func ApplyDynamicTestConfigs(t *testing.T, svc quotaservice.Service, adminID uint64) {
	t.Helper()
	// Tier-order validation requires staged updates from migration defaults (1/10/1 + 1/20/5).
	wechatTarget := DynamicTestConfigs[quotaenum.TrustTierWechatOnly]
	phoneTarget := DynamicTestConfigs[quotaenum.TrustTierPhoneBound]
	UpdateTierConfig(t, svc, adminID, string(enums.AdminRoleRootAdmin), quotaenum.TrustTierWechatOnly, TierConfig{
		MaxOwnedFamilies: 1, MaxMembersPerOwnedFamily: wechatTarget.MaxMembersPerOwnedFamily, MaxJoinedFamilies: 1,
		SupportsFeaturePreview: false,
	})
	UpdateTierConfig(t, svc, adminID, string(enums.AdminRoleSuperAdmin), quotaenum.TrustTierPhoneBound, phoneTarget)
	UpdateTierConfig(t, svc, adminID, string(enums.AdminRoleRootAdmin), quotaenum.TrustTierWechatOnly, wechatTarget)
}

func RestoreMigrationDefaults(t *testing.T, svc quotaservice.Service, adminID uint64) {
	t.Helper()
	if err := restoreConfigsStaged(svc, adminID, string(enums.AdminRoleRootAdmin), MigrationDefaults()); err != nil {
		t.Fatalf("restore migration defaults: %v", err)
	}
}
