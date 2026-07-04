package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
)

func TestWriteAdminUnbindNoWechatReturns409(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	write(ctx, nil, apperrors.New(apperrors.CodeAdminUnbindPhoneLoginNoWechat))

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected HTTP 409, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Code != int(apperrors.CodeAdminUnbindPhoneLoginNoWechat) {
		t.Fatalf("expected business code 40411, got %d", payload.Code)
	}
}
