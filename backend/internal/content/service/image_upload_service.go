package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	apperrors "tree/backend/internal/common/errors"
	"tree/backend/internal/content/vo"
)

const (
	contentImagePublicPrefix = "/api/static/content"
	maxContentImageBytes     = 5 << 20
)

func ContentImageStorageRoot() string {
	if root := strings.TrimSpace(os.Getenv("FILE_STORAGE_LOCAL_PATH")); root != "" {
		return filepath.Join(root, "content")
	}
	return "./uploads/content"
}

func (s *service) UploadImage(ctx context.Context, adminID uint64, role string, input UploadImageInput) (*vo.ImageUploadResult, *apperrors.BusinessError) {
	if input.File == nil {
		return nil, apperrors.New(apperrors.CodeInvalidParams)
	}
	if input.File.Size <= 0 || input.File.Size > maxContentImageBytes {
		return nil, contentError(CodeContentInvalidInput, "图片大小需在 5MB 以内")
	}

	src, err := input.File.Open()
	if err != nil {
		return nil, apperrors.New(apperrors.CodeUploadFailed)
	}
	defer src.Close()

	header := make([]byte, 512)
	n, err := src.Read(header)
	if err != nil && err != io.EOF {
		return nil, apperrors.New(apperrors.CodeUploadFailed)
	}
	header = header[:n]

	ext, ok := detectContentImageExt(header)
	if !ok {
		return nil, contentError(CodeContentInvalidInput, "仅支持 JPG、PNG、WEBP 图片")
	}

	now := s.now()
	dateDir := fmt.Sprintf("%04d/%02d", now.Year(), int(now.Month()))
	storageDir := filepath.Join(ContentImageStorageRoot(), filepath.FromSlash(dateDir))
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return nil, apperrors.New(apperrors.CodeUploadFailed)
	}

	filename, err := randomContentImageFilename(ext)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeSystemError)
	}
	destPath := filepath.Join(storageDir, filename)
	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeUploadFailed)
	}

	reader := io.MultiReader(bytes.NewReader(header), src)
	written, err := io.Copy(dest, io.LimitReader(reader, maxContentImageBytes+1))
	_ = dest.Close()
	if err != nil || written > maxContentImageBytes {
		_ = os.Remove(destPath)
		return nil, apperrors.New(apperrors.CodeUploadFailed)
	}

	publicPath := fmt.Sprintf("%s/%s/%s", contentImagePublicPrefix, dateDir, filename)
	_ = s.repo.WriteLog(ctx, contentLog(adminID, role, "UPLOAD_CONTENT_IMAGE", "content_image", 0, map[string]any{
		"path": publicPath,
		"size": written,
	}))
	return &vo.ImageUploadResult{
		Path:     publicPath,
		URL:      publicPath,
		Filename: filename,
		Markdown: fmt.Sprintf("![%s](%s)", strings.TrimSuffix(filename, ext), publicPath),
	}, nil
}

func detectContentImageExt(header []byte) (string, bool) {
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

func randomContentImageFilename(ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf) + ext, nil
}
