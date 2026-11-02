package vo

import "time"

type TransferRequest struct {
	RequestID         uint64     `json:"requestId"`
	FamilyID          uint64     `json:"familyId"`
	FamilyName        string     `json:"familyName,omitempty"`
	FromMemberID      uint64     `json:"fromMemberId"`
	FromUserID        uint64     `json:"fromUserId"`
	ToMemberID        uint64     `json:"toMemberId"`
	ToUserID          uint64     `json:"toUserId"`
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
	Items    []TransferRequest `json:"items"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
	Total    int64             `json:"total"`
}
