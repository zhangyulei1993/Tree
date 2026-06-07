package vo

type Node struct {
	MemberID            uint64  `json:"memberId"`
	DisplayName         string  `json:"displayName"`
	Surname             *string `json:"surname,omitempty"`
	GenerationCharacter *string `json:"generationCharacter,omitempty"`
	Gender              string  `json:"gender"`
	MemberType          string  `json:"memberType"`
	BirthDate           *string `json:"birthDate,omitempty"`
	DeathDate           *string `json:"deathDate,omitempty"`
	IsLiving            *bool   `json:"isLiving,omitempty"`
	UserBindingState    string  `json:"userBindingState"`
	CanExpand           bool    `json:"canExpand"`
	StopReason          *string `json:"stopReason,omitempty"`
}

type Edge struct {
	RelationshipID   uint64  `json:"relationshipId"`
	FromMemberID     uint64  `json:"fromMemberId"`
	ToMemberID       uint64  `json:"toMemberId"`
	RelationshipType string  `json:"relationshipType"`
	ParentLinkType   *string `json:"parentLinkType,omitempty"`
	RelationNoteType *string `json:"relationNoteType,omitempty"`
	RelationNote     *string `json:"relationNote,omitempty"`
}

type TreeItem struct {
	MemberID    uint64   `json:"memberId"`
	ParentIDs   []uint64 `json:"parentIds"`
	ChildrenIDs []uint64 `json:"childrenIds"`
	SpouseIDs   []uint64 `json:"spouseIds"`
}

type TreeResult struct {
	FamilyID     uint64     `json:"familyId"`
	TreeMode     string     `json:"treeMode"`
	GraphVersion int64      `json:"graphVersion"`
	Nodes        []Node     `json:"nodes"`
	Edges        []Edge     `json:"edges"`
	Tree         []TreeItem `json:"tree"`
}
