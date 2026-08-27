package api

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// RequestOption 声明单个接口的请求处理策略。
type RequestOption interface {
	apply(*Context, any, *Options)
}

// Options 保存当前请求解析后的缓存、限流和锁配置。
type Options struct {
	ctx         *Context
	saveCache   *cacheRule
	deleteCache []cacheRule
	locks       []lockRule
	rateLimits  []limitRule
}

type cacheRule struct {
	key string
	ttl time.Duration
}
type lockRule struct {
	key string
	ttl time.Duration
}
type limitRule struct {
	key string
	ttl time.Duration
}

type optionConfig struct {
	prefix       string
	parts        []string
	withTokenUID bool
	ttl          time.Duration
}
type saveRedisCacheOption optionConfig
type deleteRedisCacheOption optionConfig
type lockOption optionConfig
type rateLimitOption optionConfig

// WithCache 声明 Redis 读穿透缓存策略。
// 缓存默认使用 Redis；本地缓存需要显式使用 WithLocalCache。
func WithCache(prefix string, parts []string, ttl time.Duration) RequestOption {
	return saveRedisCacheOption{prefix: prefix, parts: parts, ttl: ttl}
}

// WithDeleteCache 声明业务成功后删除 Redis 缓存的策略。
// 缓存默认使用 Redis；本地缓存需要显式使用 WithDeleteLocalCache。
func WithDeleteCache(prefix string, parts []string) RequestOption {
	return deleteRedisCacheOption{prefix: prefix, parts: parts}
}

// WithLock 声明分布式锁策略。
// 获取成功后只在当前 Handler 执行期间互斥，Handler 成功或失败都会立即释放锁；
// 适合更新、删除、状态流转等需要保护临界区的操作。
// 锁被占用时返回业务错误 429，提示“当前数据正在处理中，请稍后再试”，不会执行 Handler。
// 与 WithRateLimit 的区别是：WithRateLimit 在 TTL 到期前不会释放限速标识。
func WithLock(prefix string, parts []string, ttl time.Duration) RequestOption {
	return lockOption{prefix: prefix, parts: parts, ttl: ttl}
}

// WithRateLimit 声明固定时间窗内只允许通过一次的限速策略。
// 首次通过后不主动释放限速标识，直到 TTL 到期前同 key 请求都会被拒绝；
// 适合创建、提交等需要限制短时间重复请求的操作。
// 命中限速时返回业务错误 429，提示“请求过于频繁,稍后再试”，不会执行 Handler。
// 与 WithLock 的区别是：WithLock 会在当前 Handler 执行结束后立即释放锁。
func WithRateLimit(prefix string, parts []string, withTokenUID bool, ttl time.Duration) RequestOption {
	return rateLimitOption{prefix: prefix, parts: parts, withTokenUID: withTokenUID, ttl: ttl}
}

func (v saveRedisCacheOption) apply(ctx *Context, req any, opts *Options) {
	c := optionConfig(v)
	opts.saveCache = &cacheRule{key: buildOptionKey("cache", ctx, req, c), ttl: c.ttl}
}
func (v deleteRedisCacheOption) apply(ctx *Context, req any, opts *Options) {
	c := optionConfig(v)
	opts.deleteCache = append(opts.deleteCache, cacheRule{key: buildOptionKey("cache", ctx, req, c)})
}
func (v lockOption) apply(ctx *Context, req any, opts *Options) {
	c := optionConfig(v)
	opts.locks = append(opts.locks, lockRule{key: buildOptionKey("lock", ctx, req, c), ttl: c.ttl})
}
func (v rateLimitOption) apply(ctx *Context, req any, opts *Options) {
	c := optionConfig(v)
	opts.rateLimits = append(opts.rateLimits, limitRule{key: buildOptionKey("limit", ctx, req, c), ttl: c.ttl})
}

// NewOptions 根据绑定请求生成本次调用的 Options。
func NewOptions(ctx *Context, req any, definitions ...RequestOption) *Options {
	opts := &Options{ctx: ctx}
	for _, definition := range definitions {
		if definition != nil {
			definition.apply(ctx, req, opts)
		}
	}
	return opts
}

// Execute 在完整 Handler 调用外执行请求策略。
func Execute[T any](opts *Options, fn func() (T, error)) (T, error) {
	var zero T
	if opts == nil {
		return fn()
	}
	for _, rateLimit := range opts.rateLimits {
		if err := opts.ctx.Limiter().AllowOnce(rateLimit.key, rateLimit.ttl); err != nil {
			return zero, err
		}
	}
	run := func() (T, error) {
		if opts.saveCache != nil {
			return RememberRedisStruct(opts.ctx.Redis(), opts.saveCache.key, opts.saveCache.ttl, fn)
		}
		return fn()
	}
	for i := len(opts.locks) - 1; i >= 0; i-- {
		previous := run
		lock := opts.locks[i]
		run = func() (T, error) {
			var result T
			err := opts.ctx.Locker().WithLock(lock.key, lock.ttl, func() error { var callErr error; result, callErr = previous(); return callErr })
			return result, err
		}
	}
	result, err := run()
	if err != nil {
		return zero, err
	}
	for _, cache := range opts.deleteCache {
		if err = opts.ctx.Redis().Delete(cache.key); err != nil {
			return zero, err
		}
	}
	return result, nil
}

func buildOptionKey(kind string, ctx *Context, req any, config optionConfig) string {
	values := []string{kind, config.prefix}
	if config.withTokenUID {
		values = append(values, fmt.Sprintf("%d", currentOptionUserID(ctx)))
	}
	for _, part := range config.parts {
		values = append(values, url.PathEscape(requestOptionField(req, part)))
	}
	return strings.Join(values, ":")
}

func currentOptionUserID(ctx *Context) int {
	if ctx.UserInfo != nil {
		return ctx.UserInfo.UserID
	}
	if ctx.AdminInfo != nil {
		return ctx.AdminInfo.UserID
	}
	return 0
}
func requestOptionField(req any, name string) string {
	v := reflect.ValueOf(req)
	for v.IsValid() && v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ""
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		field := v.Field(i)
		for _, tag := range []string{"json", "form", "uri"} {
			value := strings.Split(f.Tag.Get(tag), ",")[0]
			if value == name {
				return fmt.Sprint(field.Interface())
			}
		}
		if strings.EqualFold(f.Name, name) {
			return fmt.Sprint(field.Interface())
		}
	}
	return ""
}
