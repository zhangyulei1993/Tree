package vo

import "time"

type FamilySummary struct {
	ID                       uint64     `json:"id"`
	FamilyName               string     `json:"familyName"`
	FamilySurname            string     `json:"familySurname"`
	NativePlace              *string    `json:"nativePlace,omitempty"`
	RegionText               *string    `json:"regionText,omitempty"`
	AvatarURL                *string    `json:"avatarUrl,omitempty"`
	Status                   string     `json:"status"`
	DissolutionCooldownUntil *time.Time `json:"dissolutionCooldownUntil,omitempty"`
	DissolutionCooldownDays  int        `json:"dissolutionCooldownDays"`
	PublicDisplayStatus      string     `json:"publicDisplayStatus"`
	PublicDisplayEnabled     bool       `json:"publicDisplayEnabled"`
	Role                     string     `json:"role"`
}

type FamilyDetail struct {
	ID                       uint64     `json:"id"`
	FamilyName               string     `json:"familyName"`
	FamilySurname            string     `json:"familySurname"`
	NativePlace              *string    `json:"nativePlace,omitempty"`
	RegionCode               *string    `json:"regionCode,omitempty"`
	RegionText               *string    `json:"regionText,omitempty"`
	Description              *string    `json:"description,omitempty"`
	AvatarURL                *string    `json:"avatarUrl,omitempty"`
	Status                   string     `json:"status"`
	DissolutionCooldownUntil *time.Time `json:"dissolutionCooldownUntil,omitempty"`
	DissolutionCooldownDays  int        `json:"dissolutionCooldownDays"`
	DissolutionCompletedAt   *time.Time `json:"dissolutionCompletedAt,omitempty"`
	Searchable               bool       `json:"searchable"`
	PublicDisplayStatus      string     `json:"publicDisplayStatus"`
	PublicDisplayEnabled     bool       `json:"publicDisplayEnabled"`
	PublicContactName        *string    `json:"publicContactName,omitempty"`
	PublicContactPhone       *string    `json:"publicContactPhone,omitempty"`
	PublicContactWechat      *string    `json:"publicContactWechat,omitempty"`
	PublicContactNote        *string    `json:"publicContactNote,omitempty"`
	PublicContactVisible     bool       `json:"publicContactVisible"`
	CurrentFounderMemberID   *uint64    `json:"currentFounderMemberId,omitempty"`
	GraphVersion             int64      `json:"graphVersion"`
	Role                     string     `json:"role"`
}

type PublicFamily struct {
	ID                   uint64  `json:"id"`
	FamilyName           string  `json:"familyName"`
	FamilySurname        string  `json:"familySurname"`
	NativePlace          *string `json:"nativePlace,omitempty"`
	RegionText           *string `json:"regionText,omitempty"`
	Description          *string `json:"description,omitempty"`
	AvatarURL            *string `json:"avatarUrl,omitempty"`
	PublicContactName    *string `json:"publicContactName,omitempty"`
	PublicContactPhone   *string `json:"publicContactPhone,omitempty"`
	PublicContactWechat  *string `json:"publicContactWechat,omitempty"`
	PublicContactNote    *string `json:"publicContactNote,omitempty"`
	PublicContactVisible bool    `json:"publicContactVisible"`
}

type PublicFamilyListItem struct {
	ID                   uint64     `json:"id"`
	FamilyName           string     `json:"familyName"`
	FamilySurname        string     `json:"familySurname"`
	NativePlace          *string    `json:"nativePlace,omitempty"`
	RegionText           *string    `json:"regionText,omitempty"`
	Description          *string    `json:"description,omitempty"`
	AvatarURL            *string    `json:"avatarUrl,omitempty"`
	PublicContactVisible bool       `json:"publicContactVisible"`
	PublicContactName    *string    `json:"publicContactName,omitempty"`
	PublicContactNote    *string    `json:"publicContactNote,omitempty"`
	PublicApprovedAt     *time.Time `json:"publicApprovedAt,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type PublicFamilyShowcaseItem struct {
	ID               uint64     `json:"id"`
	FamilyName       string     `json:"familyName"`
	FamilySurname    string     `json:"familySurname"`
	NativePlace      *string    `json:"nativePlace,omitempty"`
	RegionText       *string    `json:"regionText,omitempty"`
	Description      *string    `json:"description,omitempty"`
	AvatarURL        *string    `json:"avatarUrl,omitempty"`
	PublicApprovedAt *time.Time `json:"publicApprovedAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type OperationLogItem struct {
	ID              uint64    `json:"id"`
	OperatorType    string    `json:"operatorType"`
	OperatorAdminID *uint64   `json:"operatorAdminId,omitempty"`
	OperatorUserID  *uint64   `json:"operatorUserId,omitempty"`
	Module          string    `json:"module"`
	Action          string    `json:"action"`
	TargetType      *string   `json:"targetType,omitempty"`
	TargetID        *uint64   `json:"targetId,omitempty"`
	MemberID        *uint64   `json:"memberId,omitempty"`
	UserID          *uint64   `json:"userId,omitempty"`
	Result          string    `json:"result"`
	ErrorMessage    *string   `json:"errorMessage,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
}

type ListOperationLogsResult struct {
	Items    []OperationLogItem `json:"items"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
	Total    int64              `json:"total"`
}

type ListPublicFamilyShowcaseResult struct {
	Items    []PublicFamilyShowcaseItem `json:"items"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"pageSize"`
	Total    int64                      `json:"total"`
}

type ListPublicFamiliesResult struct {
	Items    []PublicFamilyListItem `json:"items"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
	Total    int64                  `json:"total"`
}

type DissolutionRequest struct {
	ID                uint64     `json:"id"`
	FamilyID          uint64     `json:"familyId"`
	RequesterMemberID uint64     `json:"requesterMemberId"`
	RequesterUserID   uint64     `json:"requesterUserId"`
	RequestStatus     string     `json:"requestStatus"`
	RequestReason     *string    `json:"requestReason,omitempty"`
	CancelledAt       *time.Time `json:"cancelledAt,omitempty"`
	CancelReason      *string    `json:"cancelReason,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
}

type LeaveFamilyResult struct {
	FamilyID uint64 `json:"familyId"`
	MemberID uint64 `json:"memberId"`
	Status   string `json:"status"`
}
