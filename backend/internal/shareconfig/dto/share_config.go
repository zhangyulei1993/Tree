package dto

type UpdateRequest struct {
	TitleTemplate string   `json:"titleTemplate"`
	ImageURL      string   `json:"imageUrl"`
	ImageURLs     []string `json:"imageUrls"`
	Description   *string  `json:"description"`
}
