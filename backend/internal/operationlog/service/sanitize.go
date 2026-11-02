package service

import (
	"encoding/json"
	"strings"
	"unicode"
)

const redactedValue = "[REDACTED]"

var sensitiveDetailFields = map[string]struct{}{
	"password":         {},
	"passwordhash":     {},
	"accesstoken":      {},
	"refreshtoken":     {},
	"verificationcode": {},
	"openid":           {},
	"unionid":          {},
	"appsecret":        {},
	"sessionkey":       {},
}

func sanitizeDetailJSON(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return json.Marshal(sanitizeDetailValue(value))
}

func sanitizeDetailValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		cleaned := make(map[string]any, len(typed))
		for key, child := range typed {
			if isSensitiveDetailField(key) {
				cleaned[key] = redactedValue
				continue
			}
			cleaned[key] = sanitizeDetailValue(child)
		}
		return cleaned
	case []any:
		cleaned := make([]any, len(typed))
		for index, child := range typed {
			cleaned[index] = sanitizeDetailValue(child)
		}
		return cleaned
	default:
		return value
	}
}

func isSensitiveDetailField(field string) bool {
	_, sensitive := sensitiveDetailFields[normalizeSensitiveField(field)]
	return sensitive
}

func normalizeSensitiveField(field string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, field)
}

func sanitizeErrorMessage(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := normalizeSensitiveField(*value)
	for field := range sensitiveDetailFields {
		if strings.Contains(normalized, field) {
			redacted := redactedValue
			return &redacted
		}
	}
	return value
}
