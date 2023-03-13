package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type RedisClient interface {
	redis.UniversalClient
}

func New(address string, password string, opts ...Option) (RedisClient, error) {
	ctx := context.Background()
	redisOpts := &redis.Options{
		Addr:     address,
		Password: password,
	}
	for _, opt := range opts {
		opt(redisOpts)
	}

	redisDB := redis.NewClient(redisOpts)
	err := redisDB.Ping(ctx).Err()
	if err != nil {
		redisDB.Close()
		err = fmt.Errorf("redis.New: %w", err)
		return nil, err
	}

	return redisDB, nil
}
