package service

import (
	"testing"

	choicedto "tree/backend/internal/choicescenario/dto"
)

func TestNormalizeInputTrimsAndCollapsesWhitespace(t *testing.T) {
	title, options, err := normalizeInput(choicedto.CreateRequest{
		Title:   "  下班  去干什么 ",
		Options: []string{" 打篮球 ", "去健身房"},
	})
	if err != nil {
		t.Fatalf("normalizeInput returned error: %v", err)
	}
	if title != "下班 去干什么" {
		t.Fatalf("title = %q", title)
	}
	if len(options) != 2 || options[0] != "打篮球" {
		t.Fatalf("options = %#v", options)
	}
}

func TestNormalizeInputRejectsUnsafeShape(t *testing.T) {
	tests := []choicedto.CreateRequest{
		{Title: "", Options: []string{"A", "B"}},
		{Title: "标题", Options: []string{"A"}},
		{Title: "标题", Options: []string{"A", "A"}},
		{Title: "标题", Options: []string{"A\nB", "C"}},
	}
	for index, input := range tests {
		if _, _, err := normalizeInput(input); err == nil {
			t.Errorf("case %d was accepted", index)
		}
	}
}
