package vo

import "tree/backend/internal/common/mask"

func MaskPublicFamily(family *PublicFamily) *PublicFamily {
	if family == nil {
		return nil
	}
	masked := *family
	if masked.PublicContactVisible {
		if masked.PublicContactName != nil {
			value := mask.DisplayName(*masked.PublicContactName)
			masked.PublicContactName = &value
		}
		if masked.PublicContactPhone != nil {
			value := mask.Phone(*masked.PublicContactPhone)
			masked.PublicContactPhone = &value
		}
		if masked.PublicContactWechat != nil {
			value := mask.Wechat(*masked.PublicContactWechat)
			masked.PublicContactWechat = &value
		}
	} else {
		masked.PublicContactName = nil
		masked.PublicContactPhone = nil
		masked.PublicContactWechat = nil
	}
	masked.PublicContactNote = nil
	return &masked
}
