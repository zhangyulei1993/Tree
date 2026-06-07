package dto

type CreateJoinRequest struct {
	ApplicantRealName *string `json:"applicantRealName"`
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
	HandleComment *string         `json:"handleComment"`
}

type RejectJoinRequest struct {
	HandleComment *string `json:"handleComment"`
}

type CancelJoinRequest struct {
	CancelReason *string `json:"cancelReason"`
}
