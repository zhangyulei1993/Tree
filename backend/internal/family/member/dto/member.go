package dto

type CreateMemberRequest struct {
	Name              string  `json:"name" binding:"required"`
	Gender            *string `json:"gender"`
	BirthDate         *string `json:"birthDate"`
	BirthYear         *int    `json:"birthYear"`
	DeathDate         *string `json:"deathDate"`
	DeathYear         *int    `json:"deathYear"`
	IsAlive           *bool   `json:"isAlive"`
	AvatarURL         *string `json:"avatarUrl"`
	Description       *string `json:"description"`
	UserBindingPolicy *string `json:"userBindingPolicy"`
}

type UpdateMemberRequest struct {
	Name              *string `json:"name"`
	Gender            *string `json:"gender"`
	BirthDate         *string `json:"birthDate"`
	BirthYear         *int    `json:"birthYear"`
	DeathDate         *string `json:"deathDate"`
	DeathYear         *int    `json:"deathYear"`
	IsAlive           *bool   `json:"isAlive"`
	AvatarURL         *string `json:"avatarUrl"`
	Description       *string `json:"description"`
	UserBindingPolicy *string `json:"userBindingPolicy"`
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
