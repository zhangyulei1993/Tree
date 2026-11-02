package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/auth/vo"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/common/security"
	operationlog "tree/backend/internal/operationlog/service"
	usermodel "tree/backend/internal/user/model"
)

const (
	actionBindPhoneCredential      = "BIND_PHONE_CREDENTIAL"
	actionChangePhoneLoginPassword = "CHANGE_PHONE_LOGIN_PASSWORD"
	actionAdminUnbindPhoneLogin    = "ADMIN_UNBIND_PHONE_LOGIN"
)

type BindPhoneCredentialInput struct {
	UserID          uint64
	TokenID         string
	ExpiresAt       int64
	Phone           string
	Password        string
	ConfirmPassword string
	IP              string
	UserAgent       string
}

type ChangePhoneLoginPasswordInput struct {
	UserID          uint64
	TokenID         string
	ExpiresAt       int64
	CurrentPassword string
	NewPassword     string
	IP              string
	UserAgent       string
}

type AdminUnbindPhoneLoginInput struct {
	AdminID      uint64
	AdminRole    string
	TargetUserID uint64
	Confirm      bool
	Reason       string
	IP           string
	UserAgent    string
}

func nicknameComplete(user *usermodel.User) bool {
	return user != nil && user.Nickname != nil && strings.TrimSpace(*user.Nickname) != ""
}

func phoneCredentialConflictStatus(status string) bool {
	switch status {
	case string(enums.StatusActive), string(enums.StatusPendingClaim), string(enums.StatusPendingBind):
		return true
	default:
		return false
	}
}

func (s *PhoneAuthService) BindPhoneCredential(ctx context.Context, input BindPhoneCredentialInput) (*vo.LoginResponse, *apperrors.BusinessError) {
	if !security.ValidPhone(input.Phone) {
		return nil, apperrors.New(apperrors.CodeVerificationPhoneInvalid)
	}
	if input.Password == "" || input.ConfirmPassword == "" || input.Password != input.ConfirmPassword {
		return nil, apperrors.New(apperrors.CodeRegisterPasswordWeak)
	}
	if len(input.Password) < 6 {
		return nil, apperrors.New(apperrors.CodeRegisterPasswordWeak)
	}

	passwordHash, err := security.HashPassword(input.Password)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}

	phoneHash := security.PhoneHash(input.Phone)
	var resultUser *usermodel.User
	var businessErr *apperrors.BusinessError

	txErr := s.withTransaction(ctx, func(tx *gorm.DB) error {
		txUserRepo := s.userRepo.WithTx(tx)
		current, lockErr := txUserRepo.LockByID(ctx, input.UserID)
		if lockErr != nil {
			businessErr = apperrors.New(apperrors.CodeBindPhoneLoginRequired)
			return errors.New(errRollback)
		}
		if current.Status != string(enums.StatusActive) {
			businessErr = apperrors.New(apperrors.CodeBindPhoneStatusDenied)
			return errors.New(errRollback)
		}
		if !nicknameComplete(current) {
			businessErr = apperrors.New(apperrors.CodeBindPhoneCredentialProfileIncomplete)
			return errors.New(errRollback)
		}
		if current.PhoneLoginEnabled {
			businessErr = apperrors.New(apperrors.CodeBindPhoneCredentialAlreadyEnabled)
			return errors.New(errRollback)
		}

		if existing, findErr := txUserRepo.FindByPhoneHash(ctx, phoneHash); findErr == nil {
			if existing.ID != current.ID && phoneCredentialConflictStatus(existing.Status) {
				businessErr = apperrors.New(apperrors.CodeBindPhoneExists)
				return errors.New(errRollback)
			}
		}

		now := time.Now()
		updates := map[string]any{
			"phone":               input.Phone,
			"phone_hash":          phoneHash,
			"password_hash":       passwordHash,
			"phone_login_enabled": true,
			"updated_at":          now,
		}
		if err := txUserRepo.UpdateUser(ctx, current.ID, updates); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				businessErr = apperrors.New(apperrors.CodeBindPhoneExists)
				return errors.New(errRollback)
			}
			businessErr = apperrors.New(apperrors.CodeSystemError)
			return errors.New(errRollback)
		}
		_ = txUserRepo.CreatePhoneHistory(ctx, phoneHistory(current.ID, &input.Phone, &phoneHash, actionBindPhoneCredential, current.ID))
		var findErr error
		resultUser, findErr = txUserRepo.FindByID(ctx, current.ID)
		if findErr != nil {
			businessErr = apperrors.New(apperrors.CodeSystemError)
			return errors.New(errRollback)
		}
		return nil
	})
	if txErr != nil {
		if businessErr != nil {
			s.writeUserAction(ctx, input.UserID, actionBindPhoneCredential, input.IP, input.UserAgent, false, businessErr.Message)
			return nil, businessErr
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}

	if input.TokenID != "" && s.blacklist != nil {
		_ = s.blacklist.Revoke(ctx, input.TokenID, time.Until(time.Unix(input.ExpiresAt, 0)))
	}
	s.writeUserAction(ctx, resultUser.ID, actionBindPhoneCredential, input.IP, input.UserAgent, true, "")
	return s.issueUserToken(resultUser)
}

func (s *PhoneAuthService) ChangePhoneLoginPassword(ctx context.Context, input ChangePhoneLoginPasswordInput) (*vo.LoginResponse, *apperrors.BusinessError) {
	if len(input.NewPassword) < 6 {
		return nil, apperrors.New(apperrors.CodeRegisterPasswordWeak)
	}

	var resultUser *usermodel.User
	var businessErr *apperrors.BusinessError
	txErr := s.withTransaction(ctx, func(tx *gorm.DB) error {
		txUserRepo := s.userRepo.WithTx(tx)
		user, lockErr := txUserRepo.LockByID(ctx, input.UserID)
		if lockErr != nil {
			businessErr = apperrors.New(apperrors.CodeChangePhoneLoginPasswordInvalid)
			return errors.New(errRollback)
		}
		if !user.PhoneLoginEnabled || user.PasswordHash == nil {
			businessErr = apperrors.New(apperrors.CodeChangePhoneLoginPasswordInvalid)
			return errors.New(errRollback)
		}
		if !security.CheckPassword(input.CurrentPassword, *user.PasswordHash) {
			businessErr = apperrors.New(apperrors.CodeChangePhoneLoginPasswordInvalid)
			return errors.New(errRollback)
		}
		newHash, hashErr := security.HashPassword(input.NewPassword)
		if hashErr != nil {
			businessErr = apperrors.New(apperrors.CodeSystemError)
			return errors.New(errRollback)
		}
		if err := txUserRepo.UpdateUser(ctx, user.ID, map[string]any{
			"password_hash": newHash,
			"updated_at":    time.Now(),
		}); err != nil {
			businessErr = apperrors.New(apperrors.CodeSystemError)
			return errors.New(errRollback)
		}
		var findErr error
		resultUser, findErr = txUserRepo.FindByID(ctx, user.ID)
		if findErr != nil {
			businessErr = apperrors.New(apperrors.CodeSystemError)
			return errors.New(errRollback)
		}
		return nil
	})
	if txErr != nil {
		if businessErr != nil {
			s.writeUserAction(ctx, input.UserID, actionChangePhoneLoginPassword, input.IP, input.UserAgent, false, businessErr.Message)
			return nil, businessErr
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	revokedAt := time.Now().Unix()
	if s.blacklist != nil {
		_ = s.blacklist.RevokeUserTokens(ctx, resultUser.ID, revokedAt, 30*24*time.Hour)
	}
	s.writeUserAction(ctx, resultUser.ID, actionChangePhoneLoginPassword, input.IP, input.UserAgent, true, "")
	return s.issueUserTokenWithIssuedAt(resultUser, revokedAt+1)
}

func (s *PhoneAuthService) AdminUnbindPhoneLogin(ctx context.Context, input AdminUnbindPhoneLoginInput) *apperrors.BusinessError {
	if input.AdminRole != string(enums.AdminRoleRootAdmin) && input.AdminRole != string(enums.AdminRoleSuperAdmin) {
		return apperrors.New(apperrors.CodeForbidden)
	}
	if !input.Confirm {
		return apperrors.New(apperrors.CodeInvalidParams)
	}

	user, err := s.userRepo.FindByID(ctx, input.TargetUserID)
	if err != nil {
		return apperrors.New(apperrors.CodeResourceNotFound)
	}
	if !user.PhoneLoginEnabled {
		return nil
	}
	hasWechat, err := s.userRepo.HasActiveWechatMiniIdentity(ctx, user.ID, s.wechatAppID)
	if err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	if !hasWechat {
		return apperrors.New(apperrors.CodeAdminUnbindPhoneLoginNoWechat)
	}

	now := time.Now()
	if err := s.userRepo.UpdateUser(ctx, user.ID, map[string]any{
		"phone":               nil,
		"phone_hash":          nil,
		"password_hash":       nil,
		"phone_login_enabled": false,
		"updated_at":          now,
	}); err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	operatorType := string(enums.OperatorTypeAdmin)
	_ = s.userRepo.CreatePhoneHistory(ctx, phoneHistory(user.ID, nil, nil, actionAdminUnbindPhoneLogin, 0))
	if s.blacklist != nil {
		_ = s.blacklist.RevokeUserTokens(ctx, user.ID, now.Unix(), 30*24*time.Hour)
	}

	detail, _ := json.Marshal(map[string]any{
		"targetUserId": user.ID,
		"reason":       strings.TrimSpace(input.Reason),
	})
	targetType := "USER"
	s.write(ctx, operationlog.WriteInput{
		OperatorType:    operatorType,
		OperatorAdminID: &input.AdminID,
		OperatorRole:    &input.AdminRole,
		Module:          "USER",
		Action:          actionAdminUnbindPhoneLogin,
		TargetType:      &targetType,
		TargetID:        &user.ID,
		UserID:          &user.ID,
		DetailJSON:      detail,
		IP:              stringPtr(input.IP),
		UserAgent:       stringPtr(input.UserAgent),
	}, true)
	return nil
}
