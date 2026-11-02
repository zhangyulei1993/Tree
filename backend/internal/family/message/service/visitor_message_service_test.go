package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"tree/backend/internal/common/enums"
	familymodel "tree/backend/internal/family/core/model"
	messagedto "tree/backend/internal/family/message/dto"
	messageenum "tree/backend/internal/family/message/enum"
	messagemodel "tree/backend/internal/family/message/model"
	messagerepo "tree/backend/internal/family/message/repository"
	publicenum "tree/backend/internal/family/publicdisplay/enum"
	operationlog "tree/backend/internal/operationlog/service"
)

type fakeRepo struct {
	family   familymodel.Family
	messages map[uint64]*messagemodel.VisitorMessage
	logs     []operationlog.WriteInput
	nextID   uint64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		family: familymodel.Family{
			ID: 2, FamilyName: "Tree", Status: string(enums.StatusNormal),
			PublicDisplayStatus: publicenum.PublicApproved, PublicDisplayEnabled: true, GraphVersion: 7,
		},
		messages: map[uint64]*messagemodel.VisitorMessage{},
		nextID:   1,
	}
}

func (r *fakeRepo) WithTx(*gorm.DB) messagerepo.Repository { return r }
func (r *fakeRepo) FindFamily(context.Context, uint64, bool) (*familymodel.Family, error) {
	family := r.family
	return &family, nil
}
func (r *fakeRepo) CountByIP(_ context.Context, familyID uint64, ip string, since time.Time) (int64, error) {
	var count int64
	for _, msg := range r.messages {
		if msg.FamilyID == familyID && msg.IP != nil && *msg.IP == ip && !msg.CreatedAt.Before(since) {
			count++
		}
	}
	return count, nil
}
func (r *fakeRepo) Create(_ context.Context, msg *messagemodel.VisitorMessage) error {
	msg.ID = r.nextID
	r.nextID++
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	msg.UpdatedAt = msg.CreatedAt
	copy := *msg
	r.messages[msg.ID] = &copy
	return nil
}
func (r *fakeRepo) FindByID(_ context.Context, id uint64, _ bool) (*messagemodel.VisitorMessage, error) {
	msg, ok := r.messages[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *msg
	return &copy, nil
}
func (r *fakeRepo) List(_ context.Context, q messagerepo.ListQuery) ([]messagerepo.MessageRow, int64, error) {
	rows := make([]messagerepo.MessageRow, 0)
	for _, msg := range r.messages {
		if q.FamilyID != 0 && msg.FamilyID != q.FamilyID {
			continue
		}
		if q.PublicOnly && (msg.Status != messageenum.StatusApproved || msg.DeletedAt != nil) {
			continue
		}
		if !q.PublicOnly && q.Status != "" && msg.Status != q.Status {
			continue
		}
		rows = append(rows, messagerepo.MessageRow{VisitorMessage: *msg, FamilyName: r.family.FamilyName})
	}
	return rows, int64(len(rows)), nil
}
func (r *fakeRepo) UpdateStatus(_ context.Context, id uint64, values map[string]any) error {
	msg, ok := r.messages[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if value, ok := values["status"].(string); ok {
		msg.Status = value
	}
	if value, ok := values["reviewed_by_admin_id"].(uint64); ok {
		msg.ReviewedByAdminID = &value
	}
	if value, ok := values["reviewed_at"].(time.Time); ok {
		msg.ReviewedAt = &value
	}
	if value, ok := values["review_comment"].(*string); ok {
		msg.ReviewComment = value
	}
	if value, ok := values["deleted_at"].(time.Time); ok {
		msg.DeletedAt = &value
	}
	if value, ok := values["deleted_by_admin_id"].(uint64); ok {
		msg.DeletedByAdminID = &value
	}
	if value, ok := values["delete_reason"].(*string); ok {
		msg.DeleteReason = value
	}
	return nil
}
func (r *fakeRepo) WriteLog(_ context.Context, input operationlog.WriteInput) error {
	r.logs = append(r.logs, input)
	return nil
}

type fakeUOW struct{ repo messagerepo.Repository }

func (u fakeUOW) WithinTransaction(ctx context.Context, fn func(messagerepo.Repository) error) error {
	return fn(u.repo)
}

func newTestService(repo *fakeRepo, now time.Time) *service {
	svc := NewService(repo, fakeUOW{repo: repo}).(*service)
	svc.now = func() time.Time { return now }
	return svc
}

func addMessage(repo *fakeRepo, id uint64, status string, createdAt time.Time) {
	phone := "sensitive-phone"
	wechat := "sensitive-wechat"
	ip := "127.0.0.1"
	repo.messages[id] = &messagemodel.VisitorMessage{
		ID: id, FamilyID: 2, VisitorPhone: &phone, VisitorWechat: &wechat,
		MessageContent: "hello", Status: status, IP: &ip, CreatedAt: createdAt, UpdatedAt: createdAt,
	}
	if id >= repo.nextID {
		repo.nextID = id + 1
	}
}

func TestVisitorMessageSubmissionAndRateLimit(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	t.Run("approved family accepts pending message without operation log", func(t *testing.T) {
		repo := newFakeRepo()
		result, err := newTestService(repo, now).CreatePublic(context.Background(), 2, messagedto.CreateVisitorMessageRequest{
			MessageContent: "hello",
		}, AuditInput{IP: "127.0.0.1"})
		if err != nil || result.Status != messageenum.StatusPending || len(repo.logs) != 0 {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
	})
	t.Run("private family rejected", func(t *testing.T) {
		repo := newFakeRepo()
		repo.family.PublicDisplayStatus = publicenum.PublicPrivate
		if _, err := newTestService(repo, now).CreatePublic(context.Background(), 2, messagedto.CreateVisitorMessageRequest{MessageContent: "hello"}, AuditInput{IP: "127.0.0.1"}); err == nil || err.Code != CodeVisitorMessageFamilyPrivate {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("empty and too long content rejected", func(t *testing.T) {
		svc := newTestService(newFakeRepo(), now)
		if _, err := svc.CreatePublic(context.Background(), 2, messagedto.CreateVisitorMessageRequest{MessageContent: " "}, AuditInput{}); err == nil || err.Code != CodeVisitorMessageContentEmpty {
			t.Fatalf("unexpected %#v", err)
		}
		if _, err := svc.CreatePublic(context.Background(), 2, messagedto.CreateVisitorMessageRequest{MessageContent: strings.Repeat("字", 1001)}, AuditInput{}); err == nil || err.Code != CodeVisitorMessageContentTooLong {
			t.Fatalf("unexpected %#v", err)
		}
	})
	t.Run("one minute and daily limits enforced", func(t *testing.T) {
		repo := newFakeRepo()
		addMessage(repo, 1, messageenum.StatusPending, now.Add(-30*time.Second))
		if _, err := newTestService(repo, now).CreatePublic(context.Background(), 2, messagedto.CreateVisitorMessageRequest{MessageContent: "hello"}, AuditInput{IP: "127.0.0.1"}); err == nil || err.Code != CodeVisitorMessageRateLimited {
			t.Fatalf("minute limit: %#v", err)
		}
		repo = newFakeRepo()
		for i := uint64(1); i <= 20; i++ {
			addMessage(repo, i, messageenum.StatusPending, now.Add(-2*time.Hour))
		}
		if _, err := newTestService(repo, now).CreatePublic(context.Background(), 2, messagedto.CreateVisitorMessageRequest{MessageContent: "hello"}, AuditInput{IP: "127.0.0.1"}); err == nil || err.Code != CodeVisitorMessageRateLimited {
			t.Fatalf("daily limit: %#v", err)
		}
	})
}

func TestPublicMessageListIsApprovedAndRedacted(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	repo := newFakeRepo()
	addMessage(repo, 1, messageenum.StatusApproved, now)
	addMessage(repo, 2, messageenum.StatusPending, now)
	result, err := newTestService(repo, now).ListPublic(context.Background(), 2, messagedto.ListMessagesQuery{})
	if err != nil || len(result.Items) != 1 || result.Items[0].MessageID != 1 {
		t.Fatalf("unexpected %#v %#v", result, err)
	}
	data, _ := json.Marshal(result)
	if strings.Contains(string(data), "sensitive-phone") || strings.Contains(string(data), "sensitive-wechat") {
		t.Fatalf("public result leaked contacts: %s", data)
	}
}

func TestAdminVisitorMessageFlowAndSafeLogs(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	role := string(enums.AdminRolePlatformAdmin)
	t.Run("approve pending and rejected", func(t *testing.T) {
		for _, status := range []string{messageenum.StatusPending, messageenum.StatusRejected} {
			repo := newFakeRepo()
			addMessage(repo, 1, status, now)
			result, err := newTestService(repo, now).Approve(context.Background(), 1, role, 1, messagedto.ReviewVisitorMessageRequest{}, AuditInput{})
			if err != nil || result.Status != messageenum.StatusApproved || len(repo.logs) != 1 {
				t.Fatalf("unexpected %#v %#v", result, err)
			}
			assertSafeLogs(t, repo.logs)
		}
	})
	t.Run("reject and delete", func(t *testing.T) {
		repo := newFakeRepo()
		addMessage(repo, 1, messageenum.StatusPending, now)
		svc := newTestService(repo, now)
		result, err := svc.Reject(context.Background(), 1, role, 1, messagedto.ReviewVisitorMessageRequest{}, AuditInput{})
		if err != nil || result.Status != messageenum.StatusRejected {
			t.Fatalf("unexpected %#v %#v", result, err)
		}
		if err := svc.Delete(context.Background(), 1, role, 1, messagedto.DeleteVisitorMessageRequest{}, AuditInput{}); err != nil {
			t.Fatalf("delete error: %#v", err)
		}
		if repo.messages[1].Status != messageenum.StatusDeleted || repo.messages[1].DeletedAt == nil || len(repo.logs) != 2 {
			t.Fatal("message was not soft deleted and logged")
		}
		assertSafeLogs(t, repo.logs)
	})
	t.Run("unauthorized role denied", func(t *testing.T) {
		repo := newFakeRepo()
		addMessage(repo, 1, messageenum.StatusPending, now)
		if _, err := newTestService(repo, now).Approve(context.Background(), 1, "UNKNOWN", 1, messagedto.ReviewVisitorMessageRequest{}, AuditInput{}); err == nil || err.Code != CodeVisitorMessageForbidden {
			t.Fatalf("unexpected %#v", err)
		}
	})
}

func assertSafeLogs(t *testing.T, logs []operationlog.WriteInput) {
	t.Helper()
	data, _ := json.Marshal(logs)
	if strings.Contains(string(data), "sensitive-phone") || strings.Contains(string(data), "sensitive-wechat") {
		t.Fatalf("operation log leaked contacts: %s", data)
	}
}
