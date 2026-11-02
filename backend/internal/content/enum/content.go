package enum

const (
	ArticleStatusDraft     = "DRAFT"
	ArticleStatusPublished = "PUBLISHED"
	ArticleStatusArchived  = "ARCHIVED"

	ArticleTypeInternal       = "INTERNAL"
	ArticleTypeWechatOfficial = "WECHAT_OFFICIAL"
)

func ValidArticleStatus(status string) bool {
	switch status {
	case ArticleStatusDraft, ArticleStatusPublished, ArticleStatusArchived:
		return true
	default:
		return false
	}
}

func ValidArticleType(articleType string) bool {
	switch articleType {
	case ArticleTypeInternal, ArticleTypeWechatOfficial:
		return true
	default:
		return false
	}
}
