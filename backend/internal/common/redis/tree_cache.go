package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("tree cache miss")

type TreeCache interface {
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, []byte, time.Duration) error
}

type RedisTreeCache struct {
	client *goredis.Client
}

func NewTreeCache(client *goredis.Client) *RedisTreeCache {
	return &RedisTreeCache{client: client}
}

func (c *RedisTreeCache) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := c.client.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		return nil, ErrCacheMiss
	}
	return value, err
}

func (c *RedisTreeCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

type NoopTreeCache struct{}

func (NoopTreeCache) Get(context.Context, string) ([]byte, error) {
	return nil, ErrCacheMiss
}

func (NoopTreeCache) Set(context.Context, string, []byte, time.Duration) error {
	return nil
}
