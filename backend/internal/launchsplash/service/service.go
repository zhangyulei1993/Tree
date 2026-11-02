package service

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
	apperrors "tree/backend/internal/common/errors"
	contentservice "tree/backend/internal/content/service"
	launchdto "tree/backend/internal/launchsplash/dto"
	launchmodel "tree/backend/internal/launchsplash/model"
	launchrepo "tree/backend/internal/launchsplash/repository"
	"tree/backend/internal/launchsplash/vo"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	CodeNotFound     apperrors.Code = 49301
	CodeInvalidInput apperrors.Code = 49302
)

type Service interface {
	Public(context.Context) (*vo.Config, *apperrors.BusinessError)
	AdminGet(context.Context) (*vo.Config, *apperrors.BusinessError)
	AdminUpdate(context.Context, uint64, string, launchdto.UpdateRequest) (*vo.Config, *apperrors.BusinessError)
}

type service struct {
	repo launchrepo.Repository
	logs operationlog.Service
}

func New(repo launchrepo.Repository, logs operationlog.Service) Service {
	return &service{repo: repo, logs: logs}
}

func (s *service) Public(ctx context.Context) (*vo.Config, *apperrors.BusinessError) {
	return s.get(ctx)
}
func (s *service) AdminGet(ctx context.Context) (*vo.Config, *apperrors.BusinessError) {
	return s.get(ctx)
}

func (s *service) get(ctx context.Context) (*vo.Config, *apperrors.BusinessError) {
	row, err := s.repo.Find(ctx, "HOME")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, business(CodeNotFound, "开屏配置不存在")
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return ptr(toVO(*row)), nil
}

func (s *service) AdminUpdate(ctx context.Context, adminID uint64, role string, req launchdto.UpdateRequest) (*vo.Config, *apperrors.BusinessError) {
	if !canManage(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	title, tagline := strings.TrimSpace(req.Title), strings.TrimSpace(req.Tagline)
	imageURL := strings.TrimSpace(req.ImageURL)
	dayImageURL, nightImageURL := strings.TrimSpace(req.DayImageURL), strings.TrimSpace(req.NightImageURL)
	if title == "" || len([]rune(title)) > 120 || tagline == "" || len([]rune(tagline)) > 240 || len([]rune(imageURL)) > 500 || len([]rune(dayImageURL)) > 500 || len([]rune(nightImageURL)) > 500 || req.DurationMs < 1500 || req.DurationMs > 15000 {
		return nil, business(CodeInvalidInput, "标题、副标题不能为空，时长须在 1.5 至 15 秒之间")
	}
	row, err := s.repo.Find(ctx, "HOME")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, business(CodeNotFound, "开屏配置不存在")
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	values := map[string]any{"title": title, "tagline": tagline, "image_url": imageURL, "day_image_url": dayImageURL, "night_image_url": nightImageURL, "duration_ms": req.DurationMs, "updated_by_admin_id": adminID}
	if req.Enabled != nil {
		values["enabled"] = *req.Enabled
	}
	if req.SkipEnabled != nil {
		values["skip_enabled"] = *req.SkipEnabled
	}
	if err := s.repo.Update(ctx, "HOME", values); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	updated, err := s.repo.Find(ctx, "HOME")
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	targetType := "launch_splash_config"
	newImagePaths := map[string]struct{}{imageURL: {}, dayImageURL: {}, nightImageURL: {}}
	for _, oldImagePath := range []string{row.ImageURL, row.DayImageURL, row.NightImageURL} {
		if oldImagePath == "" {
			continue
		}
		if _, stillUsed := newImagePaths[oldImagePath]; stillUsed {
			continue
		}
		if deleteErr := contentservice.DeleteContentImageByPublicPath(oldImagePath); deleteErr != nil && s.logs != nil {
			detail := fmt.Sprintf(`{"path":%q}`, oldImagePath)
			errText := deleteErr.Error()
			_ = s.logs.WriteFailed(ctx, operationlog.WriteInput{
				OperatorType: "SYSTEM",
				Module:       "LAUNCH_SPLASH",
				Action:       "DELETE_REPLACED_IMAGE",
				TargetType:   &targetType,
				DetailJSON:   []byte(detail),
				ErrorMessage: &errText,
			})
		}
	}
	admin, roleCopy := adminID, role
	if s.logs != nil {
		_ = s.logs.WriteSuccess(ctx, operationlog.WriteInput{OperatorType: "ADMIN", OperatorAdminID: &admin, OperatorRole: &roleCopy, Module: "LAUNCH_SPLASH", Action: "UPDATE_LAUNCH_SPLASH", TargetType: &targetType, TargetID: &updated.ID})
	}
	return ptr(toVO(*updated)), nil
}

func canManage(role string) bool {
	return role == "ROOT_ADMIN" || role == "SUPER_ADMIN"
}

func toVO(row launchmodel.Config) vo.Config {
	return vo.Config{ID: row.ID, ConfigKey: row.ConfigKey, Enabled: row.Enabled, Title: row.Title, Tagline: row.Tagline, ImageURL: row.ImageURL, DayImageURL: row.DayImageURL, NightImageURL: row.NightImageURL, DurationMs: row.DurationMs, SkipEnabled: row.SkipEnabled, UpdatedAt: row.UpdatedAt}
}
func ptr(value vo.Config) *vo.Config { return &value }
func business(code apperrors.Code, message string) *apperrors.BusinessError {
	return &apperrors.BusinessError{Code: code, Message: message}
}
