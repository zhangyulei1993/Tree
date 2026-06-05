package dto

type SendCodeRequest struct {
	Phone      string `json:"phone" binding:"required"`
	Scene      string `json:"scene" binding:"required"`
	ClientType string `json:"clientType" binding:"required"`
}

type RegisterPhoneRequest struct {
	Phone      string  `json:"phone" binding:"required"`
	Code       string  `json:"code" binding:"required"`
	Password   string  `json:"password" binding:"required"`
	Nickname   *string `json:"nickname"`
	ClientType string  `json:"clientType" binding:"required"`
}

type LoginPhoneRequest struct {
	Phone      string `json:"phone" binding:"required"`
	Password   string `json:"password" binding:"required"`
	ClientType string `json:"clientType" binding:"required"`
}
