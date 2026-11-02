package dto

type CreateMemberRequest struct {
	Name                 string  `json:"name" binding:"required"`
	Gender               *string `json:"gender"`
	PortraitHairStyle    *string `json:"portraitHairStyle"`
	PortraitGlassesStyle *string `json:"portraitGlassesStyle"`
	PortraitAvatarStyle  *string `json:"portraitAvatarStyle"`
	BirthDate            *string `json:"birthDate"`
	BirthYear            *int    `json:"birthYear"`
	DeathDate            *string `json:"deathDate"`
	DeathYear            *int    `json:"deathYear"`
	IsAlive              *bool   `json:"isAlive"`
	AvatarURL            *string `json:"avatarUrl"`
	Description          *string `json:"description"`
	MemorialVisible      *bool   `json:"memorialVisible"`
	UserBindingPolicy    *string `json:"userBindingPolicy"`
}

type UpdateMemberRequest struct {
	ExpectedRevision     *int64  `json:"expectedRevision"`
	Name                 *string `json:"name"`
	Gender               *string `json:"gender"`
	PortraitHairStyle    *string `json:"portraitHairStyle"`
	PortraitGlassesStyle *string `json:"portraitGlassesStyle"`
	PortraitAvatarStyle  *string `json:"portraitAvatarStyle"`
	BirthDate            *string `json:"birthDate"`
	BirthYear            *int    `json:"birthYear"`
	DeathDate            *string `json:"deathDate"`
	DeathYear            *int    `json:"deathYear"`
	IsAlive              *bool   `json:"isAlive"`
	AvatarURL            *string `json:"avatarUrl"`
	Description          *string `json:"description"`
	MemorialVisible      *bool   `json:"memorialVisible"`
	UserBindingPolicy    *string `json:"userBindingPolicy"`
}

type ConfirmMemberProfileRequest struct {
	ExpectedRevision *int64 `json:"expectedRevision"`
	Status           string `json:"status" binding:"required"`
}

type BindUserRequest struct {
	UserID uint64 `json:"userId" binding:"required"`
}

type UnbindUserRequest struct {
	Reason *string `json:"reason"`
}

type DeleteMemberRequest struct {
	Reason *string `json:"reason"`
}
