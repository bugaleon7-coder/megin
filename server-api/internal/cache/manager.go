package cache

import (
	"errors"
	"sync"

	"github.com/go-redis/redis/v8"
)

type Manager struct {
	Local *LocalCacheManager
	Redis *RedisCacheManager
}

var (
	defaultManager     *Manager
	defaultManagerOnce sync.Once
)

func InitManager(redisClient *redis.Client) *Manager {
	defaultManagerOnce.Do(func() {
		defaultManager = &Manager{
			Local: GetLocalCacheManager(),
		}
		if redisClient != nil {
			defaultManager.Redis = InitDefaultRedisCacheManager(redisClient)
		}
	})
	return defaultManager
}

func GetManager() (*Manager, error) {
	if defaultManager == nil {
		return nil, errors.New("cache manager is not initialized")
	}
	return defaultManager, nil
}

func MustManager() *Manager {
	manager, err := GetManager()
	if err != nil {
		panic(err)
	}
	return manager
}
