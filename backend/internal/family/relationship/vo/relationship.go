package vo

import "time"

type CreatedMember struct {
	MemberID uint64 `json:"memberId"`
	FamilyID uint64 `json:"familyId"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Status   string `json:"status"`
}

type RelationshipVO struct {
	RelationshipID   uint64     `json:"relationshipId"`
	FamilyID         uint64     `json:"familyId"`
	FromMemberID     uint64     `json:"fromMemberId"`
	ToMemberID       uint64     `json:"toMemberId"`
	RelationshipType string     `json:"relationshipType"`
	ParentLinkType   *string    `json:"parentLinkType,omitempty"`
	RelationNoteType *string    `json:"relationNoteType,omitempty"`
	RelationNote     *string    `json:"relationNote,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	DeletedAt        *time.Time `json:"deletedAt,omitempty"`
}

type MutationResult struct {
	CreatedMember *CreatedMember   `json:"createdMember,omitempty"`
	Relationships []RelationshipVO `json:"relationships"`
	GraphVersion  int64            `json:"graphVersion"`
}
