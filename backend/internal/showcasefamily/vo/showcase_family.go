package vo

import "time"

type ShowcaseFamily struct {
	ID                uint64     `json:"id"`
	DisplayName       string     `json:"displayName"`
	RegionText        string     `json:"regionText"`
	Summary           string     `json:"summary"`
	CoverImageURL     string     `json:"coverImageUrl"`
	TreeImageURL      string     `json:"treeImageUrl"`
	Status            string     `json:"status"`
	SortOrder         int        `json:"sortOrder"`
	MemberCount       int        `json:"memberCount"`
	RelationshipCount int        `json:"relationshipCount"`
	TreeStatus        string     `json:"treeStatus"`
	PublishedAt       *time.Time `json:"publishedAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	RotationPinned    bool       `json:"rotationPinned"`
}

type DisplaySettings struct {
	RotationDays           int        `json:"rotationDays"`
	PinnedShowcaseFamilyID *uint64    `json:"pinnedShowcaseFamilyId"`
	PinnedUntil            *time.Time `json:"pinnedUntil"`
}

type TreeMember struct {
	MemberID     uint64 `json:"memberId"`
	DisplayName  string `json:"displayName"`
	IdentityText string `json:"identityText"`
	MemberType   string `json:"memberType"`
	Gender       string `json:"gender"`
	BirthYear    *int   `json:"birthYear"`
	SortOrder    int    `json:"sortOrder"`
}

type TreeRelationship struct {
	RelationshipID   uint64 `json:"relationshipId"`
	FromMemberID     uint64 `json:"fromMemberId"`
	ToMemberID       uint64 `json:"toMemberId"`
	RelationshipType string `json:"relationshipType"`
}

type ShowcaseFamilyTree struct {
	ShowcaseFamilyID uint64             `json:"showcaseFamilyId"`
	Nodes            []TreeMember       `json:"nodes"`
	Edges            []TreeRelationship `json:"edges"`
}

type AddRelativeResult struct {
	Member TreeMember  `json:"member"`
	Spouse *TreeMember `json:"spouse,omitempty"`
}
