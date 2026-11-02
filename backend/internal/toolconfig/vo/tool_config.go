package vo

import "time"

type ToolConfig struct {
	ID          uint64    `json:"id"`
	ToolKey     string    `json:"toolKey"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description"`
	Visible     bool      `json:"visible"`
	Enabled     bool      `json:"enabled"`
	Pinned      bool      `json:"pinned"`
	Highlighted bool      `json:"highlighted"`
	SortOrder   int       `json:"sortOrder"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
