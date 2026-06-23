package service

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"path/filepath"
	"strings"
	"testing"

	"tree/backend/internal/common/enums"
	apperrors "tree/backend/internal/common/errors"
	usermodel "tree/backend/internal/user/model"
)

func TestNormalizeNickname(t *testing.T) {
	valid := "  张三  "
	got, businessErr := normalizeNickname(&valid)
	if businessErr != nil || got == nil || *got != "张三" {
		t.Fatalf("unexpected result: %#v %#v", got, businessErr)
	}

	empty := "   "
	got, businessErr = normalizeNickname(&empty)
	if businessErr != nil || got != nil {
		t.Fatalf("expected nil nickname, got %#v %#v", got, businessErr)
	}

	long := strings.Repeat("昵", 31)
	got, businessErr = normalizeNickname(&long)
	if got != nil || businessErr == nil || businessErr.Code != apperrors.CodeProfileNicknameInvalid {
		t.Fatalf("unexpected long nickname result: %#v %#v", got, businessErr)
	}

	control := "a\x01b"
	got, businessErr = normalizeNickname(&control)
	if got != nil || businessErr == nil || businessErr.Code != apperrors.CodeProfileNicknameInvalid {
		t.Fatalf("unexpected control nickname result: %#v %#v", got, businessErr)
	}
}

func TestUpdateProfileSuccess(t *testing.T) {
	users := newBindUserRepoFake(&usermodel.User{
		ID: 501, Status: string(enums.StatusActive), PhoneVerified: true,
	})
	logs := &authLogCapture{}
	service := newBindService(t, users, &codeRepoFake{}, logs)
	nickname := "  树友一号  "

	result, businessErr := service.UpdateProfile(context.Background(), UpdateProfileInput{
		UserID: 501, Nickname: &nickname, IP: "127.0.0.1", UserAgent: "unit-test",
	})
	if businessErr != nil {
		t.Fatalf("UpdateProfile: %v", businessErr)
	}
	if result.Nickname == nil || *result.Nickname != "树友一号" {
		t.Fatalf("unexpected nickname: %#v", result.Nickname)
	}
	if len(logs.success) != 1 || logs.success[0].Action != actionUpdateProfile {
		t.Fatalf("unexpected log: %#v", logs.success)
	}
	serialized, _ := json.Marshal(logs.success)
	if strings.Contains(string(serialized), "树友一号") {
		t.Fatal("operation log should not contain nickname plaintext")
	}
}

func TestUploadAvatarStoresPublicURL(t *testing.T) {
	tmpRoot := t.TempDir()
	users := newBindUserRepoFake(&usermodel.User{
		ID: 502, Status: string(enums.StatusActive), PhoneVerified: true,
	})
	logs := &authLogCapture{}
	service := newBindService(t, users, &codeRepoFake{}, logs)
	service.avatarStorageRoot = tmpRoot

	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x01}
	fileHeader := mustMultipartFileHeader(t, "avatar.png", pngHeader)

	result, businessErr := service.UploadAvatar(context.Background(), UploadAvatarInput{
		UserID:    502,
		File:      fileHeader,
		IP:        "127.0.0.1",
		UserAgent: "unit-test",
	})
	if businessErr != nil {
		t.Fatalf("UploadAvatar: %v", businessErr)
	}
	if result.AvatarURL == nil || !strings.HasPrefix(*result.AvatarURL, avatarPublicPrefix+"/502/") {
		t.Fatalf("unexpected avatar url: %#v", result.AvatarURL)
	}
	matches, _ := filepath.Glob(filepath.Join(tmpRoot, "502", "*.png"))
	if len(matches) != 1 {
		t.Fatalf("expected one stored file, got %#v", matches)
	}
}

func TestDetectAvatarExt(t *testing.T) {
	if ext, ok := detectAvatarExt([]byte{0xFF, 0xD8, 0xFF, 0x00}); !ok || ext != ".jpg" {
		t.Fatalf("jpg not detected: %s %v", ext, ok)
	}
	if _, ok := detectAvatarExt([]byte("plain-text")); ok {
		t.Fatal("expected invalid type")
	}
}

func mustMultipartFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(1 << 20)
	if err != nil {
		t.Fatalf("ReadForm: %v", err)
	}
	files := form.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected one file, got %d", len(files))
	}
	return files[0]
}
