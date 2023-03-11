package redis

import (
	"github.com/go-redis/redis/v8"
)

type Option func(*redis.Options)

// func WithAccessTokenTTL(t time.Duration) Option {
// 	return func(opts *redis.Options) {
// 		rc.accessTokenTTL = t
// 	}
// }

// func WithRefreshTokenTTL(t time.Duration) Option {
// 	return func(rc *redisClient) {
// 		rc.refreshTokenTTL = t
// 	}
// }

func WithMaxRetries(n int) Option {
	return func(o *redis.Options) {
		o.MaxRetries = n
	}
}

// func WithReadTimeout(t time.Duration) Option {
// 	return func(o *redis.Options) {
// 		o.ReadTimeout = t
// 	}
// }

// func WithWriteTimeout(t time.Duration) Option {
// 	return func(o *redis.Options) {
// 		o.WriteTimeout = t
// 	}
// }

func WithPoolSize(n int) Option {
	return func(o *redis.Options) {
		o.PoolSize = n
	}
}

func WithDatabaseName(n int) Option {
	return func(o *redis.Options) {
		o.DB = 1
	}
}
