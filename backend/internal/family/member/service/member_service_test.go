package service

import (
	"testing"

	membermodel "tree/backend/internal/family/member/model"
	memberrepo "tree/backend/internal/family/member/repository"
)

func TestParseDateOrYearUsesExplicitDate(t *testing.T) {
	date := "1992-06-07"
	year := 1980

	value, err := parseDateOrYear(&date, &year)
	if err != nil {
		t.Fatalf("parseDateOrYear returned error: %v", err)
	}
	if got := value.Format("2006-01-02"); got != date {
		t.Fatalf("expected explicit date %s, got %s", date, got)
	}
}

func TestParseDateOrYearConvertsYearToJanuaryFirst(t *testing.T) {
	year := 1992

	value, err := parseDateOrYear(nil, &year)
	if err != nil {
		t.Fatalf("parseDateOrYear returned error: %v", err)
	}
	if got := value.Format("2006-01-02"); got != "1992-01-01" {
		t.Fatalf("expected 1992-01-01, got %s", got)
	}
}

func TestMemberValuesRejectsDeathBeforeBirth(t *testing.T) {
	birth := "2000-01-01"
	death := "1999-12-31"

	_, businessErr := memberValues(nil, &birth, nil, &death, nil, nil, nil, nil, nil, nil)
	if businessErr == nil || businessErr.Code != CodeMemberInvalidField {
		t.Fatalf("expected CodeMemberInvalidField, got %#v", businessErr)
	}
}

func TestMemberVOHidesHiddenMemorialDescriptionForRegularMember(t *testing.T) {
	alive := false
	description := "仅管理员可见"
	note, noteType, err := encodeProfile(&description, nil)
	if err != nil {
		t.Fatalf("encodeProfile returned error: %v", err)
	}
	row := &memberrepo.MemberRow{FamilyMember: membermodel.FamilyMember{
		ID: 1, FamilyID: 2, DisplayName: "亲人", IsLiving: &alive,
		LineageNote: note, LineageNoteType: noteType, MemorialVisible: false,
	}}

	regular := memberVO(row, false)
	if regular.Description != nil {
		t.Fatalf("expected hidden description for regular member, got %#v", regular.Description)
	}
	if regular.MemorialVisible {
		t.Fatalf("expected memorialVisible false")
	}

	manager := memberVO(row, true)
	if manager.Description == nil || *manager.Description != description {
		t.Fatalf("manager should still see description, got %#v", manager.Description)
	}
}

func TestProfileRoundTrip(t *testing.T) {
	description := "member description"
	avatarURL := "https://example.com/avatar.png"

	note, noteType, err := encodeProfile(&description, &avatarURL)
	if err != nil {
		t.Fatalf("encodeProfile returned error: %v", err)
	}
	profile := decodeProfile(&membermodel.FamilyMember{
		LineageNote:     note,
		LineageNoteType: noteType,
	})
	if profile.Description == nil || *profile.Description != description {
		t.Fatalf("description did not round trip: %#v", profile.Description)
	}
	if profile.AvatarURL == nil || *profile.AvatarURL != avatarURL {
		t.Fatalf("avatar URL did not round trip: %#v", profile.AvatarURL)
	}
}
