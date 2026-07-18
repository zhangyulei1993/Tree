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
	TrustTier         string `json:"trustTier"`
	GrayAccessEnabled bool   `json:"grayAccessEnabled"`
	Limits            Limits `json:"limits"`
	Usage             Usage  `json:"usage"`
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

type FeatureOverrideItem struct {
	ID         uint64 `json:"id"`
	FeatureKey string `json:"featureKey"`
	PhoneMask  string `json:"phoneMask"`
	UpdatedAt  string `json:"updatedAt"`
}
