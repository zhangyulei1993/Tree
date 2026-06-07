package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	familymodel "tree/backend/internal/family/core/model"
	membermodel "tree/backend/internal/family/member/model"
	relationshipmodel "tree/backend/internal/family/relationship/model"
	treerepo "tree/backend/internal/family/tree/repository"
	"tree/backend/internal/family/tree/vo"
)

type fakePermission struct {
	allowed bool
	err     error
}

func (p fakePermission) IsFamilyMember(context.Context, uint64, uint64) (bool, error) {
	return p.allowed, p.err
}

type fakeRepository struct {
	family        *familymodel.Family
	snapshot      *treerepo.Snapshot
	findErr       error
	snapshotErr   error
	findCalls     int
	snapshotCalls int
}

func (r *fakeRepository) FindFamily(context.Context, uint64) (*familymodel.Family, error) {
	r.findCalls++
	if r.findErr != nil {
		return nil, r.findErr
	}
	copy := *r.family
	return &copy, nil
}

func (r *fakeRepository) LoadSnapshot(context.Context, uint64) (*treerepo.Snapshot, error) {
	r.snapshotCalls++
	if r.snapshotErr != nil {
		return nil, r.snapshotErr
	}
	return r.snapshot, nil
}

type fakeCache struct {
	values  map[string][]byte
	getErr  error
	setErr  error
	getKeys []string
	setKeys []string
}

func newFakeCache() *fakeCache {
	return &fakeCache{values: make(map[string][]byte)}
}

func (c *fakeCache) Get(_ context.Context, key string) ([]byte, error) {
	c.getKeys = append(c.getKeys, key)
	if c.getErr != nil {
		return nil, c.getErr
	}
	value, ok := c.values[key]
	if !ok {
		return nil, errors.New("cache miss")
	}
	return value, nil
}

func (c *fakeCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.setKeys = append(c.setKeys, key)
	if c.setErr != nil {
		return c.setErr
	}
	c.values[key] = value
	return nil
}

func TestBuildTreeResultSupportsFamilyShapesAndFiltersInvalidRows(t *testing.T) {
	snapshot := testSnapshot()
	deletedAt := time.Now()
	snapshot.Members = append(snapshot.Members, treerepo.MemberRow{
		FamilyMember: membermodel.FamilyMember{
			ID: 99, FamilyID: 2, DisplayName: "Deleted", Status: "DELETED", DeletedAt: &deletedAt,
		},
	})
	snapshot.Relationships = append(snapshot.Relationships,
		relationshipmodel.FamilyRelationship{
			ID: 99, FamilyID: 2, FromMemberID: 1, ToMemberID: 99,
			RelationshipType: "PARENT_CHILD", Status: "ACTIVE",
		},
		relationshipmodel.FamilyRelationship{
			ID: 100, FamilyID: 2, FromMemberID: 1, ToMemberID: 2,
			RelationshipType: "SIBLING", Status: "ACTIVE",
		},
		relationshipmodel.FamilyRelationship{
			ID: 101, FamilyID: 2, FromMemberID: 1, ToMemberID: 2,
			RelationshipType: "PARENT_CHILD", Status: "DELETED", DeletedAt: &deletedAt,
		},
	)

	result := buildTreeResult(snapshot)
	if len(result.Nodes) != 7 {
		t.Fatalf("expected 7 active nodes, got %d", len(result.Nodes))
	}
	if len(result.Edges) != 5 {
		t.Fatalf("expected 5 valid active edges, got %d", len(result.Edges))
	}
	for _, edge := range result.Edges {
		if edge.RelationshipType == "SIBLING" {
			t.Fatal("SIBLING edge must never be returned")
		}
	}

	child := treeItem(result, 2)
	if child != nil {
		t.Fatal("child with parents must not be a root")
	}
	parentOne := treeItem(result, 1)
	if parentOne == nil || !contains(parentOne.ChildrenIDs, 2) ||
		!contains(parentOne.SpouseIDs, 3) || !contains(parentOne.SpouseIDs, 6) {
		t.Fatalf("unexpected parent root: %#v", parentOne)
	}
	parentTwo := treeItem(result, 4)
	if parentTwo == nil || !contains(parentTwo.ChildrenIDs, 2) {
		t.Fatalf("second parent missing child: %#v", parentTwo)
	}
	if treeItem(result, 5) == nil {
		t.Fatal("independent node must be a root")
	}
	if treeItem(result, 3) != nil || treeItem(result, 6) != nil {
		t.Fatal("attached SPOUSE nodes must not be duplicated as roots")
	}
}

func TestBuildTreeResultHandlesCycleWithoutRecursion(t *testing.T) {
	snapshot := &treerepo.Snapshot{
		Family: familymodel.Family{ID: 2, GraphVersion: 3},
		Members: []treerepo.MemberRow{
			activeMember(1, "One", "LINEAGE_MEMBER"),
			activeMember(2, "Two", "LINEAGE_MEMBER"),
		},
		Relationships: []relationshipmodel.FamilyRelationship{
			activeRelationship(1, 1, 2, "PARENT_CHILD"),
			activeRelationship(2, 2, 1, "PARENT_CHILD"),
		},
	}
	result := buildTreeResult(snapshot)
	if len(result.Tree) != 1 {
		t.Fatalf("cyclic component must receive one fallback root, got %#v", result.Tree)
	}
}

func TestPrivateTreePermissionDenied(t *testing.T) {
	repo := testRepository()
	service := NewTreeService(repo, newFakeCache(), fakePermission{allowed: false})
	_, businessErr := service.GetPrivateTree(context.Background(), 9, 2)
	if businessErr == nil || businessErr.Code != CodePrivateTreeForbidden {
		t.Fatalf("expected private-tree permission error, got %#v", businessErr)
	}
	if repo.findCalls != 0 || repo.snapshotCalls != 0 {
		t.Fatal("database tree data must not load after permission denial")
	}
}

func TestPublicTreeRequiresNormalAndApproved(t *testing.T) {
	tests := []struct {
		name   string
		status string
		public string
	}{
		{name: "non normal", status: "DISSOLUTION_PENDING", public: "APPROVED"},
		{name: "non approved", status: "NORMAL", public: "PRIVATE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := testRepository()
			repo.family.Status = tt.status
			repo.family.PublicDisplayStatus = tt.public
			service := NewTreeService(repo, newFakeCache(), fakePermission{})
			_, businessErr := service.GetPublicTree(context.Background(), 2)
			if businessErr == nil || businessErr.Code != CodePublicTreeForbidden {
				t.Fatalf("expected public-tree error, got %#v", businessErr)
			}
			if repo.snapshotCalls != 0 {
				t.Fatal("private family must be rejected before cache/data load")
			}
		})
	}
}

func TestCacheMissBuildsAndWritesVersionedKey(t *testing.T) {
	repo := testRepository()
	cache := newFakeCache()
	service := NewTreeService(repo, cache, fakePermission{allowed: true})

	result, businessErr := service.GetPrivateTree(context.Background(), 8, 2)
	if businessErr != nil {
		t.Fatalf("GetPrivateTree returned error: %v", businessErr)
	}
	expectedKey := "family:2:tree:v14"
	if result.GraphVersion != 14 || repo.snapshotCalls != 1 ||
		len(cache.setKeys) != 1 || cache.setKeys[0] != expectedKey {
		t.Fatalf("unexpected cache miss behavior: result=%#v keys=%#v", result, cache.setKeys)
	}
}

func TestCacheHitReturnsWithoutSnapshotLoad(t *testing.T) {
	repo := testRepository()
	cache := newFakeCache()
	cached := buildTreeResult(repo.snapshot)
	data, _ := json.Marshal(cached)
	cache.values["family:2:tree:v14"] = data
	service := NewTreeService(repo, cache, fakePermission{allowed: true})

	result, businessErr := service.GetPrivateTree(context.Background(), 8, 2)
	if businessErr != nil || result.GraphVersion != 14 {
		t.Fatalf("unexpected cache hit result: %#v %#v", result, businessErr)
	}
	if repo.snapshotCalls != 0 {
		t.Fatal("cache hit must not load snapshot")
	}
}

func TestRedisErrorsDegradeToDatabase(t *testing.T) {
	repo := testRepository()
	cache := newFakeCache()
	cache.getErr = errors.New("redis unavailable")
	cache.setErr = errors.New("redis unavailable")
	service := NewTreeService(repo, cache, fakePermission{allowed: true})

	result, businessErr := service.GetPrivateTree(context.Background(), 8, 2)
	if businessErr != nil || result == nil || repo.snapshotCalls != 1 {
		t.Fatalf("redis failure did not degrade to database: %#v %#v", result, businessErr)
	}
}

func TestGraphVersionChangeUsesNewCacheKey(t *testing.T) {
	repo := testRepository()
	cache := newFakeCache()
	service := NewTreeService(repo, cache, fakePermission{allowed: true})
	if _, err := service.GetPrivateTree(context.Background(), 8, 2); err != nil {
		t.Fatalf("first read failed: %v", err)
	}
	repo.family.GraphVersion = 15
	repo.snapshot.Family.GraphVersion = 15
	if _, err := service.GetPrivateTree(context.Background(), 8, 2); err != nil {
		t.Fatalf("second read failed: %v", err)
	}
	if len(cache.getKeys) != 2 ||
		cache.getKeys[0] != "family:2:tree:v14" ||
		cache.getKeys[1] != "family:2:tree:v15" {
		t.Fatalf("unexpected cache keys: %#v", cache.getKeys)
	}
}

func TestTreeJSONContainsNoSensitiveFields(t *testing.T) {
	result := buildTreeResult(testSnapshot())
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal tree: %v", err)
	}
	jsonText := strings.ToLower(string(data))
	for _, forbidden := range []string{
		"phone", "userid", "token", "openid", "unionid", "password",
		"bounduserid", "familyrole",
	} {
		if strings.Contains(jsonText, forbidden) {
			t.Fatalf("tree JSON contains forbidden field %q: %s", forbidden, jsonText)
		}
	}
}

func testRepository() *fakeRepository {
	snapshot := testSnapshot()
	family := snapshot.Family
	return &fakeRepository{family: &family, snapshot: snapshot}
}

func testSnapshot() *treerepo.Snapshot {
	return &treerepo.Snapshot{
		Family: familymodel.Family{
			ID: 2, Status: "NORMAL", PublicDisplayStatus: "APPROVED",
			TreeMode: "LIST_TREE", GraphVersion: 14,
		},
		Members: []treerepo.MemberRow{
			activeMember(1, "Parent One", "LINEAGE_MEMBER"),
			activeMember(2, "Child", "LINEAGE_MEMBER"),
			activeMember(3, "Spouse One", "SPOUSE"),
			activeMember(4, "Parent Two", "LINEAGE_MEMBER"),
			activeMember(5, "Independent", "LINEAGE_MEMBER"),
			activeMember(6, "Spouse Two", "SPOUSE"),
			{
				FamilyMember: membermodel.FamilyMember{
					ID: 7, FamilyID: 2, DisplayName: "Not Required",
					Gender: "UNKNOWN", MemberType: "LINEAGE_MEMBER",
					UserBindingPolicy: "NOT_REQUIRED", Status: "ACTIVE",
				},
			},
		},
		Relationships: []relationshipmodel.FamilyRelationship{
			activeRelationship(1, 1, 2, "PARENT_CHILD"),
			activeRelationship(2, 4, 2, "PARENT_CHILD"),
			activeRelationship(3, 1, 3, "SPOUSE"),
			activeRelationship(4, 1, 6, "SPOUSE"),
			activeRelationship(5, 2, 7, "PARENT_CHILD"),
		},
	}
}

func activeMember(id uint64, name string, memberType string) treerepo.MemberRow {
	return treerepo.MemberRow{
		FamilyMember: membermodel.FamilyMember{
			ID: id, FamilyID: 2, DisplayName: name, Gender: "UNKNOWN",
			MemberType: memberType, UserBindingPolicy: "OPTIONAL", Status: "ACTIVE",
		},
		HasActiveLink: id == 1,
	}
}

func activeRelationship(id uint64, from uint64, to uint64, relationshipType string) relationshipmodel.FamilyRelationship {
	return relationshipmodel.FamilyRelationship{
		ID: id, FamilyID: 2, FromMemberID: from, ToMemberID: to,
		RelationshipType: relationshipType, Status: "ACTIVE",
	}
}

func treeItem(result *vo.TreeResult, memberID uint64) *vo.TreeItem {
	for i := range result.Tree {
		if result.Tree[i].MemberID == memberID {
			return &result.Tree[i]
		}
	}
	return nil
}

func contains(values []uint64, target uint64) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

var _ treerepo.Repository = (*fakeRepository)(nil)
