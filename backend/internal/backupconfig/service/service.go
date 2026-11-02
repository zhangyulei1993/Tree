package service

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"strconv"
	"strings"

	backupdto "tree/backend/internal/backupconfig/dto"
	backupmodel "tree/backend/internal/backupconfig/model"
	backuprepo "tree/backend/internal/backupconfig/repository"
	backupvo "tree/backend/internal/backupconfig/vo"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	operationlog "tree/backend/internal/operationlog/service"
)

const backupDirectoryDefault = "/opt/tree-staging/shared/backups"

type Service interface {
	Overview(context.Context, string) (*backupvo.Overview, *apperrors.BusinessError)
	Update(context.Context, uint64, string, backupdto.UpdateRequest) (*backupvo.Config, *apperrors.BusinessError)
}

type service struct {
	repo            backuprepo.Repository
	logs            operationlog.Service
	backupDirectory string
}

func New(repo backuprepo.Repository, logs operationlog.Service) Service {
	directory := strings.TrimSpace(os.Getenv("TREE_BACKUP_DIRECTORY"))
	if directory == "" {
		directory = backupDirectoryDefault
	}
	return &service{repo: repo, logs: logs, backupDirectory: directory}
}

func (s *service) Overview(ctx context.Context, role string) (*backupvo.Overview, *apperrors.BusinessError) {
	if !canManage(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	row, err := s.repo.Find(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	status, err := readStatus(s.backupDirectory)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &backupvo.Overview{Config: toVO(*row), Status: status}, nil
}

func (s *service) Update(ctx context.Context, adminID uint64, role string, req backupdto.UpdateRequest) (*backupvo.Config, *apperrors.BusinessError) {
	if !canManage(role) {
		return nil, apperrors.New(apperrors.CodeForbidden)
	}
	if req.Enabled == nil || req.RetentionDays < 1 || req.RetentionDays > 90 || req.MaxStorageMB < 64 || req.MaxStorageMB > 102400 || req.MaxDiskUsagePercent < 60 || req.MaxDiskUsagePercent > 95 {
		return nil, apperrors.New(apperrors.CodeInvalidParams)
	}
	before, err := s.repo.Find(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	values := map[string]any{"enabled": *req.Enabled, "retention_days": req.RetentionDays, "max_storage_mb": req.MaxStorageMB, "max_disk_usage_percent": req.MaxDiskUsagePercent, "updated_by_admin_id": adminID}
	if err := s.repo.Update(ctx, values); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	updated, err := s.repo.Find(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if s.logs != nil {
		beforeJSON, _ := json.Marshal(toVO(*before))
		afterJSON, _ := json.Marshal(toVO(*updated))
		targetType := "backup_operation_config"
		_ = s.logs.WriteSuccess(ctx, operationlog.WriteInput{OperatorType: string(enums.OperatorTypeAdmin), OperatorAdminID: &adminID, OperatorRole: &role, Module: "BACKUP_OPERATION", Action: "UPDATE_BACKUP_OPERATION_CONFIG", TargetType: &targetType, TargetID: uint64Ptr(1), BeforeJSON: beforeJSON, AfterJSON: afterJSON})
	}
	result := toVO(*updated)
	return &result, nil
}

func canManage(role string) bool {
	return role == string(enums.AdminRoleRootAdmin) || role == string(enums.AdminRoleSuperAdmin)
}
func uint64Ptr(value uint64) *uint64 { return &value }
func toVO(row backupmodel.Config) backupvo.Config {
	return backupvo.Config{Enabled: row.Enabled, RetentionDays: row.RetentionDays, MaxStorageMB: row.MaxStorageMB, MaxDiskUsagePercent: row.MaxDiskUsagePercent, UpdatedAt: row.UpdatedAt}
}

func readStatus(directory string) (backupvo.Status, error) {
	status := backupvo.Status{BackupDirectory: directory}
	entries, err := os.ReadDir(directory)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			entries = nil
		} else {
			return status, err
		}
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql.gz") {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return status, infoErr
		}
		status.BackupCount++
		status.StorageBytes += info.Size()
		modified := info.ModTime()
		if status.LatestBackupAt == nil || modified.After(*status.LatestBackupAt) {
			status.LatestBackupAt = &modified
		}
	}
	output, err := exec.Command("df", "-P", directory).Output()
	if err != nil {
		return status, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return status, errors.New("unexpected df output")
	}
	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) < 5 {
		return status, errors.New("unexpected df fields")
	}
	usageText := strings.TrimSuffix(fields[4], "%")
	usage, err := strconv.Atoi(usageText)
	if err != nil {
		return status, err
	}
	status.DiskUsagePercent = usage
	return status, nil
}
