package service

import (
	"context"

	"gorm.io/gorm"

	"tree/backend/internal/operationlog/model"
)

const (
	ResultSuccess = "SUCCESS"
	ResultFailed  = "FAILED"
)

type WriteInput struct {
	OperatorType    string
	OperatorAdminID *uint64
	OperatorUserID  *uint64
	OperatorRole    *string
	Module          string
	Action          string
	TargetType      *string
	TargetID        *uint64
	FamilyID        *uint64
	MemberID        *uint64
	UserID          *uint64
	BeforeJSON      []byte
	AfterJSON       []byte
	DetailJSON      []byte
	ErrorMessage    *string
	IP              *string
	UserAgent       *string
}

type Service interface {
	WriteSuccess(ctx context.Context, input WriteInput) error
	WriteFailed(ctx context.Context, input WriteInput) error
}

type NoopService struct{}

func (NoopService) WriteSuccess(context.Context, WriteInput) error {
	return nil
}

func (NoopService) WriteFailed(context.Context, WriteInput) error {
	return nil
}

type GormService struct {
	db *gorm.DB
}

func NewGormService(db *gorm.DB) *GormService {
	return &GormService{db: db}
}

func (s *GormService) WriteSuccess(ctx context.Context, input WriteInput) error {
	return s.write(ctx, input, ResultSuccess)
}

func (s *GormService) WriteFailed(ctx context.Context, input WriteInput) error {
	return s.write(ctx, input, ResultFailed)
}

func (s *GormService) write(ctx context.Context, input WriteInput, result string) error {
	if s == nil || s.db == nil {
		return nil
	}

	detailJSON, err := sanitizeDetailJSON(input.DetailJSON)
	if err != nil {
		return err
	}
	record := model.OperationLog{
		OperatorType:    input.OperatorType,
		OperatorAdminID: input.OperatorAdminID,
		OperatorUserID:  input.OperatorUserID,
		OperatorRole:    input.OperatorRole,
		Module:          input.Module,
		Action:          input.Action,
		TargetType:      input.TargetType,
		TargetID:        input.TargetID,
		FamilyID:        input.FamilyID,
		MemberID:        input.MemberID,
		UserID:          input.UserID,
		BeforeJSON:      input.BeforeJSON,
		AfterJSON:       input.AfterJSON,
		DetailJSON:      detailJSON,
		Result:          result,
		ErrorMessage:    input.ErrorMessage,
		IP:              input.IP,
		UserAgent:       input.UserAgent,
	}

	return s.db.WithContext(ctx).Create(&record).Error
}
