package service

import (
	"context"
	"fmt"
)

type GraphVersionService interface {
	Increment(ctx context.Context, familyID uint64) error
	Get(ctx context.Context, familyID uint64) (int64, error)
	BuildTreeCacheKey(familyID uint64, graphVersion int64) string
}

func BuildTreeCacheKey(familyID uint64, graphVersion int64) string {
	return fmt.Sprintf("family:%d:tree:v%d", familyID, graphVersion)
}
