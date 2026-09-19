package service

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	operationlog "tree/backend/internal/operationlog/service"
	showcasedto "tree/backend/internal/showcasefamily/dto"
	showcasemodel "tree/backend/internal/showcasefamily/model"
	showcaserepo "tree/backend/internal/showcasefamily/repository"
	"tree/backend/internal/showcasefamily/vo"
)

const (
	StatusDraft     = "DRAFT"
	StatusPublished = "PUBLISHED"

	CodeNotFound     apperrors.Code = 49401
	CodeInvalidInput apperrors.Code = 49402
)

type Service interface {
	List(context.Context) ([]vo.ShowcaseFamily, *apperrors.BusinessError)
	ListPublished(context.Context) ([]vo.ShowcaseFamily, *apperrors.BusinessError)
	GetDisplaySettings(context.Context) (*vo.DisplaySettings, *apperrors.BusinessError)
	UpdateDisplaySettings(context.Context, uint64, string, showcasedto.DisplaySettingsRequest) (*vo.DisplaySettings, *apperrors.BusinessError)
	Tree(context.Context, uint64) (*vo.ShowcaseFamilyTree, *apperrors.BusinessError)
	PublicTree(context.Context, uint64) (*vo.ShowcaseFamilyTree, *apperrors.BusinessError)
	Create(context.Context, uint64, string, showcasedto.CreateRequest) (*vo.ShowcaseFamily, *apperrors.BusinessError)
	Update(context.Context, uint64, string, uint64, showcasedto.UpdateRequest) (*vo.ShowcaseFamily, *apperrors.BusinessError)
	Publish(context.Context, uint64, string, uint64, bool) (*vo.ShowcaseFamily, *apperrors.BusinessError)
	Delete(context.Context, uint64, string, uint64) *apperrors.BusinessError
	CreateMember(context.Context, uint64, string, uint64, showcasedto.MemberRequest) (*vo.TreeMember, *apperrors.BusinessError)
	UpdateMember(context.Context, uint64, string, uint64, uint64, showcasedto.MemberRequest) (*vo.TreeMember, *apperrors.BusinessError)
	DeleteMember(context.Context, uint64, string, uint64, uint64) *apperrors.BusinessError
	CreateRelationship(context.Context, uint64, string, uint64, showcasedto.RelationshipRequest) (*vo.TreeRelationship, *apperrors.BusinessError)
	DeleteRelationship(context.Context, uint64, string, uint64, uint64) *apperrors.BusinessError
	AddRelative(context.Context, uint64, string, uint64, uint64, showcasedto.AddRelativeRequest) (*vo.AddRelativeResult, *apperrors.BusinessError)
}

type service struct {
	repo showcaserepo.Repository
	logs operationlog.Service
}

func New(repo showcaserepo.Repository, logs operationlog.Service) Service {
	return &service{repo: repo, logs: logs}
}

func (s *service) List(ctx context.Context) ([]vo.ShowcaseFamily, *apperrors.BusinessError) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	items, businessErr := s.toVOs(ctx, rows)
	if businessErr != nil {
		return nil, businessErr
	}
	return s.applyRotation(ctx, items, false)
}

func (s *service) ListPublished(ctx context.Context) ([]vo.ShowcaseFamily, *apperrors.BusinessError) {
	rows, err := s.repo.ListPublished(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	items, businessErr := s.toVOs(ctx, rows)
	if businessErr != nil {
		return nil, businessErr
	}
	return s.applyRotation(ctx, items, true)
}

func (s *service) GetDisplaySettings(ctx context.Context) (*vo.DisplaySettings, *apperrors.BusinessError) {
	settings, err := s.repo.GetDisplaySettings(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return displaySettingsVO(*settings), nil
}

func (s *service) UpdateDisplaySettings(ctx context.Context, adminID uint64, role string, req showcasedto.DisplaySettingsRequest) (*vo.DisplaySettings, *apperrors.BusinessError) {
	if !canManage(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	if req.RotationDays < 1 || req.RotationDays > 30 {
		return nil, showcaseError(CodeInvalidInput, "随机置顶时长应为 1 至 30 天")
	}
	before, err := s.repo.GetDisplaySettings(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if err := s.repo.UpdateDisplaySettings(ctx, map[string]any{"rotation_days": req.RotationDays, "pinned_showcase_family_id": nil, "pinned_until": nil, "updated_by_admin_id": adminID}); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	after, err := s.repo.GetDisplaySettings(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	s.writeLog(ctx, adminID, role, "UPDATE_SHOWCASE_DISPLAY_SETTINGS", 1, before, after)
	return displaySettingsVO(*after), nil
}

func (s *service) Tree(ctx context.Context, showcaseFamilyID uint64) (*vo.ShowcaseFamilyTree, *apperrors.BusinessError) {
	if _, businessErr := s.find(ctx, showcaseFamilyID); businessErr != nil {
		return nil, businessErr
	}
	return s.tree(ctx, showcaseFamilyID)
}

func (s *service) PublicTree(ctx context.Context, showcaseFamilyID uint64) (*vo.ShowcaseFamilyTree, *apperrors.BusinessError) {
	family, businessErr := s.find(ctx, showcaseFamilyID)
	if businessErr != nil {
		return nil, businessErr
	}
	if family.Status != StatusPublished {
		return nil, showcaseError(CodeNotFound, "展示家谱不存在或未发布")
	}
	return s.tree(ctx, showcaseFamilyID)
}

func (s *service) tree(ctx context.Context, showcaseFamilyID uint64) (*vo.ShowcaseFamilyTree, *apperrors.BusinessError) {
	members, err := s.repo.ListMembers(ctx, showcaseFamilyID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	relationships, err := s.repo.ListRelationships(ctx, showcaseFamilyID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := &vo.ShowcaseFamilyTree{ShowcaseFamilyID: showcaseFamilyID, Nodes: make([]vo.TreeMember, 0, len(members)), Edges: make([]vo.TreeRelationship, 0, len(relationships))}
	for _, member := range members {
		result.Nodes = append(result.Nodes, treeMemberVO(member))
	}
	for _, relationship := range relationships {
		result.Edges = append(result.Edges, treeRelationshipVO(relationship))
	}
	return result, nil
}

func (s *service) Create(ctx context.Context, adminID uint64, role string, req showcasedto.CreateRequest) (*vo.ShowcaseFamily, *apperrors.BusinessError) {
	if !canManage(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	values, businessErr := normalize(req.DisplayName, req.RegionText, req.Summary, req.CoverImageURL, req.TreeImageURL, req.SortOrder)
	if businessErr != nil {
		return nil, businessErr
	}
	row := &showcasemodel.ShowcaseFamily{
		DisplayName: values.displayName, RegionText: values.regionText, Summary: values.summary,
		CoverImageURL: values.coverImageURL, TreeImageURL: values.treeImageURL, SortOrder: values.sortOrder,
		Status: StatusDraft, UpdatedByAdminID: &adminID,
	}
	root := &showcasemodel.ShowcaseFamilyMember{DisplayName: "待补充根节点", MemberType: "LINEAGE_MEMBER", Gender: "UNKNOWN", SortOrder: 0}
	if err := s.repo.CreateWithRoot(ctx, row, root); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	s.writeLog(ctx, adminID, role, "CREATE_SHOWCASE_FAMILY", row.ID, nil, row)
	s.writeLog(ctx, adminID, role, "CREATE_SHOWCASE_MEMBER", root.ID, nil, root)
	return s.toVO(ctx, *row)
}

func (s *service) Update(ctx context.Context, adminID uint64, role string, id uint64, req showcasedto.UpdateRequest) (*vo.ShowcaseFamily, *apperrors.BusinessError) {
	if !canManage(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	before, businessErr := s.find(ctx, id)
	if businessErr != nil {
		return nil, businessErr
	}
	values, businessErr := normalize(req.DisplayName, req.RegionText, req.Summary, req.CoverImageURL, req.TreeImageURL, req.SortOrder)
	if businessErr != nil {
		return nil, businessErr
	}
	if err := s.repo.Update(ctx, id, map[string]any{
		"display_name": values.displayName, "region_text": values.regionText, "summary": values.summary,
		"cover_image_url": values.coverImageURL, "tree_image_url": values.treeImageURL, "sort_order": values.sortOrder,
		"updated_by_admin_id": adminID,
	}); err != nil {
		return nil, mapRepoError(err)
	}
	after, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	s.writeLog(ctx, adminID, role, "UPDATE_SHOWCASE_FAMILY", id, before, after)
	return s.toVO(ctx, *after)
}

func (s *service) Publish(ctx context.Context, adminID uint64, role string, id uint64, publish bool) (*vo.ShowcaseFamily, *apperrors.BusinessError) {
	if !canManage(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	before, businessErr := s.find(ctx, id)
	if businessErr != nil {
		return nil, businessErr
	}
	values := map[string]any{"updated_by_admin_id": adminID}
	action := "UNPUBLISH_SHOWCASE_FAMILY"
	if publish {
		if businessErr := s.validatePublishReady(ctx, id); businessErr != nil {
			return nil, businessErr
		}
		now := time.Now()
		values["status"] = StatusPublished
		values["published_at"] = now
		action = "PUBLISH_SHOWCASE_FAMILY"
	} else {
		values["status"] = StatusDraft
		values["published_at"] = nil
	}
	if err := s.repo.Update(ctx, id, values); err != nil {
		return nil, mapRepoError(err)
	}
	after, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	s.writeLog(ctx, adminID, role, action, id, before, after)
	return s.toVO(ctx, *after)
}

func (s *service) Delete(ctx context.Context, adminID uint64, role string, id uint64) *apperrors.BusinessError {
	if !canManage(role) {
		return apperrors.New(apperrors.CodeForbidden)
	}
	before, businessErr := s.find(ctx, id)
	if businessErr != nil {
		return businessErr
	}
	if err := s.repo.DeleteWithTree(ctx, id); err != nil {
		return mapRepoError(err)
	}
	s.writeLog(ctx, adminID, role, "DELETE_SHOWCASE_FAMILY", id, before, nil)
	return nil
}

func (s *service) CreateMember(ctx context.Context, adminID uint64, role string, showcaseFamilyID uint64, req showcasedto.MemberRequest) (*vo.TreeMember, *apperrors.BusinessError) {
	if !canManage(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	if _, businessErr := s.find(ctx, showcaseFamilyID); businessErr != nil {
		return nil, businessErr
	}
	values, businessErr := normalizeMember(req)
	if businessErr != nil {
		return nil, businessErr
	}
	member := &showcasemodel.ShowcaseFamilyMember{ShowcaseFamilyID: showcaseFamilyID, DisplayName: values.displayName, IdentityText: values.identityText, MemberType: values.memberType, Gender: values.gender, BirthYear: values.birthYear, SortOrder: values.sortOrder}
	if err := s.repo.CreateMember(ctx, member); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	s.writeLog(ctx, adminID, role, "CREATE_SHOWCASE_MEMBER", member.ID, nil, member)
	result := treeMemberVO(*member)
	return &result, nil
}

func (s *service) UpdateMember(ctx context.Context, adminID uint64, role string, showcaseFamilyID, memberID uint64, req showcasedto.MemberRequest) (*vo.TreeMember, *apperrors.BusinessError) {
	if !canManage(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	before, businessErr := s.findMember(ctx, showcaseFamilyID, memberID)
	if businessErr != nil {
		return nil, businessErr
	}
	values, businessErr := normalizeMember(req)
	if businessErr != nil {
		return nil, businessErr
	}
	if err := s.repo.UpdateMember(ctx, showcaseFamilyID, memberID, map[string]any{"display_name": values.displayName, "identity_text": values.identityText, "member_type": values.memberType, "gender": values.gender, "birth_year": values.birthYear, "sort_order": values.sortOrder}); err != nil {
		return nil, mapRepoError(err)
	}
	after, err := s.repo.FindMemberByID(ctx, showcaseFamilyID, memberID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	s.writeLog(ctx, adminID, role, "UPDATE_SHOWCASE_MEMBER", memberID, before, after)
	result := treeMemberVO(*after)
	return &result, nil
}

func (s *service) DeleteMember(ctx context.Context, adminID uint64, role string, showcaseFamilyID, memberID uint64) *apperrors.BusinessError {
	if !canManage(role) {
		return apperrors.New(apperrors.CodeForbidden)
	}
	before, businessErr := s.findMember(ctx, showcaseFamilyID, memberID)
	if businessErr != nil {
		return businessErr
	}
	relationships, err := s.repo.ListRelationships(ctx, showcaseFamilyID)
	if err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	descendantCount := descendantCount(relationships, memberID)
	if descendantCount > 0 && before.MemberType != "SPOUSE" {
		return showcaseError(CodeInvalidInput, fmt.Sprintf("该成员有 %d 位子女，不能删除；如资料有误请直接编辑", descendantCount))
	}
	if descendantCount > 0 {
		members, err := s.repo.ListMembers(ctx, showcaseFamilyID)
		if err != nil {
			return apperrors.New(apperrors.CodeSystemError)
		}
		if hasChildWithoutOtherLineageParent(relationships, membersByID(members), memberID) {
			return showcaseError(CodeInvalidInput, "该外姓配偶存在未保留本家父母的子女，不能删除")
		}
	}
	if err := s.repo.DeleteMember(ctx, showcaseFamilyID, memberID); err != nil {
		return mapRepoError(err)
	}
	s.writeLog(ctx, adminID, role, "DELETE_SHOWCASE_MEMBER", memberID, before, nil)
	return nil
}

func (s *service) CreateRelationship(ctx context.Context, adminID uint64, role string, showcaseFamilyID uint64, req showcasedto.RelationshipRequest) (*vo.TreeRelationship, *apperrors.BusinessError) {
	if !canManage(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	if _, businessErr := s.find(ctx, showcaseFamilyID); businessErr != nil {
		return nil, businessErr
	}
	if req.FromMemberID == 0 || req.ToMemberID == 0 || req.FromMemberID == req.ToMemberID || (req.RelationshipType != "PARENT_CHILD" && req.RelationshipType != "SPOUSE") {
		return nil, showcaseError(CodeInvalidInput, "成员关系参数不正确")
	}
	fromMember, businessErr := s.findMember(ctx, showcaseFamilyID, req.FromMemberID)
	if businessErr != nil {
		return nil, businessErr
	}
	toMember, businessErr := s.findMember(ctx, showcaseFamilyID, req.ToMemberID)
	if businessErr != nil {
		return nil, businessErr
	}
	relationships, err := s.repo.ListRelationships(ctx, showcaseFamilyID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if relationshipExists(relationships, req.FromMemberID, req.ToMemberID, req.RelationshipType) {
		return nil, showcaseError(CodeInvalidInput, "该成员关系已存在")
	}
	if req.RelationshipType == "SPOUSE" {
		if businessErr := validateSpouseLink(relationships, *fromMember, *toMember); businessErr != nil {
			return nil, businessErr
		}
	}
	if req.RelationshipType == "PARENT_CHILD" {
		if businessErr := validateParentLink(relationships, *fromMember, *toMember); businessErr != nil {
			return nil, businessErr
		}
		for _, row := range relationships {
			if row.RelationshipType != "PARENT_CHILD" || row.ToMemberID != req.ToMemberID {
				continue
			}
			existingParent, businessErr := s.findMember(ctx, showcaseFamilyID, row.FromMemberID)
			if businessErr != nil {
				return nil, businessErr
			}
			if businessErr := validateParentPair(*fromMember, *existingParent); businessErr != nil {
				return nil, businessErr
			}
		}
	}
	relationship := &showcasemodel.ShowcaseFamilyRelationship{ShowcaseFamilyID: showcaseFamilyID, FromMemberID: req.FromMemberID, ToMemberID: req.ToMemberID, RelationshipType: req.RelationshipType}
	if err := s.repo.CreateRelationship(ctx, relationship); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	s.writeLog(ctx, adminID, role, "CREATE_SHOWCASE_RELATIONSHIP", relationship.ID, nil, relationship)
	result := treeRelationshipVO(*relationship)
	return &result, nil
}

func (s *service) DeleteRelationship(ctx context.Context, adminID uint64, role string, showcaseFamilyID, relationshipID uint64) *apperrors.BusinessError {
	if !canManage(role) {
		return apperrors.New(apperrors.CodeForbidden)
	}
	if _, businessErr := s.find(ctx, showcaseFamilyID); businessErr != nil {
		return businessErr
	}
	if relationshipID == 0 {
		return showcaseError(CodeNotFound, "展示家庭关系不存在")
	}
	if err := s.repo.DeleteRelationship(ctx, showcaseFamilyID, relationshipID); err != nil {
		return mapRepoError(err)
	}
	s.writeLog(ctx, adminID, role, "DELETE_SHOWCASE_RELATIONSHIP", relationshipID, nil, nil)
	return nil
}

func (s *service) AddRelative(ctx context.Context, adminID uint64, role string, showcaseFamilyID, anchorMemberID uint64, req showcasedto.AddRelativeRequest) (*vo.AddRelativeResult, *apperrors.BusinessError) {
	if !canManage(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	if _, businessErr := s.find(ctx, showcaseFamilyID); businessErr != nil {
		return nil, businessErr
	}
	anchorMember, businessErr := s.findMember(ctx, showcaseFamilyID, anchorMemberID)
	if businessErr != nil {
		return nil, businessErr
	}
	relationKind := strings.ToUpper(strings.TrimSpace(req.RelationKind))
	if relationKind != "CHILD" && relationKind != "SPOUSE" {
		return nil, showcaseError(CodeInvalidInput, "只能添加子女或配偶")
	}
	memberValues, businessErr := normalizeMember(req.Member)
	if businessErr != nil {
		return nil, businessErr
	}
	memberType := "LINEAGE_MEMBER"
	if relationKind == "SPOUSE" {
		memberType = "SPOUSE"
	}
	member := &showcasemodel.ShowcaseFamilyMember{ShowcaseFamilyID: showcaseFamilyID, DisplayName: memberValues.displayName, IdentityText: memberValues.identityText, MemberType: memberType, Gender: memberValues.gender, BirthYear: memberValues.birthYear, SortOrder: memberValues.sortOrder}
	relationships, err := s.repo.ListRelationships(ctx, showcaseFamilyID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	var existingSpouseID uint64
	var spouse *showcasemodel.ShowcaseFamilyMember
	if relationKind == "CHILD" {
		if req.ExistingSpouseMemberID != nil && *req.ExistingSpouseMemberID != 0 {
			existingSpouseID = *req.ExistingSpouseMemberID
			if existingSpouseID == anchorMemberID {
				return nil, showcaseError(CodeInvalidInput, "另一位家长不能选择当前成员")
			}
			existingSpouse, businessErr := s.findMember(ctx, showcaseFamilyID, existingSpouseID)
			if businessErr != nil {
				return nil, businessErr
			}
			if !spouseLinked(relationships, anchorMemberID, existingSpouseID) {
				return nil, showcaseError(CodeInvalidInput, "所选成员不是当前成员的配偶")
			}
			if businessErr := validateParentPair(*anchorMember, *existingSpouse); businessErr != nil {
				return nil, businessErr
			}
		}
		if req.NewSpouse != nil {
			if existingSpouseID != 0 {
				return nil, showcaseError(CodeInvalidInput, "请选择已有配偶或新建配偶其中一种")
			}
			spouseValues, businessErr := normalizeMember(*req.NewSpouse)
			if businessErr != nil {
				return nil, businessErr
			}
			spouse = &showcasemodel.ShowcaseFamilyMember{ShowcaseFamilyID: showcaseFamilyID, DisplayName: spouseValues.displayName, IdentityText: spouseValues.identityText, MemberType: "SPOUSE", Gender: spouseValues.gender, BirthYear: spouseValues.birthYear, SortOrder: spouseValues.sortOrder}
			if businessErr := validateParentPair(*anchorMember, *spouse); businessErr != nil {
				return nil, businessErr
			}
			if businessErr := validateSpouseLink(relationships, *anchorMember, *spouse); businessErr != nil {
				return nil, businessErr
			}
		}
	} else if req.ExistingSpouseMemberID != nil || req.NewSpouse != nil {
		return nil, showcaseError(CodeInvalidInput, "添加配偶时不需要设置另一位家长")
	} else if businessErr := validateSpouseLink(relationships, *anchorMember, *member); businessErr != nil {
		return nil, businessErr
	}
	if err := s.repo.CreateRelative(ctx, showcaseFamilyID, anchorMemberID, relationKind, member, existingSpouseID, spouse); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	s.writeLog(ctx, adminID, role, "ADD_SHOWCASE_RELATIVE", member.ID, nil, map[string]any{"anchorMemberID": anchorMemberID, "relationKind": relationKind, "member": treeMemberVO(*member), "newSpouse": treeMemberVOPtr(spouse)})
	result := &vo.AddRelativeResult{Member: treeMemberVO(*member), Spouse: treeMemberVOPtr(spouse)}
	return result, nil
}

type normalizedValues struct {
	displayName   string
	regionText    string
	summary       string
	coverImageURL string
	treeImageURL  string
	sortOrder     int
}

func normalize(displayName, regionText, summary, coverImageURL, treeImageURL string, sortOrder int) (normalizedValues, *apperrors.BusinessError) {
	values := normalizedValues{
		displayName: strings.TrimSpace(displayName), regionText: strings.TrimSpace(regionText), summary: strings.TrimSpace(summary),
		coverImageURL: strings.TrimSpace(coverImageURL), treeImageURL: strings.TrimSpace(treeImageURL), sortOrder: sortOrder,
	}
	if values.displayName == "" || values.summary == "" || utf8.RuneCountInString(values.displayName) > 120 || utf8.RuneCountInString(values.regionText) > 120 || utf8.RuneCountInString(values.summary) > 1000 || len(values.coverImageURL) > 500 || len(values.treeImageURL) > 500 || sortOrder < 0 || sortOrder > 9999 {
		return normalizedValues{}, showcaseError(CodeInvalidInput, "展示名称、简介、图片地址或排序值格式不正确")
	}
	return values, nil
}

type normalizedMemberValues struct {
	displayName  string
	identityText string
	memberType   string
	gender       string
	birthYear    *int
	sortOrder    int
}

func normalizeMember(req showcasedto.MemberRequest) (normalizedMemberValues, *apperrors.BusinessError) {
	values := normalizedMemberValues{displayName: strings.TrimSpace(req.DisplayName), identityText: strings.TrimSpace(req.IdentityText), memberType: strings.ToUpper(strings.TrimSpace(req.MemberType)), gender: strings.ToUpper(strings.TrimSpace(req.Gender)), birthYear: req.BirthYear, sortOrder: req.SortOrder}
	if values.memberType == "" {
		values.memberType = "LINEAGE_MEMBER"
	}
	if values.displayName == "" || utf8.RuneCountInString(values.displayName) > 80 || utf8.RuneCountInString(values.identityText) > 120 || (values.memberType != "LINEAGE_MEMBER" && values.memberType != "SPOUSE") || (values.gender != "MALE" && values.gender != "FEMALE" && values.gender != "UNKNOWN") || values.sortOrder < 0 || values.sortOrder > 9999 || (values.birthYear != nil && (*values.birthYear < 1 || *values.birthYear > 9999)) {
		return normalizedMemberValues{}, showcaseError(CodeInvalidInput, "成员姓名、身份信息、性别、出生年份或排序值格式不正确")
	}
	return values, nil
}

func (s *service) find(ctx context.Context, id uint64) (*showcasemodel.ShowcaseFamily, *apperrors.BusinessError) {
	if id == 0 {
		return nil, showcaseError(CodeNotFound, "展示家庭不存在")
	}
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err)
	}
	return row, nil
}

func (s *service) findMember(ctx context.Context, showcaseFamilyID, memberID uint64) (*showcasemodel.ShowcaseFamilyMember, *apperrors.BusinessError) {
	if memberID == 0 {
		return nil, showcaseError(CodeNotFound, "展示家庭成员不存在")
	}
	member, err := s.repo.FindMemberByID(ctx, showcaseFamilyID, memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, showcaseError(CodeNotFound, "展示家庭成员不存在")
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return member, nil
}

func canManage(role string) bool {
	return role == string(enums.AdminRoleRootAdmin) || role == string(enums.AdminRoleSuperAdmin)
}

func mapRepoError(err error) *apperrors.BusinessError {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return showcaseError(CodeNotFound, "展示家庭不存在")
	}
	return apperrors.New(apperrors.CodeSystemError)
}

func toVO(row showcasemodel.ShowcaseFamily) vo.ShowcaseFamily {
	return vo.ShowcaseFamily{ID: row.ID, DisplayName: row.DisplayName, RegionText: row.RegionText, Summary: row.Summary, CoverImageURL: row.CoverImageURL, TreeImageURL: row.TreeImageURL, Status: row.Status, SortOrder: row.SortOrder, PublishedAt: row.PublishedAt, UpdatedAt: row.UpdatedAt}
}

func (s *service) toVOs(ctx context.Context, rows []showcasemodel.ShowcaseFamily) ([]vo.ShowcaseFamily, *apperrors.BusinessError) {
	result := make([]vo.ShowcaseFamily, 0, len(rows))
	for _, row := range rows {
		item, businessErr := s.toVO(ctx, row)
		if businessErr != nil {
			return nil, businessErr
		}
		result = append(result, *item)
	}
	return result, nil
}

func (s *service) toVO(ctx context.Context, row showcasemodel.ShowcaseFamily) (*vo.ShowcaseFamily, *apperrors.BusinessError) {
	members, err := s.repo.ListMembers(ctx, row.ID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	relationships, err := s.repo.ListRelationships(ctx, row.ID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := toVO(row)
	result.MemberCount = len(members)
	result.RelationshipCount = len(relationships)
	result.TreeStatus = treeStatus(members, relationships)
	return &result, nil
}

func (s *service) applyRotation(ctx context.Context, items []vo.ShowcaseFamily, allowRotate bool) ([]vo.ShowcaseFamily, *apperrors.BusinessError) {
	if len(items) == 0 {
		return items, nil
	}
	settings, err := s.repo.GetDisplaySettings(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	currentID := uint64(0)
	if settings.PinnedShowcaseFamilyID != nil {
		currentID = *settings.PinnedShowcaseFamilyID
	}
	now := time.Now()
	if allowRotate && (currentID == 0 || settings.PinnedUntil == nil || !settings.PinnedUntil.After(now) || !containsFamily(items, currentID)) {
		index, randomErr := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(items))))
		if randomErr != nil {
			return nil, apperrors.New(apperrors.CodeSystemError)
		}
		currentID = items[index.Int64()].ID
		until := now.AddDate(0, 0, settings.RotationDays)
		if err := s.repo.UpdateDisplaySettings(ctx, map[string]any{"pinned_showcase_family_id": currentID, "pinned_until": until}); err != nil {
			return nil, apperrors.New(apperrors.CodeSystemError)
		}
	}
	if currentID == 0 || !containsFamily(items, currentID) {
		return items, nil
	}
	for index := range items {
		if items[index].ID != currentID {
			continue
		}
		items[index].RotationPinned = true
		if index > 0 {
			pinned := items[index]
			copy(items[1:index+1], items[0:index])
			items[0] = pinned
		}
		break
	}
	return items, nil
}

func containsFamily(items []vo.ShowcaseFamily, id uint64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func displaySettingsVO(row showcasemodel.ShowcaseDisplaySettings) *vo.DisplaySettings {
	return &vo.DisplaySettings{RotationDays: row.RotationDays, PinnedShowcaseFamilyID: row.PinnedShowcaseFamilyID, PinnedUntil: row.PinnedUntil}
}

func (s *service) validatePublishReady(ctx context.Context, showcaseFamilyID uint64) *apperrors.BusinessError {
	members, err := s.repo.ListMembers(ctx, showcaseFamilyID)
	if err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	relationships, err := s.repo.ListRelationships(ctx, showcaseFamilyID)
	if err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	if treeStatus(members, relationships) != "READY" {
		return showcaseError(CodeInvalidInput, "请先完成家谱：补充根节点名称，并至少添加一位亲属和一条关系")
	}
	return nil
}

func treeStatus(members []showcasemodel.ShowcaseFamilyMember, relationships []showcasemodel.ShowcaseFamilyRelationship) string {
	if len(members) < 2 || len(relationships) == 0 {
		return "BUILDING"
	}
	for _, member := range members {
		if member.DisplayName == "待补充根节点" {
			return "BUILDING"
		}
	}
	return "READY"
}

func treeMemberVO(row showcasemodel.ShowcaseFamilyMember) vo.TreeMember {
	return vo.TreeMember{MemberID: row.ID, DisplayName: row.DisplayName, IdentityText: row.IdentityText, MemberType: row.MemberType, Gender: row.Gender, BirthYear: row.BirthYear, SortOrder: row.SortOrder}
}

func treeRelationshipVO(row showcasemodel.ShowcaseFamilyRelationship) vo.TreeRelationship {
	return vo.TreeRelationship{RelationshipID: row.ID, FromMemberID: row.FromMemberID, ToMemberID: row.ToMemberID, RelationshipType: row.RelationshipType}
}

func treeMemberVOPtr(row *showcasemodel.ShowcaseFamilyMember) *vo.TreeMember {
	if row == nil {
		return nil
	}
	result := treeMemberVO(*row)
	return &result
}

func spouseLinked(rows []showcasemodel.ShowcaseFamilyRelationship, leftID, rightID uint64) bool {
	for _, row := range rows {
		if row.RelationshipType == "SPOUSE" && ((row.FromMemberID == leftID && row.ToMemberID == rightID) || (row.FromMemberID == rightID && row.ToMemberID == leftID)) {
			return true
		}
	}
	return false
}

func validateSpouseLink(rows []showcasemodel.ShowcaseFamilyRelationship, left, right showcasemodel.ShowcaseFamilyMember) *apperrors.BusinessError {
	if spouseLinked(rows, left.ID, right.ID) {
		return showcaseError(CodeInvalidInput, "该成员关系已存在")
	}
	for _, member := range []showcasemodel.ShowcaseFamilyMember{left, right} {
		if member.Gender == "FEMALE" && spouseCount(rows, member.ID) > 0 {
			return showcaseError(CodeInvalidInput, "女性成员已有配偶，不能再添加配偶")
		}
	}
	return nil
}

func spouseCount(rows []showcasemodel.ShowcaseFamilyRelationship, memberID uint64) int {
	count := 0
	for _, row := range rows {
		if row.RelationshipType == "SPOUSE" && (row.FromMemberID == memberID || row.ToMemberID == memberID) {
			count++
		}
	}
	return count
}

func memberRelationshipCount(rows []showcasemodel.ShowcaseFamilyRelationship, memberID uint64) int {
	count := 0
	for _, row := range rows {
		if row.FromMemberID == memberID || row.ToMemberID == memberID {
			count++
		}
	}
	return count
}

func descendantCount(rows []showcasemodel.ShowcaseFamilyRelationship, memberID uint64) int {
	count := 0
	for _, row := range rows {
		if row.RelationshipType == "PARENT_CHILD" && row.FromMemberID == memberID {
			count++
		}
	}
	return count
}

func membersByID(rows []showcasemodel.ShowcaseFamilyMember) map[uint64]showcasemodel.ShowcaseFamilyMember {
	result := make(map[uint64]showcasemodel.ShowcaseFamilyMember, len(rows))
	for _, row := range rows {
		result[row.ID] = row
	}
	return result
}

func hasChildWithoutOtherLineageParent(rows []showcasemodel.ShowcaseFamilyRelationship, members map[uint64]showcasemodel.ShowcaseFamilyMember, spouseMemberID uint64) bool {
	for _, row := range rows {
		if row.RelationshipType != "PARENT_CHILD" || row.FromMemberID != spouseMemberID {
			continue
		}
		hasLineageParent := false
		for _, candidate := range rows {
			if candidate.RelationshipType == "PARENT_CHILD" && candidate.ToMemberID == row.ToMemberID && candidate.FromMemberID != spouseMemberID && members[candidate.FromMemberID].MemberType == "LINEAGE_MEMBER" {
				hasLineageParent = true
				break
			}
		}
		if !hasLineageParent {
			return true
		}
	}
	return false
}

func relationshipExists(rows []showcasemodel.ShowcaseFamilyRelationship, fromID, toID uint64, relationshipType string) bool {
	for _, row := range rows {
		if row.RelationshipType == relationshipType && row.FromMemberID == fromID && row.ToMemberID == toID {
			return true
		}
	}
	return false
}

func validateParentLink(rows []showcasemodel.ShowcaseFamilyRelationship, parent, child showcasemodel.ShowcaseFamilyMember) *apperrors.BusinessError {
	parentCount := 0
	for _, row := range rows {
		if row.RelationshipType != "PARENT_CHILD" || row.ToMemberID != child.ID {
			continue
		}
		parentCount++
		if row.FromMemberID == parent.ID {
			return showcaseError(CodeInvalidInput, "该成员已是此子女的父母")
		}
	}
	if parentCount >= 2 {
		return showcaseError(CodeInvalidInput, "子女最多只能设置一位父亲和一位母亲")
	}
	return nil
}

func validateParentPair(left, right showcasemodel.ShowcaseFamilyMember) *apperrors.BusinessError {
	if left.Gender != "UNKNOWN" && right.Gender != "UNKNOWN" && left.Gender == right.Gender {
		return showcaseError(CodeInvalidInput, "子女的父母性别不能相同")
	}
	return nil
}

func (s *service) writeLog(ctx context.Context, adminID uint64, role, action string, targetID uint64, before, after any) {
	roleCopy := role
	var beforeJSON, afterJSON []byte
	if before != nil {
		beforeJSON, _ = json.Marshal(before)
	}
	if after != nil {
		afterJSON, _ = json.Marshal(after)
	}
	_ = s.logs.WriteSuccess(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, OperatorRole: &roleCopy,
		Module: "SHOWCASE_FAMILY", Action: action, TargetType: stringPtr("showcase_family"), TargetID: &targetID,
		BeforeJSON: beforeJSON, AfterJSON: afterJSON,
	})
}

func showcaseError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}

func stringPtr(value string) *string { return &value }
