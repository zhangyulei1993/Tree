package errors

import (
	"net/http"
	"testing"
)

func TestContentSafetyHTTPStatusMapping(t *testing.T) {
	cases := []struct {
		code     Code
		expected int
	}{
		{CodeContentSafetyRejected, http.StatusBadRequest},
		{CodeContentSafetyUnavailable, http.StatusServiceUnavailable},
		{CodeContentSafetyWechatRequired, http.StatusForbidden},
	}
	for _, tc := range cases {
		if got := HTTPStatus(tc.code); got != tc.expected {
			t.Fatalf("HTTPStatus(%d) = %d, want %d", tc.code, got, tc.expected)
		}
		if got := StatusOr(tc.code, http.StatusInternalServerError); got != tc.expected {
			t.Fatalf("StatusOr(%d) = %d, want %d", tc.code, got, tc.expected)
		}
	}
}
