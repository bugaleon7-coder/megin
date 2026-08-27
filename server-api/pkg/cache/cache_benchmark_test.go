package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func BenchmarkLocalStoreSetStruct(b *testing.B) {
	store := NewLocalStore(4 * 1024 * 1024)
	b.Cleanup(store.Stop)
	payload := cachePayload{Name: "alpha", Age: 20}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench:local:set:%d", i)
		if err := store.SetStruct(key, payload, time.Minute); err != nil {
			b.Fatalf("SetStruct: %v", err)
		}
	}
}

func BenchmarkLocalStoreGetStruct(b *testing.B) {
	store := NewLocalStore(4 * 1024 * 1024)
	b.Cleanup(store.Stop)
	payload := cachePayload{Name: "alpha", Age: 20}
	key := "bench:local:get"
	if err := store.SetStruct(key, payload, time.Minute); err != nil {
		b.Fatalf("SetStruct: %v", err)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var got cachePayload
		if err := store.GetStruct(key, &got); err != nil {
			b.Fatalf("GetStruct: %v", err)
		}
	}
}

func BenchmarkLocalStoreGetStructGeneric(b *testing.B) {
	store := NewLocalStore(4 * 1024 * 1024)
	b.Cleanup(store.Stop)
	payload := cachePayload{Name: "alpha", Age: 20}
	key := "bench:local:get:generic"
	if err := SetStruct(store, key, payload, time.Minute); err != nil {
		b.Fatalf("SetStruct: %v", err)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		got, err := GetStruct[cachePayload](store, key)
		if err != nil {
			b.Fatalf("GetStruct generic: %v", err)
		}
		if got == nil {
			b.Fatal("GetStruct generic returned nil")
		}
	}
}

func BenchmarkLocalStoreSetString(b *testing.B) {
	store := NewLocalStore(4 * 1024 * 1024)
	b.Cleanup(store.Stop)
	value := `{"name":"alpha","age":20}`

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench:local:set:string:%d", i)
		if err := store.SetString(key, value, time.Minute); err != nil {
			b.Fatalf("SetString: %v", err)
		}
	}
}

func BenchmarkLocalStoreGetString(b *testing.B) {
	store := NewLocalStore(4 * 1024 * 1024)
	b.Cleanup(store.Stop)
	key := "bench:local:get:string"
	value := `{"name":"alpha","age":20}`
	if err := store.SetString(key, value, time.Minute); err != nil {
		b.Fatalf("SetString: %v", err)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := store.GetString(key); err != nil {
			b.Fatalf("GetString: %v", err)
		}
	}
}

func BenchmarkRedisStoreSetStruct(b *testing.B) {
	store := newBenchmarkRedisStore(b)
	payload := cachePayload{Name: "alpha", Age: 20}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench:redis:set:%d", i)
		if err := store.SetStruct(key, payload, time.Minute); err != nil {
			b.Fatalf("SetStruct: %v", err)
		}
	}
}

func BenchmarkRedisStoreGetStruct(b *testing.B) {
	store := newBenchmarkRedisStore(b)
	payload := cachePayload{Name: "alpha", Age: 20}
	key := "bench:redis:get"
	if err := store.SetStruct(key, payload, time.Minute); err != nil {
		b.Fatalf("SetStruct: %v", err)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var got cachePayload
		if err := store.GetStruct(key, &got); err != nil {
			b.Fatalf("GetStruct: %v", err)
		}
	}
}

func BenchmarkRedisStoreGetStructGeneric(b *testing.B) {
	store := newBenchmarkRedisStore(b)
	payload := cachePayload{Name: "alpha", Age: 20}
	key := "bench:redis:get:generic"
	if err := SetStruct(store, key, payload, time.Minute); err != nil {
		b.Fatalf("SetStruct: %v", err)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		got, err := GetStruct[cachePayload](store, key)
		if err != nil {
			b.Fatalf("GetStruct generic: %v", err)
		}
		if got == nil {
			b.Fatal("GetStruct generic returned nil")
		}
	}
}

func BenchmarkRedisStoreSetString(b *testing.B) {
	store := newBenchmarkRedisStore(b)
	value := `{"name":"alpha","age":20}`

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench:redis:set:string:%d", i)
		if err := store.SetString(key, value, time.Minute); err != nil {
			b.Fatalf("SetString: %v", err)
		}
	}
}

func BenchmarkRedisStoreGetString(b *testing.B) {
	store := newBenchmarkRedisStore(b)
	key := "bench:redis:get:string"
	value := `{"name":"alpha","age":20}`
	if err := store.SetString(key, value, time.Minute); err != nil {
		b.Fatalf("SetString: %v", err)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := store.GetString(key); err != nil {
			b.Fatalf("GetString: %v", err)
		}
	}
}

func newBenchmarkRedisStore(b *testing.B) *RedisStore {
	b.Helper()

	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		b.Skipf("redis not available: %v", err)
	}

	b.Cleanup(func() {
		_ = client.Close()
	})

	return NewRedisStore(client)
}
