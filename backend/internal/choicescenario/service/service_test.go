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

func TestBuiltinChoiceScenarioDetectionRequiresExactOptions(t *testing.T) {
	if !isBuiltinChoiceScenario("今天吃什么", []string{"家常菜", "火锅", "烧烤", "面食", "外卖"}) {
		t.Fatal("expected built-in scenario to be detected")
	}
	if isBuiltinChoiceScenario("今天吃什么", []string{"家常菜", "火锅"}) {
		t.Fatal("expected incomplete options not to be detected as built-in")
	}
	if isBuiltinChoiceScenario("今天吃什么", []string{"家常菜", "烧烤", "火锅", "面食", "外卖"}) {
		t.Fatal("expected reordered options not to be detected as built-in")
	}
}
