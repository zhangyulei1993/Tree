package service_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"gorm.io/gorm"

	quotadto "tree/backend/internal/accountquota/dto"
	quotaenum "tree/backend/internal/accountquota/enum"
	quotamodel "tree/backend/internal/accountquota/model"
	quotarepo "tree/backend/internal/accountquota/repository"
	quotaservice "tree/backend/internal/accountquota/service"
	quotavo "tree/backend/internal/accountquota/vo"
	"tree/backend/internal/common/config"
	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/database"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/security"
	coredto "tree/backend/internal/family/core/dto"
	familymodel "tree/backend/internal/family/core/model"
	familyrepo "tree/backend/internal/family/core/repository"
	familyservice "tree/backend/internal/family/core/service"
	memberdto "tree/backend/internal/family/member/dto"
	membermodel "tree/backend/internal/family/member/model"
	memberrepo "tree/backend/internal/family/member/repository"
	memberservice "tree/backend/internal/family/member/service"
	rolemodel "tree/backend/internal/family/role/model"
	usermodel "tree/backend/internal/user/model"
)

func quotaTestDB(t *testing.T) *gorm.DB {
	t.Helper()
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
	seedQuotaConfigs(t, tx)
	return tx
}

func seedQuotaConfigs(t *testing.T, tx *gorm.DB) {
	t.Helper()
	if err := tx.Where("1 = 1").Delete(&quotamodel.AccountQuotaConfig{}).Error; err != nil {
		t.Fatalf("clear quota configs: %v", err)
	}
	rows := []quotamodel.AccountQuotaConfig{
		{TrustTier: quotaenum.TrustTierWechatOnly, MaxOwnedFamilies: 1, MaxMembersPerOwnedFamily: 10, MaxJoinedFamilies: 1},
		{TrustTier: quotaenum.TrustTierPhoneBound, MaxOwnedFamilies: 1, MaxMembersPerOwnedFamily: 20, MaxJoinedFamilies: 5},
	}
	if err := tx.Create(&rows).Error; err != nil {
		t.Fatalf("seed quota configs: %v", err)
	}
}

func createQuotaUser(t *testing.T, tx *gorm.DB, phoneLoginEnabled bool, nickname string) *usermodel.User {
	t.Helper()
	nick := nickname
	user := &usermodel.User{
		AccountOrigin: "P0_TEST", RegisterClient: "P0_TEST",
		Status: string(enums.StatusActive), Nickname: &nick,
	}
	if nickname == "" {
		user.Nickname = nil
	}
	if err := tx.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if phoneLoginEnabled {
		phone := fmt.Sprintf("138%08d", user.ID%100000000)
		phoneHash := security.PhoneHash(phone)
		passwordHash, err := security.HashPassword("quota-test-password")
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

func TestCapabilitiesReturnsTierLimitsAndUsage(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	user := createQuotaUser(t, tx, false, "测试用户")
	svc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	result, err := svc.GetCapabilities(ctx, user.ID)
	if err != nil || result.TrustTier != quotaenum.TrustTierWechatOnly ||
		result.Limits.MaxOwnedFamilies != 1 || result.Limits.MaxMembersPerOwnedFamily != 10 {
		t.Fatalf("unexpected capabilities: %#v %#v", result, err)
	}
}

func TestPhoneVerifiedUserUsesHigherTier(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	user := createQuotaUser(t, tx, true, "高级用户")
	svc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	result, err := svc.GetCapabilities(ctx, user.ID)
	if err != nil || result.TrustTier != quotaenum.TrustTierPhoneBound ||
		result.Limits.MaxMembersPerOwnedFamily != 20 || result.Limits.MaxJoinedFamilies != 5 {
		t.Fatalf("unexpected phone verified capabilities: %#v %#v", result, err)
	}
}

func TestOwnedFamilyQuotaBlocksSecondCreate(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	user := createQuotaUser(t, tx, false, "创始人")
	quotaSvc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	familySvc := familyservice.NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), nil, quotaSvc, contentsafety.AlwaysPass())
	gender := string(enums.GenderMale)
	if _, err := familySvc.Create(ctx, user.ID, coredto.CreateFamilyRequest{Surname: "张", FounderGender: &gender}, familyservice.AuditInput{}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if _, err := familySvc.Create(ctx, user.ID, coredto.CreateFamilyRequest{Surname: "李", FounderGender: &gender}, familyservice.AuditInput{}); err == nil || err.Code != 41502 {
		t.Fatalf("expected owned quota error, got %#v", err)
	}
}

func TestMemberQuotaBoundaryWechatOnly(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	user := createQuotaUser(t, tx, false, "成员上限")
	quotaSvc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	familySvc := familyservice.NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), nil, quotaSvc, contentsafety.AlwaysPass())
	memberSvc := memberservice.NewMemberService(tx, memberrepo.NewMemberRepository(tx), allowAllMemberPerm{}, quotaSvc, contentsafety.AlwaysPass())
	gender := string(enums.GenderMale)
	created, err := familySvc.Create(ctx, user.ID, coredto.CreateFamilyRequest{Surname: "王", FounderGender: &gender}, familyservice.AuditInput{})
	if err != nil {
		t.Fatalf("create family: %v", err)
	}
	for i := 0; i < 9; i++ {
		if _, err := memberSvc.Create(ctx, user.ID, created.ID, memberdto.CreateMemberRequest{Name: "成员" + string(rune('A'+i))}, memberservice.AuditInput{}); err != nil {
			t.Fatalf("create member %d: %v", i, err)
		}
	}
	if _, err := memberSvc.Create(ctx, user.ID, created.ID, memberdto.CreateMemberRequest{Name: "超额成员"}, memberservice.AuditInput{}); err == nil || err.Code != 41504 {
		t.Fatalf("expected member quota error, got %#v", err)
	}
}

func TestProfileIncompleteBlocksCreateFamily(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	user := createQuotaUser(t, tx, false, "")
	quotaSvc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	familySvc := familyservice.NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), nil, quotaSvc, contentsafety.AlwaysPass())
	gender := string(enums.GenderMale)
	if _, err := familySvc.Create(ctx, user.ID, coredto.CreateFamilyRequest{Surname: "赵", FounderGender: &gender}, familyservice.AuditInput{}); err == nil || err.Code != 41501 {
		t.Fatalf("expected profile incomplete, got %#v", err)
	}
}

func TestConcurrentCreateFamilyDoesNotExceedQuota(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	user := createQuotaUser(t, tx, false, "并发创始人")
	quotaSvc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	gender := string(enums.GenderMale)
	var wg sync.WaitGroup
	successes := 0
	var mu sync.Mutex
	var dbMu sync.Mutex
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dbMu.Lock()
			familySvc := familyservice.NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), nil, quotaSvc, contentsafety.AlwaysPass())
			_, err := familySvc.Create(ctx, user.ID, coredto.CreateFamilyRequest{Surname: "并发", FounderGender: &gender}, familyservice.AuditInput{})
			dbMu.Unlock()
			if err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if successes != 1 {
		t.Fatalf("expected exactly one successful family create, got %d", successes)
	}
}

func quotaValues(owned, members, joined int) quotadto.ConfigValues {
	return quotadto.ConfigValues{
		MaxOwnedFamilies: owned, MaxMembersPerOwnedFamily: members, MaxJoinedFamilies: joined,
	}
}

func TestPlatformAdminCannotUpdateConfig(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	svc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	_, err := svc.UpdateConfig(ctx, 2, string(enums.AdminRolePlatformAdmin), quotaenum.TrustTierWechatOnly, quotaValues(1, 10, 1), quotaservice.AuditInput{})
	if err == nil || err.Code != 41506 {
		t.Fatalf("expected forbidden, got %#v", err)
	}
}

func TestPhoneVerifiedConfigCannotBeLowerThanWechatOnly(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	svc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	_, err := svc.UpdateConfig(ctx, 1, string(enums.AdminRoleSuperAdmin), quotaenum.TrustTierPhoneBound, quotaValues(1, 5, 5), quotaservice.AuditInput{})
	if err == nil || err.Code != 41507 {
		t.Fatalf("expected tier order error, got %#v", err)
	}
}

func TestLoweringConfigDoesNotDeleteExistingFamilies(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	user := createQuotaUser(t, tx, false, "保留数据")
	quotaSvc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	familySvc := familyservice.NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), allowAllFamilyPerm{}, quotaSvc, contentsafety.AlwaysPass())
	gender := string(enums.GenderMale)
	created, err := familySvc.Create(ctx, user.ID, coredto.CreateFamilyRequest{Surname: "陈", FounderGender: &gender}, familyservice.AuditInput{})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := quotaSvc.UpdateConfig(ctx, 1, string(enums.AdminRoleRootAdmin), quotaenum.TrustTierWechatOnly, quotaValues(0, 1, 0), quotaservice.AuditInput{}); err != nil {
		t.Fatalf("lower config: %v", err)
	}
	detail, err := familySvc.Detail(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("detail after lower config should still work: %v", err)
	}
	if detail == nil || detail.ID != created.ID {
		t.Fatalf("family should remain readable")
	}
	if _, err := familySvc.Create(ctx, user.ID, coredto.CreateFamilyRequest{Surname: "新", FounderGender: &gender}, familyservice.AuditInput{}); err == nil || err.Code != 41502 {
		t.Fatalf("new create should be blocked after lowering, got %#v", err)
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

type allowAllMemberPerm struct{}

func (allowAllMemberPerm) IsFamilyAdmin(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllMemberPerm) IsFamilyMember(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllMemberPerm) CanManageFamily(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllMemberPerm) CanCreateMember(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllMemberPerm) CanEditMember(context.Context, uint64, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllMemberPerm) CanDeleteMember(context.Context, uint64, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllMemberPerm) CanManageRelationships(context.Context, uint64, uint64) (bool, error) {
	return true, nil
}
func (allowAllMemberPerm) GetActiveLink(context.Context, uint64, uint64) (*rolemodel.FamilyMemberUserLink, error) {
	return nil, nil
}

func TestDefaultLimitsWhenConfigMissing(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	if err := tx.Where("1 = 1").Delete(&quotamodel.AccountQuotaConfig{}).Error; err != nil {
		t.Fatalf("delete configs: %v", err)
	}
	repo := quotarepo.NewRepository(tx)
	limits, err := repo.GetLimitsForTier(ctx, quotaenum.TrustTierWechatOnly)
	if err != nil || limits.MaxMembersPerOwnedFamily != 10 {
		t.Fatalf("unexpected default limits: %#v %v", limits, err)
	}
}

func TestNicknameTrimmedForProfileComplete(t *testing.T) {
	user := &usermodel.User{Nickname: stringPtr("  ")}
	if quotaservice.IsProfileComplete(user) {
		t.Fatalf("whitespace nickname should be incomplete")
	}
	user.Nickname = stringPtr("  松山  ")
	if !quotaservice.IsProfileComplete(user) || strings.TrimSpace(*user.Nickname) != "松山" {
		t.Fatalf("trimmed nickname should be complete")
	}
}

func stringPtr(v string) *string { return &v }

func TestWechatOnlyConfigCannotExceedPhoneVerified(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	svc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	_, err := svc.UpdateConfig(ctx, 1, string(enums.AdminRoleRootAdmin), quotaenum.TrustTierWechatOnly, quotaValues(2, 25, 6), quotaservice.AuditInput{})
	if err == nil || err.Code != 41507 {
		t.Fatalf("expected tier order error when raising wechat only above phone verified, got %#v", err)
	}
}

func TestPreviewImpactDedupesUsers(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	repo := quotarepo.NewRepository(tx)
	baselineDistinct, err := repo.CountUsersExceedingDistinct(ctx, quotaenum.TrustTierWechatOnly, 0, 0)
	if err != nil {
		t.Fatalf("baseline distinct: %v", err)
	}
	user := createQuotaUser(t, tx, false, "重复统计")
	other := createQuotaUser(t, tx, false, "其他创始人")
	family1 := &familymodel.Family{FamilyName: "周家", FamilySurname: "周", Status: string(enums.StatusNormal), PublicDisplayStatus: string(enums.StatusPrivate), Searchable: true, GraphVersion: 1}
	family2 := &familymodel.Family{FamilyName: "吴家", FamilySurname: "吴", Status: string(enums.StatusNormal), PublicDisplayStatus: string(enums.StatusPrivate), Searchable: true, GraphVersion: 1}
	if err := tx.Create(family1).Error; err != nil {
		t.Fatalf("create family1: %v", err)
	}
	if err := tx.Create(family2).Error; err != nil {
		t.Fatalf("create family2: %v", err)
	}
	founderMember1 := &membermodel.FamilyMember{FamilyID: family1.ID, DisplayName: "周创始人", Status: string(enums.StatusActive)}
	founderMember2 := &membermodel.FamilyMember{FamilyID: family2.ID, DisplayName: "吴创始人", Status: string(enums.StatusActive)}
	member2 := &membermodel.FamilyMember{FamilyID: family2.ID, DisplayName: "访客成员", Status: string(enums.StatusActive)}
	for _, row := range []*membermodel.FamilyMember{founderMember1, founderMember2, member2} {
		if err := tx.Create(row).Error; err != nil {
			t.Fatalf("create member: %v", err)
		}
	}
	links := []rolemodel.FamilyMemberUserLink{
		{FamilyID: family1.ID, MemberID: founderMember1.ID, UserID: user.ID, LinkStatus: string(enums.StatusActive), LinkSource: "TEST", FamilyRole: string(enums.FamilyRoleFounder)},
		{FamilyID: family2.ID, MemberID: member2.ID, UserID: user.ID, LinkStatus: string(enums.StatusActive), LinkSource: "TEST", FamilyRole: string(enums.FamilyRoleMember)},
		{FamilyID: family2.ID, MemberID: founderMember2.ID, UserID: other.ID, LinkStatus: string(enums.StatusActive), LinkSource: "TEST", FamilyRole: string(enums.FamilyRoleFounder)},
	}
	if err := tx.Create(&links).Error; err != nil {
		t.Fatalf("create links: %v", err)
	}
	owned, err := repo.CountUsersExceedingOwned(ctx, quotaenum.TrustTierWechatOnly, 0)
	if err != nil {
		t.Fatalf("count owned: %v", err)
	}
	joined, err := repo.CountUsersExceedingJoined(ctx, quotaenum.TrustTierWechatOnly, 0)
	if err != nil {
		t.Fatalf("count joined: %v", err)
	}
	distinct, err := repo.CountUsersExceedingDistinct(ctx, quotaenum.TrustTierWechatOnly, 0, 0)
	if err != nil {
		t.Fatalf("count distinct: %v", err)
	}
	if owned < 1 || joined < 1 {
		t.Fatalf("expected seeded user to exceed owned and joined, got owned=%d joined=%d", owned, joined)
	}
	if distinct >= owned+joined {
		t.Fatalf("distinct should dedupe users appearing in both owned and joined sets, got distinct=%d owned=%d joined=%d", distinct, owned, joined)
	}
	if distinct < baselineDistinct+1 {
		t.Fatalf("expected at least one new affected user after seeding, got distinct=%d baseline=%d", distinct, baselineDistinct)
	}
}

func TestFindFounderNotFoundReturnsError(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	svc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	err := svc.AssertCanAddMember(ctx, tx, 99999, 1)
	if !errors.Is(err, quotarepo.ErrFounderNotFound) {
		t.Fatalf("expected ErrFounderNotFound, got %v", err)
	}
	businessErr := svc.MapQuotaError(err)
	if businessErr == nil || businessErr.Code != apperrors.CodeResourceNotFound {
		t.Fatalf("expected resource not found mapping, got %#v", businessErr)
	}
}

func TestExplicitZeroConfigAccepted(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	svc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	item, err := svc.UpdateConfig(ctx, 1, string(enums.AdminRoleRootAdmin), quotaenum.TrustTierWechatOnly, quotaValues(1, 10, 0), quotaservice.AuditInput{})
	if err != nil {
		t.Fatalf("update with explicit zero joined: %v", err)
	}
	if item.MaxJoinedFamilies != 0 {
		t.Fatalf("expected zero joined families saved, got %d", item.MaxJoinedFamilies)
	}
}

func quotaTestDBDirect(t *testing.T) *gorm.DB {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	db, err := database.Init(context.Background(), cfg.MySQL)
	if err != nil {
		t.Skipf("mysql unavailable: %v", err)
	}
	seedQuotaConfigs(t, db)
	t.Cleanup(func() { seedQuotaConfigs(t, db) })
	return db
}

func TestConcurrentUpdateConfigPreservesTierOrdering(t *testing.T) {
	ctx := context.Background()
	db := quotaTestDBDirect(t)
	svc := quotaservice.NewService(db, quotarepo.NewRepository(db))
	type updateOutcome struct {
		item *quotavo.ConfigItem
		err  *apperrors.BusinessError
	}
	outcomes := make([]updateOutcome, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		item, err := svc.UpdateConfig(ctx, 1, string(enums.AdminRoleRootAdmin), quotaenum.TrustTierWechatOnly, quotaValues(1, 19, 1), quotaservice.AuditInput{})
		outcomes[0] = updateOutcome{item: item, err: err}
	}()
	go func() {
		defer wg.Done()
		item, err := svc.UpdateConfig(ctx, 1, string(enums.AdminRoleRootAdmin), quotaenum.TrustTierPhoneBound, quotaValues(1, 17, 5), quotaservice.AuditInput{})
		outcomes[1] = updateOutcome{item: item, err: err}
	}()
	wg.Wait()

	successes := 0
	tierOrderFailures := 0
	for i, outcome := range outcomes {
		if outcome.err == nil {
			if outcome.item == nil {
				t.Fatalf("goroutine %d succeeded without config item", i)
			}
			successes++
			continue
		}
		if outcome.err.Code == apperrors.CodeQuotaConfigTierOrder {
			tierOrderFailures++
			continue
		}
		t.Fatalf("goroutine %d unexpected error: %#v", i, outcome.err)
	}
	if successes != 1 {
		t.Fatalf("expected exactly 1 successful update, got %d (outcomes=%#v)", successes, outcomes)
	}
	if tierOrderFailures != 1 {
		t.Fatalf("expected exactly 1 tier order failure (41507), got %d (outcomes=%#v)", tierOrderFailures, outcomes)
	}

	repo := quotarepo.NewRepository(db)
	wechat, err := repo.FindConfigByTier(ctx, quotaenum.TrustTierWechatOnly)
	if err != nil {
		t.Fatalf("load wechat config: %v", err)
	}
	phone, err := repo.FindConfigByTier(ctx, quotaenum.TrustTierPhoneBound)
	if err != nil {
		t.Fatalf("load phone config: %v", err)
	}
	if phone.MaxOwnedFamilies < wechat.MaxOwnedFamilies ||
		phone.MaxMembersPerOwnedFamily < wechat.MaxMembersPerOwnedFamily ||
		phone.MaxJoinedFamilies < wechat.MaxJoinedFamilies {
		t.Fatalf("tier order violated after concurrent updates: wechat=%#v phone=%#v", wechat, phone)
	}
}

func seedDissolvedFamily(t *testing.T, tx *gorm.DB, founder *usermodel.User, extraMembers int, joinedUsers []*usermodel.User) uint64 {
	t.Helper()
	family := &familymodel.Family{
		FamilyName: "待恢复家", FamilySurname: "测", Status: string(enums.StatusDissolved),
		PublicDisplayStatus: string(enums.StatusPrivate), Searchable: false, GraphVersion: 1,
	}
	if err := tx.Create(family).Error; err != nil {
		t.Fatalf("create dissolved family: %v", err)
	}
	founderMember := &membermodel.FamilyMember{
		FamilyID: family.ID, DisplayName: "创始人", Status: string(enums.StatusActive), Gender: string(enums.GenderMale),
	}
	if err := tx.Create(founderMember).Error; err != nil {
		t.Fatalf("create founder member: %v", err)
	}
	founderLink := &rolemodel.FamilyMemberUserLink{
		FamilyID: family.ID, MemberID: founderMember.ID, UserID: founder.ID,
		LinkStatus: string(enums.StatusActive), LinkSource: "TEST", FamilyRole: string(enums.FamilyRoleFounder),
	}
	if err := tx.Create(founderLink).Error; err != nil {
		t.Fatalf("create founder link: %v", err)
	}
	for i := 0; i < extraMembers; i++ {
		member := &membermodel.FamilyMember{
			FamilyID: family.ID, DisplayName: "额外成员", Status: string(enums.StatusActive), Gender: string(enums.GenderMale),
		}
		if err := tx.Create(member).Error; err != nil {
			t.Fatalf("create extra member: %v", err)
		}
	}
	for i, user := range joinedUsers {
		member := &membermodel.FamilyMember{
			FamilyID: family.ID, DisplayName: "加入成员", Status: string(enums.StatusActive), Gender: string(enums.GenderMale),
		}
		if err := tx.Create(member).Error; err != nil {
			t.Fatalf("create joined member %d: %v", i, err)
		}
		link := &rolemodel.FamilyMemberUserLink{
			FamilyID: family.ID, MemberID: member.ID, UserID: user.ID,
			LinkStatus: string(enums.StatusActive), LinkSource: "TEST", FamilyRole: string(enums.FamilyRoleMember),
		}
		if err := tx.Create(link).Error; err != nil {
			t.Fatalf("create joined link %d: %v", i, err)
		}
	}
	return family.ID
}

func seedFounderFamily(t *testing.T, tx *gorm.DB, founder *usermodel.User) *familymodel.Family {
	t.Helper()
	family := &familymodel.Family{
		FamilyName: "现有家", FamilySurname: "李", Status: string(enums.StatusNormal),
		PublicDisplayStatus: string(enums.StatusPrivate), Searchable: true, GraphVersion: 1,
	}
	if err := tx.Create(family).Error; err != nil {
		t.Fatalf("create family: %v", err)
	}
	member := &membermodel.FamilyMember{
		FamilyID: family.ID, DisplayName: "创始人", Status: string(enums.StatusActive), Gender: string(enums.GenderMale),
	}
	if err := tx.Create(member).Error; err != nil {
		t.Fatalf("create member: %v", err)
	}
	link := &rolemodel.FamilyMemberUserLink{
		FamilyID: family.ID, MemberID: member.ID, UserID: founder.ID,
		LinkStatus: string(enums.StatusActive), LinkSource: "TEST", FamilyRole: string(enums.FamilyRoleFounder),
	}
	if err := tx.Create(link).Error; err != nil {
		t.Fatalf("create link: %v", err)
	}
	return family
}

func seedJoinedMembership(t *testing.T, tx *gorm.DB, user *usermodel.User) {
	t.Helper()
	family := &familymodel.Family{
		FamilyName: "已加入家", FamilySurname: "王", Status: string(enums.StatusNormal),
		PublicDisplayStatus: string(enums.StatusPrivate), Searchable: true, GraphVersion: 1,
	}
	if err := tx.Create(family).Error; err != nil {
		t.Fatalf("create joined family: %v", err)
	}
	other := createQuotaUser(t, tx, false, "其他创始人")
	founderMember := &membermodel.FamilyMember{
		FamilyID: family.ID, DisplayName: "王创始人", Status: string(enums.StatusActive), Gender: string(enums.GenderMale),
	}
	member := &membermodel.FamilyMember{
		FamilyID: family.ID, DisplayName: "加入者", Status: string(enums.StatusActive), Gender: string(enums.GenderMale),
	}
	for _, row := range []*membermodel.FamilyMember{founderMember, member} {
		if err := tx.Create(row).Error; err != nil {
			t.Fatalf("create member: %v", err)
		}
	}
	links := []rolemodel.FamilyMemberUserLink{
		{FamilyID: family.ID, MemberID: founderMember.ID, UserID: other.ID, LinkStatus: string(enums.StatusActive), LinkSource: "TEST", FamilyRole: string(enums.FamilyRoleFounder)},
		{FamilyID: family.ID, MemberID: member.ID, UserID: user.ID, LinkStatus: string(enums.StatusActive), LinkSource: "TEST", FamilyRole: string(enums.FamilyRoleMember)},
	}
	if err := tx.Create(&links).Error; err != nil {
		t.Fatalf("create links: %v", err)
	}
}

func assertQuotaCode(t *testing.T, svc quotaservice.Service, err error, code apperrors.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected quota error %d, got nil", code)
	}
	businessErr := svc.MapQuotaError(err)
	if businessErr == nil || businessErr.Code != code {
		t.Fatalf("expected quota error %d, got %#v from %v", code, businessErr, err)
	}
}

func TestAssertCanRestoreFamilyOwnedExceeded(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	svc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	founder := createQuotaUser(t, tx, false, "双家庭创始人")
	seedFounderFamily(t, tx, founder)
	dissolvedID := seedDissolvedFamily(t, tx, founder, 0, nil)
	assertQuotaCode(t, svc, svc.AssertCanRestoreFamily(ctx, tx, dissolvedID), apperrors.CodeQuotaOwnedFamiliesExceeded)
}

func TestAssertCanRestoreFamilyMembersExceeded(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	svc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	founder := createQuotaUser(t, tx, false, "成员超限创始人")
	dissolvedID := seedDissolvedFamily(t, tx, founder, 10, nil)
	assertQuotaCode(t, svc, svc.AssertCanRestoreFamily(ctx, tx, dissolvedID), apperrors.CodeQuotaMembersExceeded)
}

func TestAssertCanRestoreFamilyJoinedExceeded(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	svc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	founder := createQuotaUser(t, tx, false, "恢复创始人")
	joinedUser := createQuotaUser(t, tx, false, "加入已满用户")
	seedJoinedMembership(t, tx, joinedUser)
	dissolvedID := seedDissolvedFamily(t, tx, founder, 0, []*usermodel.User{joinedUser})
	assertQuotaCode(t, svc, svc.AssertCanRestoreFamily(ctx, tx, dissolvedID), apperrors.CodeQuotaJoinedFamiliesExceeded)
}

func TestAssertCanRestoreFamilyBoundarySuccess(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	svc := quotaservice.NewService(tx, quotarepo.NewRepository(tx))
	founder := createQuotaUser(t, tx, false, "边界创始人")
	joinedUser := createQuotaUser(t, tx, false, "边界加入者")
	dissolvedID := seedDissolvedFamily(t, tx, founder, 8, []*usermodel.User{joinedUser})
	if err := svc.AssertCanRestoreFamily(ctx, tx, dissolvedID); err != nil {
		t.Fatalf("expected restore quota check to pass at boundary, got %v", err)
	}
}

func TestJoinedCountIncludesDissolutionPending(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	repo := quotarepo.NewRepository(tx)
	user := createQuotaUser(t, tx, false, "待解散成员")
	other := createQuotaUser(t, tx, false, "待解散创始人")
	family := &familymodel.Family{
		FamilyName: "待解散家", FamilySurname: "赵", Status: "DISSOLUTION_PENDING",
		PublicDisplayStatus: string(enums.StatusPrivate), Searchable: true, GraphVersion: 1,
	}
	if err := tx.Create(family).Error; err != nil {
		t.Fatalf("create family: %v", err)
	}
	founderMember := &membermodel.FamilyMember{
		FamilyID: family.ID, DisplayName: "创始人", Status: string(enums.StatusActive), Gender: string(enums.GenderMale),
	}
	member := &membermodel.FamilyMember{
		FamilyID: family.ID, DisplayName: "成员", Status: string(enums.StatusActive), Gender: string(enums.GenderMale),
	}
	for _, row := range []*membermodel.FamilyMember{founderMember, member} {
		if err := tx.Create(row).Error; err != nil {
			t.Fatalf("create member: %v", err)
		}
	}
	links := []rolemodel.FamilyMemberUserLink{
		{FamilyID: family.ID, MemberID: founderMember.ID, UserID: other.ID, LinkStatus: string(enums.StatusActive), LinkSource: "TEST", FamilyRole: string(enums.FamilyRoleFounder)},
		{FamilyID: family.ID, MemberID: member.ID, UserID: user.ID, LinkStatus: string(enums.StatusActive), LinkSource: "TEST", FamilyRole: string(enums.FamilyRoleMember)},
	}
	if err := tx.Create(&links).Error; err != nil {
		t.Fatalf("create links: %v", err)
	}
	count, err := repo.CountJoinedFamilies(ctx, user.ID)
	if err != nil {
		t.Fatalf("count joined: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected DISSOLUTION_PENDING family to count toward joined usage, got %d", count)
	}
}

func TestCooldownFamiliesStillCountTowardQuotaUsage(t *testing.T) {
	ctx := context.Background()
	tx := quotaTestDB(t)
	repo := quotarepo.NewRepository(tx)
	founder := createQuotaUser(t, tx, false, "冷静期创始人")
	memberUser := createQuotaUser(t, tx, false, "冷静期成员")
	family := &familymodel.Family{
		FamilyName: "冷静期家庭", FamilySurname: "周", Status: "DISSOLUTION_COOLDOWN",
		PublicDisplayStatus: string(enums.StatusPrivate), Searchable: false, GraphVersion: 1,
	}
	if err := tx.Create(family).Error; err != nil {
		t.Fatalf("create family: %v", err)
	}
	founderMember := &membermodel.FamilyMember{
		FamilyID: family.ID, DisplayName: "创建者", Status: string(enums.StatusActive), Gender: string(enums.GenderMale),
	}
	member := &membermodel.FamilyMember{
		FamilyID: family.ID, DisplayName: "成员", Status: string(enums.StatusActive), Gender: string(enums.GenderFemale),
	}
	for _, row := range []*membermodel.FamilyMember{founderMember, member} {
		if err := tx.Create(row).Error; err != nil {
			t.Fatalf("create member: %v", err)
		}
	}
	links := []rolemodel.FamilyMemberUserLink{
		{FamilyID: family.ID, MemberID: founderMember.ID, UserID: founder.ID, LinkStatus: string(enums.StatusActive), LinkSource: "TEST", FamilyRole: string(enums.FamilyRoleFounder)},
		{FamilyID: family.ID, MemberID: member.ID, UserID: memberUser.ID, LinkStatus: string(enums.StatusActive), LinkSource: "TEST", FamilyRole: string(enums.FamilyRoleMember)},
	}
	if err := tx.Create(&links).Error; err != nil {
		t.Fatalf("create links: %v", err)
	}
	owned, err := repo.CountOwnedFamilies(ctx, founder.ID)
	if err != nil {
		t.Fatalf("count owned: %v", err)
	}
	joined, err := repo.CountJoinedFamilies(ctx, memberUser.ID)
	if err != nil {
		t.Fatalf("count joined: %v", err)
	}
	if owned != 1 || joined != 1 {
		t.Fatalf("expected cooldown family to count toward quota usage, got owned=%d joined=%d", owned, joined)
	}
}
