package service

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"gorm.io/gorm"
	apperrors "tree/backend/internal/common/errors"
	operationlog "tree/backend/internal/operationlog/service"
	sharedto "tree/backend/internal/shareconfig/dto"
	sharemodel "tree/backend/internal/shareconfig/model"
	sharerepo "tree/backend/internal/shareconfig/repository"
	"tree/backend/internal/shareconfig/vo"
)

const (
	CodeNotFound     apperrors.Code = 49201
	CodeInvalidInput apperrors.Code = 49202
)

var allowedKeys = map[string]bool{
	"HOME": true, "PUBLIC_FAMILY": true, "INVITATION_NODE": true, "INVITATION_PENDING_MEMBER": true,
}

type Service interface {
	List(context.Context) ([]vo.ShareConfig, *apperrors.BusinessError)
	Update(context.Context, uint64, string, string, sharedto.UpdateRequest) (*vo.ShareConfig, *apperrors.BusinessError)
}

type service struct {
	repo sharerepo.Repository
	logs operationlog.Service
}

func New(repo sharerepo.Repository, logs operationlog.Service) Service {
	return &service{repo: repo, logs: logs}
}

func (s *service) List(ctx context.Context) ([]vo.ShareConfig, *apperrors.BusinessError) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := make([]vo.ShareConfig, 0, len(rows))
	for _, row := range rows {
		result = append(result, toVO(row))
	}
	return result, nil
}

func (s *service) Update(ctx context.Context, adminID uint64, role, key string, req sharedto.UpdateRequest) (*vo.ShareConfig, *apperrors.BusinessError) {
	key = strings.ToUpper(strings.TrimSpace(key))
	if !allowedKeys[key] {
		return nil, shareError(CodeNotFound, "分享配置不存在")
	}
	title := strings.TrimSpace(req.TitleTemplate)
	imageURL := strings.TrimSpace(req.ImageURL)
	imageURLs := normalizeImageURLs(req.ImageURLs, imageURL)
	if title == "" || len([]rune(title)) > 300 || imageURL == "" || len([]rune(imageURL)) > 500 || len(imageURLs) == 0 || len(imageURLs) > 5 {
		return nil, shareError(CodeInvalidInput, "分享标题和图片地址不能为空，且图片必须填写 HTTPS 地址")
	}
	for _, image := range imageURLs {
		parsed, err := url.Parse(image)
		if err != nil || !strings.EqualFold(parsed.Scheme, "https") || parsed.Host == "" {
			return nil, shareError(CodeInvalidInput, "分享图片必须是有效的 HTTPS 地址")
		}
	}
	if strings.Contains(title, "{{") && !strings.Contains(title, "}}") {
		return nil, shareError(CodeInvalidInput, "分享标题变量格式不完整")
	}
	row, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shareError(CodeNotFound, "分享配置不存在")
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	imageURLsJSON, _ := json.Marshal(imageURLs)
	imageURLsText := string(imageURLsJSON)
	values := map[string]any{"title_template": title, "image_url": imageURL, "image_urls": imageURLsText, "description": cleanDescription(req.Description), "updated_by_admin_id": adminID}
	if err := s.repo.Update(ctx, key, values); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	updated, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	before, _ := json.Marshal(toVO(*row))
	after, _ := json.Marshal(toVO(*updated))
	admin := adminID
	roleCopy := role
	_ = s.logs.WriteSuccess(ctx, operationlog.WriteInput{OperatorType: "ADMIN", OperatorAdminID: &admin, OperatorRole: &roleCopy, Module: "SHARE_CONFIG", Action: "UPDATE_SHARE_CONFIG", TargetType: stringPtr("share_config"), TargetID: &updated.ID, BeforeJSON: before, AfterJSON: after})
	return ptrVO(toVO(*updated)), nil
}

func toVO(row sharemodel.ShareConfig) vo.ShareConfig {
	return vo.ShareConfig{ID: row.ID, ConfigKey: row.ConfigKey, TitleTemplate: row.TitleTemplate, ImageURL: row.ImageURL, ImageURLs: parseImageURLs(row.ImageURLs, row.ImageURL), Description: row.Description, UpdatedAt: row.UpdatedAt}
}
func ptrVO(value vo.ShareConfig) *vo.ShareConfig { return &value }
func cleanDescription(value *string) *string {
	if value == nil {
		return nil
	}
	text := strings.TrimSpace(*value)
	if text == "" {
		return nil
	}
	return &text
}
func stringPtr(value string) *string { return &value }

func normalizeImageURLs(values []string, active string) []string {
	result := make([]string, 0, len(values)+1)
	seen := make(map[string]struct{})
	for _, value := range append([]string{active}, values...) {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func parseImageURLs(value *string, active string) []string {
	if value != nil {
		var result []string
		if json.Unmarshal([]byte(*value), &result) == nil {
			return normalizeImageURLs(result, active)
		}
	}
	return normalizeImageURLs(nil, active)
}
func shareError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}
