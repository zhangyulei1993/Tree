package dto

import relationshipdto "tree/backend/internal/family/relationship/dto"

type CreateJoinRequest struct {
	ApplicantRealName *string `json:"applicantRealName"`
	ApplicantGender   *string `json:"applicantGender"`
	ApplicantMessage  *string `json:"applicantMessage"`
}

type NewMemberInput struct {
	Name              string  `json:"name"`
	Gender            *string `json:"gender"`
	BirthDate         *string `json:"birthDate"`
	BirthYear         *int    `json:"birthYear"`
	IsAlive           *bool   `json:"isAlive"`
	UserBindingPolicy *string `json:"userBindingPolicy"`
}

type ApproveJoinRequest struct {
	ApproveMode   string          `json:"approveMode" binding:"required"`
	MemberID      *uint64         `json:"memberId"`
	NewMember     *NewMemberInput `json:"newMember"`
	Location      *MemberLocation `json:"location"`
	HandleComment *string         `json:"handleComment"`
}

type MemberLocation struct {
	BaseMemberID uint64                            `json:"baseMemberId" binding:"required"`
	AddType      string                            `json:"addType" binding:"required"`
	MemberType   *string                           `json:"memberType"`
	Relationship relationshipdto.RelationshipInput `json:"relationship"`
}

type RejectJoinRequest struct {
	HandleComment *string `json:"handleComment"`
}

type CancelJoinRequest struct {
	CancelReason *string `json:"cancelReason"`
}
