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
	Delete(context.Context, string) error
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

func (c *RedisTreeCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

type NoopTreeCache struct{}

func (NoopTreeCache) Get(context.Context, string) ([]byte, error) {
	return nil, ErrCacheMiss
}

func (NoopTreeCache) Set(context.Context, string, []byte, time.Duration) error {
	return nil
}

func (NoopTreeCache) Delete(context.Context, string) error {
	return nil
}
