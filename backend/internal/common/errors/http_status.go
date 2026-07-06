package errors

import "net/http"

func HTTPStatus(code Code) int {
	switch code {
	case CodeContentSafetyRejected:
		return http.StatusBadRequest
	case CodeContentSafetyUnavailable:
		return http.StatusServiceUnavailable
	case CodeContentSafetyWechatRequired:
		return http.StatusForbidden
	default:
		return 0
	}
}

func StatusOr(code Code, fallback int) int {
	if status := HTTPStatus(code); status != 0 {
		return status
	}
	return fallback
}
