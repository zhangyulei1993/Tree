package service

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"gorm.io/gorm"

	apperrors "tree/backend/internal/common/errors"
	commonredis "tree/backend/internal/common/redis"
	relationshipmodel "tree/backend/internal/family/relationship/model"
	treerepo "tree/backend/internal/family/tree/repository"
	"tree/backend/internal/family/tree/vo"
)

const (
	CodePrivateTreeForbidden apperrors.Code = 42401
	CodePublicTreeForbidden  apperrors.Code = 42402

	treeModeListTree = "LIST_TREE"
	treeCacheTTL     = 24 * time.Hour
)

type TreeService interface {
	GetPrivateTree(context.Context, uint64, uint64) (*vo.TreeResult, *apperrors.BusinessError)
	GetPublicTree(context.Context, uint64) (*vo.TreeResult, *apperrors.BusinessError)
}

type familyPermission interface {
	IsFamilyMember(context.Context, uint64, uint64) (bool, error)
}

type treeService struct {
	repo        treerepo.Repository
	cache       commonredis.TreeCache
	permissions familyPermission
}

func NewTreeService(repo treerepo.Repository, cache commonredis.TreeCache, permissions familyPermission) TreeService {
	if cache == nil {
		cache = commonredis.NoopTreeCache{}
	}
	return &treeService{repo: repo, cache: cache, permissions: permissions}
}

func (s *treeService) GetPrivateTree(ctx context.Context, actorID uint64, familyID uint64) (*vo.TreeResult, *apperrors.BusinessError) {
	allowed, err := s.permissions.IsFamilyMember(ctx, actorID, familyID)
	if err != nil || !allowed {
		return nil, treeError(CodePrivateTreeForbidden, "无权查看家庭树")
	}
	family, err := s.repo.FindFamily(ctx, familyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, treeError(CodePrivateTreeForbidden, "无权查看家庭树")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return s.loadTree(ctx, familyID, family.GraphVersion, false)
}

func (s *treeService) GetPublicTree(ctx context.Context, familyID uint64) (*vo.TreeResult, *apperrors.BusinessError) {
	family, err := s.repo.FindFamily(ctx, familyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, treeError(CodePublicTreeForbidden, "公开家庭树不可访问")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if !publicFamily(family.Status, family.PublicDisplayStatus) {
		return nil, treeError(CodePublicTreeForbidden, "公开家庭树不可访问")
	}
	return s.loadTree(ctx, familyID, family.GraphVersion, true)
}

func (s *treeService) loadTree(ctx context.Context, familyID uint64, graphVersion int64, public bool) (*vo.TreeResult, *apperrors.BusinessError) {
	key := BuildTreeCacheKey(familyID, graphVersion)
	if data, err := s.cache.Get(ctx, key); err == nil {
		var result vo.TreeResult
		if json.Unmarshal(data, &result) == nil &&
			result.FamilyID == familyID &&
			result.GraphVersion == graphVersion {
			return &result, nil
		}
	}

	snapshot, err := s.repo.LoadSnapshot(ctx, familyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if public {
			return nil, treeError(CodePublicTreeForbidden, "公开家庭树不可访问")
		}
		return nil, treeError(CodePrivateTreeForbidden, "无权查看家庭树")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if public && !publicFamily(snapshot.Family.Status, snapshot.Family.PublicDisplayStatus) {
		return nil, treeError(CodePublicTreeForbidden, "公开家庭树不可访问")
	}

	result := buildTreeResult(snapshot)
	actualKey := BuildTreeCacheKey(familyID, result.GraphVersion)
	if data, err := json.Marshal(result); err == nil {
		_ = s.cache.Set(ctx, actualKey, data, treeCacheTTL)
	}
	return result, nil
}

func buildTreeResult(snapshot *treerepo.Snapshot) *vo.TreeResult {
	result := &vo.TreeResult{
		FamilyID:     snapshot.Family.ID,
		TreeMode:     treeModeListTree,
		GraphVersion: snapshot.Family.GraphVersion,
		Nodes:        make([]vo.Node, 0, len(snapshot.Members)),
		Edges:        make([]vo.Edge, 0, len(snapshot.Relationships)),
		Tree:         make([]vo.TreeItem, 0),
	}

	memberByID := make(map[uint64]treerepo.MemberRow, len(snapshot.Members))
	memberOrder := make([]uint64, 0, len(snapshot.Members))
	for _, row := range snapshot.Members {
		if row.Status != "ACTIVE" || row.DeletedAt != nil {
			continue
		}
		memberByID[row.ID] = row
		memberOrder = append(memberOrder, row.ID)
		result.Nodes = append(result.Nodes, nodeVO(row))
	}

	parents := make(map[uint64][]uint64)
	children := make(map[uint64][]uint64)
	spouses := make(map[uint64][]uint64)
	for _, relationship := range snapshot.Relationships {
		if relationship.Status != "ACTIVE" || relationship.DeletedAt != nil {
			continue
		}
		if _, ok := memberByID[relationship.FromMemberID]; !ok {
			continue
		}
		if _, ok := memberByID[relationship.ToMemberID]; !ok {
			continue
		}
		switch relationship.RelationshipType {
		case "PARENT_CHILD":
			children[relationship.FromMemberID] = appendUnique(children[relationship.FromMemberID], relationship.ToMemberID)
			parents[relationship.ToMemberID] = appendUnique(parents[relationship.ToMemberID], relationship.FromMemberID)
		case "SPOUSE":
			spouses[relationship.FromMemberID] = appendUnique(spouses[relationship.FromMemberID], relationship.ToMemberID)
			spouses[relationship.ToMemberID] = appendUnique(spouses[relationship.ToMemberID], relationship.FromMemberID)
		default:
			continue
		}
		result.Edges = append(result.Edges, edgeVO(relationship))
	}

	for memberID := range children {
		sort.Slice(children[memberID], func(i, j int) bool { return children[memberID][i] < children[memberID][j] })
	}
	for memberID := range parents {
		sort.Slice(parents[memberID], func(i, j int) bool { return parents[memberID][i] < parents[memberID][j] })
	}
	for memberID := range spouses {
		sort.Slice(spouses[memberID], func(i, j int) bool { return spouses[memberID][i] < spouses[memberID][j] })
	}

	rootIDs := make([]uint64, 0)
	for _, memberID := range memberOrder {
		if len(parents[memberID]) == 0 && isSpouseGroupRoot(memberID, parents, spouses) {
			rootIDs = append(rootIDs, memberID)
		}
	}
	visited := make(map[uint64]bool, len(memberByID))
	visiting := make(map[uint64]bool, len(memberByID))
	var walk func(uint64)
	walk = func(memberID uint64) {
		if visited[memberID] || visiting[memberID] {
			return
		}
		visiting[memberID] = true
		for _, childID := range children[memberID] {
			walk(childID)
		}
		for _, spouseID := range spouses[memberID] {
			walk(spouseID)
		}
		delete(visiting, memberID)
		visited[memberID] = true
	}
	for _, rootID := range rootIDs {
		walk(rootID)
	}
	for _, memberID := range memberOrder {
		if !visited[memberID] {
			rootIDs = append(rootIDs, memberID)
			walk(memberID)
		}
	}

	for _, rootID := range rootIDs {
		result.Tree = append(result.Tree, vo.TreeItem{
			MemberID:    rootID,
			ParentIDs:   nonNilIDs(parents[rootID]),
			ChildrenIDs: nonNilIDs(children[rootID]),
			SpouseIDs:   nonNilIDs(spouses[rootID]),
		})
	}
	return result
}

func isSpouseGroupRoot(memberID uint64, parents map[uint64][]uint64, spouses map[uint64][]uint64) bool {
	for _, spouseID := range spouses[memberID] {
		if len(parents[spouseID]) == 0 && spouseID < memberID {
			return false
		}
	}
	return true
}

func nodeVO(row treerepo.MemberRow) vo.Node {
	result := vo.Node{
		MemberID: row.ID, DisplayName: row.DisplayName, Surname: row.Surname,
		GenerationCharacter: row.GenerationCharacter, Gender: row.Gender,
		MemberType: row.MemberType, IsLiving: row.IsLiving,
		UserBindingState: bindingState(row), CanExpand: !row.IsTerminalNode,
		StopReason: row.TerminalReason,
	}
	if row.BirthDate != nil {
		value := row.BirthDate.Format("2006-01-02")
		result.BirthDate = &value
	}
	if row.DeathDate != nil {
		value := row.DeathDate.Format("2006-01-02")
		result.DeathDate = &value
	}
	return result
}

func edgeVO(relationship relationshipmodel.FamilyRelationship) vo.Edge {
	return vo.Edge{
		RelationshipID:   relationship.ID,
		FromMemberID:     relationship.FromMemberID,
		ToMemberID:       relationship.ToMemberID,
		RelationshipType: relationship.RelationshipType,
		ParentLinkType:   relationship.ParentLinkType,
		RelationNoteType: relationship.RelationNoteType,
		RelationNote:     relationship.RelationNote,
	}
}

func bindingState(row treerepo.MemberRow) string {
	if row.HasActiveLink {
		return "BOUND"
	}
	if row.UserBindingPolicy == "NOT_REQUIRED" {
		return "NOT_REQUIRED"
	}
	return "UNBOUND"
}

func publicFamily(status string, publicDisplayStatus string) bool {
	return status == "NORMAL" && publicDisplayStatus == "APPROVED"
}

func appendUnique(values []uint64, value uint64) []uint64 {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func nonNilIDs(values []uint64) []uint64 {
	if values == nil {
		return []uint64{}
	}
	return values
}

func treeError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}
