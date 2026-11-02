package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	operationlog "tree/backend/internal/operationlog/service"
	tooldto "tree/backend/internal/toolconfig/dto"
	toolmodel "tree/backend/internal/toolconfig/model"
	toolrepo "tree/backend/internal/toolconfig/repository"
	"tree/backend/internal/toolconfig/vo"
)

const (
	CodeNotFound     apperrors.Code = 49301
	CodeInvalidInput apperrors.Code = 49302
	CodeDuplicate    apperrors.Code = 49303
)

type Service interface {
	List(context.Context) ([]vo.ToolConfig, *apperrors.BusinessError)
	Create(context.Context, uint64, string, string, string, string) (*vo.ToolConfig, *apperrors.BusinessError)
	Update(context.Context, uint64, string, string, tooldto.UpdateRequest) (*vo.ToolConfig, *apperrors.BusinessError)
}

type service struct {
	repo toolrepo.Repository
	logs operationlog.Service
}

func New(repo toolrepo.Repository, logs operationlog.Service) Service {
	return &service{repo: repo, logs: logs}
}

func (s *service) List(ctx context.Context) ([]vo.ToolConfig, *apperrors.BusinessError) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := make([]vo.ToolConfig, 0, len(rows))
	for _, row := range rows {
		result = append(result, toVO(row))
	}
	return result, nil
}

func (s *service) Create(ctx context.Context, adminID uint64, role, key, displayName, description string) (*vo.ToolConfig, *apperrors.BusinessError) {
	if !canManage(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	key = strings.ToUpper(strings.TrimSpace(key))
	displayName = strings.TrimSpace(displayName)
	description = strings.TrimSpace(description)
	if !validKey(key) || displayName == "" || description == "" || utf8.RuneCountInString(displayName) > 120 || utf8.RuneCountInString(description) > 300 {
		return nil, toolError(CodeInvalidInput, "工具标识、名称和说明格式不正确")
	}
	row, err := s.repo.FindByKey(ctx, key)
	if err == nil && row != nil {
		return nil, toolError(CodeDuplicate, "工具标识已存在")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	created := &toolmodel.ToolConfig{ToolKey: key, DisplayName: displayName, Description: description, Visible: false, Enabled: false, UpdatedByAdminID: &adminID}
	if err := s.repo.Create(ctx, created); err != nil {
		if isDuplicateError(err) {
			return nil, toolError(CodeDuplicate, "工具标识已存在")
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	roleCopy := role
	after, _ := json.Marshal(toVO(*created))
	_ = s.logs.WriteSuccess(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, OperatorRole: &roleCopy,
		Module: "TOOL_CONFIG", Action: "CREATE_TOOL_CONFIG", TargetType: stringPtr("tool_config"), TargetID: &created.ID,
		AfterJSON: after,
	})
	result := toVO(*created)
	return &result, nil
}

func (s *service) Update(ctx context.Context, adminID uint64, role, key string, req tooldto.UpdateRequest) (*vo.ToolConfig, *apperrors.BusinessError) {
	if !canManage(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	key = strings.ToUpper(strings.TrimSpace(key))
	if key == "" {
		return nil, toolError(CodeNotFound, "常用工具配置不存在")
	}
	row, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, toolError(CodeNotFound, "常用工具配置不存在")
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	values := map[string]any{"updated_by_admin_id": adminID}
	if req.Visible != nil {
		values["visible"] = *req.Visible
	}
	if req.Enabled != nil {
		values["enabled"] = *req.Enabled
	}
	if req.Pinned != nil {
		values["pinned"] = *req.Pinned
	}
	if req.Highlighted != nil {
		values["highlighted"] = *req.Highlighted
	}
	if req.SortOrder != nil {
		if *req.SortOrder < 0 || *req.SortOrder > 9999 {
			return nil, toolError(CodeInvalidInput, "排序值必须在 0 到 9999 之间")
		}
		values["sort_order"] = *req.SortOrder
	}
	if len(values) == 1 {
		return nil, toolError(CodeInvalidInput, "至少需要修改一项配置")
	}
	if err := s.repo.Update(ctx, key, values); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, toolError(CodeNotFound, "常用工具配置不存在")
		}
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
	_ = s.logs.WriteSuccess(ctx, operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &admin, OperatorRole: &roleCopy,
		Module: "TOOL_CONFIG", Action: "UPDATE_TOOL_CONFIG", TargetType: stringPtr("tool_config"), TargetID: &updated.ID,
		BeforeJSON: before, AfterJSON: after,
	})
	result := toVO(*updated)
	return &result, nil
}

func canManage(role string) bool {
	return role == string(enums.AdminRoleRootAdmin) || role == string(enums.AdminRoleSuperAdmin)
}

func validKey(key string) bool {
	if key == "" || len(key) > 80 || key[0] < 'A' || key[0] > 'Z' {
		return false
	}
	for _, char := range key[1:] {
		if (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '_' {
			return false
		}
	}
	return true
}

func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "unique")
}

func toVO(row toolmodel.ToolConfig) vo.ToolConfig {
	return vo.ToolConfig{ID: row.ID, ToolKey: row.ToolKey, DisplayName: row.DisplayName, Description: row.Description, Visible: row.Visible, Enabled: row.Enabled, Pinned: row.Pinned, Highlighted: row.Highlighted, SortOrder: row.SortOrder, UpdatedAt: row.UpdatedAt}
}

func toolError(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}

func stringPtr(value string) *string { return &value }
