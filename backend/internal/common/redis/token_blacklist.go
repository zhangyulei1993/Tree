package redis

import (
	"context"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type TokenBlacklist interface {
	Revoke(ctx context.Context, tokenID string, ttl time.Duration) error
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
	RevokeUserTokens(ctx context.Context, userID uint64, revokedAt int64, ttl time.Duration) error
	IsUserTokenRevoked(ctx context.Context, userID uint64, issuedAt int64) (bool, error)
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

func (b *RedisTokenBlacklist) RevokeUserTokens(ctx context.Context, userID uint64, revokedAt int64, ttl time.Duration) error {
	if userID == 0 {
		return nil
	}
	if ttl <= 0 {
		ttl = 30 * 24 * time.Hour
	}
	return b.client.Set(ctx, userRevocationKey(userID), revokedAt, ttl).Err()
}

func (b *RedisTokenBlacklist) IsUserTokenRevoked(ctx context.Context, userID uint64, issuedAt int64) (bool, error) {
	if userID == 0 {
		return false, nil
	}
	revokedAt, err := b.client.Get(ctx, userRevocationKey(userID)).Int64()
	if err == goredis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return issuedAt <= revokedAt, nil
}

type NoopTokenBlacklist struct{}

func (NoopTokenBlacklist) Revoke(context.Context, string, time.Duration) error {
	return nil
}

func (NoopTokenBlacklist) IsRevoked(context.Context, string) (bool, error) {
	return false, nil
}

func (NoopTokenBlacklist) RevokeUserTokens(context.Context, uint64, int64, time.Duration) error {
	return nil
}

func (NoopTokenBlacklist) IsUserTokenRevoked(context.Context, uint64, int64) (bool, error) {
	return false, nil
}

func tokenBlacklistKey(tokenID string) string {
	return "token:blacklist:" + tokenID
}

func userRevocationKey(userID uint64) string {
	return "token:user:revoked_at:" + strconv.FormatUint(userID, 10)
}
