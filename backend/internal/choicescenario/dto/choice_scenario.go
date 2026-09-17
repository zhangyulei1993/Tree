package dto

type CreateRequest struct {
	Title   string   `json:"title"`
	Options []string `json:"options"`
}

type ListQuery struct {
	Period string
}
