package dto

type CreateRequest struct {
	ToolKey     string `json:"toolKey" binding:"required"`
	DisplayName string `json:"displayName" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type UpdateRequest struct {
	Visible     *bool `json:"visible"`
	Enabled     *bool `json:"enabled"`
	Pinned      *bool `json:"pinned"`
	Highlighted *bool `json:"highlighted"`
	SortOrder   *int  `json:"sortOrder"`
}
