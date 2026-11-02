package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tree/backend/internal/auth/vo"
	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	usermodel "tree/backend/internal/user/model"
)

const (
	avatarPublicPrefix = "/api/static/avatars"
	maxAvatarBytes     = 2 << 20
)

type UpdateProfileInput struct {
	UserID    uint64
	Nickname  *string
	IP        string
	UserAgent string
}

type UploadAvatarInput struct {
	UserID    uint64
	File      *multipart.FileHeader
	IP        string
	UserAgent string
}

func AvatarStorageRoot() string {
	if root := strings.TrimSpace(os.Getenv("FILE_STORAGE_LOCAL_PATH")); root != "" {
		return filepath.Join(root, "avatars")
	}
	return "./uploads/avatars"
}

func (s *PhoneAuthService) GetMe(ctx context.Context, userID uint64) (*vo.UserInfo, *apperrors.BusinessError) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeUnauthorized)
	}
	if businessErr := profileReadableStatus(user); businessErr != nil {
		return nil, businessErr
	}
	info := userInfo(user)
	return &info, nil
}

func (s *PhoneAuthService) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*vo.UserInfo, *apperrors.BusinessError) {
	user, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeUnauthorized)
	}
	if businessErr := profileWritableStatus(user); businessErr != nil {
		s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "status denied")
		return nil, businessErr
	}

	nickname, businessErr := normalizeNickname(input.Nickname)
	if businessErr != nil {
		s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "invalid nickname")
		return nil, businessErr
	}
	if input.Nickname != nil && nickname != nil {
		if businessErr := s.contentSafety.CheckTexts(ctx, contentsafety.CheckInput{
			UserID: input.UserID,
			Scene:  contentsafety.SceneProfile,
			Fields: contentsafety.StringField("nickname", *nickname),
			IP:     input.IP, UserAgent: input.UserAgent,
		}); businessErr != nil {
			s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "content safety rejected")
			return nil, businessErr
		}
	}

	updates := map[string]any{"updated_at": time.Now()}
	if input.Nickname != nil {
		updates["nickname"] = nickname
	}
	if err := s.userRepo.UpdateUser(ctx, input.UserID, updates); err != nil {
		s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "update failed")
		return nil, apperrors.New(apperrors.CodeSystemError)
	}

	updated, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, true, "")
	info := userInfo(updated)
	return &info, nil
}

func (s *PhoneAuthService) UploadAvatar(ctx context.Context, input UploadAvatarInput) (*vo.UserInfo, *apperrors.BusinessError) {
	user, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeUnauthorized)
	}
	if businessErr := profileWritableStatus(user); businessErr != nil {
		s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "status denied")
		return nil, businessErr
	}
	if input.File == nil {
		s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "file missing")
		return nil, apperrors.New(apperrors.CodeInvalidParams)
	}
	if input.File.Size <= 0 || input.File.Size > maxAvatarBytes {
		s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "file size invalid")
		return nil, apperrors.New(apperrors.CodeUploadFailed)
	}

	src, err := input.File.Open()
	if err != nil {
		s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "open file failed")
		return nil, apperrors.New(apperrors.CodeUploadFailed)
	}
	defer src.Close()

	header := make([]byte, 512)
	n, err := src.Read(header)
	if err != nil && err != io.EOF {
		s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "read file failed")
		return nil, apperrors.New(apperrors.CodeUploadFailed)
	}
	header = header[:n]

	ext, ok := detectAvatarExt(header)
	if !ok {
		s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "file type invalid")
		return nil, apperrors.New(apperrors.CodeUploadFailed)
	}

	userDir := filepath.Join(s.avatarStorageRoot, fmt.Sprintf("%d", input.UserID))
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "mkdir failed")
		return nil, apperrors.New(apperrors.CodeUploadFailed)
	}

	filename, err := randomAvatarFilename(ext)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	destPath := filepath.Join(userDir, filename)

	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "create file failed")
		return nil, apperrors.New(apperrors.CodeUploadFailed)
	}

	reader := io.MultiReader(bytes.NewReader(header), src)
	written, err := io.Copy(dest, io.LimitReader(reader, maxAvatarBytes+1))
	_ = dest.Close()
	if err != nil || written > maxAvatarBytes {
		_ = os.Remove(destPath)
		s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "write file failed")
		return nil, apperrors.New(apperrors.CodeUploadFailed)
	}

	publicURL := fmt.Sprintf("%s/%d/%s", avatarPublicPrefix, input.UserID, filename)
	if err := s.userRepo.UpdateUser(ctx, input.UserID, map[string]any{
		"avatar_url": publicURL,
		"updated_at": time.Now(),
	}); err != nil {
		_ = os.Remove(destPath)
		s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, false, "update user failed")
		return nil, apperrors.New(apperrors.CodeSystemError)
	}

	removeStoredAvatarFile(s.avatarStorageRoot, user.AvatarURL)

	updated, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	s.writeUserAction(ctx, input.UserID, actionUpdateProfile, input.IP, input.UserAgent, true, "")
	info := userInfo(updated)
	return &info, nil
}

func profileReadableStatus(user *usermodel.User) *apperrors.BusinessError {
	switch user.Status {
	case string(enums.StatusActive), string(enums.StatusPendingBind):
		return nil
	case string(enums.StatusDisabled), string(enums.StatusDeleted):
		return apperrors.New(apperrors.CodeLoginDisabled)
	case string(enums.StatusCancelled):
		return apperrors.New(apperrors.CodeLoginCancelled)
	default:
		return apperrors.New(apperrors.CodeAccountStatusInvalid)
	}
}

func profileWritableStatus(user *usermodel.User) *apperrors.BusinessError {
	return profileReadableStatus(user)
}

func normalizeNickname(raw *string) (*string, *apperrors.BusinessError) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	runes := []rune(trimmed)
	if len(runes) < 1 || len(runes) > 30 {
		return nil, apperrors.New(apperrors.CodeProfileNicknameInvalid)
	}
	for _, ch := range runes {
		if ch < 32 || ch == 127 {
			return nil, apperrors.New(apperrors.CodeProfileNicknameInvalid)
		}
	}
	return &trimmed, nil
}

func detectAvatarExt(header []byte) (string, bool) {
	if len(header) >= 3 && header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		return ".jpg", true
	}
	if len(header) >= 8 && header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47 {
		return ".png", true
	}
	if len(header) >= 12 && string(header[0:4]) == "RIFF" && string(header[8:12]) == "WEBP" {
		return ".webp", true
	}
	return "", false
}

func randomAvatarFilename(ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf) + ext, nil
}

func removeStoredAvatarFile(storageRoot string, avatarURL *string) {
	if avatarURL == nil || *avatarURL == "" {
		return
	}
	prefix := avatarPublicPrefix + "/"
	if !strings.HasPrefix(*avatarURL, prefix) {
		return
	}
	rel := strings.TrimPrefix(*avatarURL, prefix)
	full := filepath.Join(storageRoot, filepath.FromSlash(rel))
	_ = os.Remove(full)
}
