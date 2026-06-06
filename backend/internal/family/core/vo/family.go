package vo

import "time"

type FamilySummary struct {
	ID                  uint64  `json:"id"`
	FamilyName          string  `json:"familyName"`
	FamilySurname       string  `json:"familySurname"`
	NativePlace         *string `json:"nativePlace,omitempty"`
	RegionText          *string `json:"regionText,omitempty"`
	AvatarURL           *string `json:"avatarUrl,omitempty"`
	Status              string  `json:"status"`
	PublicDisplayStatus string  `json:"publicDisplayStatus"`
	Role                string  `json:"role"`
}

type FamilyDetail struct {
	ID                     uint64  `json:"id"`
	FamilyName             string  `json:"familyName"`
	FamilySurname          string  `json:"familySurname"`
	NativePlace            *string `json:"nativePlace,omitempty"`
	RegionCode             *string `json:"regionCode,omitempty"`
	RegionText             *string `json:"regionText,omitempty"`
	Description            *string `json:"description,omitempty"`
	AvatarURL              *string `json:"avatarUrl,omitempty"`
	Status                 string  `json:"status"`
	Searchable             bool    `json:"searchable"`
	PublicDisplayStatus    string  `json:"publicDisplayStatus"`
	PublicContactName      *string `json:"publicContactName,omitempty"`
	PublicContactPhone     *string `json:"publicContactPhone,omitempty"`
	PublicContactWechat    *string `json:"publicContactWechat,omitempty"`
	PublicContactNote      *string `json:"publicContactNote,omitempty"`
	PublicContactVisible   bool    `json:"publicContactVisible"`
	CurrentFounderMemberID *uint64 `json:"currentFounderMemberId,omitempty"`
	GraphVersion           int64   `json:"graphVersion"`
	Role                   string  `json:"role"`
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
