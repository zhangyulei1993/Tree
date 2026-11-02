package dto

type Tier struct {
	RequiredInvites         int `json:"requiredInvites"`
	CumulativeRewardMembers int `json:"cumulativeRewardMembers"`
}
type Config struct {
	Enabled                  bool   `json:"enabled"`
	ActivationRemainingSlots int    `json:"activationRemainingSlots"`
	MaxRewardMembers         int    `json:"maxRewardMembers"`
	RequirePhoneBound        bool   `json:"requirePhoneBound"`
	Tiers                    []Tier `json:"tiers"`
}
type Summary struct {
	Enabled             bool   `json:"enabled"`
	Active              bool   `json:"active"`
	FamilyID            uint64 `json:"familyId"`
	UsedMembers         int    `json:"usedMembers"`
	EffectiveLimit      int    `json:"effectiveLimit"`
	ValidInvites        int    `json:"validInvites"`
	RewardedMembers     int    `json:"rewardedMembers"`
	MaxRewardMembers    int    `json:"maxRewardMembers"`
	NextRequiredInvites int    `json:"nextRequiredInvites,omitempty"`
	NextAddedMembers    int    `json:"nextAddedMembers,omitempty"`
	Tiers               []Tier `json:"tiers"`
}
type ShareResult struct {
	ReferralCode string  `json:"referralCode"`
	Summary      Summary `json:"summary"`
}
