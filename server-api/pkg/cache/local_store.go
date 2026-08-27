package cache

import (
	"context"
	"errors"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type LocalStore struct {
	cache *ttlcache.Cache[string, any]
}

func NewLocalStore(_ int) *LocalStore {
	cache := ttlcache.New[string, any]()
	go cache.Start()
	return &LocalStore{cache: cache}
}

func (s *LocalStore) Raw() *ttlcache.Cache[string, any] {
	return s.cache
}

func (s *LocalStore) Get(_ context.Context, key string) (string, error) {
	item := s.cache.Get(key)
	if item == nil {
		return "", ErrCacheMiss
	}
	value, ok := item.Value().(string)
	if !ok {
		return "", errors.New("cache value is not string")
	}
	return value, nil
}

func (s *LocalStore) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	s.cache.Set(key, value, ttl)
	return nil
}

func (s *LocalStore) Delete(_ context.Context, key string) error {
	s.cache.Delete(key)
	return nil
}

func (s *LocalStore) SetStruct(key string, value any, ttl time.Duration) error {
	return s.setValue(key, value, ttl)
}

func (s *LocalStore) SetStructList(key string, values any, ttl time.Duration) error {
	return s.setValue(key, values, ttl)
}

func (s *LocalStore) SetString(key string, value string, ttl time.Duration) error {
	return SetString(s, key, value, ttl)
}

func (s *LocalStore) GetStruct(key string, target any) error {
	value, err := s.getValue(key)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			return nil
		}
		return err
	}
	return assignStoredValue(target, value)
}

func (s *LocalStore) GetString(key string) (string, error) {
	return GetString(s, key)
}

func (s *LocalStore) DeleteKey(key string) bool {
	return DeleteKey(s, key)
}

func (s *LocalStore) Stop() {
	s.cache.Stop()
}

func (s *LocalStore) setValue(key string, value any, ttl time.Duration) error {
	s.cache.Set(key, value, ttl)
	return nil
}

func (s *LocalStore) getValue(key string) (any, error) {
	item := s.cache.Get(key)
	if item == nil {
		return nil, ErrCacheMiss
	}
	return item.Value(), nil
}
