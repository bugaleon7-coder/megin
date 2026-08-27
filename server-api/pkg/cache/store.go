package cache

import (
	"context"
	"errors"
	"time"
)

var ErrCacheMiss = errors.New("cache miss")

type BasicStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type IncrementStore interface {
	IncrBy(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error)
}

type LockStore interface {
	SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)
}
