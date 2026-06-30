package vo

import "time"

type Invitation struct {
	InvitationID          uint64     `json:"invitationId"`
	FamilyID              uint64     `json:"familyId"`
	FamilyName            string     `json:"familyName"`
	TargetMemberID        uint64     `json:"targetMemberId"`
	TargetMemberName      string     `json:"targetMemberName"`
	InviterDisplayName    string     `json:"inviterDisplayName"`
	InviterRole           string     `json:"inviterRole"`
	InviteChannel         string     `json:"inviteChannel"`
	InviteMessage         *string    `json:"inviteMessage,omitempty"`
	FamilyRoleAfterAccept string     `json:"familyRoleAfterAccept"`
	Status                string     `json:"status"`
	ExpiredAt             time.Time  `json:"expiredAt"`
	AcceptedAt            *time.Time `json:"acceptedAt,omitempty"`
	RejectedAt            *time.Time `json:"rejectedAt,omitempty"`
	CancelledAt           *time.Time `json:"cancelledAt,omitempty"`
	CreatedAt             time.Time  `json:"createdAt"`
}

type CreatedInvitation struct {
	Invitation  Invitation `json:"invitation"`
	InviteToken string     `json:"inviteToken"`
}
