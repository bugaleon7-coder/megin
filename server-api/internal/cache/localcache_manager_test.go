package cache

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

type cachePayload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestLocalCacheManagerSingleton(t *testing.T) {
	first := GetLocalCacheManager()
	second := GetLocalCacheManager()
	if first != second {
		t.Fatal("local cache manager should be singleton")
	}
	if first.DefaultStore == nil || first.PayTypeConfigStore == nil || first.ChatPriceConfigStore == nil || first.AppConfigStore == nil {
		t.Fatal("local cache manager should initialize all configured caches")
	}
}

func TestLocalCacheStructRoundTrip(t *testing.T) {
	manager := GetLocalCacheManager()
	key := "pay_type:test"
	payload := &cachePayload{Name: "vip", Age: 18}

	if err := SetLocalStoreStruct(manager.PayTypeConfigStore, key, payload, 5*time.Second); err != nil {
		t.Fatalf("SetStruct: %v", err)
	}

	got, err := GetLocalStoreStruct[cachePayload](manager.PayTypeConfigStore, key)
	if err != nil {
		t.Fatalf("GetStruct: %v", err)
	}
	if got == nil || got.Name != payload.Name || got.Age != payload.Age {
		t.Fatalf("got = %#v, want %#v", got, payload)
	}

	if !DeleteLocalStoreKey(manager.PayTypeConfigStore, key) {
		t.Fatal("DeleteKey should return true when key exists")
	}
	got, err = GetLocalStoreStruct[cachePayload](manager.PayTypeConfigStore, key)
	if err != nil {
		t.Fatalf("GetStruct after delete: %v", err)
	}
	if got != nil {
		t.Fatalf("got = %#v, want nil after delete", got)
	}
}

func TestLocalCacheStringRoundTrip(t *testing.T) {
	key := "app_config:test"

	if err := SetLocalStoreString(GetLocalCacheManager().AppConfigStore, key, "enabled", 5*time.Second); err != nil {
		t.Fatalf("SetString: %v", err)
	}

	got, err := GetLocalStoreString(GetLocalCacheManager().AppConfigStore, key)
	if err != nil {
		t.Fatalf("GetString: %v", err)
	}
	if got != "enabled" {
		t.Fatalf("got %q, want enabled", got)
	}
}

func TestRedisCacheManagerSingleton(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	first := InitDefaultRedisCacheManager(client)
	second, err := DefaultRedisCacheManager()
	if err != nil {
		t.Fatalf("DefaultRedisCacheManager: %v", err)
	}
	if first != second {
		t.Fatal("redis cache manager should be singleton")
	}
	if first.DefaultStore == nil {
		t.Fatal("redis cache manager should initialize default store")
	}
}

func TestManagerSingleton(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	first := InitManager(client)
	second, err := GetManager()
	if err != nil {
		t.Fatalf("GetManager: %v", err)
	}
	if first != second {
		t.Fatal("manager should be singleton")
	}
	if first.Local == nil {
		t.Fatal("manager should initialize local manager")
	}
	if first.Redis == nil || first.Redis.DefaultStore == nil {
		t.Fatal("manager should initialize redis manager")
	}
}

func TestConvenienceAccessors(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	InitManager(client)

	if Local() == nil {
		t.Fatal("Local() should return default local store")
	}
	if Redis() == nil {
		t.Fatal("Redis() should return default redis store")
	}
	if LocalManager() == nil || RedisManager() == nil || MustManager() == nil {
		t.Fatal("manager accessors should return initialized instances")
	}
}
