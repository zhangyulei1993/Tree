package vo

import "time"

type SendCodeResponse struct {
	ExpireSeconds   int     `json:"expireSeconds"`
	CooldownSeconds int     `json:"cooldownSeconds"`
	DevCode         *string `json:"devCode,omitempty"`
}

type UserInfo struct {
	ID                uint64  `json:"id"`
	Phone             *string `json:"phone,omitempty"`
	PhoneVerified     bool    `json:"phoneVerified"`
	PhoneLoginEnabled bool    `json:"phoneLoginEnabled"`
	Nickname          *string `json:"nickname,omitempty"`
	AvatarURL         *string `json:"avatarUrl,omitempty"`
	Status            string  `json:"status"`
	PasswordSet       bool    `json:"passwordSet"`
}

type LoginResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	User        UserInfo `json:"user"`
}

type AssistanceCodeGenerated struct {
	Code             string    `json:"code"`
	ExpiresAt        time.Time `json:"expiresAt"`
	ValidDays        int       `json:"validDays"`
	ReplacedPrevious bool      `json:"replacedPrevious"`
}
type AssistanceCodeStatus struct {
	ID         uint64     `json:"id"`
	Status     string     `json:"status"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
}
