package service

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"gorm.io/gorm"

	apperrors "tree/backend/internal/common/errors"
	contentmodel "tree/backend/internal/content/model"
	contentrepo "tree/backend/internal/content/repository"
	"tree/backend/internal/content/vo"
	operationlog "tree/backend/internal/operationlog/service"
)

func categoriesVO(rows []contentmodel.ContentCategory) []vo.Category {
	result := make([]vo.Category, 0, len(rows))
	for _, row := range rows {
		result = append(result, categoryVO(row))
	}
	return result
}

func categoryVO(row contentmodel.ContentCategory) vo.Category {
	return vo.Category{
		ID:          row.ID,
		Key:         row.Key,
		Name:        row.Name,
		Description: row.Description,
		SortOrder:   row.SortOrder,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func articleSummariesVO(rows []contentrepo.ArticleRow) []vo.ArticleSummary {
	result := make([]vo.ArticleSummary, 0, len(rows))
	for _, row := range rows {
		result = append(result, articleSummaryVO(&row))
	}
	return result
}

func articleDetailVO(row *contentrepo.ArticleRow) *vo.ArticleDetail {
	summary := articleSummaryVO(row)
	return &vo.ArticleDetail{ArticleSummary: summary, Body: row.Body}
}

func articleSummaryVO(row *contentrepo.ArticleRow) vo.ArticleSummary {
	return vo.ArticleSummary{
		ID:           row.ID,
		CategoryID:   row.CategoryID,
		CategoryKey:  row.CategoryKey,
		CategoryName: row.CategoryName,
		Title:        row.Title,
		Slug:         row.Slug,
		Summary:      row.Summary,
		CoverURL:     row.CoverURL,
		AuthorName:   row.AuthorName,
		Source:       row.Source,
		Status:       row.Status,
		IsFeatured:   row.IsFeatured,
		SortOrder:    row.SortOrder,
		PublishedAt:  row.PublishedAt,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func cleanValue(value string) string {
	return strings.TrimSpace(value)
}

func normalizeKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeStatus(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func cleanPtr(value *string) *string {
	if value == nil {
		return nil
	}
	cleaned := strings.TrimSpace(*value)
	if cleaned == "" {
		return nil
	}
	return &cleaned
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func contentError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}

func mapCategoryErr(err error) *apperrors.BusinessError {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return contentError(CodeContentCategoryNotFound, "内容分类不存在")
	}
	return apperrors.New(apperrors.CodeSystemError)
}

func mapArticleErr(err error) *apperrors.BusinessError {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return contentError(CodeContentArticleNotFound, "内容不存在")
	}
	return apperrors.New(apperrors.CodeSystemError)
}

func duplicate(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}

func keys(values map[string]any) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func contentLog(adminID uint64, role string, action string, targetType string, targetID uint64, detail map[string]any) operationlog.WriteInput {
	roleValue := role
	targetTypeValue := targetType
	detailJSON, _ := json.Marshal(detail)
	return operationlog.WriteInput{
		OperatorType:    "ADMIN",
		OperatorAdminID: &adminID,
		OperatorRole:    &roleValue,
		Module:          "CONTENT",
		Action:          action,
		TargetType:      &targetTypeValue,
		TargetID:        &targetID,
		DetailJSON:      detailJSON,
	}
}
