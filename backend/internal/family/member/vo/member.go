package vo

import "time"

type Member struct {
	MemberID                  uint64     `json:"memberId"`
	FamilyID                  uint64     `json:"familyId"`
	Name                      string     `json:"name"`
	Gender                    string     `json:"gender"`
	PortraitHairStyle         string     `json:"portraitHairStyle"`
	PortraitGlassesStyle      string     `json:"portraitGlassesStyle"`
	PortraitAvatarStyle       string     `json:"portraitAvatarStyle"`
	BirthDate                 *string    `json:"birthDate,omitempty"`
	BirthYear                 *int       `json:"birthYear,omitempty"`
	DeathDate                 *string    `json:"deathDate,omitempty"`
	DeathYear                 *int       `json:"deathYear,omitempty"`
	IsAlive                   *bool      `json:"isAlive,omitempty"`
	AvatarURL                 *string    `json:"avatarUrl,omitempty"`
	Description               *string    `json:"description,omitempty"`
	MemorialVisible           bool       `json:"memorialVisible"`
	Status                    string     `json:"status"`
	Revision                  int64      `json:"revision"`
	ProfileSourceType         string     `json:"profileSourceType"`
	ProfileVerificationStatus string     `json:"profileVerificationStatus"`
	ProfileConfirmedAt        *time.Time `json:"profileConfirmedAt,omitempty"`
	ProfileConfirmedByUserID  *uint64    `json:"profileConfirmedByUserId,omitempty"`
	UserBindingPolicy         string     `json:"userBindingPolicy"`
	BoundUserID               *uint64    `json:"boundUserId,omitempty"`
	BoundFamilyRole           *string    `json:"boundFamilyRole,omitempty"`
	CreatedAt                 time.Time  `json:"createdAt"`
	UpdatedAt                 time.Time  `json:"updatedAt"`
}

type RecoverableMember struct {
	MemberID         uint64    `json:"memberId"`
	Name             string    `json:"name"`
	DeletedAt        time.Time `json:"deletedAt"`
	RecoverableUntil time.Time `json:"recoverableUntil"`
	DeleteReason     *string   `json:"deleteReason,omitempty"`
}
