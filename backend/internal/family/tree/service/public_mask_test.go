package service

import (
	"encoding/json"
	"strings"
	"testing"

	"tree/backend/internal/family/tree/vo"
)

func TestMaskPublicTreeResultDoesNotMutateSource(t *testing.T) {
	surname := "张"
	genChar := "建"
	stopReason := "隐私截止"
	relationNote := "建国备注"
	source := &vo.TreeResult{
		Nodes: []vo.Node{{
			MemberID: 1, DisplayName: "张建国", Surname: &surname,
			GenerationCharacter: &genChar, StopReason: &stopReason,
		}},
		Edges: []vo.Edge{{
			RelationshipID: 1, FromMemberID: 1, ToMemberID: 2,
			RelationshipType: "PARENT_CHILD", RelationNote: &relationNote,
		}},
	}

	masked := maskPublicTreeResult(source)
	if source.Nodes[0].DisplayName != "张建国" {
		t.Fatalf("source displayName mutated: %#v", source.Nodes[0].DisplayName)
	}
	if source.Nodes[0].GenerationCharacter == nil || *source.Nodes[0].GenerationCharacter != "建" {
		t.Fatalf("source generationCharacter mutated: %#v", source.Nodes[0].GenerationCharacter)
	}
	if source.Nodes[0].StopReason == nil || *source.Nodes[0].StopReason != stopReason {
		t.Fatalf("source stopReason mutated: %#v", source.Nodes[0].StopReason)
	}
	if source.Edges[0].RelationNote == nil || *source.Edges[0].RelationNote != relationNote {
		t.Fatalf("source relationNote mutated: %#v", source.Edges[0].RelationNote)
	}
	if masked.Nodes[0].DisplayName != "张*国" {
		t.Fatalf("masked displayName = %#v", masked.Nodes[0].DisplayName)
	}
	if masked.Nodes[0].GenerationCharacter != nil || masked.Nodes[0].StopReason != nil {
		t.Fatalf("masked node must clear reconstructable fields: %#v", masked.Nodes[0])
	}
	if masked.Edges[0].RelationNote != nil {
		t.Fatalf("masked edge must clear relationNote: %#v", masked.Edges[0])
	}
	if masked.Nodes[0].Surname == nil || *masked.Nodes[0].Surname != "张" {
		t.Fatalf("masked surname should remain: %#v", masked.Nodes[0].Surname)
	}
}

func TestMaskPublicTreeResultJSONOmitsSensitiveKeys(t *testing.T) {
	surname := "张"
	genChar := "建"
	relationNote := "建国备注"
	source := &vo.TreeResult{
		Nodes: []vo.Node{{
			MemberID: 1, DisplayName: "张建国", Surname: &surname, GenerationCharacter: &genChar,
		}},
		Edges: []vo.Edge{{
			RelationshipID: 1, FromMemberID: 1, ToMemberID: 2,
			RelationshipType: "PARENT_CHILD", RelationNote: &relationNote,
		}},
	}
	payload, err := json.Marshal(maskPublicTreeResult(source))
	if err != nil {
		t.Fatalf("marshal masked tree: %v", err)
	}
	body := string(payload)
	for _, forbidden := range []string{"张建国", "generationCharacter", "relationNote"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("masked JSON leaked %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, "张*国") {
		t.Fatalf("masked JSON must contain masked displayName: %s", body)
	}
}
