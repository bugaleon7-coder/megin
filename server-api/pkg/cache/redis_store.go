package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

func (s *RedisStore) Raw() *redis.Client {
	return s.client
}

func (s *RedisStore) Get(ctx context.Context, key string) (string, error) {
	value, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	return value, err
}

func (s *RedisStore) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

func (s *RedisStore) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s *RedisStore) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, value, ttl).Result()
}

func (s *RedisStore) IncrBy(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error) {
	pipe := s.client.TxPipeline()
	incr := pipe.IncrBy(ctx, key, delta)
	if ttl > 0 {
		pipe.Expire(ctx, key, ttl)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

func (s *RedisStore) SetStruct(key string, value any, ttl time.Duration) error {
	return SetStruct(s, key, value, ttl)
}

func (s *RedisStore) SetStructList(key string, values any, ttl time.Duration) error {
	switch typed := values.(type) {
	case []string:
		return SetStructList(s, key, typed, ttl)
	default:
		return SetStruct(s, key, values, ttl)
	}
}

func (s *RedisStore) SetString(key string, value string, ttl time.Duration) error {
	return SetString(s, key, value, ttl)
}

func (s *RedisStore) GetStruct(key string, target any) error {
	value, err := s.Get(context.Background(), key)
	if err != nil {
		if err == ErrCacheMiss {
			return nil
		}
		return err
	}
	return decodeInto(value, target)
}

func (s *RedisStore) GetString(key string) (string, error) {
	return GetString(s, key)
}

func (s *RedisStore) DeleteKey(key string) bool {
	return DeleteKey(s, key)
}
