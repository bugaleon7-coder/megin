package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"megin/pkg/errs"
	"strings"
	"time"

	pkgcache "megin/pkg/cache"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RedisAccessor 封装当前请求下的 Redis 访问能力。
// 所有读写都会自动绑定请求 context，简化业务层调用。
type RedisAccessor struct {
	ctx   context.Context
	store *pkgcache.RedisStore
}

// NewRedisAccessor 创建 Redis 访问器。
func NewRedisAccessor(ctx context.Context, store *pkgcache.RedisStore) *RedisAccessor {
	return &RedisAccessor{
		ctx:   ctx,
		store: store,
	}
}

// SetString 写入字符串缓存。
func (a *RedisAccessor) SetString(key string, value string, ttl time.Duration) error {
	if a == nil || a.store == nil {
		return errors.New("redis store is nil")
	}
	return a.store.Set(a.ctx, key, value, ttl)
}

// GetString 读取字符串缓存。
// 返回值：
// - string: 缓存内容，未命中时返回空字符串
// - error: 读取失败时返回错误
func (a *RedisAccessor) GetString(key string) (string, error) {
	if a == nil || a.store == nil {
		return "", errors.New("redis store is nil")
	}
	value, err := a.store.Get(a.ctx, key)
	if errors.Is(err, pkgcache.ErrCacheMiss) {
		return "", nil
	}
	return value, err
}

// SetStruct 写入结构体缓存。
func (a *RedisAccessor) SetStruct(key string, value any, ttl time.Duration) error {
	if a == nil || a.store == nil {
		return errors.New("redis store is nil")
	}
	return pkgcache.SetStruct(a.store, key, value, ttl)
}

// SetStructList 写入结构体切片缓存。
func (a *RedisAccessor) SetStructList(key string, values any, ttl time.Duration) error {
	if a == nil || a.store == nil {
		return errors.New("redis store is nil")
	}
	switch typed := values.(type) {
	case []string:
		return pkgcache.SetStructList(a.store, key, typed, ttl)
	default:
		return pkgcache.SetStruct(a.store, key, values, ttl)
	}
}

// GetStruct 读取结构体缓存。
// 请求参数：
// - key: 缓存 key
// - target: 目标结构体指针，必须传非 nil 指针
// 返回值：
// - error: 读取或反序列化失败时返回错误
func (a *RedisAccessor) GetStruct(key string, target any) error {
	if a == nil || a.store == nil {
		return errors.New("redis store is nil")
	}
	value, err := a.store.Get(a.ctx, key)
	if errors.Is(err, pkgcache.ErrCacheMiss) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), target)
}

// RememberRedisStruct 优先读取 Redis 结构体缓存，未命中时执行加载函数并自动回填缓存。
// 请求参数：
// - accessor: Redis 缓存访问器
// - key: 缓存 key
// - ttl: 缓存过期时间
// - loader: 缓存未命中时执行的数据加载闭包，返回需要缓存的结构体数据
// 返回值：
// - T: 缓存命中或回源加载后的结构体数据
// - error: 读取缓存、反序列化、加载函数执行、回填缓存失败时返回错误
func RememberRedisStruct[T any](accessor *RedisAccessor, key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	var zero T
	if accessor == nil || accessor.store == nil {
		return zero, errors.New("redis store is nil")
	}
	if loader == nil {
		return zero, errors.New("redis remember loader is nil")
	}

	value, err := accessor.store.Get(accessor.ctx, key)
	if err == nil && value != "" {
		var cached T
		if unmarshalErr := json.Unmarshal([]byte(value), &cached); unmarshalErr != nil {
			return zero, unmarshalErr
		}
		return cached, nil
	}
	if err != nil && !errors.Is(err, pkgcache.ErrCacheMiss) {
		return zero, err
	}

	loaded, err := loader()
	if err != nil {
		return zero, err
	}
	if err = pkgcache.SetStruct(accessor.store, key, loaded, ttl); err != nil {
		return zero, err
	}
	return loaded, nil
}

// Delete 删除指定 key。
func (a *RedisAccessor) Delete(key string) error {
	if a == nil || a.store == nil {
		return errors.New("redis store is nil")
	}
	return a.store.Delete(a.ctx, key)
}

// Raw 返回底层 redis client，便于处理特殊命令。
func (a *RedisAccessor) Raw() *redis.Client {
	if a == nil || a.store == nil {
		return nil
	}
	return a.store.Raw()
}

// LocalCacheAccessor 封装当前请求下的本地缓存访问能力。
type LocalCacheAccessor struct {
	ctx   context.Context
	store *pkgcache.LocalStore
}

// requestGuardAccessor 封装基于 Redis SetNX + TTL 的通用控制能力。
// 该结构只提供底层原子能力，不直接暴露给业务层。
// 上层可分别封装为“限流”和“分布式锁”两种语义接口。
type requestGuardAccessor struct {
	ctx              context.Context
	store            *pkgcache.RedisStore
	requestKeyPrefix string
}

// NewLocalCacheAccessor 创建本地缓存访问器。
func NewLocalCacheAccessor(ctx context.Context, store *pkgcache.LocalStore) *LocalCacheAccessor {
	return &LocalCacheAccessor{
		ctx:   ctx,
		store: store,
	}
}

// SetString 写入字符串本地缓存。
func (a *LocalCacheAccessor) SetString(key string, value string, ttl time.Duration) error {
	if a == nil || a.store == nil {
		return errors.New("local cache store is nil")
	}
	return a.store.Set(a.ctx, key, value, ttl)
}

// GetString 读取字符串本地缓存。
// 返回值：
// - string: 缓存内容，未命中时返回空字符串
// - error: 读取失败时返回错误
func (a *LocalCacheAccessor) GetString(key string) (string, error) {
	if a == nil || a.store == nil {
		return "", errors.New("local cache store is nil")
	}
	value, err := a.store.Get(a.ctx, key)
	if errors.Is(err, pkgcache.ErrCacheMiss) {
		return "", nil
	}
	return value, err
}

// SetStruct 写入结构体本地缓存。
func (a *LocalCacheAccessor) SetStruct(key string, value any, ttl time.Duration) error {
	if a == nil || a.store == nil {
		return errors.New("local cache store is nil")
	}
	return a.store.SetStruct(key, value, ttl)
}

// SetStructList 写入结构体切片本地缓存。
func (a *LocalCacheAccessor) SetStructList(key string, values any, ttl time.Duration) error {
	if a == nil || a.store == nil {
		return errors.New("local cache store is nil")
	}
	return a.store.SetStructList(key, values, ttl)
}

// GetStruct 读取结构体本地缓存。
// 请求参数：
// - key: 缓存 key
// - target: 目标结构体指针，必须传非 nil 指针
// 返回值：
// - error: 读取失败时返回错误
func (a *LocalCacheAccessor) GetStruct(key string, target any) error {
	if a == nil || a.store == nil {
		return errors.New("local cache store is nil")
	}
	return a.store.GetStruct(key, target)
}

// RememberLocalStruct 优先读取本地结构体缓存，未命中时执行加载函数并自动回填缓存。
// 请求参数：
// - accessor: 本地缓存访问器
// - key: 缓存 key
// - ttl: 缓存过期时间
// - loader: 缓存未命中时执行的数据加载闭包，返回需要缓存的结构体数据
// 返回值：
// - T: 缓存命中或回源加载后的结构体数据
// - error: 读取缓存、加载函数执行、回填缓存失败时返回错误
func RememberLocalStruct[T any](accessor *LocalCacheAccessor, key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	var zero T
	if accessor == nil || accessor.store == nil {
		return zero, errors.New("local cache store is nil")
	}
	if loader == nil {
		return zero, errors.New("local cache remember loader is nil")
	}

	cached, err := pkgcache.GetStruct[T](accessor.store, key)
	if err != nil {
		return zero, err
	}
	if cached != nil {
		return *cached, nil
	}

	loaded, err := loader()
	if err != nil {
		return zero, err
	}
	if err = accessor.store.SetStruct(key, loaded, ttl); err != nil {
		return zero, err
	}
	return loaded, nil
}

// Delete 删除指定本地缓存 key。
func (a *LocalCacheAccessor) Delete(key string) error {
	if a == nil || a.store == nil {
		return errors.New("local cache store is nil")
	}
	return a.store.Delete(a.ctx, key)
}

// Raw 返回底层本地缓存实例。
func (a *LocalCacheAccessor) Raw() *pkgcache.LocalStore {
	if a == nil {
		return nil
	}
	return a.store
}

// LockerAccessor 封装基于 Redis 的分布式锁能力。
type LockerAccessor struct {
	ctx              context.Context
	store            *pkgcache.RedisStore
	requestKeyPrefix string
}

// LimiterAccessor 封装基于 Redis 的简单限流能力。
// 当前实现适合按钮防重复提交、短时间内重复请求拦截等场景。
type LimiterAccessor struct {
	ctx              context.Context
	store            *pkgcache.RedisStore
	requestKeyPrefix string
}

// LockHandle 表示一次成功获取的锁。
// 调用方应在使用完成后显式执行 Release。
type LockHandle struct {
	ctx   context.Context
	store *pkgcache.RedisStore
	key   string
	token string
}

// newRequestGuardAccessor 创建通用请求控制访问器。
func newRequestGuardAccessor(ctx context.Context, store *pkgcache.RedisStore, requestKeyPrefix string) requestGuardAccessor {
	return requestGuardAccessor{
		ctx:              ctx,
		store:            store,
		requestKeyPrefix: requestKeyPrefix,
	}
}

// requestKey 根据当前请求上下文自动拼装控制 key。
func (a *requestGuardAccessor) requestKey(args ...any) string {
	parts := make([]string, 0, len(args)+1)
	if a != nil && a.requestKeyPrefix != "" {
		parts = append(parts, a.requestKeyPrefix)
	}
	for _, arg := range args {
		parts = append(parts, fmt.Sprint(arg))
	}
	return strings.Join(parts, ":")
}

// acquire 使用 Redis SetNX + TTL 获取一次请求控制句柄。
func (a *requestGuardAccessor) acquire(key string, ttl time.Duration) (*LockHandle, error) {
	if a == nil || a.store == nil {
		return nil, errors.New("guard store is nil")
	}
	token := uuid.NewString()
	locked, err := a.store.SetNX(a.ctx, key, token, ttl)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, nil
	}
	return &LockHandle{
		ctx:   a.ctx,
		store: a.store,
		key:   key,
		token: token,
	}, nil
}

// NewLockerAccessor 创建锁访问器。
func NewLockerAccessor(ctx context.Context, store *pkgcache.RedisStore, requestKeyPrefix string) *LockerAccessor {
	return &LockerAccessor{
		ctx:              ctx,
		store:            store,
		requestKeyPrefix: requestKeyPrefix,
	}
}

// NewLimiterAccessor 创建限流访问器。
func NewLimiterAccessor(ctx context.Context, store *pkgcache.RedisStore, requestKeyPrefix string) *LimiterAccessor {
	return &LimiterAccessor{
		ctx:              ctx,
		store:            store,
		requestKeyPrefix: requestKeyPrefix,
	}
}

// Acquire 尝试获取锁。
// 请求参数：
// - key: 锁 key
// - ttl: 锁过期时间
// 返回值：
// - *LockHandle: 获取成功后返回锁句柄，失败时返回 nil
// - error: 获取异常时返回错误
func (a *LockerAccessor) Acquire(key string, ttl time.Duration) (*LockHandle, error) {
	guard := newRequestGuardAccessor(a.ctx, a.store, a.requestKeyPrefix)
	return guard.acquire(key, ttl)
}

// WithLock 在获取到分布式锁后执行回调逻辑，并在结束后自动释放锁。
// 请求参数：
// - key: 锁 key，建议带业务前缀和主键标识
// - ttl: 锁过期时间
// - fn: 获取锁成功后执行的业务逻辑
// 返回值：
// - error: 获取锁失败、执行回调失败、释放锁失败时返回错误
func (a *LockerAccessor) WithLock(key string, ttl time.Duration, fn func() error) error {
	if fn == nil {
		return errors.New("lock callback is nil")
	}
	lock, err := a.Acquire(key, ttl)
	if err != nil {
		return err
	}
	if lock == nil {
		return errs.NewBusinessError(429, "当前数据正在处理中，请稍后再试")
	}
	defer func() {
		_ = lock.Release()
	}()
	return fn()
}

// RequestKey 根据当前请求上下文自动拼装分布式锁 key。
// 拼装规则：
// - 自动前缀：lock + HTTP method + 路由模板
// - 手动参数：按顺序追加到 key 末尾
func (a *LockerAccessor) RequestKey(args ...any) string {
	guard := newRequestGuardAccessor(a.ctx, a.store, a.requestKeyPrefix)
	return guard.requestKey(args...)
}

// WithRequestLock 使用“当前请求前缀 + 手动参数”自动拼装 key，并执行分布式锁保护的业务逻辑。
func (a *LockerAccessor) WithRequestLock(ttl time.Duration, fn func() error, args ...any) error {
	return a.WithLock(a.RequestKey(args...), ttl, fn)
}

// WithRequestResourceLock 使用“当前请求前缀 + 单个资源标识”自动拼装 key，并执行分布式锁保护的业务逻辑。
// 这个方法适合最常见的“按 ID 锁定单条数据”场景，避免把资源参数放在回调后面导致遗漏。
// 请求参数：
// - ttl: 锁过期时间
// - resourceID: 资源唯一标识，例如文章 ID、订单 ID
// - fn: 获取锁成功后执行的业务逻辑
// 返回值：
// - error: 获取锁失败、执行回调失败、释放锁失败时返回错误
func (a *LockerAccessor) WithRequestResourceLock(ttl time.Duration, resourceID any, fn func() error) error {
	return a.WithLock(a.RequestKey(resourceID), ttl, fn)
}

// AcquireRequestLock 使用“当前请求前缀 + 手动参数”自动拼装 key，并返回锁句柄。
// 请求参数：
// - ttl: 锁过期时间
// - args: 参与拼装锁 key 的业务参数
// 返回值：
// - *LockHandle: 获取成功后返回锁句柄，调用方需负责 defer Release
// - error: 获取锁失败或锁已被占用时返回错误
func (a *LockerAccessor) AcquireRequestLock(ttl time.Duration, args ...any) (*LockHandle, error) {
	lock, err := a.Acquire(a.RequestKey(args...), ttl)
	if err != nil {
		return nil, err
	}
	if lock == nil {
		return nil, errs.NewBusinessError(429, "当前数据正在处理中，请稍后再试")
	}
	return lock, nil
}

// Release 释放锁。
// 只有 token 匹配时才会删除锁，避免释放掉其他请求重新抢到的锁。
func (h *LockHandle) Release() error {
	if h == nil || h.store == nil {
		return errors.New("lock handle is nil")
	}
	const releaseScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`
	return h.store.Raw().Eval(h.ctx, releaseScript, []string{h.key}, h.token).Err()
}

// Key 返回当前锁 key。
func (h *LockHandle) Key() string {
	if h == nil {
		return ""
	}
	return h.key
}

// Token 返回当前锁 token。
func (h *LockHandle) Token() string {
	if h == nil {
		return ""
	}
	return h.token
}

// Allow 判断指定 key 在当前时间窗内是否允许继续执行。
// 请求参数：
// - key: 限流 key，建议带上业务前缀和用户标识
// - ttl: 限流时间窗，例如 5*time.Second
// 返回值：
// - error: 通过时返回 nil，命中限流或 Redis 操作失败时返回错误
func (a *LimiterAccessor) AllowOnce(key string, ttl time.Duration) error {
	guard := newRequestGuardAccessor(a.ctx, a.store, a.requestKeyPrefix)
	handle, err := guard.acquire(key, ttl)
	if err != nil {
		return err
	}
	if handle == nil {
		return errs.NewBusinessError(429, "请求过于频繁,稍后再试")
	}
	return nil
}

// RequestKey 根据当前请求上下文自动拼装限流 key。
// 拼装规则：
// - 自动前缀：limit + HTTP method + 路由模板
// - 手动参数：按顺序追加到 key 末尾
func (a *LimiterAccessor) RequestKey(args ...any) string {
	guard := newRequestGuardAccessor(a.ctx, a.store, a.requestKeyPrefix)
	return guard.requestKey(args...)
}

// AllowRequestOnce 使用“当前请求前缀 + 手动参数”自动拼装 key，并执行单次限流校验。
func (a *LimiterAccessor) AllowRequestOnce(ttl time.Duration, args ...any) error {
	return a.AllowOnce(a.RequestKey(args...), ttl)
}

// WithRequestLimit 先执行请求级限流，再执行回调逻辑。
func (a *LimiterAccessor) WithRequestLimit(ttl time.Duration, fn func() error, args ...any) error {
	if fn == nil {
		return errors.New("limit callback is nil")
	}
	if err := a.AllowRequestOnce(ttl, args...); err != nil {
		return err
	}
	return fn()
}

// AllowOncePer5Seconds 判断指定 key 是否允许每 5 秒通过一次。
// 请求参数：
// - key: 限流 key，建议带上业务前缀和用户标识
// 返回值：
// - error: 通过时返回 nil，命中限流或 Redis 操作失败时返回错误
func (a *LimiterAccessor) AllowOncePer5Seconds(key string) error {
	return a.AllowOnce(key, 5*time.Second)
}

// AllowRequestOncePer5Seconds 使用请求级自动拼装 key，并限制每 5 秒只允许一次。
func (a *LimiterAccessor) AllowRequestOncePer5Seconds(args ...any) error {
	return a.AllowRequestOnce(5*time.Second, args...)
}
