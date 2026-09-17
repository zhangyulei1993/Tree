package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"gorm.io/gorm"

	choicedto "tree/backend/internal/choicescenario/dto"
	choicemodel "tree/backend/internal/choicescenario/model"
	choicerepo "tree/backend/internal/choicescenario/repository"
	"tree/backend/internal/choicescenario/vo"
	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	MaxTitleLength       = 20
	MaxOptionLength      = 30
	MaxOptionCount       = 20
	MaxOptionsTextLength = 500
	MaxPublishPerDay     = 10
	ScenarioTTL          = 30 * 24 * time.Hour
	ListLimit            = 30
)

type builtinChoiceScenario struct {
	title   string
	options []string
}

var builtinChoiceScenarios = []builtinChoiceScenario{
	{title: "今天吃什么", options: []string{"家常菜", "火锅", "烧烤", "面食", "外卖"}},
	{title: "下班去干什么", options: []string{"打篮球", "去健身房", "逛街", "看电影", "回家休息"}},
	{title: "周末去哪儿", options: []string{"公园散步", "看电影", "逛街", "周边短途游", "在家休息"}},
	{title: "今晚看什么", options: []string{"纪录片", "喜剧片", "电视剧", "综艺", "读书"}},
	{title: "先做哪一件事", options: []string{"工作任务", "运动锻炼", "整理房间", "回复消息", "休息一下"}},
}

type AuditInput struct {
	IP        string
	UserAgent string
}

type Service interface {
	List(context.Context, choicedto.ListQuery) ([]vo.ChoiceScenario, *apperrors.BusinessError)
	Create(context.Context, uint64, choicedto.CreateRequest, AuditInput) (*vo.ChoiceScenario, *apperrors.BusinessError)
	Use(context.Context, uint64, uint64) *apperrors.BusinessError
}

type service struct {
	uow           choicerepo.UnitOfWork
	repo          choicerepo.Repository
	contentSafety contentsafety.Service
	logs          operationlog.Service
	now           func() time.Time
}

func New(db *gorm.DB, repo choicerepo.Repository, contentSafety contentsafety.Service, logs operationlog.Service) Service {
	if contentSafety == nil {
		contentSafety = contentsafety.FailClosed()
	}
	return &service{
		uow:           choicerepo.NewUnitOfWork(db, repo),
		repo:          repo,
		contentSafety: contentSafety,
		logs:          logs,
		now:           time.Now,
	}
}

func (s *service) List(ctx context.Context, query choicedto.ListQuery) ([]vo.ChoiceScenario, *apperrors.BusinessError) {
	now := s.now()
	from, ok := periodStart(now, query.Period)
	if !ok {
		return nil, apperrors.New(apperrors.CodeChoiceScenarioInvalidInput)
	}
	rows, err := s.repo.List(ctx, from, now, ListLimit)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result := make([]vo.ChoiceScenario, 0, len(rows))
	for _, row := range rows {
		item, err := toVO(row)
		if err != nil {
			continue
		}
		if isBuiltinChoiceScenario(item.Title, item.Options) {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *service) Create(ctx context.Context, userID uint64, req choicedto.CreateRequest, audit AuditInput) (*vo.ChoiceScenario, *apperrors.BusinessError) {
	title, options, err := normalizeInput(req)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeChoiceScenarioInvalidInput)
	}
	if isBuiltinChoiceScenario(title, options) {
		return nil, apperrors.New(apperrors.CodeChoiceScenarioInvalidInput)
	}
	dayStart, _ := periodStart(s.now(), "today")
	count, err := s.repo.CountCreatedByUserSince(ctx, userID, dayStart)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if count >= MaxPublishPerDay {
		return nil, apperrors.New(apperrors.CodeChoiceScenarioRateLimited)
	}
	fields := []contentsafety.Field{{Label: "场景标题", Value: title}}
	for index, option := range options {
		fields = append(fields, contentsafety.Field{Label: "选项" + string(rune('A'+index)), Value: option})
	}
	if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
		UserID: userID, Scene: contentsafety.SceneSocial, Fields: fields,
		LogAction: "CREATE_CHOICE_SCENARIO", IP: audit.IP, UserAgent: audit.UserAgent,
	}); businessErr != nil {
		return nil, businessErr
	}
	optionsJSON, _ := json.Marshal(options)
	now := s.now()
	created := &choicemodel.ChoiceScenario{
		UserID: userID, Title: title, OptionsJSON: optionsJSON, Status: "ACTIVE",
		ExpiresAt: now.Add(ScenarioTTL), CreatedAt: now, UpdatedAt: now,
	}
	err = s.uow.WithinTransaction(ctx, func(repo choicerepo.Repository) error {
		if err := repo.Create(ctx, created); err != nil {
			return err
		}
		if s.logs == nil {
			return nil
		}
		return repo.WriteLog(ctx, operationlog.WriteInput{
			OperatorType: string(enums.OperatorTypeUser), OperatorUserID: &userID,
			Module: "CHOICE_SCENARIO", Action: "CREATE", TargetType: stringPtr("CHOICE_SCENARIO"),
			TargetID: &created.ID, DetailJSON: detailJSON(title, len(options)),
			IP: optionalString(audit.IP), UserAgent: optionalString(audit.UserAgent),
		})
	})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	result, err := toVO(*created)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &result, nil
}

func (s *service) Use(ctx context.Context, userID, scenarioID uint64) *apperrors.BusinessError {
	if userID == 0 || scenarioID == 0 {
		return apperrors.New(apperrors.CodeChoiceScenarioInvalidInput)
	}
	if err := s.repo.IncrementUseCount(ctx, scenarioID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.CodeChoiceScenarioNotFound)
		}
		return apperrors.New(apperrors.CodeSystemError)
	}
	return nil
}

func normalizeInput(req choicedto.CreateRequest) (string, []string, error) {
	if hasControl(strings.TrimSpace(req.Title)) {
		return "", nil, errors.New("invalid title")
	}
	title := normalizeText(req.Title)
	if title == "" || utf8.RuneCountInString(title) > MaxTitleLength {
		return "", nil, errors.New("invalid title")
	}
	if len(req.Options) < 2 || len(req.Options) > MaxOptionCount {
		return "", nil, errors.New("invalid option count")
	}
	options := make([]string, 0, len(req.Options))
	seen := make(map[string]struct{}, len(req.Options))
	for _, value := range req.Options {
		if hasControl(strings.TrimSpace(value)) {
			return "", nil, errors.New("invalid option")
		}
		option := normalizeText(value)
		if option == "" || utf8.RuneCountInString(option) > MaxOptionLength {
			return "", nil, errors.New("invalid option")
		}
		if _, exists := seen[option]; exists {
			return "", nil, errors.New("duplicate option")
		}
		seen[option] = struct{}{}
		options = append(options, option)
	}
	encoded, _ := json.Marshal(options)
	if utf8.RuneCountInString(string(encoded)) > MaxOptionsTextLength {
		return "", nil, errors.New("options too long")
	}
	return title, options, nil
}

func normalizeText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func hasControl(value string) bool {
	for _, char := range value {
		if unicode.IsControl(char) {
			return true
		}
	}
	return false
}

func periodStart(now time.Time, period string) (time.Time, bool) {
	switch period {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), true
	case "7d":
		return now.AddDate(0, 0, -7), true
	case "month":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), true
	default:
		return time.Time{}, false
	}
}

func toVO(row choicemodel.ChoiceScenario) (vo.ChoiceScenario, error) {
	var options []string
	if err := json.Unmarshal(row.OptionsJSON, &options); err != nil {
		return vo.ChoiceScenario{}, err
	}
	return vo.ChoiceScenario{ID: row.ID, Title: row.Title, Options: options, UseCount: row.UseCount, CreatedAt: row.CreatedAt}, nil
}

func isBuiltinChoiceScenario(title string, options []string) bool {
	for _, builtin := range builtinChoiceScenarios {
		if builtin.title != title || len(builtin.options) != len(options) {
			continue
		}
		matched := true
		for index, option := range builtin.options {
			if option != options[index] {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

func detailJSON(title string, optionCount int) []byte {
	data, _ := json.Marshal(map[string]any{"title": title, "optionCount": optionCount})
	return data
}

func stringPtr(value string) *string { return &value }

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
