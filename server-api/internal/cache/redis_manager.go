package cache

import (
	"errors"
	"sync"

	pkgcache "megin/pkg/cache"

	"github.com/go-redis/redis/v8"
)

type RedisCacheManager struct {
	DefaultStore *pkgcache.RedisStore
}

var (
	defaultRedisCacheManager     *RedisCacheManager
	defaultRedisCacheManagerOnce sync.Once
)

func InitDefaultRedisCacheManager(client *redis.Client) *RedisCacheManager {
	defaultRedisCacheManagerOnce.Do(func() {
		defaultRedisCacheManager = &RedisCacheManager{
			DefaultStore: pkgcache.NewRedisStore(client),
		}
	})
	return defaultRedisCacheManager
}

func DefaultRedisCacheManager() (*RedisCacheManager, error) {
	if defaultRedisCacheManager == nil {
		return nil, errors.New("default redis cache manager is not initialized")
	}
	return defaultRedisCacheManager, nil
}

func Redis() *pkgcache.RedisStore {
	manager, err := DefaultRedisCacheManager()
	if err != nil {
		panic(err)
	}
	return manager.DefaultStore
}

func RedisManager() *RedisCacheManager {
	manager, err := DefaultRedisCacheManager()
	if err != nil {
		panic(err)
	}
	return manager
}
