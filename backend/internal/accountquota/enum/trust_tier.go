package enum

const (
	TrustTierWechatOnly    = "WECHAT_ONLY"
	TrustTierPhoneVerified = "PHONE_VERIFIED"
)

var AllTiers = []string{TrustTierWechatOnly, TrustTierPhoneVerified}

func IsValidTier(tier string) bool {
	switch tier {
	case TrustTierWechatOnly, TrustTierPhoneVerified:
		return true
	default:
		return false
	}
}
