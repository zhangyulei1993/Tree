package vo

import "time"

type ChoiceScenario struct {
	ID        uint64    `json:"id"`
	Title     string    `json:"title"`
	Options   []string  `json:"options"`
	UseCount  uint      `json:"useCount"`
	CreatedAt time.Time `json:"createdAt"`
}
