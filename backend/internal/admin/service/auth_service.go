package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/admin/model"
	"tree/backend/internal/admin/repository"
	"tree/backend/internal/admin/vo"
	"tree/backend/internal/common/config"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	commonjwt "tree/backend/internal/common/jwt"
	"tree/backend/internal/common/redis"
	"tree/backend/internal/common/security"
	operationlog "tree/backend/internal/operationlog/service"
)

const (
	adminLogModule = "ADMIN_AUTH"
	actionLogin    = "LOGIN"
	actionLogout   = "LOGOUT"
	actionPassword = "CHANGE_PASSWORD"
)

type LoginInput struct {
	Username  string
	Password  string
	IP        string
	UserAgent string
}

type LogoutInput struct {
	AdminID   uint64
	Role      string
	TokenID   string
	ExpiresAt int64
	IP        string
	UserAgent string
}

type ChangePasswordInput struct {
	AdminID     uint64
	Role        string
	OldPassword string
	NewPassword string
	TokenID     string
	ExpiresAt   int64
	IP          string
	UserAgent   string
}

type AuthService interface {
	Login(ctx context.Context, input LoginInput) (*vo.LoginResponse, *apperrors.BusinessError)
	Logout(ctx context.Context, input LogoutInput) *apperrors.BusinessError
	Me(ctx context.Context, adminID uint64) (*vo.AdminInfo, *apperrors.BusinessError)
	ChangePassword(ctx context.Context, input ChangePasswordInput) *apperrors.BusinessError
}

type AdminAuthService struct {
	repo            repository.AdminUserRepository
	jwtManager      *commonjwt.Manager
	blacklist       redis.TokenBlacklist
	operationLog    operationlog.Service
	lockMaxFailures int
	lockDuration    time.Duration
}

func NewAdminAuthService(
	repo repository.AdminUserRepository,
	jwtManager *commonjwt.Manager,
	blacklist redis.TokenBlacklist,
	operationLog operationlog.Service,
	adminSecurity config.AdminSecurityConfig,
) *AdminAuthService {
	lockMaxFailures := adminSecurity.LockMaxFailures
	if lockMaxFailures <= 0 {
		lockMaxFailures = 5
	}
	lockDuration := time.Duration(adminSecurity.LockMinutes) * time.Minute
	if lockDuration <= 0 {
		lockDuration = 30 * time.Minute
	}

	return &AdminAuthService{
		repo:            repo,
		jwtManager:      jwtManager,
		blacklist:       blacklist,
		operationLog:    operationLog,
		lockMaxFailures: lockMaxFailures,
		lockDuration:    lockDuration,
	}
}

func (s *AdminAuthService) Login(ctx context.Context, input LoginInput) (*vo.LoginResponse, *apperrors.BusinessError) {
	admin, err := s.repo.FindByUsername(ctx, input.Username)
	if err != nil {
		s.writeLoginFailed(ctx, nil, input, "admin not found")
		return nil, apperrors.New(apperrors.CodeAdminUsernameOrPasswordInvalid)
	}

	if admin.DeletedAt != nil || admin.Status == string(enums.StatusDeleted) {
		s.writeLoginFailed(ctx, admin, input, "admin deleted")
		return nil, apperrors.New(apperrors.CodeAdminDeleted)
	}
	if admin.Status == string(enums.StatusDisabled) {
		s.writeLoginFailed(ctx, admin, input, "admin disabled")
		return nil, apperrors.New(apperrors.CodeAdminDisabled)
	}
	if admin.LockedUntil != nil && admin.LockedUntil.After(time.Now()) {
		s.writeLoginFailed(ctx, admin, input, "admin locked")
		return nil, apperrors.New(apperrors.CodeAdminLocked)
	}

	if !security.CheckPassword(input.Password, admin.PasswordHash) {
		_ = s.recordFailedPassword(ctx, admin)
		s.writeLoginFailed(ctx, admin, input, "invalid credentials")
		return nil, apperrors.New(apperrors.CodeAdminUsernameOrPasswordInvalid)
	}

	now := time.Now()
	if err := s.repo.RecordLoginSuccess(ctx, admin.ID, input.IP, now); err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}

	tokenID, err := newTokenID()
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	token, err := s.jwtManager.GenerateAccessToken(admin.ID, commonjwt.TokenTypeAdmin, admin.Role, tokenID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}

	s.writeLoginSuccess(ctx, admin, input)
	return &vo.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		Admin:       adminInfo(admin),
	}, nil
}

func (s *AdminAuthService) Logout(ctx context.Context, input LogoutInput) *apperrors.BusinessError {
	if input.TokenID != "" && s.blacklist != nil {
		ttl := time.Until(time.Unix(input.ExpiresAt, 0))
		if err := s.blacklist.Revoke(ctx, input.TokenID, ttl); err != nil {
			return apperrors.New(apperrors.CodeSystemError)
		}
	}

	s.writeAdminActionSuccess(ctx, input.AdminID, input.Role, actionLogout, input.IP, input.UserAgent)
	return nil
}

func (s *AdminAuthService) Me(ctx context.Context, adminID uint64) (*vo.AdminInfo, *apperrors.BusinessError) {
	admin, err := s.repo.FindByID(ctx, adminID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeAdminDeleted)
		}
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	if businessErr := checkAdminUsable(admin); businessErr != nil {
		return nil, businessErr
	}
	info := adminInfo(admin)
	return &info, nil
}

func (s *AdminAuthService) ChangePassword(ctx context.Context, input ChangePasswordInput) *apperrors.BusinessError {
	admin, err := s.repo.FindByID(ctx, input.AdminID)
	if err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	if businessErr := checkAdminUsable(admin); businessErr != nil {
		return businessErr
	}
	if !security.CheckPassword(input.OldPassword, admin.PasswordHash) {
		s.writeAdminActionFailed(ctx, input.AdminID, input.Role, actionPassword, input.IP, input.UserAgent, "old password invalid")
		return apperrors.New(apperrors.CodeAdminUsernameOrPasswordInvalid)
	}

	hash, err := security.HashPassword(input.NewPassword)
	if err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}
	if err := s.repo.UpdatePassword(ctx, input.AdminID, hash, time.Now()); err != nil {
		return apperrors.New(apperrors.CodeSystemError)
	}

	if input.TokenID != "" && s.blacklist != nil {
		// TODO(M4+): revoke all old admin tokens issued before password_changed_at.
		_ = s.blacklist.Revoke(ctx, input.TokenID, time.Until(time.Unix(input.ExpiresAt, 0)))
	}
	s.writeAdminActionSuccess(ctx, input.AdminID, input.Role, actionPassword, input.IP, input.UserAgent)
	return nil
}

func (s *AdminAuthService) recordFailedPassword(ctx context.Context, admin *model.AdminUser) error {
	if admin.Role == string(enums.AdminRoleRootAdmin) {
		return s.repo.RecordLoginFailure(ctx, admin.ID, admin.FailedLoginCount+1, nil)
	}

	failedCount := admin.FailedLoginCount + 1
	var lockedUntil *time.Time
	if failedCount >= s.lockMaxFailures {
		until := time.Now().Add(s.lockDuration)
		lockedUntil = &until
	}
	return s.repo.RecordLoginFailure(ctx, admin.ID, failedCount, lockedUntil)
}

func checkAdminUsable(admin *model.AdminUser) *apperrors.BusinessError {
	if admin.DeletedAt != nil || admin.Status == string(enums.StatusDeleted) {
		return apperrors.New(apperrors.CodeAdminDeleted)
	}
	if admin.Status == string(enums.StatusDisabled) {
		return apperrors.New(apperrors.CodeAdminDisabled)
	}
	if admin.LockedUntil != nil && admin.LockedUntil.After(time.Now()) {
		return apperrors.New(apperrors.CodeAdminLocked)
	}
	return nil
}

func adminInfo(admin *model.AdminUser) vo.AdminInfo {
	return vo.AdminInfo{
		ID:          admin.ID,
		Username:    admin.Username,
		DisplayName: admin.DisplayName,
		Phone:       admin.Phone,
		Email:       admin.Email,
		Role:        admin.Role,
		Status:      admin.Status,
	}
}

func (s *AdminAuthService) writeLoginSuccess(ctx context.Context, admin *model.AdminUser, input LoginInput) {
	s.write(ctx, operationlog.WriteInput{
		OperatorType:    string(enums.OperatorTypeAdmin),
		OperatorAdminID: &admin.ID,
		OperatorRole:    &admin.Role,
		Module:          adminLogModule,
		Action:          actionLogin,
		IP:              stringPtr(input.IP),
		UserAgent:       stringPtr(input.UserAgent),
	}, true)
}

func (s *AdminAuthService) writeLoginFailed(ctx context.Context, admin *model.AdminUser, input LoginInput, reason string) {
	logInput := operationlog.WriteInput{
		OperatorType: string(enums.OperatorTypeAdmin),
		Module:       adminLogModule,
		Action:       actionLogin,
		ErrorMessage: stringPtr(reason),
		IP:           stringPtr(input.IP),
		UserAgent:    stringPtr(input.UserAgent),
	}
	if admin != nil {
		logInput.OperatorAdminID = &admin.ID
		logInput.OperatorRole = &admin.Role
	}
	s.write(ctx, logInput, false)
}

func (s *AdminAuthService) writeAdminActionSuccess(ctx context.Context, adminID uint64, role string, action string, ip string, userAgent string) {
	s.write(ctx, operationlog.WriteInput{
		OperatorType:    string(enums.OperatorTypeAdmin),
		OperatorAdminID: &adminID,
		OperatorRole:    &role,
		Module:          adminLogModule,
		Action:          action,
		IP:              stringPtr(ip),
		UserAgent:       stringPtr(userAgent),
	}, true)
}

func (s *AdminAuthService) writeAdminActionFailed(ctx context.Context, adminID uint64, role string, action string, ip string, userAgent string, reason string) {
	s.write(ctx, operationlog.WriteInput{
		OperatorType:    string(enums.OperatorTypeAdmin),
		OperatorAdminID: &adminID,
		OperatorRole:    &role,
		Module:          adminLogModule,
		Action:          action,
		ErrorMessage:    stringPtr(reason),
		IP:              stringPtr(ip),
		UserAgent:       stringPtr(userAgent),
	}, false)
}

func (s *AdminAuthService) write(ctx context.Context, input operationlog.WriteInput, success bool) {
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
