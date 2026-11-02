package dto

type CreateInvitationRequest struct {
	InviteChannel         string  `json:"inviteChannel" binding:"required"`
	TargetUserID          *uint64 `json:"targetUserId"`
	InviteMessage         *string `json:"inviteMessage"`
	FamilyRoleAfterAccept *string `json:"familyRoleAfterAccept"`
}

type RejectInvitationRequest struct {
	Reason *string `json:"reason"`
}

type CancelInvitationRequest struct {
	Reason *string `json:"reason"`
}
