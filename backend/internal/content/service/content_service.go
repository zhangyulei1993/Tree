package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	apperrors "tree/backend/internal/common/errors"
	contentdto "tree/backend/internal/content/dto"
	contentenum "tree/backend/internal/content/enum"
	contentmodel "tree/backend/internal/content/model"
	contentrepo "tree/backend/internal/content/repository"
	"tree/backend/internal/content/vo"
)

const (
	CodeContentCategoryNotFound apperrors.Code = 49001
	CodeContentArticleNotFound  apperrors.Code = 49002
	CodeContentInvalidStatus    apperrors.Code = 49003
	CodeContentDuplicateKey     apperrors.Code = 49004
	CodeContentForbidden        apperrors.Code = 49005
	CodeContentInvalidInput     apperrors.Code = 49006
)

type Service interface {
	ListPublicCategories(context.Context) ([]vo.Category, *apperrors.BusinessError)
	ListPublicArticles(context.Context, contentdto.ListArticlesQuery) (*vo.ListArticlesResult, *apperrors.BusinessError)
	PublicArticleDetail(context.Context, string) (*vo.ArticleDetail, *apperrors.BusinessError)
	ListAdminCategories(context.Context, bool) ([]vo.Category, *apperrors.BusinessError)
	CreateCategory(context.Context, uint64, string, contentdto.CreateCategoryRequest) (*vo.Category, *apperrors.BusinessError)
	UpdateCategory(context.Context, uint64, string, uint64, contentdto.UpdateCategoryRequest) (*vo.Category, *apperrors.BusinessError)
	DeleteCategory(context.Context, uint64, string, uint64) *apperrors.BusinessError
	ListAdminArticles(context.Context, contentdto.ListArticlesQuery) (*vo.ListArticlesResult, *apperrors.BusinessError)
	AdminArticleDetail(context.Context, uint64) (*vo.ArticleDetail, *apperrors.BusinessError)
	CreateArticle(context.Context, uint64, string, contentdto.CreateArticleRequest) (*vo.ArticleDetail, *apperrors.BusinessError)
	UpdateArticle(context.Context, uint64, string, uint64, contentdto.UpdateArticleRequest) (*vo.ArticleDetail, *apperrors.BusinessError)
	PublishArticle(context.Context, uint64, string, uint64) (*vo.ArticleDetail, *apperrors.BusinessError)
	UnpublishArticle(context.Context, uint64, string, uint64) (*vo.ArticleDetail, *apperrors.BusinessError)
	DeleteArticle(context.Context, uint64, string, uint64) *apperrors.BusinessError
}

type service struct {
	repo contentrepo.Repository
	uow  contentrepo.UnitOfWork
	now  func() time.Time
}

func NewService(repo contentrepo.Repository, uow contentrepo.UnitOfWork) Service {
	return &service{repo: repo, uow: uow, now: time.Now}
}

func (s *service) ListPublicCategories(ctx context.Context) ([]vo.Category, *apperrors.BusinessError) {
	rows, err := s.repo.ListCategories(ctx, false)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return categoriesVO(rows), nil
}

func (s *service) ListPublicArticles(ctx context.Context, req contentdto.ListArticlesQuery) (*vo.ListArticlesResult, *apperrors.BusinessError) {
	req.Page, req.PageSize = normalizePage(req.Page, req.PageSize)
	rows, total, err := s.repo.ListArticles(ctx, contentrepo.ListArticlesQuery{
		CategoryKey: cleanValue(req.CategoryKey),
		Keyword:     cleanValue(req.Keyword),
		Featured:    req.Featured,
		PublicOnly:  true,
		Page:        req.Page,
		PageSize:    req.PageSize,
	})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &vo.ListArticlesResult{Items: articleSummariesVO(rows), Page: req.Page, PageSize: req.PageSize, Total: total}, nil
}

func (s *service) PublicArticleDetail(ctx context.Context, idOrSlug string) (*vo.ArticleDetail, *apperrors.BusinessError) {
	row, err := s.repo.FindArticleByIDOrSlug(ctx, strings.TrimSpace(idOrSlug), true)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contentError(CodeContentArticleNotFound, "内容不存在或未发布")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return articleDetailVO(row), nil
}

func (s *service) ListAdminCategories(ctx context.Context, includeInactive bool) ([]vo.Category, *apperrors.BusinessError) {
	rows, err := s.repo.ListCategories(ctx, includeInactive)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return categoriesVO(rows), nil
}

func (s *service) CreateCategory(ctx context.Context, adminID uint64, role string, req contentdto.CreateCategoryRequest) (*vo.Category, *apperrors.BusinessError) {
	key := normalizeKey(req.Key)
	name := cleanValue(req.Name)
	if key == "" || name == "" {
		return nil, contentError(CodeContentInvalidInput, "分类标识和名称不能为空")
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	category := &contentmodel.ContentCategory{Key: key, Name: name, Description: cleanPtr(req.Description), SortOrder: req.SortOrder, IsActive: isActive}
	if err := s.repo.CreateCategory(ctx, category); err != nil {
		if duplicate(err) {
			return nil, contentError(CodeContentDuplicateKey, "内容分类已存在")
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	_ = s.repo.WriteLog(ctx, contentLog(adminID, role, "CREATE_CONTENT_CATEGORY", "content_category", category.ID, map[string]any{"categoryKey": key}))
	return s.categoryResult(ctx, category.ID)
}

func (s *service) UpdateCategory(ctx context.Context, adminID uint64, role string, categoryID uint64, req contentdto.UpdateCategoryRequest) (*vo.Category, *apperrors.BusinessError) {
	values := map[string]any{}
	if req.Name != nil {
		name := cleanValue(*req.Name)
		if name == "" {
			return nil, contentError(CodeContentInvalidInput, "分类名称不能为空")
		}
		values["name"] = name
	}
	if req.Description != nil {
		values["description"] = cleanPtr(req.Description)
	}
	if req.SortOrder != nil {
		values["sort_order"] = *req.SortOrder
	}
	if req.IsActive != nil {
		values["is_active"] = *req.IsActive
	}
	if len(values) == 0 {
		return s.categoryResult(ctx, categoryID)
	}
	if err := s.repo.UpdateCategory(ctx, categoryID, values); err != nil {
		return nil, mapCategoryErr(err)
	}
	_ = s.repo.WriteLog(ctx, contentLog(adminID, role, "UPDATE_CONTENT_CATEGORY", "content_category", categoryID, map[string]any{"updatedFields": keys(values)}))
	return s.categoryResult(ctx, categoryID)
}

func (s *service) DeleteCategory(ctx context.Context, adminID uint64, role string, categoryID uint64) *apperrors.BusinessError {
	if err := s.repo.DeleteCategory(ctx, categoryID, s.now()); err != nil {
		return mapCategoryErr(err)
	}
	_ = s.repo.WriteLog(ctx, contentLog(adminID, role, "DELETE_CONTENT_CATEGORY", "content_category", categoryID, nil))
	return nil
}

func (s *service) ListAdminArticles(ctx context.Context, req contentdto.ListArticlesQuery) (*vo.ListArticlesResult, *apperrors.BusinessError) {
	req.Page, req.PageSize = normalizePage(req.Page, req.PageSize)
	rows, total, err := s.repo.ListArticles(ctx, contentrepo.ListArticlesQuery{
		CategoryKey: cleanValue(req.CategoryKey),
		Status:      cleanValue(req.Status),
		Keyword:     cleanValue(req.Keyword),
		Featured:    req.Featured,
		Page:        req.Page,
		PageSize:    req.PageSize,
	})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &vo.ListArticlesResult{Items: articleSummariesVO(rows), Page: req.Page, PageSize: req.PageSize, Total: total}, nil
}

func (s *service) AdminArticleDetail(ctx context.Context, articleID uint64) (*vo.ArticleDetail, *apperrors.BusinessError) {
	row, err := s.repo.FindArticleByID(ctx, articleID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contentError(CodeContentArticleNotFound, "内容不存在")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return articleDetailVO(row), nil
}

func (s *service) CreateArticle(ctx context.Context, adminID uint64, role string, req contentdto.CreateArticleRequest) (*vo.ArticleDetail, *apperrors.BusinessError) {
	category, businessErr := s.categoryByKey(ctx, req.CategoryKey)
	if businessErr != nil {
		return nil, businessErr
	}
	title, slug := cleanValue(req.Title), normalizeKey(req.Slug)
	if title == "" {
		return nil, contentError(CodeContentInvalidInput, "标题不能为空")
	}
	articleType := normalizeArticleType(req.ContentType)
	if articleType == "" {
		articleType = contentenum.ArticleTypeInternal
	}
	if !contentenum.ValidArticleType(articleType) {
		return nil, contentError(CodeContentInvalidInput, "文章类型不合法")
	}
	body, externalURL, contentErr := validateArticleContent(articleType, req.Body, req.ExternalURL)
	if contentErr != nil {
		return nil, contentErr
	}
	if slug == "" {
		if articleType != contentenum.ArticleTypeWechatOfficial || externalURL == nil {
			return nil, contentError(CodeContentInvalidInput, "站内文章路径不能为空")
		}
		slug = wechatArticleSlug(*externalURL)
	}
	status := normalizeStatus(req.Status)
	if status == "" {
		status = contentenum.ArticleStatusDraft
	}
	if !contentenum.ValidArticleStatus(status) {
		return nil, contentError(CodeContentInvalidStatus, "内容状态不合法")
	}
	var publishedAt *time.Time
	if status == contentenum.ArticleStatusPublished {
		now := s.now()
		publishedAt = &now
	}
	article := &contentmodel.ContentArticle{
		CategoryID: category.ID, CategoryKey: category.Key, ContentType: articleType, Title: title, Slug: slug,
		Summary: cleanPtr(req.Summary), CoverURL: cleanPtr(req.CoverURL), Body: body, ExternalURL: externalURL,
		AuthorName: cleanPtr(req.AuthorName), Source: cleanPtr(req.Source), Status: status,
		IsFeatured: req.IsFeatured, SortOrder: req.SortOrder, PublishedAt: publishedAt,
		CreatedByAdminID: &adminID, UpdatedByAdminID: &adminID,
	}
	if err := s.repo.CreateArticle(ctx, article); err != nil {
		if duplicate(err) {
			return nil, contentError(CodeContentDuplicateKey, "内容路径已存在")
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	_ = s.repo.WriteLog(ctx, contentLog(adminID, role, "CREATE_CONTENT_ARTICLE", "content_article", article.ID, map[string]any{"slug": slug, "status": status}))
	return s.AdminArticleDetail(ctx, article.ID)
}

func (s *service) UpdateArticle(ctx context.Context, adminID uint64, role string, articleID uint64, req contentdto.UpdateArticleRequest) (*vo.ArticleDetail, *apperrors.BusinessError) {
	current, err := s.repo.FindArticleByID(ctx, articleID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contentError(CodeContentArticleNotFound, "内容不存在")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}

	values := map[string]any{"updated_by_admin_id": adminID}
	if req.CategoryKey != nil {
		category, businessErr := s.categoryByKey(ctx, *req.CategoryKey)
		if businessErr != nil {
			return nil, businessErr
		}
		values["category_id"] = category.ID
		values["category_key"] = category.Key
	}
	if req.Title != nil {
		title := cleanValue(*req.Title)
		if title == "" {
			return nil, contentError(CodeContentInvalidInput, "标题不能为空")
		}
		values["title"] = title
	}
	if req.Slug != nil {
		slug := normalizeKey(*req.Slug)
		if slug == "" {
			return nil, contentError(CodeContentInvalidInput, "路径不能为空")
		}
		values["slug"] = slug
	}
	if req.Summary != nil {
		values["summary"] = cleanPtr(req.Summary)
	}
	if req.CoverURL != nil {
		values["cover_url"] = cleanPtr(req.CoverURL)
	}
	if req.AuthorName != nil {
		values["author_name"] = cleanPtr(req.AuthorName)
	}
	if req.Source != nil {
		values["source"] = cleanPtr(req.Source)
	}
	if req.Status != nil {
		status := normalizeStatus(*req.Status)
		if !contentenum.ValidArticleStatus(status) {
			return nil, contentError(CodeContentInvalidStatus, "内容状态不合法")
		}
		values["status"] = status
		if status == contentenum.ArticleStatusPublished {
			now := s.now()
			values["published_at"] = &now
		}
		if status != contentenum.ArticleStatusPublished {
			values["published_at"] = nil
		}
	}
	if req.IsFeatured != nil {
		values["is_featured"] = *req.IsFeatured
	}
	if req.SortOrder != nil {
		values["sort_order"] = *req.SortOrder
	}
	articleType := current.ContentType
	if articleType == "" {
		articleType = contentenum.ArticleTypeInternal
	}
	if req.ContentType != nil {
		articleType = normalizeArticleType(*req.ContentType)
	}
	if !contentenum.ValidArticleType(articleType) {
		return nil, contentError(CodeContentInvalidInput, "文章类型不合法")
	}
	body := current.Body
	if req.Body != nil {
		body = *req.Body
	}
	externalURL := current.ExternalURL
	if req.ExternalURL != nil {
		externalURL = req.ExternalURL
	}
	body, externalURL, contentErr := validateArticleContent(articleType, body, externalURL)
	if contentErr != nil {
		return nil, contentErr
	}
	values["content_type"] = articleType
	values["body"] = body
	values["external_url"] = externalURL
	if err := s.repo.UpdateArticle(ctx, articleID, values); err != nil {
		if duplicate(err) {
			return nil, contentError(CodeContentDuplicateKey, "内容路径已存在")
		}
		return nil, mapArticleErr(err)
	}
	_ = s.repo.WriteLog(ctx, contentLog(adminID, role, "UPDATE_CONTENT_ARTICLE", "content_article", articleID, map[string]any{"updatedFields": keys(values)}))
	return s.AdminArticleDetail(ctx, articleID)
}

func (s *service) PublishArticle(ctx context.Context, adminID uint64, role string, articleID uint64) (*vo.ArticleDetail, *apperrors.BusinessError) {
	now := s.now()
	if err := s.repo.UpdateArticle(ctx, articleID, map[string]any{"status": contentenum.ArticleStatusPublished, "published_at": now, "updated_by_admin_id": adminID}); err != nil {
		return nil, mapArticleErr(err)
	}
	_ = s.repo.WriteLog(ctx, contentLog(adminID, role, "PUBLISH_CONTENT_ARTICLE", "content_article", articleID, nil))
	return s.AdminArticleDetail(ctx, articleID)
}

func (s *service) UnpublishArticle(ctx context.Context, adminID uint64, role string, articleID uint64) (*vo.ArticleDetail, *apperrors.BusinessError) {
	if err := s.repo.UpdateArticle(ctx, articleID, map[string]any{"status": contentenum.ArticleStatusDraft, "published_at": nil, "updated_by_admin_id": adminID}); err != nil {
		return nil, mapArticleErr(err)
	}
	_ = s.repo.WriteLog(ctx, contentLog(adminID, role, "UNPUBLISH_CONTENT_ARTICLE", "content_article", articleID, nil))
	return s.AdminArticleDetail(ctx, articleID)
}

func (s *service) DeleteArticle(ctx context.Context, adminID uint64, role string, articleID uint64) *apperrors.BusinessError {
	if err := s.repo.DeleteArticle(ctx, articleID, s.now()); err != nil {
		return mapArticleErr(err)
	}
	_ = s.repo.WriteLog(ctx, contentLog(adminID, role, "DELETE_CONTENT_ARTICLE", "content_article", articleID, nil))
	return nil
}

func (s *service) categoryByKey(ctx context.Context, key string) (*contentmodel.ContentCategory, *apperrors.BusinessError) {
	category, err := s.repo.FindCategoryByKey(ctx, normalizeKey(key), false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contentError(CodeContentCategoryNotFound, "内容分类不存在")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return category, nil
}

func (s *service) categoryResult(ctx context.Context, categoryID uint64) (*vo.Category, *apperrors.BusinessError) {
	category, err := s.repo.FindCategoryByID(ctx, categoryID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contentError(CodeContentCategoryNotFound, "内容分类不存在")
	}
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := categoryVO(*category)
	return &result, nil
}
