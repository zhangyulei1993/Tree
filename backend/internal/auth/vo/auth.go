package vo

type SendCodeResponse struct {
	ExpireSeconds   int     `json:"expireSeconds"`
	CooldownSeconds int     `json:"cooldownSeconds"`
	DevCode         *string `json:"devCode,omitempty"`
}

type UserInfo struct {
	ID            uint64  `json:"id"`
	Phone         *string `json:"phone,omitempty"`
	PhoneVerified bool    `json:"phoneVerified"`
	Nickname      *string `json:"nickname,omitempty"`
	Status        string  `json:"status"`
}

type LoginResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	User        UserInfo `json:"user"`
}
