package enum

const (
	TrustTierWechatOnly = "WECHAT_ONLY"
	TrustTierPhoneBound = "PHONE_BOUND"
)

var AllTiers = []string{TrustTierWechatOnly, TrustTierPhoneBound}

func IsValidTier(tier string) bool {
	switch tier {
	case TrustTierWechatOnly, TrustTierPhoneBound:
		return true
	default:
		return false
	}
}
