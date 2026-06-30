package service

import (
	"context"

	"tree/backend/internal/admin/repository"
	apperrors "tree/backend/internal/common/errors"
)

type PageResult struct {
	Items    any   `json:"items"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}
type ManagementService struct {
	repo *repository.ManagementRepository
}

func NewManagementService(repo *repository.ManagementRepository) *ManagementService {
	return &ManagementService{repo: repo}
}
func normalize(query repository.PageQuery) repository.PageQuery {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	return query
}
func (s *ManagementService) Dashboard(ctx context.Context) (any, *apperrors.BusinessError) {
	value, err := s.repo.Dashboard(ctx)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return value, nil
}
func (s *ManagementService) Users(ctx context.Context, q repository.PageQuery) (any, *apperrors.BusinessError) {
	q = normalize(q)
	rows, total, err := s.repo.Users(ctx, q)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return PageResult{rows, q.Page, q.PageSize, total}, nil
}
func (s *ManagementService) User(ctx context.Context, id uint64) (any, *apperrors.BusinessError) {
	value, err := s.repo.UserDetail(ctx, id)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeResourceNotFound)
	}
	return value, nil
}
func (s *ManagementService) Families(ctx context.Context, q repository.PageQuery) (any, *apperrors.BusinessError) {
	q = normalize(q)
	rows, total, err := s.repo.Families(ctx, q)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return PageResult{rows, q.Page, q.PageSize, total}, nil
}
func (s *ManagementService) Family(ctx context.Context, id uint64) (any, *apperrors.BusinessError) {
	value, err := s.repo.Family(ctx, id)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeResourceNotFound)
	}
	return value, nil
}
func (s *ManagementService) Members(ctx context.Context, id uint64) (any, *apperrors.BusinessError) {
	value, err := s.repo.Members(ctx, id)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return value, nil
}
func (s *ManagementService) Admins(ctx context.Context, q repository.PageQuery) (any, *apperrors.BusinessError) {
	q = normalize(q)
	rows, total, err := s.repo.Admins(ctx, q)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return PageResult{rows, q.Page, q.PageSize, total}, nil
}
func (s *ManagementService) Logs(ctx context.Context, q repository.PageQuery) (any, *apperrors.BusinessError) {
	q = normalize(q)
	rows, total, err := s.repo.Logs(ctx, q)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return PageResult{rows, q.Page, q.PageSize, total}, nil
}
