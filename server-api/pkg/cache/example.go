package cache

import (
	"context"
	"time"
)

// ExampleLoginFailCounter shows a typical statistics usage:
//
//	redisStore := cache.NewRedisStore(config.GetRedis().GetDB())
//	counter := cache.InitDefaultCounter("login_fail", redisStore, redisStore, "admin:login:fail")
//	result, err := counter.HitAndCheck(ctx, "user:password:admin", 5, 15*time.Minute)
//	if err != nil {
//	    return err
//	}
//	if result.Limited {
//	    // lock or deny login
//	}
//
// Business-side local cache manager usage:
//
//	local := bizcache.GetLocalCacheManager()
//	err := cache.SetStruct(local.PayTypeConfigStore, "pay_type:wx", payType, 300*time.Second)
//	val, err := cache.GetStruct[PayType](local.PayTypeConfigStore, "pay_type:wx")
func ExampleLoginFailCounter(ctx context.Context, counter *Counter) (*CounterResult, error) {
	return counter.HitAndCheck(ctx, "user:password:admin", 5, 15*time.Minute)
}
