package dto

type CreatePublicApplicationRequest struct {
	ApplicationReason *string `json:"applicationReason"`
}

type ReviewPublicApplicationRequest struct {
	ReviewComment *string `json:"reviewComment"`
}

type CancelPublicApplicationRequest struct {
	CancelReason *string `json:"cancelReason"`
}

type TakeDownPublicFamilyRequest struct {
	Reason *string `json:"reason"`
}

type ListApplicationsQuery struct {
	Status   string
	Page     int
	PageSize int
}
