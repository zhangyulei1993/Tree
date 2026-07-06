package contentsafety

import "strings"

func OptionalField(label string, value *string) []Field {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return []Field{{Label: label, Value: trimmed}}
}

func StringField(label, value string) []Field {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return []Field{{Label: label, Value: trimmed}}
}

func MergeFields(parts ...[]Field) []Field {
	total := 0
	for _, part := range parts {
		total += len(part)
	}
	if total == 0 {
		return nil
	}
	out := make([]Field, 0, total)
	for _, part := range parts {
		out = append(out, part...)
	}
	return out
}
