package enum

const (
	ArticleStatusDraft     = "DRAFT"
	ArticleStatusPublished = "PUBLISHED"
	ArticleStatusArchived  = "ARCHIVED"
)

func ValidArticleStatus(status string) bool {
	switch status {
	case ArticleStatusDraft, ArticleStatusPublished, ArticleStatusArchived:
		return true
	default:
		return false
	}
}
