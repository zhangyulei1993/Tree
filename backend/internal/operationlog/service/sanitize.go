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
	normalized := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, field)
	_, sensitive := sensitiveDetailFields[normalized]
	return sensitive
}
