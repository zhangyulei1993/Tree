package dto

type CreateVisitorMessageRequest struct {
	VisitorName    *string `json:"visitorName"`
	VisitorPhone   *string `json:"visitorPhone"`
	VisitorWechat  *string `json:"visitorWechat"`
	MessageContent string  `json:"messageContent" binding:"required"`
}

type ReviewVisitorMessageRequest struct {
	ReviewComment *string `json:"reviewComment"`
}

type DeleteVisitorMessageRequest struct {
	DeleteReason *string `json:"deleteReason"`
}

type ListMessagesQuery struct {
	Status   string
	FamilyID uint64
	Page     int
	PageSize int
}
