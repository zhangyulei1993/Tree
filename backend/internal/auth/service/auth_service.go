package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"

	accountmodel "tree/backend/internal/account/model"
	"tree/backend/internal/auth/repository"
	"tree/backend/internal/auth/vo"
	"tree/backend/internal/common/config"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	commonjwt "tree/backend/internal/common/jwt"
	"tree/backend/internal/common/redis"
	"tree/backend/internal/common/security"
	"tree/backend/internal/common/wechat"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"
)

const (
	SceneRegister       = "REGISTER"
	SceneLogin          = "LOGIN"
	SceneBindPhone      = "BIND_PHONE"
	SceneChangePhoneOld = "CHANGE_PHONE_OLD"
	SceneChangePhoneNew = "CHANGE_PHONE_NEW"
	SceneCancelAccount  = "CANCEL_ACCOUNT"

	ClientPCWeb      = "PC_WEB"
	ClientH5Web      = "H5_WEB"
	ClientWechatMini = "WECHAT_MINI_PROGRAM"

	authLogModule       = "USER_AUTH"
	actionSendCode      = "SEND_CODE"
	actionRegister      = "REGISTER_PHONE"
	actionLogin         = "LOGIN_PHONE"
	actionLogout        = "LOGOUT"
	actionWechatLogin   = "WECHAT_MINI_LOGIN"
	actionBindPhone     = "BIND_PHONE"
	actionChangePhone   = "CHANGE_PHONE"
	actionCancelAccount = "CANCEL_ACCOUNT"
	actionMergeAccount  = "ACCOUNT_MERGE"
	actionClaimAccount  = "ACCOUNT_CLAIM"

	accountOriginPhoneRegister = "PHONE_REGISTER"
	accountOriginWechatMini    = "WECHAT_MINI_PROGRAM"
	providerWechatMini         = "WECHAT_MINI"
	claimViaWechatBindPhone    = "WECHAT_MINI_PROGRAM_BIND_PHONE"
	mergeSourceWechatBindPhone = "WECHAT_MINI_PROGRAM_BIND_PHONE"
	errRollback                = "rollback"
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

type WechatMiniLoginInput struct {
	Code       string
	ClientType string
	IP         string
	UserAgent  string
}

type BindPhoneInput struct {
	UserID    uint64
	TokenID   string
	Phone     string
	Code      string
	IP        string
	UserAgent string
}

type ChangePhoneInput struct {
	UserID       uint64
	OldPhoneCode string
	NewPhone     string
	NewPhoneCode string
	IP           string
	UserAgent    string
}

type CancelAccountInput struct {
	UserID       uint64
	TokenID      string
	ExpiresAt    int64
	PhoneCode    string
	CancelReason string
	IP           string
	UserAgent    string
}

type AuthService interface {
	SendCode(ctx context.Context, input SendCodeInput) (*vo.SendCodeResponse, *apperrors.BusinessError)
	RegisterPhone(ctx context.Context, input RegisterPhoneInput) (*vo.LoginResponse, *apperrors.BusinessError)
	LoginPhone(ctx context.Context, input LoginPhoneInput) (*vo.LoginResponse, *apperrors.BusinessError)
	Logout(ctx context.Context, input LogoutInput) *apperrors.BusinessError
	WechatMiniLogin(ctx context.Context, input WechatMiniLoginInput) (*vo.LoginResponse, *apperrors.BusinessError)
	BindPhone(ctx context.Context, input BindPhoneInput) (*vo.LoginResponse, *apperrors.BusinessError)
	ChangePhone(ctx context.Context, input ChangePhoneInput) *apperrors.BusinessError
	CancelAccount(ctx context.Context, input CancelAccountInput) *apperrors.BusinessError
}

type PhoneAuthService struct {
	db              *gorm.DB
	userRepo        repository.UserRepository
	codeRepo        repository.VerificationCodeRepository
	wechatClient    wechat.Client
	jwtManager      *commonjwt.Manager
	blacklist       redis.TokenBlacklist
	operationLog    operationlog.Service
	appEnv          string
	wechatAppID     string
	expireSeconds   int
	cooldownSeconds int
}

func NewPhoneAuthService(
	db *gorm.DB,
	userRepo repository.UserRepository,
	codeRepo repository.VerificationCodeRepository,
	wechatClient wechat.Client,
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
		db:              db,
		userRepo:        userRepo,
		codeRepo:        codeRepo,
		wechatClient:    wechatClient,
		jwtManager:      jwtManager,
		blacklist:       blacklist,
		operationLog:    operationLog,
		appEnv:          cfg.App.Env,
		wechatAppID:     cfg.Wechat.MiniAppID,
		expireSeconds:   expireSeconds,
		cooldownSeconds: cooldownSeconds,
	}
}

func (s *PhoneAuthService) SendCode(ctx context.Context, input SendCodeInput) (*vo.SendCodeResponse, *apperrors.BusinessError) {
	if !security.ValidPhone(input.Phone) {
		s.writeSystemAction(ctx, actionSendCode, input.IP, input.UserAgent, false, "invalid phone")
		return nil, apperrors.New(apperrors.CodeVerificationPhoneInvalid)
	}
	if !allowedCodeScene(input.Scene) {
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
	} else if input.Scene == SceneLogin {
		if err != nil || user.Status != string(enums.StatusActive) {
			s.writeSystemAction(ctx, actionSendCode, input.IP, input.UserAgent, false, "phone missing")
			return nil, apperrors.New(apperrors.CodeVerificationPhoneMissing)
		}
	} else if input.Scene == SceneChangePhoneNew {
		if err == nil && user.Status == string(enums.StatusActive) {
			s.writeSystemAction(ctx, actionSendCode, input.IP, input.UserAgent, false, "new phone exists")
			return nil, apperrors.New(apperrors.CodeChangeNewPhoneExists)
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

func (s *PhoneAuthService) WechatMiniLogin(ctx context.Context, input WechatMiniLoginInput) (*vo.LoginResponse, *apperrors.BusinessError) {
	result, err := s.wechatClient.Code2Session(ctx, input.Code)
	if err != nil {
		s.writeSystemAction(ctx, actionWechatLogin, input.IP, input.UserAgent, false, "wechat code2session failed")
		return nil, apperrors.New(apperrors.CodeWechatOpenIDFetchFailed)
	}

	openIDHash := security.HashPlain(result.OpenID)
	appID := s.wechatAppID
	identity, err := s.userRepo.FindIdentityByOpenIDHash(ctx, providerWechatMini, appID, openIDHash)
	if err == nil {
		user, err := s.userRepo.FindByID(ctx, identity.UserID)
		if err != nil {
			return nil, apperrors.New(apperrors.CodeWechatIdentityExpired)
		}
		if user.Status == string(enums.StatusDisabled) || user.Status == string(enums.StatusDeleted) || user.Status == string(enums.StatusCancelled) {
			s.writeUserAction(ctx, user.ID, actionWechatLogin, input.IP, input.UserAgent, false, "status denied")
			return nil, apperrors.New(apperrors.CodeWechatAccountInvalid)
		}
		now := time.Now()
		_ = s.userRepo.UpdateIdentity(ctx, identity.ID, map[string]any{"last_login_at": now, "updated_at": now})
		_ = s.userRepo.RecordLoginSuccess(ctx, user.ID, input.IP, input.ClientType, now)
		s.writeUserAction(ctx, user.ID, actionWechatLogin, input.IP, input.UserAgent, true, "")
		return s.issueUserToken(user)
	}

	var created *usermodel.User
	var businessErr *apperrors.BusinessError
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txUserRepo := s.userRepo.WithTx(tx)
		now := time.Now()
		user := &usermodel.User{
			AccountOrigin:   accountOriginWechatMini,
			RegisterClient:  input.ClientType,
			Status:          string(enums.StatusPendingBind),
			PhoneVerified:   false,
			LastLoginAt:     &now,
			LastLoginIP:     stringPtr(input.IP),
			LastLoginClient: &input.ClientType,
		}
		if err := txUserRepo.Create(ctx, user); err != nil {
			businessErr = apperrors.New(apperrors.CodeSystemError)
			return errors.New(errRollback)
		}
		openID := result.OpenID
		unionID := result.UnionID
		unionHash := ""
		if unionID != "" {
			unionHash = security.HashPlain(unionID)
		}
		identity := &usermodel.UserAuthIdentity{
			UserID:         user.ID,
			Provider:       providerWechatMini,
			ProviderAppID:  &appID,
			OpenID:         &openID,
			OpenIDHash:     &openIDHash,
			UnionID:        stringPtr(unionID),
			UnionIDHash:    stringPtr(unionHash),
			IdentityStatus: string(enums.StatusActive),
			BoundAt:        &now,
			LastLoginAt:    &now,
		}
		if err := txUserRepo.CreateIdentity(ctx, identity); err != nil {
			businessErr = apperrors.New(apperrors.CodeSystemError)
			return errors.New(errRollback)
		}
		created = user
		return nil
	})
	if err != nil {
		if businessErr != nil {
			return nil, businessErr
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}

	s.writeUserAction(ctx, created.ID, actionWechatLogin, input.IP, input.UserAgent, true, "")
	return s.issueUserToken(created)
}

func (s *PhoneAuthService) BindPhone(ctx context.Context, input BindPhoneInput) (*vo.LoginResponse, *apperrors.BusinessError) {
	if !security.ValidPhone(input.Phone) {
		return nil, apperrors.New(apperrors.CodeVerificationPhoneInvalid)
	}
	current, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeBindPhoneLoginRequired)
	}
	if current.Status != string(enums.StatusPendingBind) && current.Status != string(enums.StatusActive) {
		return nil, apperrors.New(apperrors.CodeBindPhoneStatusDenied)
	}

	phoneHash := security.PhoneHash(input.Phone)
	var resultUser *usermodel.User
	var businessErr *apperrors.BusinessError
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txUserRepo := s.userRepo.WithTx(tx)
		txCodeRepo := s.codeRepo.WithTx(tx)
		if businessErr = s.consumeCodeWithRepo(ctx, txCodeRepo, input.Phone, phoneHash, SceneBindPhone, input.Code, apperrors.CodeBindPhoneCodeInvalid, apperrors.CodeBindPhoneCodeInvalid); businessErr != nil {
			return errors.New(errRollback)
		}
		target, findErr := txUserRepo.FindByPhoneHash(ctx, phoneHash)
		if findErr != nil {
			now := time.Now()
			if err := txUserRepo.UpdateUser(ctx, current.ID, map[string]any{
				"phone":          input.Phone,
				"phone_hash":     phoneHash,
				"phone_verified": true,
				"status":         string(enums.StatusActive),
				"updated_at":     now,
			}); err != nil {
				businessErr = apperrors.New(apperrors.CodeSystemError)
				return errors.New(errRollback)
			}
			_ = txUserRepo.CreatePhoneHistory(ctx, phoneHistory(current.ID, &input.Phone, &phoneHash, "BIND_PHONE", current.ID))
			resultUser, _ = txUserRepo.FindByID(ctx, current.ID)
			return nil
		}
		switch target.Status {
		case string(enums.StatusActive):
			resultUser, businessErr = s.mergeUsers(ctx, txUserRepo, current, target, input.IP, input.UserAgent)
		case string(enums.StatusPendingClaim):
			resultUser, businessErr = s.claimUser(ctx, txUserRepo, current, target, input.Phone, phoneHash, input.IP, input.UserAgent)
		default:
			businessErr = apperrors.New(apperrors.CodeBindPhoneStatusDenied)
		}
		if businessErr != nil {
			return errors.New(errRollback)
		}
		return nil
	})
	if err != nil {
		if businessErr != nil {
			s.writeUserAction(ctx, current.ID, actionBindPhone, input.IP, input.UserAgent, false, businessErr.Message)
			return nil, businessErr
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}

	if resultUser.ID != current.ID {
		_ = s.blacklist.RevokeUserTokens(ctx, current.ID, time.Now().Unix(), 30*24*time.Hour)
	}
	s.writeUserAction(ctx, resultUser.ID, actionBindPhone, input.IP, input.UserAgent, true, "")
	return s.issueUserToken(resultUser)
}

func (s *PhoneAuthService) ChangePhone(ctx context.Context, input ChangePhoneInput) *apperrors.BusinessError {
	if !security.ValidPhone(input.NewPhone) {
		return apperrors.New(apperrors.CodeVerificationPhoneInvalid)
	}
	user, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil || user.Status != string(enums.StatusActive) || user.Phone == nil || user.PhoneHash == nil {
		return apperrors.New(apperrors.CodeAccountStatusInvalid)
	}
	newPhoneHash := security.PhoneHash(input.NewPhone)
	var businessErr *apperrors.BusinessError
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txUserRepo := s.userRepo.WithTx(tx)
		txCodeRepo := s.codeRepo.WithTx(tx)
		if businessErr = s.consumeCodeWithRepo(ctx, txCodeRepo, *user.Phone, *user.PhoneHash, SceneChangePhoneOld, input.OldPhoneCode, apperrors.CodeChangeOldPhoneCodeInvalid, apperrors.CodeChangeOldPhoneCodeInvalid); businessErr != nil {
			return errors.New(errRollback)
		}
		if businessErr = s.consumeCodeWithRepo(ctx, txCodeRepo, input.NewPhone, newPhoneHash, SceneChangePhoneNew, input.NewPhoneCode, apperrors.CodeChangeNewPhoneCodeInvalid, apperrors.CodeChangeNewPhoneCodeInvalid); businessErr != nil {
			return errors.New(errRollback)
		}
		if existing, err := txUserRepo.FindByPhoneHash(ctx, newPhoneHash); err == nil && existing.ID != user.ID && existing.Status == string(enums.StatusActive) {
			businessErr = apperrors.New(apperrors.CodeChangeNewPhoneExists)
			return errors.New(errRollback)
		}
		now := time.Now()
		oldPhone := *user.Phone
		oldHash := *user.PhoneHash
		if err := txUserRepo.UpdateUser(ctx, user.ID, map[string]any{
			"phone":      input.NewPhone,
			"phone_hash": newPhoneHash,
			"updated_at": now,
		}); err != nil {
			businessErr = apperrors.New(apperrors.CodeSystemError)
			return errors.New(errRollback)
		}
		_ = txUserRepo.CreatePhoneHistory(ctx, phoneHistory(user.ID, &oldPhone, &oldHash, "CHANGE_PHONE_OLD", user.ID))
		_ = txUserRepo.CreatePhoneHistory(ctx, phoneHistory(user.ID, &input.NewPhone, &newPhoneHash, "CHANGE_PHONE_NEW", user.ID))
		return nil
	})
	if err != nil {
		if businessErr != nil {
			s.writeUserAction(ctx, user.ID, actionChangePhone, input.IP, input.UserAgent, false, businessErr.Message)
			return businessErr
		}
		return apperrors.New(apperrors.CodeSystemError)
	}
	s.writeUserAction(ctx, user.ID, actionChangePhone, input.IP, input.UserAgent, true, "")
	return nil
}

func (s *PhoneAuthService) CancelAccount(ctx context.Context, input CancelAccountInput) *apperrors.BusinessError {
	user, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil || user.Status != string(enums.StatusActive) || user.Phone == nil || user.PhoneHash == nil {
		return apperrors.New(apperrors.CodeCancelStatusDenied)
	}
	if businessErr := s.checkCanCancel(ctx, user.ID); businessErr != nil {
		return businessErr
	}

	var businessErr *apperrors.BusinessError
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txUserRepo := s.userRepo.WithTx(tx)
		txCodeRepo := s.codeRepo.WithTx(tx)
		if businessErr = s.consumeCodeWithRepo(ctx, txCodeRepo, *user.Phone, *user.PhoneHash, SceneCancelAccount, input.PhoneCode, apperrors.CodeCancelCodeInvalid, apperrors.CodeCancelCodeInvalid); businessErr != nil {
			return errors.New(errRollback)
		}
		now := time.Now()
		oldPhone := *user.Phone
		oldHash := *user.PhoneHash
		if err := txUserRepo.UpdateUser(ctx, user.ID, map[string]any{
			"status":         string(enums.StatusCancelled),
			"phone":          nil,
			"phone_hash":     nil,
			"phone_verified": false,
			"nickname":       nil,
			"avatar_url":     nil,
			"real_name":      nil,
			"password_hash":  nil,
			"cancelled_at":   now,
			"cancel_reason":  input.CancelReason,
			"updated_at":     now,
		}); err != nil {
			businessErr = apperrors.New(apperrors.CodeSystemError)
			return errors.New(errRollback)
		}
		if err := txUserRepo.MoveIdentities(ctx, user.ID, user.ID); err != nil {
			businessErr = apperrors.New(apperrors.CodeSystemError)
			return errors.New(errRollback)
		}
		_ = txUserRepo.UpdateUser(ctx, user.ID, map[string]any{"updated_at": now})
		_ = txUserRepo.CreatePhoneHistory(ctx, phoneHistory(user.ID, &oldPhone, &oldHash, "CANCEL_ACCOUNT", user.ID))
		return tx.Model(&usermodel.UserAuthIdentity{}).Where("user_id = ? AND deleted_at IS NULL", user.ID).Updates(map[string]any{
			"identity_status": string(enums.StatusCancelled),
			"unbound_at":      now,
			"updated_at":      now,
		}).Error
	})
	if err != nil {
		if businessErr != nil {
			s.writeUserAction(ctx, user.ID, actionCancelAccount, input.IP, input.UserAgent, false, businessErr.Message)
			return businessErr
		}
		return apperrors.New(apperrors.CodeSystemError)
	}
	_ = s.blacklist.Revoke(ctx, input.TokenID, time.Until(time.Unix(input.ExpiresAt, 0)))
	_ = s.blacklist.RevokeUserTokens(ctx, user.ID, time.Now().Unix(), 30*24*time.Hour)
	s.writeUserAction(ctx, user.ID, actionCancelAccount, input.IP, input.UserAgent, true, "")
	return nil
}

func (s *PhoneAuthService) consumeCode(ctx context.Context, phone string, phoneHash string, scene string, code string, register bool) *apperrors.BusinessError {
	invalidCode := apperrors.CodeVerificationCodeInvalid
	expiredCode := apperrors.CodeVerificationCodeExpired
	if register {
		invalidCode = apperrors.CodeRegisterCodeInvalid
		expiredCode = apperrors.CodeRegisterCodeExpired
	}
	return s.consumeCodeWithRepo(ctx, s.codeRepo, phone, phoneHash, scene, code, invalidCode, expiredCode)
}

func (s *PhoneAuthService) consumeCodeWithRepo(ctx context.Context, codeRepo repository.VerificationCodeRepository, phone string, phoneHash string, scene string, code string, invalidCode apperrors.Code, expiredCode apperrors.Code) *apperrors.BusinessError {
	record, err := codeRepo.FindLatestPending(ctx, phoneHash, scene)
	if err != nil {
		return apperrors.New(invalidCode)
	}
	now := time.Now()
	if record.ExpiredAt.Before(now) {
		return apperrors.New(expiredCode)
	}
	if record.UsedAt != nil || record.Status == string(enums.StatusUsed) {
		return apperrors.New(apperrors.CodeVerificationCodeUsed)
	}
	if !security.CheckVerificationCode(phone, scene, code, record.CodeHash) {
		return apperrors.New(invalidCode)
	}
	if err := codeRepo.MarkUsed(ctx, record.ID, now); err != nil {
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

func (s *PhoneAuthService) mergeUsers(ctx context.Context, repo repository.UserRepository, source *usermodel.User, target *usermodel.User, ip string, userAgent string) (*usermodel.User, *apperrors.BusinessError) {
	conflict, err := repo.HasMemberBindingConflict(ctx, source.ID, target.ID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if conflict {
		return nil, apperrors.New(apperrors.CodeAccountMergeMemberConflict)
	}

	now := time.Now()
	reason := "wechat mini bind phone matched active phone user"
	if err := repo.MoveFamilyLinks(ctx, source.ID, target.ID); err != nil {
		return nil, apperrors.New(apperrors.CodeAccountMergeFailed)
	}
	if err := repo.MoveIdentities(ctx, source.ID, target.ID); err != nil {
		return nil, apperrors.New(apperrors.CodeAccountMergeFailed)
	}
	if err := repo.UpdateUser(ctx, source.ID, map[string]any{
		"status":            string(enums.StatusMerged),
		"merged_to_user_id": target.ID,
		"merged_at":         now,
		"phone":             nil,
		"phone_hash":        nil,
		"phone_verified":    false,
		"updated_at":        now,
	}); err != nil {
		return nil, apperrors.New(apperrors.CodeAccountMergeFailed)
	}
	operatorType := string(enums.OperatorTypeUser)
	if err := repo.CreateMergeLog(ctx, &accountmodel.UserAccountMergeLog{
		SourceUserID:   source.ID,
		TargetUserID:   target.ID,
		MergeReason:    &reason,
		MergeSource:    mergeSourceWechatBindPhone,
		OperatorType:   &operatorType,
		OperatorUserID: &source.ID,
	}); err != nil {
		return nil, apperrors.New(apperrors.CodeAccountMergeFailed)
	}
	s.writeUserAction(ctx, target.ID, actionMergeAccount, ip, userAgent, true, "")
	updated, err := repo.FindByID(ctx, target.ID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeAccountMergeFailed)
	}
	return updated, nil
}

func (s *PhoneAuthService) claimUser(ctx context.Context, repo repository.UserRepository, temp *usermodel.User, target *usermodel.User, phone string, phoneHash string, ip string, userAgent string) (*usermodel.User, *apperrors.BusinessError) {
	now := time.Now()
	if err := repo.MoveIdentities(ctx, temp.ID, target.ID); err != nil {
		return nil, apperrors.New(apperrors.CodeAccountClaimFailed)
	}
	if err := repo.UpdateUser(ctx, target.ID, map[string]any{
		"status":         string(enums.StatusActive),
		"phone":          phone,
		"phone_hash":     phoneHash,
		"phone_verified": true,
		"claimed_at":     now,
		"claimed_via":    claimViaWechatBindPhone,
		"updated_at":     now,
	}); err != nil {
		return nil, apperrors.New(apperrors.CodeAccountClaimFailed)
	}
	if err := repo.UpdateUser(ctx, temp.ID, map[string]any{
		"status":            string(enums.StatusMerged),
		"merged_to_user_id": target.ID,
		"merged_at":         now,
		"phone":             nil,
		"phone_hash":        nil,
		"phone_verified":    false,
		"updated_at":        now,
	}); err != nil {
		return nil, apperrors.New(apperrors.CodeAccountClaimFailed)
	}
	if err := repo.CreateClaimLog(ctx, &accountmodel.UserAccountClaimLog{
		ClaimedUserID: target.ID,
		TempUserID:    &temp.ID,
		Phone:         &phone,
		PhoneHash:     &phoneHash,
		ClaimVia:      claimViaWechatBindPhone,
	}); err != nil {
		return nil, apperrors.New(apperrors.CodeAccountClaimFailed)
	}
	_ = repo.CreatePhoneHistory(ctx, phoneHistory(target.ID, &phone, &phoneHash, "ACCOUNT_CLAIM", temp.ID))
	s.writeUserAction(ctx, target.ID, actionClaimAccount, ip, userAgent, true, "")
	updated, err := repo.FindByID(ctx, target.ID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeAccountClaimFailed)
	}
	return updated, nil
}

func (s *PhoneAuthService) checkCanCancel(ctx context.Context, userID uint64) *apperrors.BusinessError {
	if count, err := s.userRepo.CountFounderLinks(ctx, userID); err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	} else if count > 0 {
		return apperrors.New(apperrors.CodeCancelFounder)
	}
	if count, err := s.userRepo.CountActiveFamilyLinks(ctx, userID); err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	} else if count > 0 {
		return apperrors.New(apperrors.CodeCancelActiveFamily)
	}
	if count, err := s.userRepo.CountPendingFounderTransfer(ctx, userID); err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	} else if count > 0 {
		return apperrors.New(apperrors.CodeCancelPendingTransfer)
	}
	if count, err := s.userRepo.CountPendingDissolution(ctx, userID); err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	} else if count > 0 {
		return apperrors.New(apperrors.CodeCancelPendingDissolution)
	}
	return nil
}

func allowedCodeScene(scene string) bool {
	switch scene {
	case SceneRegister, SceneLogin, SceneBindPhone, SceneChangePhoneOld, SceneChangePhoneNew, SceneCancelAccount:
		return true
	default:
		return false
	}
}

func phoneHistory(userID uint64, phone *string, phoneHash *string, action string, operatorUserID uint64) *accountmodel.UserPhoneHistory {
	operatorType := string(enums.OperatorTypeUser)
	return &accountmodel.UserPhoneHistory{
		UserID:         userID,
		Phone:          phone,
		PhoneHash:      phoneHash,
		ActionType:     action,
		OperatorType:   &operatorType,
		OperatorUserID: &operatorUserID,
	}
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
