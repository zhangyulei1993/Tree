package freshquota

import (
	"context"
	"testing"

	"gorm.io/gorm"

	quotarepo "tree/backend/internal/accountquota/repository"
	quotaservice "tree/backend/internal/accountquota/service"
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

func NewQuotaService(tx *gorm.DB) quotaservice.Service {
	return quotaservice.NewService(tx, quotarepo.NewRepository(tx))
}

func NewFamilyService(tx *gorm.DB, quota quotaservice.Service) familyservice.FamilyService {
	return familyservice.NewFamilyService(tx, familyrepo.NewFamilyRepository(tx), AllowAllFamilyPerm{}, quota)
}

func NewMemberService(tx *gorm.DB, quota quotaservice.Service) memberservice.MemberService {
	return memberservice.NewMemberService(tx, memberrepo.NewMemberRepository(tx), AllowAllMemberPerm{}, quota)
}

func CreateFamily(t *testing.T, familySvc familyservice.FamilyService, userID uint64, surname string) uint64 {
	t.Helper()
	gender := "MALE"
	created, err := familySvc.Create(context.Background(), userID, coredto.CreateFamilyRequest{
		Surname: surname, FounderGender: &gender,
	}, familyservice.AuditInput{IP: "127.0.0.1", UserAgent: "fresh-quota-test"})
	if err != nil {
		t.Fatalf("create family %s: %v", surname, err)
	}
	return created.ID
}

func AddMember(t *testing.T, memberSvc memberservice.MemberService, actorID, familyID uint64, name string) uint64 {
	t.Helper()
	created, err := memberSvc.Create(context.Background(), actorID, familyID, memberdto.CreateMemberRequest{Name: name}, memberservice.AuditInput{})
	if err != nil {
		t.Fatalf("create member %s: %v", name, err)
	}
	return created.MemberID
}

// SeedJoinedFamily gives user a non-owned active membership (counts toward joined quota).
func SeedJoinedFamily(t *testing.T, tx *gorm.DB, user *usermodel.User, label string) uint64 {
	t.Helper()
	host := CreateActiveUser(t, tx, label+"创始人", false)
	familyID := seedFamilyWithFounder(t, tx, host, label+"家", label)
	memberID := insertMember(t, tx, familyID, label+"加入者", "MALE", "OPTIONAL")
	link := rolemodel.FamilyMemberUserLink{
		FamilyID: familyID, MemberID: memberID, UserID: user.ID,
		LinkStatus: "ACTIVE", LinkSource: "FRESH_QUOTA_TEST", FamilyRole: "MEMBER",
	}
	if err := tx.Create(&link).Error; err != nil {
		t.Fatalf("create joined link: %v", err)
	}
	return familyID
}

func seedFamilyWithFounder(t *testing.T, tx *gorm.DB, founder *usermodel.User, familyName, surname string) uint64 {
	t.Helper()
	family := &familymodel.Family{
		FamilyName: familyName, FamilySurname: surname, Status: "NORMAL",
		PublicDisplayStatus: "PRIVATE", Searchable: true, GraphVersion: 1,
	}
	if err := tx.Create(family).Error; err != nil {
		t.Fatalf("create family: %v", err)
	}
	memberID := insertMember(t, tx, family.ID, surname+"创始人", "MALE", "OPTIONAL")
	link := rolemodel.FamilyMemberUserLink{
		FamilyID: family.ID, MemberID: memberID, UserID: founder.ID,
		LinkStatus: "ACTIVE", LinkSource: "FRESH_QUOTA_TEST", FamilyRole: "FOUNDER",
	}
	if err := tx.Create(&link).Error; err != nil {
		t.Fatalf("create founder link: %v", err)
	}
	return family.ID
}

func insertMember(t *testing.T, tx *gorm.DB, familyID uint64, name, gender, bindingPolicy string) uint64 {
	t.Helper()
	member := &membermodel.FamilyMember{
		FamilyID: familyID, DisplayName: name, Status: "ACTIVE",
		Gender: gender, UserBindingPolicy: bindingPolicy,
	}
	if err := tx.Create(member).Error; err != nil {
		t.Fatalf("create member: %v", err)
	}
	return member.ID
}

// SeedHostFamilyWithGuest creates a host-owned family with one invitable unbound member.
func SeedHostFamilyWithGuest(t *testing.T, tx *gorm.DB, host *usermodel.User, label string) (familyID, guestMemberID uint64) {
	t.Helper()
	familyID = seedFamilyWithFounder(t, tx, host, label+"家", label)
	guestMemberID = insertMember(t, tx, familyID, label+"待绑定", "MALE", "OPTIONAL")
	return familyID, guestMemberID
}
