package vo

import "time"

type PublicApplication struct {
	ApplicationID     uint64     `json:"applicationId"`
	FamilyID          uint64     `json:"familyId"`
	FamilyName        string     `json:"familyName,omitempty"`
	ApplicantUserID   *uint64    `json:"applicantUserId,omitempty"`
	ApplicantAdminID  *uint64    `json:"applicantAdminId,omitempty"`
	Status            string     `json:"status"`
	Reason            *string    `json:"reason,omitempty"`
	ReviewResult      *string    `json:"reviewResult,omitempty"`
	ReviewedByAdminID *uint64    `json:"reviewedByAdminId,omitempty"`
	ReviewedAt        *time.Time `json:"reviewedAt,omitempty"`
	ReviewComment     *string    `json:"reviewComment,omitempty"`
	CancelledAt       *time.Time `json:"cancelledAt,omitempty"`
	CancelReason      *string    `json:"cancelReason,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type ListResult struct {
	Items    []PublicApplication `json:"items"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
	Total    int64               `json:"total"`
}

type FamilyPublicStatus struct {
	FamilyID            uint64     `json:"familyId"`
	PublicDisplayStatus string     `json:"publicDisplayStatus"`
	PublicAppliedAt     *time.Time `json:"publicAppliedAt,omitempty"`
	PublicApprovedAt    *time.Time `json:"publicApprovedAt,omitempty"`
	PublicTakenDownAt   *time.Time `json:"publicTakenDownAt,omitempty"`
}
