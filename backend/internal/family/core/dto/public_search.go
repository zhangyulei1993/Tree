package dto

type ListPublicFamiliesQuery struct {
	Keyword       string
	FamilySurname string
	RegionText    string
	Page          int
	PageSize      int
}
