package dto

type UpdateRequest struct {
	Enabled       *bool  `json:"enabled"`
	Title         string `json:"title" binding:"required"`
	Tagline       string `json:"tagline" binding:"required"`
	ImageURL      string `json:"imageUrl"`
	DayImageURL   string `json:"dayImageUrl"`
	NightImageURL string `json:"nightImageUrl"`
	DurationMs    int    `json:"durationMs" binding:"required"`
	SkipEnabled   *bool  `json:"skipEnabled"`
}
