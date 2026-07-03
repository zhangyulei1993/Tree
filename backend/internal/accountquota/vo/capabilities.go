package vo

type Limits struct {
	MaxOwnedFamilies         int `json:"maxOwnedFamilies"`
	MaxMembersPerOwnedFamily int `json:"maxMembersPerOwnedFamily"`
	MaxJoinedFamilies        int `json:"maxJoinedFamilies"`
}

type Usage struct {
	OwnedFamilies         int            `json:"ownedFamilies"`
	JoinedFamilies        int            `json:"joinedFamilies"`
	MembersPerOwnedFamily map[string]int `json:"membersPerOwnedFamily"`
}

type Capabilities struct {
	TrustTier string `json:"trustTier"`
	Limits    Limits `json:"limits"`
	Usage     Usage  `json:"usage"`
}

type ConfigItem struct {
	TrustTier                string  `json:"trustTier"`
	MaxOwnedFamilies         int     `json:"maxOwnedFamilies"`
	MaxMembersPerOwnedFamily int     `json:"maxMembersPerOwnedFamily"`
	MaxJoinedFamilies        int     `json:"maxJoinedFamilies"`
	UpdatedByAdminID         *uint64 `json:"updatedByAdminId,omitempty"`
	UpdatedAt                string  `json:"updatedAt"`
}

type ImpactPreview struct {
	TrustTier                string `json:"trustTier"`
	AffectedUsers            int    `json:"affectedUsers"`
	AffectedFamilies         int    `json:"affectedFamilies"`
	MaxOwnedFamilies         int    `json:"maxOwnedFamilies"`
	MaxMembersPerOwnedFamily int    `json:"maxMembersPerOwnedFamily"`
	MaxJoinedFamilies        int    `json:"maxJoinedFamilies"`
}
