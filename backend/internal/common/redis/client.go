package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"tree/backend/internal/common/config"
)

func NewClient(ctx context.Context, cfg config.RedisConfig) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
