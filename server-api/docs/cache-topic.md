# Cache 专题

## 1. 文档定位

本文描述本项目当前缓存模块的设计、职责边界、使用方式和性能结论。

目标不是介绍通用缓存理论，而是明确本仓库里应该怎么用缓存、什么时候用本地缓存、什么时候用 Redis，以及当前封装的取舍。

相关代码：

- `pkg/cache`
- `internal/cache`
- `docs/login-security-topic.md`

## 2. 设计目标

当前缓存模块主要解决四个问题：

- 抽象 Redis 和本地缓存的基础能力，保证基础读写可替换
- 保留不同缓存介质的独有能力，避免为了统一而过度抽象
- 在业务代码侧提供足够直接的调用方式
- 给登录安全、配置缓存、热点只读数据提供统一基础设施

对应的设计原则：

- 基础操作统一
- 特有能力分层暴露
- 基础设施和业务装配分离
- 使用方式尽量简单

## 3. 当前目录划分

### 3.1 `pkg/cache`

`pkg/cache` 是基础设施层，只放和具体业务无关的缓存抽象与实现。

当前职责：

- 定义基础接口
- 封装 Redis Store
- 封装 Local Store
- 提供结构化读写辅助函数
- 提供计数器能力

它不负责：

- 管理整个项目有哪些本地缓存实例
- 维护业务命名的缓存池
- 决定某个业务到底该用本地还是 Redis

### 3.2 `internal/cache`

`internal/cache` 是业务装配层，负责把项目里实际会用到的缓存实例组织起来。

当前职责：

- 管理本地缓存单例
- 管理 Redis Store 单例
- 提供业务侧统一入口
- 维护特定本地缓存池，例如消息已读缓存

这层的存在是必要的，因为“大型项目有哪些缓存实例、各自大小和用途是什么”本质上属于业务装配问题，不属于通用基础设施。

## 4. 当前接口设计

### 4.1 基础接口

当前最核心的基础接口在 `pkg/cache/store.go`：

```go
type BasicStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
```

这组接口只覆盖“所有缓存介质都应该有”的最小能力：

- `Get`
- `Set`
- `Delete`

这样做的目的很直接：

- 保证本地缓存和 Redis 在基础读写层可以互换
- 避免把 Redis 独有语义强行塞进统一接口

### 4.2 扩展接口

对 Redis 这类更强能力的缓存，额外定义独立接口：

```go
type IncrementStore interface {
	IncrBy(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error)
}

type LockStore interface {
	SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)
}
```

这样分层后：

- 基础逻辑依赖 `BasicStore`
- 计数逻辑依赖 `IncrementStore`
- 分布式锁逻辑依赖 `LockStore`

不会因为某个能力只有 Redis 支持，就污染所有缓存实现。

## 5. 当前实现

### 5.1 LocalStore

当前默认本地缓存使用 `ttlcache`，封装在 `pkg/cache/local_store.go`。


当前特点：

- 进程内存缓存
- 单机可见
- 无网络开销
- 支持 TTL
- 适合热点只读、小范围临时态

要注意的一点：

- `ttlcache` 可以直接存对象
- 所以本地缓存的 `SetStruct/GetStruct` 不再需要走 JSON 序列化
- 当前默认 LocalStore 是“对象缓存”，不是“JSON 快照缓存”

这带来的收益是本地结构体缓存明显更快，但也引入了一个语义差异：

- 如果缓存的是指针，后续原对象被修改，缓存中的对象也可能随之变化
- 因此业务上更建议缓存值类型，或者在写入前自行复制对象

### 5.2 RedisStore

Redis 缓存封装在 `pkg/cache/redis_store.go`。

当前特点：

- 支持跨实例共享
- 支持计数、TTL、锁等更强语义
- 更适合登录失败计数、锁定状态、全局共享配置

RedisStore 除了基础 `Get/Set/Delete`，还额外支持：

- `SetNX`
- `IncrBy`
- `Raw() *redis.Client`

这部分能力没有强行塞进 `BasicStore`，而是保留在 Redis 实现本身。

## 6. 结构化辅助函数

当前 `pkg/cache/helper.go` 提供统一辅助函数：

- `SetStruct`
- `GetStruct`
- `SetStructList`
- `GetStructList`
- `SetString`
- `GetString`
- `DeleteKey`
- `GetInt64`

这层的作用是把常见的“字符串缓存”和“结构体缓存”收口成统一语义。

当前约定：

- `SetString/GetString` 直接读写字符串
- `RedisStore` 的 `SetStruct/GetStruct` 仍然走 JSON 编解码
- `LocalStore` 的 `SetStruct/GetStruct` 对对象缓存走直接存取
- 读不到 key 时，`GetString` 返回空字符串，`GetStruct` 返回 `nil`

## 7. 业务侧使用方式

### 7.1 本地缓存

业务侧推荐通过 `internal/cache` 的单例入口使用：

```go
import bizcache "megin/internal/cache"

err := bizcache.SetLocalStruct("app:config", config, 5*time.Minute)
if err != nil {
	return err
}

out, err := bizcache.GetLocalStruct[AppConfig]("app:config")
if err != nil {
	return err
}
_ = out
```

如果需要操作指定的本地缓存池，而不是默认 `Local()`，推荐使用 store 版辅助函数：

```go
import bizcache "megin/internal/cache"

store := bizcache.LocalManager().PayTypeConfigStore

err := bizcache.SetLocalStoreStruct(store, "pay:type:wx", payType, 10*time.Minute)
if err != nil {
	return err
}

payType, err := bizcache.GetLocalStoreStruct[PayType](store, "pay:type:wx")
if err != nil {
	return err
}
_ = payType
```

字符串缓存同理：

```go
import bizcache "megin/internal/cache"

store := bizcache.LocalManager().AppConfigStore

err := bizcache.SetLocalStoreString(store, "app:switch:gift", "enabled", 5*time.Minute)
if err != nil {
	return err
}

value, err := bizcache.GetLocalStoreString(store, "app:switch:gift")
if err != nil {
	return err
}
_ = value
```

### 7.2 Redis 缓存

Redis 侧同样通过业务层入口获取：

```go
import bizcache "megin/internal/cache"

err := bizcache.SetRedisString("admin:login:lock:user:root", "1", 15*time.Minute)
if err != nil {
	return err
}

locked, err := bizcache.GetRedisString("admin:login:lock:user:root")
if err != nil {
	return err
}
_ = locked
```

### 7.3 计数器

对于失败次数、限流次数这类场景，优先使用 `Counter`：

```go
ctx := context.Background()
store := bizcache.Redis()
counter := cache.NewCounter(store, store, "admin:login:fail")

result, err := counter.HitAndCheck(ctx, "user:password:admin", 5, 15*time.Minute)
if err != nil {
	return err
}
if result.Limited {
	return errs.NewBusinessError(403, "登录受限，请稍后再试")
}
```

设计原因：

- 计数依赖 `IncrBy`
- 读取和重置依赖 `BasicStore`
- LocalStore 没有原生原子自增能力，所以不能拿它直接做分布式失败计数

## 8. 什么时候用 Local，什么时候用 Redis

### 8.1 适合 Local 的场景

- 单机热点配置
- 读多写少的展示配置
- 只要求当前进程可见的临时缓存
- 可接受重启丢失的数据

例如：

- 支付方式配置缓存
- 礼物配置缓存
- 聊天价格配置缓存
- App 配置缓存

### 8.2 适合 Redis 的场景

- 多实例共享状态
- 登录失败计数
- 锁定状态
- 限流计数
- 分布式锁
- 需要跨服务读取的数据

例如：

- `admin:login:fail:user:*`
- `admin:login:fail:ip:*`
- `admin:login:lock:user:*`
- `admin:login:lock:ip:*`

### 8.3 不建议混用判断的场景

如果一个数据同时满足以下条件，就不要优先放本地缓存：

- 需要多实例一致
- 写后要立即被其他实例感知
- 不能接受进程重启丢失

这类数据应该直接使用 Redis。

## 9. ttlcache 最佳实践

当前默认本地缓存已经切到 `ttlcache`。这类缓存如果只配置 TTL，不配置容量，长期运行时 key 数量仍可能持续增长。

因此本项目的最佳实践不是“只配过期时间”，而是：

- 默认同时配置 `TTL + Capacity`
- 保留 `Start()`，让过期项自动清理
- 根据缓存用途决定容量大小，而不是所有缓存池都用一个值

### 9.1 为什么要同时配置 TTL 和 Capacity

- `TTL` 控制单个 key 存活多久
- `Capacity` 控制缓存池最多保留多少条
- `Start()` 负责自动清理到期元素

三者职责不同：

- 只有 TTL，没有 Capacity：key 种类持续增长时，缓存池仍可能膨胀
- 只有 Capacity，没有 TTL：旧数据会靠淘汰退出，但语义上不适合配置类缓存
- 不启动 `Start()`：已过期数据不会被后台自动清理

因此对于服务端本地缓存，推荐默认使用：

- `TTL + Capacity + Start()`

### 9.2 容量配置原则

`ttlcache.WithCapacity(...)` 限制的是条目数，不是字节数。

因此容量规划要按“预计会有多少不同 key”来评估，而不是按 MB 直接换算。

推荐原则：

- 配置类缓存：小容量，较长 TTL
- 高频临时态缓存：较大容量，较短 TTL
- 高基数缓存：必须有明确容量上限

### 9.3 本项目推荐策略

按当前项目现状，推荐这样使用：

- `DefaultStore`：中等容量，适合普通默认本地缓存
- `PayTypeConfigStore`：小容量，配置种类有限
- `GiftConfigStore`：小容量，配置种类有限
- `AppConfigStore`：小到中等容量
- `ChatPriceConfigStore`：中等容量
- `MessageReadedStore`：较大容量，但必须限制上限

核心原则：

- key 种类有限的缓存池，不需要大容量
- 用户态、消息态这类高基数缓存池，容量一定要有上限
- 跨实例共享数据不应该依赖本地缓存容量策略，应直接走 Redis

### 9.4 推荐调用方式

默认缓存池推荐使用普通函数入口：

```go
import bizcache "megin/internal/cache"

err := bizcache.SetLocalStruct("home:banner", bannerConfig, 10*time.Minute)
if err != nil {
	return err
}

bannerConfig, err := bizcache.GetLocalStruct[BannerConfig]("home:banner")
if err != nil {
	return err
}
_ = bannerConfig
```

指定缓存池推荐显式传入 store：

```go
import bizcache "megin/internal/cache"

store := bizcache.LocalManager().GiftConfigStore

err := bizcache.SetLocalStoreStruct(store, "gift:list", giftList, 10*time.Minute)
if err != nil {
	return err
}

giftList, err := bizcache.GetLocalStoreStruct[[]Gift](store, "gift:list")
if err != nil {
	return err
}
_ = giftList
```

Redis 推荐使用对称的普通函数入口：

```go
import bizcache "megin/internal/cache"

err := bizcache.SetRedisStruct("admin:user:1", profile, 15*time.Minute)
if err != nil {
	return err
}

profile, err := bizcache.GetRedisStruct[AdminProfile]("admin:user:1")
if err != nil {
	return err
}
_ = profile
```

### 9.5 指针缓存注意事项

由于 `ttlcache` 本地缓存现在是对象缓存，不是 JSON 快照缓存，因此：

- 尽量优先缓存值类型
- 如果缓存指针，要明确接受共享可变状态
- 对需要隔离修改影响的数据，写入前先复制对象

否则可能出现“原对象改了，缓存内容也一起变”的情况。

## 10. 性能测试结论

项目内已经增加 benchmark，对比当前 `LocalStore(ttlcache)` 和 `RedisStore` 各自执行 1000 次时的性能。

测试环境：

- `darwin/arm64`
- `Apple M4`
- `go test ./pkg/cache -run '^$' -bench ... -benchtime=1000x -benchmem`

### 10.1 结构体缓存

`LocalStore(ttlcache)` 直接存对象；`RedisStore` 走 JSON 编解码：

- `LocalStore SetStruct`: `327.9 ns/op`
- `LocalStore GetStruct`: `199.1 ns/op`
- `RedisStore SetStruct`: `34761 ns/op`
- `RedisStore GetStruct`: `16275 ns/op`

### 10.2 字符串缓存

`SetString/GetString` 不包含结构体序列化：

- `LocalStore SetString`: `302.9 ns/op`
- `LocalStore GetString`: `148.6 ns/op`
- `RedisStore SetString`: `16368 ns/op`
- `RedisStore GetString`: `15698 ns/op`

### 10.3 结果解读

- 本地缓存明显快于 Redis
- 去掉本地结构体序列化后，结构体缓存性能提升非常明显
- 主导差异仍然是本地内存访问和 Redis 网络往返、协议处理之间的差异

本次替换后的实际收益：


因此：

- 追求极致访问速度、且数据只需单机可见时，用 Local
- 追求共享一致性和分布式语义时，用 Redis

## 11. 当前方案的取舍

### 11.1 当前优点

- 基础读写抽象简单，理解成本低
- 本地缓存和 Redis 基础能力可替换
- Redis 特有能力没有被错误抽象进统一接口
- 业务侧调用方式直接
- `internal/cache` 和 `pkg/cache` 职责边界清晰

### 11.2 当前限制

- LocalStore 和 RedisStore 的结构体缓存语义不再完全一致
- LocalStore 是对象缓存，RedisStore 是 JSON 快照缓存
- 如果本地缓存直接存指针，可能出现共享可变状态问题
- LocalStore 不具备 Redis 那种天然分布式共享能力
- 计数器在强一致限流场景下仍然应优先依赖 Redis

这些限制当前是可接受的，因为它们换来了更简单、稳定和可替换的接口。

## 12. 使用规范

- 共享状态默认优先 Redis
- 单机热点只读数据优先 Local
- 不要把业务缓存实例继续散落在各模块内，统一收口到 `internal/cache`
- 不要在业务代码里直接依赖 `redis.Client`，优先走当前封装
- 需要原生 Redis 能力时，再通过 `RedisStore` 的扩展能力或 `Raw()` 访问
- 新增缓存时，先判断它属于“业务装配问题”还是“基础设施能力问题”

判断原则：

- 通用抽象、通用辅助函数，放 `pkg/cache`
- 项目级缓存池管理、单例实例组织，放 `internal/cache`

## 13. 结论

当前缓存方案的核心不是“把所有缓存能力抽象成一个大而全的接口”，而是：

- 用最小公共接口保证基础可替换
- 用独立扩展能力保留 Redis 等实现的特性
- 用 `internal/cache` 负责项目级缓存装配

这套方案更适合当前项目规模，也更利于后续继续演进登录安全、配置缓存和限流能力。
