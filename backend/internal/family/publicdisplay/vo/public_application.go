package vo

import "time"

type PublicApplication struct {
	ApplicationID             uint64     `json:"applicationId"`
	FamilyID                  uint64     `json:"familyId"`
	FamilyName                string     `json:"familyName,omitempty"`
	FamilyPublicDisplayStatus string     `json:"familyPublicDisplayStatus,omitempty"`
	ApplicantUserID           *uint64    `json:"applicantUserId,omitempty"`
	ApplicantAdminID          *uint64    `json:"applicantAdminId,omitempty"`
	Status                    string     `json:"status"`
	Reason                    *string    `json:"reason,omitempty"`
	ReviewResult              *string    `json:"reviewResult,omitempty"`
	ReviewedByAdminID         *uint64    `json:"reviewedByAdminId,omitempty"`
	ReviewedAt                *time.Time `json:"reviewedAt,omitempty"`
	ReviewComment             *string    `json:"reviewComment,omitempty"`
	CancelledAt               *time.Time `json:"cancelledAt,omitempty"`
	CancelReason              *string    `json:"cancelReason,omitempty"`
	CreatedAt                 time.Time  `json:"createdAt"`
	UpdatedAt                 time.Time  `json:"updatedAt"`
}

type ListResult struct {
	Items    []PublicApplication `json:"items"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
	Total    int64               `json:"total"`
}

type QualityEligibility struct {
	Eligible             bool                      `json:"eligible"`
	MemberCount          int64                     `json:"memberCount"`
	RelationshipCount    int64                     `json:"relationshipCount"`
	BoundMemberCount     int64                     `json:"boundMemberCount"`
	ProfileCompleteCount int64                     `json:"profileCompleteCount"`
	UnknownGenderCount   int64                     `json:"unknownGenderCount"`
	ProfileComplete      bool                      `json:"profileComplete"`
	Requirements         []string                  `json:"requirements"`
	Checks               []QualityEligibilityCheck `json:"checks"`
}

type QualityEligibilityCheck struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Current  string `json:"current"`
	Required string `json:"required"`
	Met      bool   `json:"met"`
}

type QualityEligibilityConfig struct {
	ID                     uint64 `json:"id"`
	Enabled                bool   `json:"enabled"`
	MinMemberCount         int    `json:"minMemberCount"`
	MinRelationshipCount   int    `json:"minRelationshipCount"`
	MinBoundMemberCount    int    `json:"minBoundMemberCount"`
	MinProfileCompleteRate int    `json:"minProfileCompleteRate"`
	MinBoundMemberRate     int    `json:"minBoundMemberRate"`
	RequireDescription     bool   `json:"requireDescription"`
	RequireLocation        bool   `json:"requireLocation"`
	UpdatedAt              string `json:"updatedAt"`
}

type FamilyPublicStatus struct {
	FamilyID             uint64     `json:"familyId"`
	PublicDisplayStatus  string     `json:"publicDisplayStatus"`
	PublicDisplayEnabled bool       `json:"publicDisplayEnabled"`
	PublicAppliedAt      *time.Time `json:"publicAppliedAt,omitempty"`
	PublicApprovedAt     *time.Time `json:"publicApprovedAt,omitempty"`
	PublicEnabledAt      *time.Time `json:"publicEnabledAt,omitempty"`
	PublicTakenDownAt    *time.Time `json:"publicTakenDownAt,omitempty"`
}
