package kinship

import "testing"

func TestResolveBasicKinship(t *testing.T) {
	tests := []struct {
		name string
		req  ResolveRequest
		want string
	}{
		{
			name: "grandfather",
			req: ResolveRequest{
				Self: Person{Name: "我", Gender: "male", Birthday: "1990-01-01"},
				Rows: []PathRow{
					{Relation: "parent", TargetGender: "male", TargetBirthday: "1965-01-01"},
					{Relation: "parent", TargetGender: "male", TargetBirthday: "1940-01-01"},
				},
			},
			want: "爷爷",
		},
		{
			name: "younger brother wife",
			req: ResolveRequest{
				Self: Person{Name: "我", Gender: "male", Birthday: "1990-01-01"},
				Rows: []PathRow{
					{Relation: "sibling", TargetGender: "male", TargetBirthday: "1995-01-01"},
					{Relation: "spouse", TargetGender: "female", TargetBirthday: "1996-01-01"},
				},
			},
			want: "弟媳",
		},
		{
			name: "paternal male cousin older",
			req: ResolveRequest{
				Self: Person{Name: "我", Gender: "male", Birthday: "1990-01-01"},
				Rows: []PathRow{
					{Relation: "parent", TargetGender: "male", TargetBirthday: "1965-01-01"},
					{Relation: "sibling", TargetGender: "male", TargetBirthday: "1960-01-01"},
					{Relation: "child", TargetGender: "male", TargetBirthday: "1988-01-01"},
				},
			},
			want: "堂哥",
		},
		{
			name: "paternal uncle spouse son younger",
			req: ResolveRequest{
				Self: Person{Name: "我", Gender: "male", Birthday: "1990-01-01"},
				Rows: []PathRow{
					{Relation: "parent", TargetGender: "male", TargetAge: 50},
					{Relation: "sibling", TargetGender: "male", TargetAge: 55},
					{Relation: "spouse", TargetGender: "unknown", TargetAge: 50},
					{Relation: "child", TargetGender: "male", TargetAge: 30},
				},
			},
			want: "堂弟",
		},
		{
			name: "paternal elder male cousin wife",
			req: ResolveRequest{
				Self: Person{Name: "我", Gender: "male", Birthday: "1990-01-01"},
				Rows: []PathRow{
					{Relation: "parent", TargetGender: "male", TargetBirthday: "1965-01-01"},
					{Relation: "sibling", TargetGender: "male", TargetBirthday: "1960-01-01"},
					{Relation: "child", TargetGender: "male", TargetBirthday: "1988-01-01"},
					{Relation: "spouse", TargetGender: "female", TargetBirthday: "1990-01-01"},
				},
			},
			want: "堂嫂",
		},
		{
			name: "paternal younger female cousin husband",
			req: ResolveRequest{
				Self: Person{Name: "我", Gender: "male", Birthday: "1990-01-01"},
				Rows: []PathRow{
					{Relation: "parent", TargetGender: "male", TargetBirthday: "1965-01-01"},
					{Relation: "sibling", TargetGender: "male", TargetBirthday: "1960-01-01"},
					{Relation: "child", TargetGender: "female", TargetBirthday: "1995-01-01"},
					{Relation: "spouse", TargetGender: "male", TargetBirthday: "1993-01-01"},
				},
			},
			want: "堂妹夫",
		},
		{
			name: "paternal female cousin spouse son",
			req: ResolveRequest{
				Self: Person{Name: "我", Gender: "male", Birthday: "1990-01-01"},
				Rows: []PathRow{
					{Relation: "parent", TargetGender: "male", TargetBirthday: "1965-01-01"},
					{Relation: "sibling", TargetGender: "male", TargetBirthday: "1960-01-01"},
					{Relation: "child", TargetGender: "female", TargetBirthday: "1995-01-01"},
					{Relation: "spouse", TargetGender: "male", TargetBirthday: "1993-01-01"},
					{Relation: "child", TargetGender: "male", TargetBirthday: "2020-01-01"},
				},
			},
			want: "堂外甥",
		},
		{
			name: "maternal male cousin daughter",
			req: ResolveRequest{
				Self: Person{Name: "我", Gender: "male", Birthday: "1990-01-01"},
				Rows: []PathRow{
					{Relation: "parent", TargetGender: "female", TargetBirthday: "1968-01-01"},
					{Relation: "sibling", TargetGender: "female", TargetBirthday: "1965-01-01"},
					{Relation: "child", TargetGender: "male", TargetBirthday: "1996-01-01"},
					{Relation: "child", TargetGender: "female", TargetBirthday: "2022-01-01"},
				},
			},
			want: "表侄女",
		},
		{
			name: "maternal female cousin younger",
			req: ResolveRequest{
				Self: Person{Name: "我", Gender: "male", Birthday: "1990-01-01"},
				Rows: []PathRow{
					{Relation: "parent", TargetGender: "female", TargetBirthday: "1968-01-01"},
					{Relation: "sibling", TargetGender: "female", TargetBirthday: "1965-01-01"},
					{Relation: "child", TargetGender: "female", TargetBirthday: "1996-01-01"},
				},
			},
			want: "表妹",
		},
		{
			name: "wife elder brother",
			req: ResolveRequest{
				Self: Person{Name: "我", Gender: "male", Birthday: "1990-01-01"},
				Rows: []PathRow{
					{Relation: "spouse", TargetGender: "female", TargetBirthday: "1992-01-01"},
					{Relation: "sibling", TargetGender: "male", TargetBirthday: "1988-01-01"},
				},
			},
			want: "大舅子",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(tt.req)
			if got.Name != tt.want {
				t.Fatalf("Resolve() name = %q, want %q; detail=%v", got.Name, tt.want, got.DetailPath)
			}
			if !got.Matched {
				t.Fatalf("Resolve() matched = false, want true")
			}
		})
	}
}

func TestRankTitle(t *testing.T) {
	tests := map[int]string{
		1:  "大哥",
		2:  "二哥",
		6:  "六哥",
		11: "十一哥",
		21: "二十一哥",
	}

	for rank, want := range tests {
		if got := RankTitle(rank, "哥"); got != want {
			t.Fatalf("RankTitle(%d, 哥) = %q, want %q", rank, got, want)
		}
	}
}
