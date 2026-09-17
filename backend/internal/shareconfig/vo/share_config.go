package vo

import "time"

type ShareConfig struct {
	ID            uint64    `json:"id"`
	ConfigKey     string    `json:"configKey"`
	TitleTemplate string    `json:"titleTemplate"`
	ImageURL      string    `json:"imageUrl"`
	ImageURLs     []string  `json:"imageUrls"`
	Description   *string   `json:"description,omitempty"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
