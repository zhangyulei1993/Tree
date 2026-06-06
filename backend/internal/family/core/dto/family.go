package dto

type PublicContact struct {
	Name    *string `json:"name"`
	Phone   *string `json:"phone"`
	Wechat  *string `json:"wechat"`
	Note    *string `json:"note"`
	Visible *bool   `json:"visible"`
}

type CreateFamilyRequest struct {
	Surname       string         `json:"surname" binding:"required"`
	FamilyName    *string        `json:"familyName"`
	Name          *string        `json:"name"`
	NativePlace   *string        `json:"nativePlace"`
	RegionCode    *string        `json:"regionCode"`
	RegionText    *string        `json:"regionText"`
	Description   *string        `json:"description"`
	AvatarURL     *string        `json:"avatarUrl"`
	PublicContact *PublicContact `json:"publicContact"`
}

type UpdateFamilyRequest struct {
	FamilyName           *string `json:"familyName"`
	NativePlace          *string `json:"nativePlace"`
	RegionCode           *string `json:"regionCode"`
	RegionText           *string `json:"regionText"`
	Description          *string `json:"description"`
	AvatarURL            *string `json:"avatarUrl"`
	Searchable           *bool   `json:"searchable"`
	PublicContactName    *string `json:"publicContactName"`
	PublicContactPhone   *string `json:"publicContactPhone"`
	PublicContactWechat  *string `json:"publicContactWechat"`
	PublicContactNote    *string `json:"publicContactNote"`
	PublicContactVisible *bool   `json:"publicContactVisible"`
}

type CreateDissolutionRequest struct {
	RequestReason *string `json:"requestReason"`
}

type CancelDissolutionRequest struct {
	CancelReason *string `json:"cancelReason"`
}
