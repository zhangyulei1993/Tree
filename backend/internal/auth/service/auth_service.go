package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/auth/repository"
	"tree/backend/internal/auth/vo"
	"tree/backend/internal/common/config"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	commonjwt "tree/backend/internal/common/jwt"
	"tree/backend/internal/common/redis"
	"tree/backend/internal/common/security"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"
)

const (
	SceneRegister = "REGISTER"
	SceneLogin    = "LOGIN"

	ClientPCWeb = "PC_WEB"
	ClientH5Web = "H5_WEB"

	authLogModule  = "USER_AUTH"
	actionSendCode = "SEND_CODE"
	actionRegister = "REGISTER_PHONE"
	actionLogin    = "LOGIN_PHONE"
	actionLogout   = "LOGOUT"

	accountOriginPhoneRegister = "PHONE_REGISTER"
)

type SendCodeInput struct {
	Phone      string
	Scene      string
	ClientType string
	IP         string
	UserAgent  string
}

type RegisterPhoneInput struct {
	Phone      string
	Code       string
	Password   string
	Nickname   *string
	ClientType string
	IP         string
	UserAgent  string
}

type LoginPhoneInput struct {
	Phone      string
	Password   string
	ClientType string
	IP         string
	UserAgent  string
}

type LogoutInput struct {
	UserID    uint64
	TokenID   string
	ExpiresAt int64
	IP        string
	UserAgent string
}

type AuthService interface {
	SendCode(ctx context.Context, input SendCodeInput) (*vo.SendCodeResponse, *apperrors.BusinessError)
	RegisterPhone(ctx context.Context, input RegisterPhoneInput) (*vo.LoginResponse, *apperrors.BusinessError)
	LoginPhone(ctx context.Context, input LoginPhoneInput) (*vo.LoginResponse, *apperrors.BusinessError)
	Logout(ctx context.Context, input LogoutInput) *apperrors.BusinessError
}

type PhoneAuthService struct {
	userRepo        repository.UserRepository
	codeRepo        repository.VerificationCodeRepository
	jwtManager      *commonjwt.Manager
	blacklist       redis.TokenBlacklist
	operationLog    operationlog.Service
	appEnv          string
	expireSeconds   int
	cooldownSeconds int
}

func NewPhoneAuthService(
	userRepo repository.UserRepository,
	codeRepo repository.VerificationCodeRepository,
	jwtManager *commonjwt.Manager,
	blacklist redis.TokenBlacklist,
	operationLog operationlog.Service,
	cfg *config.Config,
) *PhoneAuthService {
	expireSeconds := cfg.VerifyCode.ExpireSeconds
	if expireSeconds <= 0 {
		expireSeconds = 300
	}
	cooldownSeconds := cfg.VerifyCode.CooldownSeconds
	if cooldownSeconds <= 0 {
		cooldownSeconds = 60
	}

	return &PhoneAuthService{
		userRepo:        userRepo,
		codeRepo:        codeRepo,
		jwtManager:      jwtManager,
		blacklist:       blacklist,
		operationLog:    operationLog,
		appEnv:          cfg.App.Env,
		expireSeconds:   expireSeconds,
		cooldownSeconds: cooldownSeconds,
	}
}

func (s *PhoneAuthService) SendCode(ctx context.Context, input SendCodeInput) (*vo.SendCodeResponse, *apperrors.BusinessError) {
	if !security.ValidPhone(input.Phone) {
		s.writeSystemAction(ctx, actionSendCode, input.IP, input.UserAgent, false, "invalid phone")
		return nil, apperrors.New(apperrors.CodeVerificationPhoneInvalid)
	}
	if input.Scene != SceneRegister && input.Scene != SceneLogin {
		s.writeSystemAction(ctx, actionSendCode, input.IP, input.UserAgent, false, "invalid scene")
		return nil, apperrors.New(apperrors.CodeVerificationSceneInvalid)
	}

	phoneHash := security.PhoneHash(input.Phone)
	user, err := s.userRepo.FindByPhoneHash(ctx, phoneHash)
	if input.Scene == SceneRegister {
		if err == nil && user.Status == string(enums.StatusActive) {
			s.writeSystemAction(ctx, actionSendCode, input.IP, input.UserAgent, false, "phone exists")
			return nil, apperrors.New(apperrors.CodeVerificationPhoneExists)
		}
	} else {
		if err != nil || user.Status != string(enums.StatusActive) {
			s.writeSystemAction(ctx, actionSendCode, input.IP, input.UserAgent, false, "phone missing")
			return nil, apperrors.New(apperrors.CodeVerificationPhoneMissing)
		}
	}

	if _, err := s.codeRepo.FindLatestCreatedAfter(ctx, phoneHash, input.Scene, time.Now().Add(-time.Duration(s.cooldownSeconds)*time.Second)); err == nil {
		s.writeSystemAction(ctx, actionSendCode, input.IP, input.UserAgent, false, "too frequent")
		return nil, apperrors.New(apperrors.CodeVerificationTooFrequent)
	}

	code, err := s.newVerificationCode()
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	codeHash, err := security.HashVerificationCode(input.Phone, input.Scene, code)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}

	record := &usermodel.VerificationCode{
		Phone:      input.Phone,
		PhoneHash:  phoneHash,
		CodeHash:   codeHash,
		Scene:      input.Scene,
		ClientType: input.ClientType,
		Status:     string(enums.StatusPending),
		ExpiredAt:  time.Now().Add(time.Duration(s.expireSeconds) * time.Second),
		IP:         stringPtr(input.IP),
		UserAgent:  stringPtr(input.UserAgent),
	}
	if err := s.codeRepo.Create(ctx, record); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}

	s.writeSystemAction(ctx, actionSendCode, input.IP, input.UserAgent, true, "")
	resp := &vo.SendCodeResponse{
		ExpireSeconds:   s.expireSeconds,
		CooldownSeconds: s.cooldownSeconds,
	}
	if s.appEnv != "prod" {
		resp.DevCode = &code
	}
	return resp, nil
}

func (s *PhoneAuthService) RegisterPhone(ctx context.Context, input RegisterPhoneInput) (*vo.LoginResponse, *apperrors.BusinessError) {
	if !security.ValidPhone(input.Phone) {
		return nil, apperrors.New(apperrors.CodeVerificationPhoneInvalid)
	}
	if len(input.Password) < 6 {
		return nil, apperrors.New(apperrors.CodeRegisterPasswordWeak)
	}

	phoneHash := security.PhoneHash(input.Phone)
	existing, err := s.userRepo.FindByPhoneHash(ctx, phoneHash)
	if err == nil {
		if existing.Status == string(enums.StatusActive) {
			s.writeUserAction(ctx, existing.ID, actionRegister, input.IP, input.UserAgent, false, "phone exists")
			return nil, apperrors.New(apperrors.CodeRegisterPhoneExists)
		}
		if existing.Status == string(enums.StatusPendingClaim) {
			s.writeUserAction(ctx, existing.ID, actionRegister, input.IP, input.UserAgent, false, "pending claim")
			return nil, apperrors.New(apperrors.CodeRegisterClaimFailed)
		}
	}

	if businessErr := s.consumeCode(ctx, input.Phone, phoneHash, SceneRegister, input.Code, true); businessErr != nil {
		s.writeSystemAction(ctx, actionRegister, input.IP, input.UserAgent, false, "invalid verification code")
		return nil, businessErr
	}

	passwordHash, err := security.HashPassword(input.Password)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	user := &usermodel.User{
		Phone:          &input.Phone,
		PhoneHash:      &phoneHash,
		PhoneVerified:  true,
		PasswordHash:   &passwordHash,
		Nickname:       input.Nickname,
		AccountOrigin:  accountOriginPhoneRegister,
		RegisterClient: input.ClientType,
		Status:         string(enums.StatusActive),
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}

	if err := s.userRepo.RecordLoginSuccess(ctx, user.ID, input.IP, input.ClientType, time.Now()); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	s.writeUserAction(ctx, user.ID, actionRegister, input.IP, input.UserAgent, true, "")
	return s.issueUserToken(user)
}

func (s *PhoneAuthService) LoginPhone(ctx context.Context, input LoginPhoneInput) (*vo.LoginResponse, *apperrors.BusinessError) {
	if !security.ValidPhone(input.Phone) {
		return nil, apperrors.New(apperrors.CodeLoginPhonePasswordInvalid)
	}

	user, err := s.userRepo.FindByPhoneHash(ctx, security.PhoneHash(input.Phone))
	if err != nil {
		s.writeSystemAction(ctx, actionLogin, input.IP, input.UserAgent, false, "user not found")
		return nil, apperrors.New(apperrors.CodeLoginPhonePasswordInvalid)
	}
	if businessErr := loginStatusError(user); businessErr != nil {
		s.writeUserAction(ctx, user.ID, actionLogin, input.IP, input.UserAgent, false, "status denied")
		return nil, businessErr
	}
	if user.PasswordHash == nil || !security.CheckPassword(input.Password, *user.PasswordHash) {
		s.writeUserAction(ctx, user.ID, actionLogin, input.IP, input.UserAgent, false, "invalid credentials")
		return nil, apperrors.New(apperrors.CodeLoginPhonePasswordInvalid)
	}

	if err := s.userRepo.RecordLoginSuccess(ctx, user.ID, input.IP, input.ClientType, time.Now()); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	s.writeUserAction(ctx, user.ID, actionLogin, input.IP, input.UserAgent, true, "")
	return s.issueUserToken(user)
}

func (s *PhoneAuthService) Logout(ctx context.Context, input LogoutInput) *apperrors.BusinessError {
	if input.TokenID != "" && s.blacklist != nil {
		if err := s.blacklist.Revoke(ctx, input.TokenID, time.Until(time.Unix(input.ExpiresAt, 0))); err != nil {
			return apperrors.New(apperrors.CodeSystemError)
		}
	}
	s.writeUserAction(ctx, input.UserID, actionLogout, input.IP, input.UserAgent, true, "")
	return nil
}

func (s *PhoneAuthService) consumeCode(ctx context.Context, phone string, phoneHash string, scene string, code string, register bool) *apperrors.BusinessError {
	record, err := s.codeRepo.FindLatestPending(ctx, phoneHash, scene)
	if err != nil {
		if register {
			return apperrors.New(apperrors.CodeRegisterCodeInvalid)
		}
		return apperrors.New(apperrors.CodeVerificationCodeInvalid)
	}
	now := time.Now()
	if record.ExpiredAt.Before(now) {
		if register {
			return apperrors.New(apperrors.CodeRegisterCodeExpired)
		}
		return apperrors.New(apperrors.CodeVerificationCodeExpired)
	}
	if record.UsedAt != nil || record.Status == string(enums.StatusUsed) {
		return apperrors.New(apperrors.CodeVerificationCodeUsed)
	}
	if !security.CheckVerificationCode(phone, scene, code, record.CodeHash) {
		if register {
			return apperrors.New(apperrors.CodeRegisterCodeInvalid)
		}
		return apperrors.New(apperrors.CodeVerificationCodeInvalid)
	}
	if err := s.codeRepo.MarkUsed(ctx, record.ID, now); err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	return nil
}

func (s *PhoneAuthService) issueUserToken(user *usermodel.User) (*vo.LoginResponse, *apperrors.BusinessError) {
	tokenID, err := newTokenID()
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	token, err := s.jwtManager.GenerateAccessToken(user.ID, commonjwt.TokenTypeUser, "", tokenID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	return &vo.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		User:        userInfo(user),
	}, nil
}

func (s *PhoneAuthService) newVerificationCode() (string, error) {
	if s.appEnv != "prod" {
		return "123456", nil
	}
	return security.GenerateNumericCode(6)
}

func loginStatusError(user *usermodel.User) *apperrors.BusinessError {
	switch user.Status {
	case string(enums.StatusActive):
		return nil
	case string(enums.StatusDisabled), string(enums.StatusDeleted):
		return apperrors.New(apperrors.CodeLoginDisabled)
	case string(enums.StatusCancelled):
		return apperrors.New(apperrors.CodeLoginCancelled)
	case string(enums.StatusMerged):
		return apperrors.New(apperrors.CodeLoginMerged)
	case string(enums.StatusPendingClaim), string(enums.StatusPendingBind):
		return apperrors.New(apperrors.CodeLoginPendingClaim)
	default:
		return apperrors.New(apperrors.CodeAccountStatusInvalid)
	}
}

func userInfo(user *usermodel.User) vo.UserInfo {
	return vo.UserInfo{
		ID:            user.ID,
		Phone:         user.Phone,
		PhoneVerified: user.PhoneVerified,
		Nickname:      user.Nickname,
		Status:        user.Status,
	}
}

func (s *PhoneAuthService) writeSystemAction(ctx context.Context, action string, ip string, userAgent string, success bool, reason string) {
	input := operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeSystem),
		Module:       authLogModule,
		Action:       action,
		IP:           stringPtr(ip),
		UserAgent:    stringPtr(userAgent),
	}
	if reason != "" {
		input.ErrorMessage = &reason
	}
	s.write(ctx, input, success)
}

func (s *PhoneAuthService) writeUserAction(ctx context.Context, userID uint64, action string, ip string, userAgent string, success bool, reason string) {
	input := operationlog.WriteInput{
		OperatorType:   string(enums.OperatorTypeUser),
		OperatorUserID: &userID,
		UserID:         &userID,
		Module:         authLogModule,
		Action:         action,
		IP:             stringPtr(ip),
		UserAgent:      stringPtr(userAgent),
	}
	if reason != "" {
		input.ErrorMessage = &reason
	}
	s.write(ctx, input, success)
}

func (s *PhoneAuthService) write(ctx context.Context, input operationlog.WriteInput, success bool) {
	if s.operationLog == nil {
		return
	}
	if success {
		_ = s.operationLog.WriteSuccess(ctx, input)
		return
	}
	_ = s.operationLog.WriteFailed(ctx, input)
}

func newTokenID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes[:]), nil
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
