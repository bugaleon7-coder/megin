package cache

import (
	"fmt"
	"sync"

	pkgcache "megin/pkg/cache"
)

const (
	DefaultSize = 512 * 1024
	MemSizeMB   = 1024 * 1024

	MemSizeMessageReaded = 10 * MemSizeMB
)

const (
	CacheKeyFormatMessageReaded = "MessageReaded_%s_%s_%s"
)

type LocalCacheManager struct {
	DefaultStore         *pkgcache.LocalStore
	MessageReadedStore   *pkgcache.LocalStore
	PayTypeConfigStore   *pkgcache.LocalStore
	ChatPriceConfigStore *pkgcache.LocalStore
	GiftConfigStore      *pkgcache.LocalStore
	ClosenessConfigStore *pkgcache.LocalStore
	AppConfigStore       *pkgcache.LocalStore
}

var (
	localCacheManager     *LocalCacheManager
	localCacheManagerOnce sync.Once
)

func GetLocalCacheManager() *LocalCacheManager {
	if localCacheManager == nil {
		localCacheManagerOnce.Do(func() {
			localCacheManager = &LocalCacheManager{
				DefaultStore:         pkgcache.NewLocalStore(MemSizeMB),
				MessageReadedStore:   pkgcache.NewLocalStore(MemSizeMessageReaded),
				PayTypeConfigStore:   pkgcache.NewLocalStore(DefaultSize),
				ChatPriceConfigStore: pkgcache.NewLocalStore(MemSizeMB),
				GiftConfigStore:      pkgcache.NewLocalStore(DefaultSize),
				ClosenessConfigStore: pkgcache.NewLocalStore(MemSizeMB),
				AppConfigStore:       pkgcache.NewLocalStore(MemSizeMB),
			}
		})
	}
	return localCacheManager
}

func Local() *pkgcache.LocalStore {
	return GetLocalCacheManager().DefaultStore
}

func LocalManager() *LocalCacheManager {
	return GetLocalCacheManager()
}

func (m *LocalCacheManager) GetMessageReadedStore() *pkgcache.LocalStore {
	return m.MessageReadedStore
}

func (m *LocalCacheManager) GetMessageReadedKey(loginUid, channelId, messageId string) string {
	return fmt.Sprintf(CacheKeyFormatMessageReaded, loginUid, channelId, messageId)
}
