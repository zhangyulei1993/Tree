package wechat

import "testing"

func TestSubscribeDataKeysKeepsTemplateOrder(t *testing.T) {
	keys := subscribeDataKeys("商品:{{thing28.DATA}}\n订单:{{character_string4.DATA}}\n订单:{{character_string9.DATA}}\n金额:{{amount3.DATA}}\n付款人:{{name1.DATA}}\n日期:{{date2.DATA}}")
	want := []string{"thing28", "character_string4", "character_string9", "amount3", "name1", "date2"}
	if len(keys) != len(want) {
		t.Fatalf("key count = %d, want %d", len(keys), len(want))
	}
	for index := range want {
		if keys[index] != want[index] {
			t.Fatalf("key %d = %q, want %q", index, keys[index], want[index])
		}
	}
}

func TestNormalizeSubscribeValueForWechatFieldTypes(t *testing.T) {
	cases := []struct {
		key   string
		value string
		want  string
	}{
		{key: "date2", value: "2026-09-29 15:23", want: "2026-09-29"},
		{key: "time1", value: "2026-09-29 15:23", want: "15:23"},
		{key: "thing2", value: "家庭扩容退款申请审核通知内容", want: "家庭扩容退款申请审核通知内容"},
		{key: "phrase5", value: "审核已通过并完成", want: "审核已通过"},
	}
	for _, item := range cases {
		if got := normalizeSubscribeValue(item.key, item.value); got != item.want {
			t.Fatalf("normalize(%s) = %q, want %q", item.key, got, item.want)
		}
	}
}
