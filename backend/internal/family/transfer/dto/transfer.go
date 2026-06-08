package dto

type CreateTransferRequest struct {
	ToMemberID    uint64  `json:"toMemberId" binding:"required"`
	RequestReason *string `json:"requestReason"`
}

type CancelTransferRequest struct {
	CancelReason *string `json:"cancelReason"`
}

type ReviewTransferRequest struct {
	ReviewComment *string `json:"reviewComment"`
}

type ListTransferQuery struct {
	Status   string
	FamilyID uint64
	Page     int
	PageSize int
}
