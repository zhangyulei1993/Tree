package mask

import "testing"

func TestDisplayNameMasking(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{input: "张三", expected: "张*"},
		{input: "张建国", expected: "张*国"},
		{input: "欧阳娜娜", expected: "欧*娜娜"},
		{input: "张", expected: "张*"},
		{input: "", expected: ""},
		{input: "  ", expected: ""},
		{input: "𠮷野家", expected: "𠮷*家"},
	}
	for _, tc := range cases {
		if got := DisplayName(tc.input); got != tc.expected {
			t.Fatalf("DisplayName(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestPhoneMasking(t *testing.T) {
	if got := Phone("13812345678"); got != "138****5678" {
		t.Fatalf("Phone() = %q", got)
	}
	if got := Phone("16600020101"); got != "166****0101" {
		t.Fatalf("Phone() = %q", got)
	}
	for _, raw := range []string{"123", "1381234", "abc", "138123456789"} {
		if got := Phone(raw); got != "****" {
			t.Fatalf("Phone(%q) = %q, want ****", raw, got)
		}
	}
}

func TestWechatMasking(t *testing.T) {
	if got := Wechat("wx_secret_01"); got != "wx****01" {
		t.Fatalf("Wechat() = %q", got)
	}
	if got := Wechat("ab"); got != "****" {
		t.Fatalf("Wechat() = %q", got)
	}
}

func TestEmailMasking(t *testing.T) {
	if got := Email("zhang@example.com"); got != "z***@example.com" {
		t.Fatalf("Email() = %q", got)
	}
	if got := Email("bad"); got != "****" {
		t.Fatalf("Email() = %q", got)
	}
}
