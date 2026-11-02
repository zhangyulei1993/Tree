package dto

type CreateRequest struct {
	DisplayName   string `json:"displayName" binding:"required"`
	RegionText    string `json:"regionText"`
	Summary       string `json:"summary" binding:"required"`
	CoverImageURL string `json:"coverImageUrl"`
	TreeImageURL  string `json:"treeImageUrl"`
	SortOrder     int    `json:"sortOrder"`
}

type UpdateRequest struct {
	DisplayName   string `json:"displayName" binding:"required"`
	RegionText    string `json:"regionText"`
	Summary       string `json:"summary" binding:"required"`
	CoverImageURL string `json:"coverImageUrl"`
	TreeImageURL  string `json:"treeImageUrl"`
	SortOrder     int    `json:"sortOrder"`
}

type MemberRequest struct {
	DisplayName  string `json:"displayName" binding:"required"`
	IdentityText string `json:"identityText"`
	MemberType   string `json:"memberType"`
	Gender       string `json:"gender" binding:"required"`
	BirthYear    *int   `json:"birthYear"`
	SortOrder    int    `json:"sortOrder"`
}

type RelationshipRequest struct {
	FromMemberID     uint64 `json:"fromMemberId" binding:"required"`
	ToMemberID       uint64 `json:"toMemberId" binding:"required"`
	RelationshipType string `json:"relationshipType" binding:"required"`
}

type AddRelativeRequest struct {
	RelationKind           string         `json:"relationKind" binding:"required"`
	Member                 MemberRequest  `json:"member" binding:"required"`
	ExistingSpouseMemberID *uint64        `json:"existingSpouseMemberId"`
	NewSpouse              *MemberRequest `json:"newSpouse"`
}

type DisplaySettingsRequest struct {
	RotationDays int `json:"rotationDays" binding:"required"`
}
