package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"

	authmodel "tree/backend/internal/auth/model"
	"tree/backend/internal/auth/repository"
	"tree/backend/internal/auth/vo"
	apperrors "tree/backend/internal/common/errors"
	commonjwt "tree/backend/internal/common/jwt"
	"tree/backend/internal/common/redis"
	"tree/backend/internal/common/security"
	operationlog "tree/backend/internal/operationlog/service"
)

const assistanceModule = "USER_ASSISTANCE"
const assistanceMaxFailures = 5

type AssistanceService struct {
	db         *gorm.DB
	users      repository.UserRepository
	codes      repository.AssistanceCodeRepository
	jwt        *commonjwt.Manager
	blacklist  redis.TokenBlacklist
	logs       operationlog.Service
	hashSecret string
}
type GenerateAssistanceInput struct {
	UserID        uint64
	ValidDays     int
	IP, UserAgent string
}
type RedeemAssistanceInput struct{ Phone, Code, IP, UserAgent string }
type RevokeAssistanceInput struct {
	UserID, CodeID uint64
	IP, UserAgent  string
}

func NewAssistanceService(db *gorm.DB, users repository.UserRepository, codes repository.AssistanceCodeRepository, jwt *commonjwt.Manager, blacklist redis.TokenBlacklist, logs operationlog.Service, hashSecret string) *AssistanceService {
	return &AssistanceService{db: db, users: users, codes: codes, jwt: jwt, blacklist: blacklist, logs: logs, hashSecret: hashSecret}
}

func (s *AssistanceService) Generate(ctx context.Context, in GenerateAssistanceInput) (*vo.AssistanceCodeGenerated, *apperrors.BusinessError) {
	if in.ValidDays == 0 {
		in.ValidDays = 1
	}
	if in.ValidDays != 1 && in.ValidDays != 3 && in.ValidDays != 7 {
		return nil, apperrors.New(apperrors.CodeInvalidParams)
	}
	user, err := s.users.FindByID(ctx, in.UserID)
	if err != nil || user.Status != "ACTIVE" || !user.PhoneVerified {
		return nil, apperrors.New(apperrors.CodePhoneBindRequired)
	}
	code, err := generateAssistanceCode()
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	now := time.Now()
	expires := now.Add(time.Duration(in.ValidDays) * 24 * time.Hour)
	hash := security.HashAssistanceCode(s.hashSecret, code)
	var replaced []authmodel.AssistanceCode
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var txErr error
		replaced, txErr = s.codes.WithTx(tx).ReplaceActiveForOwner(ctx, in.UserID, now)
		if txErr != nil {
			return txErr
		}
		return s.codes.WithTx(tx).Create(ctx, &authmodel.AssistanceCode{OwnerUserID: in.UserID, CodeHash: hash, CodeStatus: authmodel.AssistanceCodeActive, ExpiresAt: expires})
	})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	for _, old := range replaced {
		if old.SessionTokenID != nil {
			_ = s.blacklist.Revoke(ctx, *old.SessionTokenID, time.Until(old.ExpiresAt))
		}
	}
	s.log(ctx, in.UserID, "GENERATE_ASSISTANCE_CODE", in.IP, in.UserAgent, true)
	return &vo.AssistanceCodeGenerated{Code: code, ExpiresAt: expires, ValidDays: in.ValidDays, ReplacedPrevious: len(replaced) > 0}, nil
}

func (s *AssistanceService) List(ctx context.Context, userID uint64) ([]vo.AssistanceCodeStatus, *apperrors.BusinessError) {
	items, err := s.codes.ListByOwner(ctx, userID, 20)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	now := time.Now()
	out := make([]vo.AssistanceCodeStatus, 0, len(items))
	for _, v := range items {
		status := v.CodeStatus
		if status == authmodel.AssistanceCodeActive && !v.ExpiresAt.After(now) {
			status = authmodel.AssistanceCodeExpired
			_ = s.codes.Update(ctx, v.ID, map[string]any{"code_status": status})
		}
		out = append(out, vo.AssistanceCodeStatus{ID: v.ID, Status: status, ExpiresAt: v.ExpiresAt, CreatedAt: v.CreatedAt, LastUsedAt: v.LastUsedAt})
	}
	return out, nil
}

func (s *AssistanceService) Revoke(ctx context.Context, in RevokeAssistanceInput) *apperrors.BusinessError {
	var item *authmodel.AssistanceCode
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var e error
		item, e = s.codes.WithTx(tx).FindByIDForUpdate(ctx, in.CodeID)
		if e != nil {
			return e
		}
		if item.OwnerUserID != in.UserID {
			return gorm.ErrRecordNotFound
		}
		if item.CodeStatus != authmodel.AssistanceCodeActive {
			return nil
		}
		now := time.Now()
		return s.codes.WithTx(tx).Update(ctx, item.ID, map[string]any{"code_status": authmodel.AssistanceCodeRevoked, "revoked_at": now})
	})
	if err != nil {
		return apperrors.New(apperrors.CodeNotFound)
	}
	if item != nil && item.SessionTokenID != nil {
		_ = s.blacklist.Revoke(ctx, *item.SessionTokenID, time.Until(item.ExpiresAt))
	}
	s.log(ctx, in.UserID, "REVOKE_ASSISTANCE_CODE", in.IP, in.UserAgent, true)
	return nil
}

func (s *AssistanceService) Redeem(ctx context.Context, in RedeemAssistanceInput) (*vo.LoginResponse, *apperrors.BusinessError) {
	if len(in.Code) != 8 || !security.ValidPhone(in.Phone) {
		return nil, apperrors.New(apperrors.CodeUnauthorized)
	}
	now := time.Now()
	hash := security.HashAssistanceCode(s.hashSecret, in.Code)
	var item *authmodel.AssistanceCode
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		item, err = s.codes.WithTx(tx).FindByHashForUpdate(ctx, hash)
		if err != nil {
			return err
		}
		if item.CodeStatus != authmodel.AssistanceCodeActive || !item.ExpiresAt.After(now) {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.New(apperrors.CodeUnauthorized)
	}
	user, err := s.users.FindByID(ctx, item.OwnerUserID)
	if err != nil || user.Status != "ACTIVE" || user.Phone == nil || *user.Phone != in.Phone {
		return nil, apperrors.New(apperrors.CodeUnauthorized)
	}
	tokenID, err := newAssistanceTokenID()
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if item.SessionTokenID != nil {
		_ = s.blacklist.Revoke(ctx, *item.SessionTokenID, time.Until(item.ExpiresAt))
	}
	if err = s.codes.Update(ctx, item.ID, map[string]any{"session_token_id": tokenID, "last_used_at": now}); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	token, err := s.jwt.GenerateToken(user.ID, commonjwt.TokenTypeUser, "ASSISTANCE", tokenID, time.Until(item.ExpiresAt))
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	s.log(ctx, user.ID, "REDEEM_ASSISTANCE_CODE", in.IP, in.UserAgent, true)
	return &vo.LoginResponse{AccessToken: token, TokenType: "Bearer", User: userInfo(user)}, nil
}

func generateAssistanceCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(100000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08d", n.Int64()), nil
}
func newAssistanceTokenID() (string, error) {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return fmt.Sprintf("assist-%x", raw), nil
}
func (s *AssistanceService) log(ctx context.Context, userID uint64, action, ip, ua string, ok bool) {
	id := userID
	in := operationlog.WriteInput{OperatorType: "USER", OperatorUserID: &id, UserID: &id, Module: assistanceModule, Action: action, IP: &ip, UserAgent: &ua}
	if ok {
		_ = s.logs.WriteSuccess(ctx, in)
	} else {
		_ = s.logs.WriteFailed(ctx, in)
	}
}
