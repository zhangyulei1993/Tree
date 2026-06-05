package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type TokenBlacklist interface {
	Revoke(ctx context.Context, tokenID string, ttl time.Duration) error
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
}

type RedisTokenBlacklist struct {
	client *goredis.Client
}

func NewTokenBlacklist(client *goredis.Client) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{client: client}
}

func (b *RedisTokenBlacklist) Revoke(ctx context.Context, tokenID string, ttl time.Duration) error {
	if tokenID == "" {
		return nil
	}
	if ttl <= 0 {
		ttl = time.Hour
	}
	return b.client.Set(ctx, tokenBlacklistKey(tokenID), "1", ttl).Err()
}

func (b *RedisTokenBlacklist) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	if tokenID == "" {
		return false, nil
	}
	result, err := b.client.Exists(ctx, tokenBlacklistKey(tokenID)).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

type NoopTokenBlacklist struct{}

func (NoopTokenBlacklist) Revoke(context.Context, string, time.Duration) error {
	return nil
}

func (NoopTokenBlacklist) IsRevoked(context.Context, string) (bool, error) {
	return false, nil
}

func tokenBlacklistKey(tokenID string) string {
	return "token:blacklist:" + tokenID
}
