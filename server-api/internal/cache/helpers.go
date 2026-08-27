package cache

import (
	"time"

	pkgcache "megin/pkg/cache"
)

func SetLocalStoreStruct[T any](store *pkgcache.LocalStore, key string, value T, ttl time.Duration) error {
	return pkgcache.SetStruct(store, key, value, ttl)
}

func SetLocalStruct[T any](key string, value T, ttl time.Duration) error {
	return SetLocalStoreStruct(Local(), key, value, ttl)
}

func GetLocalStoreStruct[T any](store *pkgcache.LocalStore, key string) (*T, error) {
	return pkgcache.GetStruct[T](store, key)
}

func GetLocalStruct[T any](key string) (*T, error) {
	return GetLocalStoreStruct[T](Local(), key)
}

func SetLocalStoreStructList[T any](store *pkgcache.LocalStore, key string, values []T, ttl time.Duration) error {
	return pkgcache.SetStructList(store, key, values, ttl)
}

func SetLocalStructList[T any](key string, values []T, ttl time.Duration) error {
	return SetLocalStoreStructList(Local(), key, values, ttl)
}

func GetLocalStoreStructList[T any](store *pkgcache.LocalStore, key string) ([]T, error) {
	return pkgcache.GetStructList[T](store, key)
}

func GetLocalStructList[T any](key string) ([]T, error) {
	return GetLocalStoreStructList[T](Local(), key)
}

func SetLocalStoreString(store *pkgcache.LocalStore, key string, value string, ttl time.Duration) error {
	return pkgcache.SetString(store, key, value, ttl)
}

func SetLocalString(key string, value string, ttl time.Duration) error {
	return SetLocalStoreString(Local(), key, value, ttl)
}

func GetLocalStoreString(store *pkgcache.LocalStore, key string) (string, error) {
	return pkgcache.GetString(store, key)
}

func GetLocalString(key string) (string, error) {
	return GetLocalStoreString(Local(), key)
}

func DeleteLocalStoreKey(store *pkgcache.LocalStore, key string) bool {
	return pkgcache.DeleteKey(store, key)
}

func DeleteLocalKey(key string) bool {
	return DeleteLocalStoreKey(Local(), key)
}

func SetRedisStruct[T any](key string, value T, ttl time.Duration) error {
	return pkgcache.SetStruct(Redis(), key, value, ttl)
}

func GetRedisStruct[T any](key string) (*T, error) {
	return pkgcache.GetStruct[T](Redis(), key)
}

func SetRedisStructList[T any](key string, values []T, ttl time.Duration) error {
	return pkgcache.SetStructList(Redis(), key, values, ttl)
}

func GetRedisStructList[T any](key string) ([]T, error) {
	return pkgcache.GetStructList[T](Redis(), key)
}

func SetRedisString(key string, value string, ttl time.Duration) error {
	return pkgcache.SetString(Redis(), key, value, ttl)
}

func GetRedisString(key string) (string, error) {
	return pkgcache.GetString(Redis(), key)
}

func DeleteRedisKey(key string) bool {
	return pkgcache.DeleteKey(Redis(), key)
}
