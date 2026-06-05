package vo

type AdminInfo struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"displayName"`
	Phone       *string `json:"phone,omitempty"`
	Email       *string `json:"email,omitempty"`
	Role        string  `json:"role"`
	Status      string  `json:"status"`
}

type LoginResponse struct {
	AccessToken string    `json:"accessToken"`
	TokenType   string    `json:"tokenType"`
	Admin       AdminInfo `json:"admin"`
}
