package vo

import "time"

type RoleChangeResult struct {
	FamilyID     uint64    `json:"familyId"`
	MemberID     uint64    `json:"memberId"`
	UserID       uint64    `json:"userId"`
	FamilyRole   string    `json:"familyRole"`
	GraphVersion int64     `json:"graphVersion"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
