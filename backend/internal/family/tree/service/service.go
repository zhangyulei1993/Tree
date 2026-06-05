package service

import "context"

type GraphVersionService interface {
	Increment(ctx context.Context, familyID uint64) error
	Get(ctx context.Context, familyID uint64) (int64, error)
	BuildTreeCacheKey(familyID uint64, graphVersion int64) string
}
