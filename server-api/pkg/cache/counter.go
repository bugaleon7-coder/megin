package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Counter struct {
	store  IncrementStore
	base   BasicStore
	prefix string
}

type CounterResult struct {
	Key     string
	Count   int64
	Limit   int64
	Window  time.Duration
	Limited bool
}

var (
	defaultCounterMu sync.RWMutex
	defaultCounters  = make(map[string]*Counter)
)

func NewCounter(store IncrementStore, base BasicStore, prefix string) *Counter {
	return &Counter{
		store:  store,
		base:   base,
		prefix: prefix,
	}
}

func InitDefaultCounter(name string, store IncrementStore, base BasicStore, prefix string) *Counter {
	defaultCounterMu.Lock()
	defer defaultCounterMu.Unlock()
	if counter, ok := defaultCounters[name]; ok {
		return counter
	}
	counter := NewCounter(store, base, prefix)
	defaultCounters[name] = counter
	return counter
}

func DefaultCounter(name string) (*Counter, error) {
	defaultCounterMu.RLock()
	defer defaultCounterMu.RUnlock()
	counter, ok := defaultCounters[name]
	if !ok {
		return nil, errors.New("default counter is not initialized")
	}
	return counter, nil
}

func (c *Counter) Increment(ctx context.Context, name string, window time.Duration) (int64, error) {
	return c.store.IncrBy(ctx, c.key(name), 1, window)
}

func (c *Counter) IncrementBy(ctx context.Context, name string, delta int64, window time.Duration) (int64, error) {
	return c.store.IncrBy(ctx, c.key(name), delta, window)
}

func (c *Counter) Get(ctx context.Context, name string) (int64, error) {
	value, err := GetInt64(ctx, c.base, c.key(name))
	if err == ErrCacheMiss {
		return 0, nil
	}
	return value, err
}

func (c *Counter) Reset(ctx context.Context, name string) error {
	return c.base.Delete(ctx, c.key(name))
}

func (c *Counter) HitAndCheck(ctx context.Context, name string, limit int64, window time.Duration) (*CounterResult, error) {
	count, err := c.Increment(ctx, name, window)
	if err != nil {
		return nil, err
	}
	return &CounterResult{
		Key:     c.key(name),
		Count:   count,
		Limit:   limit,
		Window:  window,
		Limited: limit > 0 && count >= limit,
	}, nil
}

func (c *Counter) key(name string) string {
	if c.prefix == "" {
		return name
	}
	return fmt.Sprintf("%s:%s", c.prefix, name)
}
