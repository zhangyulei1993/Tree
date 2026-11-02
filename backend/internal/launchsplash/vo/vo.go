package vo

import "time"

type Config struct {
	ID            uint64    `json:"id"`
	ConfigKey     string    `json:"configKey"`
	Enabled       bool      `json:"enabled"`
	Title         string    `json:"title"`
	Tagline       string    `json:"tagline"`
	ImageURL      string    `json:"imageUrl"`
	DayImageURL   string    `json:"dayImageUrl"`
	NightImageURL string    `json:"nightImageUrl"`
	DurationMs    int       `json:"durationMs"`
	SkipEnabled   bool      `json:"skipEnabled"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
