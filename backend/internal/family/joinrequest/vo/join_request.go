package vo

import "time"

type JoinRequest struct {
	RequestID         uint64     `json:"requestId"`
	FamilyID          uint64     `json:"familyId"`
	FamilyName        string     `json:"familyName,omitempty"`
	ApplicantUserID   uint64     `json:"applicantUserId"`
	ApplicantRealName *string    `json:"applicantRealName,omitempty"`
	ApplicantMessage  *string    `json:"applicantMessage,omitempty"`
	RequestStatus     string     `json:"requestStatus"`
	ApproveMode       *string    `json:"approveMode,omitempty"`
	BoundMemberID     *uint64    `json:"boundMemberId,omitempty"`
	CreatedMemberID   *uint64    `json:"createdMemberId,omitempty"`
	HandleComment     *string    `json:"handleComment,omitempty"`
	HandledAt         *time.Time `json:"handledAt,omitempty"`
	CancelledAt       *time.Time `json:"cancelledAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	GraphVersion      *int64     `json:"graphVersion,omitempty"`
}
