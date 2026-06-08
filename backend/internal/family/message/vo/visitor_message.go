package vo

import "time"

type VisitorMessage struct {
	MessageID         uint64     `json:"messageId"`
	FamilyID          uint64     `json:"familyId"`
	FamilyName        string     `json:"familyName,omitempty"`
	VisitorName       *string    `json:"visitorName,omitempty"`
	VisitorPhone      *string    `json:"visitorPhone,omitempty"`
	VisitorWechat     *string    `json:"visitorWechat,omitempty"`
	MessageContent    string     `json:"messageContent"`
	Status            string     `json:"status"`
	ReviewedByAdminID *uint64    `json:"reviewedByAdminId,omitempty"`
	ReviewedAt        *time.Time `json:"reviewedAt,omitempty"`
	ReviewComment     *string    `json:"reviewComment,omitempty"`
	DeletedAt         *time.Time `json:"deletedAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type PublicVisitorMessage struct {
	MessageID      uint64     `json:"messageId"`
	FamilyID       uint64     `json:"familyId"`
	VisitorName    *string    `json:"visitorName,omitempty"`
	MessageContent string     `json:"messageContent"`
	CreatedAt      time.Time  `json:"createdAt"`
	ReviewedAt     *time.Time `json:"reviewedAt,omitempty"`
}

type ListResult struct {
	Items    []VisitorMessage `json:"items"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
	Total    int64            `json:"total"`
}

type PublicListResult struct {
	Items    []PublicVisitorMessage `json:"items"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
	Total    int64                  `json:"total"`
}
