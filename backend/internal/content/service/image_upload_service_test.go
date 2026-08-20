package service

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	contentmodel "tree/backend/internal/content/model"
	contentrepo "tree/backend/internal/content/repository"
	operationlog "tree/backend/internal/operationlog/service"
)

func TestUploadImageStoresPublicContentPath(t *testing.T) {
	tmpRoot := t.TempDir()
	t.Setenv("FILE_STORAGE_LOCAL_PATH", tmpRoot)
	repo := &contentImageUploadRepoFake{}
	service := &service{
		repo: repo,
		now:  func() time.Time { return time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC) },
	}
	fileHeader := mustContentMultipartFileHeader(t, "guide.png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x01})

	result, businessErr := service.UploadImage(context.Background(), 1, "ROOT_ADMIN", UploadImageInput{File: fileHeader})
	if businessErr != nil {
		t.Fatalf("UploadImage: %v", businessErr)
	}
	if !strings.HasPrefix(result.Path, "/api/static/content/2026/08/") || !strings.HasSuffix(result.Path, ".png") {
		t.Fatalf("unexpected public path: %q", result.Path)
	}
	if !strings.Contains(result.Markdown, result.Path) {
		t.Fatalf("markdown must contain public path: %q", result.Markdown)
	}
	matches, _ := filepath.Glob(filepath.Join(tmpRoot, "content", "2026", "08", "*.png"))
	if len(matches) != 1 {
		t.Fatalf("expected one stored image, got %#v", matches)
	}
	if len(repo.logs) != 1 || repo.logs[0].Action != "UPLOAD_CONTENT_IMAGE" {
		t.Fatalf("expected upload operation log, got %#v", repo.logs)
	}
}

func TestDetectContentImageExt(t *testing.T) {
	cases := []struct {
		name   string
		header []byte
		ext    string
		ok     bool
	}{
		{name: "jpg", header: []byte{0xFF, 0xD8, 0xFF, 0x00}, ext: ".jpg", ok: true},
		{name: "png", header: []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00, 0x00, 0x00}, ext: ".png", ok: true},
		{name: "webp", header: []byte("RIFFxxxxWEBP"), ext: ".webp", ok: true},
		{name: "text", header: []byte("plain text"), ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ext, ok := detectContentImageExt(tc.header)
			if ok != tc.ok || ext != tc.ext {
				t.Fatalf("detectContentImageExt()=(%q,%v), want (%q,%v)", ext, ok, tc.ext, tc.ok)
			}
		})
	}
}

type contentImageUploadRepoFake struct {
	logs []operationlog.WriteInput
}

func (r *contentImageUploadRepoFake) WithTx(*gorm.DB) contentrepo.Repository { return r }
func (r *contentImageUploadRepoFake) CreateCategory(context.Context, *contentmodel.ContentCategory) error {
	return errors.New("not implemented")
}
func (r *contentImageUploadRepoFake) UpdateCategory(context.Context, uint64, map[string]any) error {
	return errors.New("not implemented")
}
func (r *contentImageUploadRepoFake) DeleteCategory(context.Context, uint64, time.Time) error {
	return errors.New("not implemented")
}
func (r *contentImageUploadRepoFake) FindCategoryByID(context.Context, uint64, bool) (*contentmodel.ContentCategory, error) {
	return nil, errors.New("not implemented")
}
func (r *contentImageUploadRepoFake) FindCategoryByKey(context.Context, string, bool) (*contentmodel.ContentCategory, error) {
	return nil, errors.New("not implemented")
}
func (r *contentImageUploadRepoFake) ListCategories(context.Context, bool) ([]contentmodel.ContentCategory, error) {
	return nil, errors.New("not implemented")
}
func (r *contentImageUploadRepoFake) CreateArticle(context.Context, *contentmodel.ContentArticle) error {
	return errors.New("not implemented")
}
func (r *contentImageUploadRepoFake) UpdateArticle(context.Context, uint64, map[string]any) error {
	return errors.New("not implemented")
}
func (r *contentImageUploadRepoFake) DeleteArticle(context.Context, uint64, time.Time) error {
	return errors.New("not implemented")
}
func (r *contentImageUploadRepoFake) FindArticleByID(context.Context, uint64, bool) (*contentrepo.ArticleRow, error) {
	return nil, errors.New("not implemented")
}
func (r *contentImageUploadRepoFake) FindArticleByIDOrSlug(context.Context, string, bool) (*contentrepo.ArticleRow, error) {
	return nil, errors.New("not implemented")
}
func (r *contentImageUploadRepoFake) ListArticles(context.Context, contentrepo.ListArticlesQuery) ([]contentrepo.ArticleRow, int64, error) {
	return nil, 0, errors.New("not implemented")
}
func (r *contentImageUploadRepoFake) WriteLog(_ context.Context, input operationlog.WriteInput) error {
	r.logs = append(r.logs, input)
	return nil
}

func mustContentMultipartFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
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
	form, err := reader.ReadForm(6 << 20)
	if err != nil {
		t.Fatalf("ReadForm: %v", err)
	}
	files := form.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected one file, got %d", len(files))
	}
	return files[0]
}
