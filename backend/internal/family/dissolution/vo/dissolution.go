package vo

import "time"

type DissolutionRequest struct {
	RequestID         uint64     `json:"requestId"`
	FamilyID          uint64     `json:"familyId"`
	FamilyName        string     `json:"familyName,omitempty"`
	FamilyStatus      string     `json:"familyStatus,omitempty"`
	RequesterMemberID uint64     `json:"requesterMemberId"`
	RequesterUserID   uint64     `json:"requesterUserId"`
	RequestStatus     string     `json:"requestStatus"`
	RequestReason     *string    `json:"requestReason,omitempty"`
	ReviewResult      *string    `json:"reviewResult,omitempty"`
	ReviewedByAdminID *uint64    `json:"reviewedByAdminId,omitempty"`
	ReviewedAt        *time.Time `json:"reviewedAt,omitempty"`
	ReviewComment     *string    `json:"reviewComment,omitempty"`
	CompletedAt       *time.Time `json:"completedAt,omitempty"`
	CancelledAt       *time.Time `json:"cancelledAt,omitempty"`
	CancelReason      *string    `json:"cancelReason,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type ListResult struct {
	Items    []DissolutionRequest `json:"items"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
	Total    int64                `json:"total"`
}

type RestoreResult struct {
	FamilyID            uint64     `json:"familyId"`
	Status              string     `json:"status"`
	PublicDisplayStatus string     `json:"publicDisplayStatus"`
	Searchable          bool       `json:"searchable"`
	GraphVersion        int64      `json:"graphVersion"`
	RestoredAt          *time.Time `json:"restoredAt,omitempty"`
}

type FinalizeResult struct {
	FamilyID               uint64     `json:"familyId"`
	Status                 string     `json:"status"`
	PublicDisplayStatus    string     `json:"publicDisplayStatus"`
	Searchable             bool       `json:"searchable"`
	GraphVersion           int64      `json:"graphVersion"`
	DissolutionCompletedAt *time.Time `json:"dissolutionCompletedAt,omitempty"`
}
