package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/family/core/dto"
	familymodel "tree/backend/internal/family/core/model"
	familyrepo "tree/backend/internal/family/core/repository"
)

func publicSearchToken(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("PUBSRCH_%d", time.Now().UnixNano())
}

func insertPublicSearchFamily(t *testing.T, tx *gorm.DB, family familymodel.Family) familymodel.Family {
	t.Helper()
	if family.Status == "" {
		family.Status = "NORMAL"
	}
	if family.PublicDisplayStatus == "" {
		family.PublicDisplayStatus = "PRIVATE"
	}
	if family.TreeMode == "" {
		family.TreeMode = "LIST_TREE"
	}
	if family.GraphVersion == 0 {
		family.GraphVersion = 1
	}

	wantContactVisible := family.PublicContactVisible
	wantSearchable := family.Searchable

	if err := tx.Create(&family).Error; err != nil {
		t.Fatalf("insert family: %v", err)
	}

	// GORM Create 会跳过 bool 零值，导致 DB 落库为 default:true。
	updates := map[string]any{}
	if !wantContactVisible {
		updates["public_contact_visible"] = false
		family.PublicContactVisible = false
	}
	if !wantSearchable {
		updates["searchable"] = false
		family.Searchable = false
	}
	if len(updates) > 0 {
		if err := tx.Model(&familymodel.Family{}).Where("id = ?", family.ID).Updates(updates).Error; err != nil {
			t.Fatalf("patch family bool columns: %v", err)
		}
	}

	return family
}

func publicSearchService(tx *gorm.DB) FamilyService {
	return NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), nil, nil)
}

func TestListPublicFamiliesVisibilityRules(t *testing.T) {
	ctx := context.Background()
	tx := transactionalTestDB(t)
	token := publicSearchToken(t)
	now := time.Now().UTC().Truncate(time.Second)
	approvedAt := now.Add(-2 * time.Hour)

	visible := insertPublicSearchFamily(t, tx, familymodel.Family{
		FamilyName: token + "_visible_张氏家族", FamilySurname: "张", NativePlace: strPtr("山东济南"),
		RegionText: strPtr("山东省济南市"), Description: strPtr(token + " 公开张氏家族"),
		Status: "NORMAL", Searchable: true, PublicDisplayStatus: "APPROVED", PublicApprovedAt: &approvedAt,
	})
	insertPublicSearchFamily(t, tx, familymodel.Family{
		FamilyName: token + "_private_李氏宗亲", FamilySurname: "李", Description: strPtr(token),
		Status: "NORMAL", Searchable: true, PublicDisplayStatus: "PRIVATE",
	})
	insertPublicSearchFamily(t, tx, familymodel.Family{
		FamilyName: token + "_pending_王氏待审", FamilySurname: "王", Description: strPtr(token),
		Status: "NORMAL", Searchable: true, PublicDisplayStatus: "PENDING",
	})
	insertPublicSearchFamily(t, tx, familymodel.Family{
		FamilyName: token + "_rejected_赵氏驳回", FamilySurname: "赵", Description: strPtr(token),
		Status: "NORMAL", Searchable: true, PublicDisplayStatus: "REJECTED",
	})
	insertPublicSearchFamily(t, tx, familymodel.Family{
		FamilyName: token + "_taken_down_陈氏下架", FamilySurname: "陈", Description: strPtr(token),
		Status: "NORMAL", Searchable: true, PublicDisplayStatus: "TAKEN_DOWN",
	})
	insertPublicSearchFamily(t, tx, familymodel.Family{
		FamilyName: token + "_hidden_周氏隐藏", FamilySurname: "周", Description: strPtr(token),
		Status: "NORMAL", Searchable: false, PublicDisplayStatus: "APPROVED",
	})
	insertPublicSearchFamily(t, tx, familymodel.Family{
		FamilyName: token + "_disabled_吴氏停用", FamilySurname: "吴", Description: strPtr(token),
		Status: "DISABLED", Searchable: true, PublicDisplayStatus: "APPROVED",
	})

	service := publicSearchService(tx)
	result, businessErr := service.ListPublicFamilies(ctx, dto.ListPublicFamiliesQuery{
		Keyword: token, Page: 1, PageSize: 20,
	})
	if businessErr != nil {
		t.Fatalf("ListPublicFamilies: %v", businessErr)
	}
	if result.Total != 1 || len(result.Items) != 1 {
		t.Fatalf("expected only one searchable approved family for token %s, got total=%d items=%d", token, result.Total, len(result.Items))
	}
	if result.Items[0].ID != visible.ID || result.Items[0].FamilyName != visible.FamilyName {
		t.Fatalf("unexpected visible family: %#v", result.Items[0])
	}
}

func TestListPublicFamiliesFiltersAndPagination(t *testing.T) {
	ctx := context.Background()
	tx := transactionalTestDB(t)
	token := publicSearchToken(t)
	now := time.Now().UTC().Truncate(time.Second)

	for index, name := range []string{"张甲家族", "张乙家族", "李甲家族"} {
		approvedAt := now.Add(time.Duration(index) * time.Hour)
		surname := "张"
		region := strPtr("浙江省杭州市")
		if index == 2 {
			surname = "李"
			region = strPtr("河南省郑州市")
		}
		insertPublicSearchFamily(t, tx, familymodel.Family{
			FamilyName: token + "_" + name, FamilySurname: surname, NativePlace: strPtr("浙江杭州"),
			RegionText: region, Description: strPtr(token + " 测试家庭 " + name),
			Status: "NORMAL", Searchable: true, PublicDisplayStatus: "APPROVED", PublicApprovedAt: &approvedAt,
		})
	}

	service := publicSearchService(tx)
	scope := dto.ListPublicFamiliesQuery{Keyword: token, Page: 1, PageSize: 20}

	keywordResult, err := service.ListPublicFamilies(ctx, dto.ListPublicFamiliesQuery{
		Keyword: token + "_张乙", Page: 1, PageSize: 20,
	})
	if err != nil || keywordResult.Total != 1 || len(keywordResult.Items) != 1 || keywordResult.Items[0].FamilyName != token+"_张乙家族" {
		t.Fatalf("keyword filter failed: err=%v result=%#v", err, keywordResult)
	}

	surnameResult, err := service.ListPublicFamilies(ctx, dto.ListPublicFamiliesQuery{
		Keyword: token, FamilySurname: "张", Page: 1, PageSize: 20,
	})
	if err != nil || surnameResult.Total != 2 {
		t.Fatalf("familySurname filter failed: err=%v total=%d", err, surnameResult.Total)
	}

	regionResult, err := service.ListPublicFamilies(ctx, dto.ListPublicFamiliesQuery{
		Keyword: token, RegionText: "河南", Page: 1, PageSize: 20,
	})
	if err != nil || regionResult.Total != 1 || regionResult.Items[0].FamilySurname != "李" {
		t.Fatalf("regionText filter failed: err=%v result=%#v", err, regionResult)
	}

	pageOne, err := service.ListPublicFamilies(ctx, dto.ListPublicFamiliesQuery{
		Keyword: scope.Keyword, Page: 1, PageSize: 2,
	})
	if err != nil || pageOne.Total != 3 || len(pageOne.Items) != 2 || pageOne.Page != 1 || pageOne.PageSize != 2 {
		t.Fatalf("page 1 failed: err=%v result=%#v", err, pageOne)
	}
	if pageOne.Items[0].FamilyName != token+"_李甲家族" || pageOne.Items[1].FamilyName != token+"_张乙家族" {
		t.Fatalf("unexpected sort order: %#v", pageOne.Items)
	}

	pageTwo, err := service.ListPublicFamilies(ctx, dto.ListPublicFamiliesQuery{
		Keyword: scope.Keyword, Page: 2, PageSize: 2,
	})
	if err != nil || len(pageTwo.Items) != 1 || pageTwo.Items[0].FamilyName != token+"_张甲家族" {
		t.Fatalf("page 2 failed: err=%v result=%#v", err, pageTwo)
	}

	oversized, err := service.ListPublicFamilies(ctx, dto.ListPublicFamiliesQuery{
		Keyword: scope.Keyword, Page: 1, PageSize: 100,
	})
	if err != nil || oversized.PageSize != 50 {
		t.Fatalf("expected pageSize capped to 50, got %#v", oversized)
	}
}

func TestListPublicFamiliesHidesSensitiveContactAndFields(t *testing.T) {
	ctx := context.Background()
	tx := transactionalTestDB(t)
	token := publicSearchToken(t)
	now := time.Now().UTC().Truncate(time.Second)

	hiddenContact := insertPublicSearchFamily(t, tx, familymodel.Family{
		FamilyName: token + "_hidden_contact_张氏", FamilySurname: "张", Description: strPtr(token),
		PublicContactVisible: false,
		PublicContactName:    strPtr("张三"),
		PublicContactPhone:   strPtr("13800138000"),
		PublicContactWechat:  strPtr("wx_secret"),
		PublicContactNote:    strPtr("请勿展示"),
		CreatorUserID:        uint64Ptr(999),
		Status:               "NORMAL",
		Searchable:           true,
		PublicDisplayStatus:  "APPROVED",
		PublicApprovedAt:     &now,
	})

	var storedHidden familymodel.Family
	if err := tx.First(&storedHidden, hiddenContact.ID).Error; err != nil {
		t.Fatalf("reload hidden contact family: %v", err)
	}
	if storedHidden.PublicContactVisible {
		t.Fatalf("expected public_contact_visible=false in DB, got true for family %#v", storedHidden)
	}

	visibleContact := insertPublicSearchFamily(t, tx, familymodel.Family{
		FamilyName: token + "_visible_contact_李氏", FamilySurname: "李", Description: strPtr(token),
		PublicContactVisible: true,
		PublicContactName:    strPtr("李四"),
		PublicContactPhone:   strPtr("13900139000"),
		PublicContactWechat:  strPtr("wx_visible"),
		PublicContactNote:    strPtr("欢迎联系"),
		CreatorUserID:        uint64Ptr(1000),
		Status:               "NORMAL",
		Searchable:           true,
		PublicDisplayStatus:  "APPROVED",
		PublicApprovedAt:     &now,
	})

	service := publicSearchService(tx)
	result, businessErr := service.ListPublicFamilies(ctx, dto.ListPublicFamiliesQuery{
		Keyword: token, Page: 1, PageSize: 20,
	})
	if businessErr != nil {
		t.Fatalf("ListPublicFamilies: %v", businessErr)
	}
	if result.Total != 2 || len(result.Items) != 2 {
		t.Fatalf("expected exactly two token-scoped families, got total=%d items=%d", result.Total, len(result.Items))
	}

	var hiddenItem, visibleItem map[string]any
	for _, item := range result.Items {
		payload, marshalErr := json.Marshal(item)
		if marshalErr != nil {
			t.Fatalf("marshal item: %v", marshalErr)
		}
		var decoded map[string]any
		if unmarshalErr := json.Unmarshal(payload, &decoded); unmarshalErr != nil {
			t.Fatalf("unmarshal item: %v", unmarshalErr)
		}
		for _, forbidden := range []string{
			"userId", "memberId", "openid", "unionid", "token", "password",
			"creatorUserId", "currentFounderMemberId", "publicContactPhone", "publicContactWechat",
		} {
			if _, exists := decoded[forbidden]; exists {
				t.Fatalf("forbidden field %s present in %#v", forbidden, decoded)
			}
		}
		if item.ID == hiddenContact.ID {
			hiddenItem = decoded
		}
		if item.ID == visibleContact.ID {
			visibleItem = decoded
		}
	}

	if hiddenItem == nil || visibleItem == nil {
		t.Fatalf("expected both families in result: %#v", result.Items)
	}
	if hiddenItem["publicContactVisible"] != false {
		t.Fatalf("hidden contact family should keep publicContactVisible=false: %#v", hiddenItem)
	}
	if _, exists := hiddenItem["publicContactName"]; exists {
		t.Fatalf("hidden contact should not expose publicContactName: %#v", hiddenItem)
	}
	if visibleItem["publicContactName"] != "李四" || visibleItem["publicContactNote"] != "欢迎联系" {
		t.Fatalf("visible contact fields missing: %#v", visibleItem)
	}
	if _, exists := visibleItem["publicContactPhone"]; exists {
		t.Fatalf("search list must not expose phone: %#v", visibleItem)
	}
}

func TestListPublicFamiliesKeywordMatchesNativePlace(t *testing.T) {
	ctx := context.Background()
	tx := transactionalTestDB(t)
	token := publicSearchToken(t)
	now := time.Now().UTC().Truncate(time.Second)
	insertPublicSearchFamily(t, tx, familymodel.Family{
		FamilyName: token + "_闽南林氏", FamilySurname: "林",
		NativePlace: strPtr(token + "_福建厦门"), Description: strPtr(token),
		Status: "NORMAL", Searchable: true, PublicDisplayStatus: "APPROVED", PublicApprovedAt: &now,
	})

	result, businessErr := publicSearchService(tx).ListPublicFamilies(ctx, dto.ListPublicFamiliesQuery{
		Keyword: token + "_福建", Page: 1, PageSize: 20,
	})
	if businessErr != nil {
		t.Fatalf("ListPublicFamilies: %v", businessErr)
	}
	if result.Total != 1 || len(result.Items) != 1 || !strings.Contains(result.Items[0].FamilyName, "林") {
		t.Fatalf("nativePlace keyword search failed: %#v", result)
	}
}

func strPtr(value string) *string {
	return &value
}

func uint64Ptr(value uint64) *uint64 {
	return &value
}
