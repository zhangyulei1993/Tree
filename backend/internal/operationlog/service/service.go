package service

import "context"

type WriteInput struct {
	OperatorType string
	OperatorID   uint64
	Module       string
	Action       string
	TargetType   string
	TargetID     uint64
	Result       string
	IP           string
	UserAgent    string
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
