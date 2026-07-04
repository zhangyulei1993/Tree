package repository

import (
	"testing"

	usermodel "tree/backend/internal/user/model"
)

func TestResolveLoginMethodLabels(t *testing.T) {
	cases := []struct {
		hasWechat, phoneEnabled bool
		want                    string
	}{
		{true, true, "微信 + 手机号"},
		{true, false, "仅微信"},
		{false, true, "仅手机号"},
		{false, false, "未设置"},
	}
	for _, tc := range cases {
		if got := resolveLoginMethod(tc.hasWechat, tc.phoneEnabled); got != tc.want {
			t.Fatalf("resolveLoginMethod(%v,%v)=%q want %q", tc.hasWechat, tc.phoneEnabled, got, tc.want)
		}
	}
}

func TestUserRowCanUnbindPhoneLogin(t *testing.T) {
	row := userRow(usermodel.User{ID: 1, PhoneLoginEnabled: true, Status: "ACTIVE"}, true)
	if !row.HasWechatLogin || !row.CanUnbindPhoneLogin || row.LoginMethod != "微信 + 手机号" {
		t.Fatalf("unexpected wechat+phone row: %#v", row)
	}
	phoneOnly := userRow(usermodel.User{ID: 2, PhoneLoginEnabled: true, Status: "ACTIVE"}, false)
	if phoneOnly.CanUnbindPhoneLogin || phoneOnly.LoginMethod != "仅手机号" {
		t.Fatalf("unexpected phone-only row: %#v", phoneOnly)
	}
}
