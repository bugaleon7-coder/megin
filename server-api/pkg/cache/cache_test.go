package cache

import (
	"context"
	"testing"
	"time"
)

type cachePayload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestBasicStoreReplaceability(t *testing.T) {
	local := NewLocalStore(1024 * 1024)
	var store BasicStore = local

	if err := store.Set(context.Background(), "gift:test", "1", time.Minute); err != nil {
		t.Fatalf("local set: %v", err)
	}
	got, err := store.Get(context.Background(), "gift:test")
	if err != nil {
		t.Fatalf("local get: %v", err)
	}
	if got != "1" {
		t.Fatalf("got %q, want 1", got)
	}
}

func TestLocalStoreMethodHelpers(t *testing.T) {
	local := NewLocalStore(1024 * 1024)
	key := "local:struct"
	payload := cachePayload{Name: "alpha", Age: 20}

	if err := local.SetStruct(key, payload, time.Minute); err != nil {
		t.Fatalf("SetStruct: %v", err)
	}

	var got cachePayload
	if err := local.GetStruct(key, &got); err != nil {
		t.Fatalf("GetStruct: %v", err)
	}
	if got.Name != payload.Name || got.Age != payload.Age {
		t.Fatalf("got = %#v, want %#v", got, payload)
	}
}

func TestCounterWithLocalAndRedisStyleInterfaces(t *testing.T) {
	local := NewLocalStore(1024 * 1024)
	counter := NewCounter(nil, local, "gift")
	counter.store = incrementStoreStub{counts: map[string]int64{}}

	result, err := counter.HitAndCheck(context.Background(), "send", 3, time.Minute)
	if err != nil {
		t.Fatalf("HitAndCheck: %v", err)
	}
	if result.Count != 1 {
		t.Fatalf("count = %d, want 1", result.Count)
	}
}

type incrementStoreStub struct {
	counts map[string]int64
}

func (s incrementStoreStub) IncrBy(_ context.Context, key string, delta int64, _ time.Duration) (int64, error) {
	s.counts[key] += delta
	return s.counts[key], nil
}
