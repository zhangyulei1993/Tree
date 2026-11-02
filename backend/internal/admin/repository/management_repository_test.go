package repository

import (
	"testing"

	usermodel "tree/backend/internal/user/model"
)

func TestResolveLoginMethodLabels(t *testing.T) {
	cases := []struct {
		hasWechat bool
		want      string
	}{
		{true, "微信登录"},
		{false, "未设置"},
	}
	for _, tc := range cases {
		if got := resolveLoginMethod(tc.hasWechat); got != tc.want {
			t.Fatalf("resolveLoginMethod(%v)=%q want %q", tc.hasWechat, got, tc.want)
		}
	}
}

func TestUserRowUsesPhoneBindingForTrustTier(t *testing.T) {
	row := userRow(usermodel.User{ID: 1, PhoneVerified: true, Status: "ACTIVE"}, true)
	if !row.HasWechatLogin || row.TrustTier != "PHONE_BOUND" || row.LoginMethod != "微信登录" {
		t.Fatalf("unexpected phone-bound row: %#v", row)
	}
}
