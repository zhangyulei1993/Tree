package dto

type NewMemberInput struct {
	Name              string  `json:"name" binding:"required"`
	Gender            *string `json:"gender"`
	BirthDate         *string `json:"birthDate"`
	BirthYear         *int    `json:"birthYear"`
	DeathDate         *string `json:"deathDate"`
	DeathYear         *int    `json:"deathYear"`
	IsAlive           *bool   `json:"isAlive"`
	UserBindingPolicy *string `json:"userBindingPolicy"`
}

type RelationshipInput struct {
	RelationshipType *string `json:"relationshipType"`
	ParentLinkType   *string `json:"parentLinkType"`
	RelationNoteType *string `json:"relationNoteType"`
	RelationNote     *string `json:"relationNote"`
}

type CreateRelationshipRequest struct {
	BaseMemberID uint64            `json:"baseMemberId" binding:"required"`
	AddType      string            `json:"addType" binding:"required"`
	NewMember    NewMemberInput    `json:"newMember" binding:"required"`
	Relationship RelationshipInput `json:"relationship"`
}

type UpdateRelationshipRequest struct {
	ParentLinkType   *string `json:"parentLinkType"`
	RelationNoteType *string `json:"relationNoteType"`
	RelationNote     *string `json:"relationNote"`
	FromMemberID     *uint64 `json:"fromMemberId"`
	ToMemberID       *uint64 `json:"toMemberId"`
	RelationshipType *string `json:"relationshipType"`
}

type DeleteRelationshipRequest struct {
	Reason *string `json:"reason"`
}
