package dto

type ReviewDissolutionRequest struct {
	ReviewComment *string `json:"reviewComment"`
}

type RestoreFamilyRequest struct {
	Searchable    *bool   `json:"searchable"`
	RestoreReason *string `json:"restoreReason"`
	RestoreStatus *string `json:"restoreStatus"`
	PublicStatus  *string `json:"publicDisplayStatus"`
}

type ListDissolutionQuery struct {
	Status   string
	FamilyID uint64
	Page     int
	PageSize int
}
