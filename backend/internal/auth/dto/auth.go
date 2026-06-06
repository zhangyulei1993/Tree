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

type WechatMiniLoginRequest struct {
	Code       string `json:"code" binding:"required"`
	ClientType string `json:"clientType" binding:"required"`
}

type BindPhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type ChangePhoneRequest struct {
	OldPhoneCode string `json:"oldPhoneCode" binding:"required"`
	NewPhone     string `json:"newPhone" binding:"required"`
	NewPhoneCode string `json:"newPhoneCode" binding:"required"`
}

type CancelAccountRequest struct {
	PhoneCode    string `json:"phoneCode" binding:"required"`
	CancelReason string `json:"cancelReason"`
}
