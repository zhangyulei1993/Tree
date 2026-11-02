package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"strings"
	"time"

	authrepo "tree/backend/internal/auth/repository"
	"tree/backend/internal/common/config"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/wechat"
	"tree/backend/internal/notification/model"
	"tree/backend/internal/notification/repository"
)

type Input struct {
	UserID      uint64
	Category    string
	Title       string
	Content     string
	Route       string
	RelatedType string
	RelatedID   uint64
	Subscribe   *SubscribeInput
}

type SubscribeInput struct {
	Kind   string
	Page   string
	Values []string
}

type SubscribeSender interface {
	Send(context.Context, uint64, SubscribeInput) error
}

type Item struct {
	ID        uint64 `json:"id"`
	Category  string `json:"category"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Route     string `json:"route,omitempty"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"createdAt"`
}

type Page struct {
	Items       []Item `json:"items"`
	Total       int64  `json:"total"`
	UnreadCount int64  `json:"unreadCount"`
	Page        int    `json:"page"`
	PageSize    int    `json:"pageSize"`
}

type Service interface {
	Create(context.Context, Input)
	List(context.Context, uint64, int, int) (*Page, *apperrors.BusinessError)
	MarkRead(context.Context, uint64, uint64) *apperrors.BusinessError
	MarkAllRead(context.Context, uint64) *apperrors.BusinessError
}

type CreateOnceService interface{ CreateOnce(context.Context, Input) }

type service struct {
	repo      repository.Repository
	push      SubscribeSender
	logger    *zap.Logger
	pushQueue chan pushJob
}

type pushJob struct {
	jobID  uint64
	userID uint64
	input  SubscribeInput
}

func New(repo repository.Repository) Service { return &service{repo: repo} }

func NewWithSubscribe(repo repository.Repository, users authrepo.UserRepository, client *wechat.MiniProgramClient, cfg config.WechatConfig, logger *zap.Logger) Service {
	push := newSubscribeSender(users, client, cfg, logger)
	s := &service{repo: repo, push: push, logger: logger}
	if push != nil {
		s.pushQueue = make(chan pushJob, 256)
		go s.runPushWorker()
		if store, ok := repo.(repository.DeliveryJobStore); ok {
			if jobs, err := store.ListReadyDeliveryJobs(context.Background(), 256); err == nil {
				for _, job := range jobs {
					var values []string
					if json.Unmarshal(job.ValuesJSON, &values) == nil {
						page := ""
						if job.Page != nil {
							page = *job.Page
						}
						s.pushQueue <- pushJob{jobID: job.ID, userID: job.UserID, input: SubscribeInput{Kind: job.Kind, Page: page, Values: values}}
					}
				}
			}
		}
	}
	return s
}

func (s *service) Create(ctx context.Context, input Input) {
	if input.UserID == 0 || strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Content) == "" {
		return
	}
	row := &model.Notification{UserID: input.UserID, Category: truncate(input.Category, 40), Title: truncate(input.Title, 120), Content: truncate(input.Content, 500), Route: optional(input.Route), RelatedType: optional(input.RelatedType)}
	if input.RelatedID != 0 {
		row.RelatedID = &input.RelatedID
	}
	if err := s.repo.Create(ctx, row); err != nil {
		if s.logger != nil {
			s.logger.Error("create in-app notification failed", zap.Error(err), zap.Uint64("userId", input.UserID), zap.String("category", row.Category), zap.String("relatedType", optionalValue(row.RelatedType)))
		}
		return
	}
	if s.push != nil && input.Subscribe != nil {
		jobID := uint64(0)
		if store, ok := s.repo.(repository.DeliveryJobStore); ok {
			jobID, _ = store.CreateDeliveryJob(ctx, input.UserID, input.Subscribe.Kind, input.Subscribe.Page, input.Subscribe.Values)
		}
		job := pushJob{jobID: jobID, userID: input.UserID, input: *input.Subscribe}
		select {
		case s.pushQueue <- job:
		default:
			if s.logger != nil {
				s.logger.Warn("wechat subscription queue is full; notification kept in app only", zap.Uint64("userId", input.UserID), zap.String("kind", input.Subscribe.Kind))
			}
		}
	}
}

func (s *service) CreateOnce(ctx context.Context, input Input) {
	if input.RelatedType != "" && input.RelatedID != 0 {
		exists, err := s.repo.Exists(ctx, input.UserID, input.RelatedType, input.RelatedID)
		if err != nil || exists {
			return
		}
	}
	s.Create(ctx, input)
}

func (s *service) runPushWorker() {
	refresh := time.NewTicker(time.Minute)
	defer refresh.Stop()
	for {
		select {
		case job := <-s.pushQueue:
			for attempt := 1; attempt <= 3; attempt++ {
				ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
				err := s.push.Send(ctx, job.userID, job.input)
				cancel()
				if err == nil || strings.Contains(err.Error(), "43101") {
					if job.jobID != 0 {
						_ = s.markDeliverySent(job.jobID)
					}
					break
				}
				if attempt < 3 {
					time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
				} else if s.logger != nil {
					if job.jobID != 0 {
						_ = s.markDeliveryRetry(job.jobID, err.Error(), time.Now().Add(time.Minute))
					}
					s.logger.Warn("wechat subscription notification failed after retries", zap.Error(err), zap.Uint64("userId", job.userID), zap.String("kind", job.input.Kind))
				}
			}
		case <-refresh.C:
			s.enqueueReadyJobs()
		}
	}
}

func (s *service) enqueueReadyJobs() {
	store, ok := s.repo.(repository.DeliveryJobStore)
	if !ok {
		return
	}
	jobs, err := store.ListReadyDeliveryJobs(context.Background(), 256)
	if err != nil {
		return
	}
	for _, job := range jobs {
		var values []string
		if json.Unmarshal(job.ValuesJSON, &values) != nil {
			continue
		}
		page := ""
		if job.Page != nil {
			page = *job.Page
		}
		select {
		case s.pushQueue <- pushJob{jobID: job.ID, userID: job.UserID, input: SubscribeInput{Kind: job.Kind, Page: page, Values: values}}:
		default:
			return
		}
	}
}

func (s *service) markDeliverySent(id uint64) error {
	store, ok := s.repo.(repository.DeliveryJobStore)
	if !ok {
		return nil
	}
	return store.MarkDeliverySent(context.Background(), id)
}

func (s *service) markDeliveryRetry(id uint64, message string, next time.Time) error {
	store, ok := s.repo.(repository.DeliveryJobStore)
	if !ok {
		return nil
	}
	return store.MarkDeliveryRetry(context.Background(), id, message, next)
}

type subscribeSender struct {
	users  authrepo.UserRepository
	client *wechat.MiniProgramClient
	appID  string
	cfg    config.WechatSubscribeConfig
	logger *zap.Logger
}

func newSubscribeSender(users authrepo.UserRepository, client *wechat.MiniProgramClient, cfg config.WechatConfig, logger *zap.Logger) SubscribeSender {
	if !cfg.Subscribe.Enabled || users == nil || client == nil {
		return nil
	}
	return &subscribeSender{users: users, client: client, appID: cfg.MiniAppID, cfg: cfg.Subscribe, logger: logger}
}

func (s *subscribeSender) Send(ctx context.Context, userID uint64, input SubscribeInput) error {
	if userID == 0 || len(input.Values) == 0 {
		return nil
	}
	templateID := ""
	switch input.Kind {
	case "PAYMENT":
		templateID = s.cfg.PaymentTemplateID
	case "REFUND":
		templateID = s.cfg.RefundTemplateID
	case "AUDIT":
		templateID = s.cfg.AuditTemplateID
	case "MEMBER":
		templateID = s.cfg.MemberTemplateID
	case "MEMBER_CHANGE":
		templateID = s.cfg.MemberChangeTemplateID
	}
	if templateID == "" {
		if s.logger != nil {
			s.logger.Warn("wechat subscription skipped: template is not configured", zap.String("kind", input.Kind), zap.Uint64("userId", userID))
		}
		return nil
	}
	openID, err := s.users.FindActiveWechatMiniOpenID(ctx, userID, s.appID)
	if err != nil || openID == "" {
		if s.logger != nil {
			s.logger.Warn("wechat subscription skipped: active mini program identity not found", zap.Uint64("userId", userID), zap.Error(err))
		}
		return err
	}
	if err := s.client.SendSubscribeMessage(ctx, openID, templateID, input.Page, s.cfg.MiniProgramState, input.Values); err != nil {
		if strings.Contains(err.Error(), "43101") {
			if s.logger != nil {
				s.logger.Warn("wechat subscription send skipped: user has not accepted or has already consumed the subscription", zap.Uint64("userId", userID), zap.String("kind", input.Kind))
			}
			return err
		}
		if s.logger != nil {
			s.logger.Warn("wechat subscription send failed", zap.Uint64("userId", userID), zap.String("kind", input.Kind), zap.Error(err))
		}
		return err
	}
	return nil
}

func (s *service) List(ctx context.Context, userID uint64, page, pageSize int) (*Page, *apperrors.BusinessError) {
	if userID == 0 {
		return nil, apperrors.New(apperrors.CodeUnauthorized)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	rows, total, err := s.repo.List(ctx, userID, page, pageSize)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	unreadCount, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		item := Item{ID: row.ID, Category: row.Category, Title: row.Title, Content: row.Content, Read: row.ReadAt != nil, CreatedAt: row.CreatedAt.Format(time.RFC3339)}
		if row.Route != nil {
			item.Route = *row.Route
		}
		items = append(items, item)
	}
	return &Page{Items: items, Total: total, UnreadCount: unreadCount, Page: page, PageSize: pageSize}, nil
}

func (s *service) MarkRead(ctx context.Context, userID, id uint64) *apperrors.BusinessError {
	if userID == 0 || id == 0 {
		return apperrors.New(apperrors.CodeInvalidParams)
	}
	if err := s.repo.MarkRead(ctx, userID, id, time.Now()); err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	return nil
}

func (s *service) MarkAllRead(ctx context.Context, userID uint64) *apperrors.BusinessError {
	if userID == 0 {
		return apperrors.New(apperrors.CodeUnauthorized)
	}
	if err := s.repo.MarkAllRead(ctx, userID, time.Now()); err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	return nil
}

func optional(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	chars := []rune(value)
	if len(chars) > limit {
		return string(chars[:limit])
	}
	return value
}
func CapacityRoute(familyID uint64, orderNo string) string {
	return fmt.Sprintf("/pages/family/capacity-orders?familyId=%d&orderNo=%s", familyID, orderNo)
}

func CapacityPageRoute(familyID uint64, orderNo string) string {
	return fmt.Sprintf("pages/family/capacity-orders?familyId=%d&orderNo=%s", familyID, orderNo)
}

func optionalValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
