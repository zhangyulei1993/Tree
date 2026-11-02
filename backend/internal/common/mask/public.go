package mask

import (
	"strings"
	"unicode"
)

const PublicContactInviteHint = "请通过家庭成员分享的邀请联系。"

func DisplayName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) == 1 {
		return string(runes[0]) + "*"
	}
	masked := make([]rune, len(runes))
	copy(masked, runes)
	masked[1] = '*'
	return string(masked)
}

func Phone(value string) string {
	digits := extractDigits(value)
	if len(digits) == 11 && digits[0] == '1' {
		return digits[:3] + "****" + digits[7:]
	}
	return "****"
}

func Wechat(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= 4 {
		return "****"
	}
	return string(runes[:2]) + "****" + string(runes[len(runes)-2:])
}

func Email(value string) string {
	trimmed := strings.TrimSpace(value)
	at := strings.Index(trimmed, "@")
	if at <= 0 {
		return "****"
	}
	local := []rune(trimmed[:at])
	if len(local) == 0 {
		return "****"
	}
	return string(local[0]) + "***" + trimmed[at:]
}

func extractDigits(value string) string {
	var builder strings.Builder
	for _, r := range value {
		if unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
