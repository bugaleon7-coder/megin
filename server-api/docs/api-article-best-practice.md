# api/article.go 开发规范与最佳实践

本文以 [internal/api/article.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/api/article.go) 为例，约束前端业务接口（`/api/...`）的推荐写法。

注意：

- `article.go` 里的缓存、限流、锁主要用于演示这些能力怎么接
- 普通接口默认不强制加缓存
- 只有在明确要求使用 Redis 缓存或本地缓存时，才建议接入缓存

重点说明三类能力：

- `Detail` 的缓存用法
- `Create` 的限流用法
- `Update/Delete` 的防重复提交用法

## 1. Handler 层职责

`internal/api` 层只做 4 件事：

1. 接收请求参数
2. 调用 `ctx` 提供的通用能力，例如缓存、限流、分布式锁
3. 调用 `service`
4. 返回统一结果

不建议在 `api` 层做下面这些事：

- 不要写复杂业务判断，业务规则应放在 `service`
- 不要手写大段 `model -> dto` 转换，转换应放在 `service`
- 不要直接操作底层 Redis client，优先走 `ctx.Redis()`、`ctx.LocalCache()`、`ctx.Limiter()`、`ctx.Locker()`

## 2. Detail 缓存用法

### 2.1 推荐写法

`Detail` 的目标是“先查缓存，未命中再查 service，并自动回填缓存”。推荐直接使用闭包式缓存助手：

```go
func (this *Article) Detail(ctx *api.Context, req *base.BaseId) (*api.Result[dto.Article], error) {
	cacheKey := fmt.Sprintf("article:detail:%d", req.ID)

	data, err := api.RememberRedisStruct(ctx.Redis(), cacheKey, 60*time.Second, func() (dto.Article, error) {
		return service.NewArticle(ctx).GetById(req.ID)
	})
	if err != nil {
		return nil, err
	}
	return api.ResultData(data)
}
```

### 2.2 为什么这样写

这种写法把“缓存命中”和“回源加载”收敛成一个调用，优点是：

- `api` 层代码短，不需要手写 `Get -> miss -> service -> Set`
- 缓存逻辑集中，后续统一调整更容易
- `service` 仍然只负责业务，不和缓存耦合

### 2.3 缓存 key 规范

推荐格式：

```text
业务名:场景:主键
```

例如：

```text
article:detail:123
```

要求：

- key 必须能看出业务含义
- key 必须带主键或唯一条件，避免串数据
- key 不要直接拼整段 JSON 请求参数

### 2.4 Redis 和本地缓存怎么选

- `ctx.Redis()`：适合多实例共享缓存、部署到多机器后仍要求缓存一致的场景
- `ctx.LocalCache()`：适合单机热点数据、本地快速命中、允许短时不一致的场景

当前 `Detail` 示例使用 Redis，更适合作为通用标准示例。

### 2.5 缓存使用注意事项

- 缓存只放查询接口，不要直接给写接口套缓存返回
- 缓存 TTL 要明确，不能无限期缓存业务详情
- 更新和删除接口如果影响详情数据，要考虑同步删缓存或延迟过期策略

## 3. Create 限流用法

### 3.1 当前示例写法

`Create` 当前示例是：

```go
if err := ctx.Limiter().AllowRequestOnce(5*time.Second, "11111"); err != nil {
	return nil, err
}
```

它表示：在 5 秒窗口内，同一个限流 key 只允许通过一次。

### 3.2 限流器适合什么场景

`ctx.Limiter()` 适合：

- 按钮防连续点击
- 短时间重复提交拦截
- 匿名接口短时间防刷

它不适合：

- 精确 QPS 控制
- 复杂配额系统
- 长周期计费型限额

### 3.3 规范写法

`"11111"` 只能当演示，不能作为正式业务 key。正式开发必须把限流 key 设计成有业务意义的粒度。

推荐做法：

```go
if err := ctx.Limiter().AllowRequestOnce(5*time.Second, "user", userID); err != nil {
	return nil, err
}
```

或者匿名接口：

```go
if err := ctx.Limiter().AllowRequestOnce(5*time.Second, "ip", clientIP); err != nil {
	return nil, err
}
```

### 3.4 限流 key 设计原则

- 按用户限流：适合登录后接口
- 按 IP 限流：适合匿名接口
- 按资源限流：适合同一资源短时间重复创建或重复触发

推荐 key 组成：

```text
limit:请求方法:路由:业务维度
```

`ctx.Limiter()` 会自动带上请求级前缀，所以调用方只需要补充业务维度参数即可。

### 3.5 限流注意事项

- 不要把所有用户都打到同一个固定 key 上
- 不要把限流时间写得过长，避免误伤正常请求
- 限流失败属于业务拒绝，应直接返回错误，不要继续执行 `service`

## 4. Update/Delete 防重用法

### 4.1 推荐写法

更新和删除属于写操作，必须防止并发重复提交。推荐先获取锁，再执行业务，最后 `defer` 释放：

```go
lock, err := ctx.Locker().AcquireRequestLock(5*time.Second, req.ID)
if err != nil {
	return nil, err
}
defer lock.Release()

updatedArticle, err := service.NewArticle(ctx).Update(req)
if err != nil {
	return nil, err
}
```

删除接口同理：

```go
lock, err := ctx.Locker().AcquireRequestLock(5*time.Second, req.ID)
if err != nil {
	return nil, err
}
defer lock.Release()

err = service.NewArticle(ctx).Delete(req.ID)
if err != nil {
	return nil, err
}
```

### 4.2 为什么用锁而不是只用限流

限流解决的是“短时间不允许重复进来”。

分布式锁解决的是“同一资源在同一时刻只允许一个请求执行”。

这两个场景不一样：

- 限流更适合拦截重复点击
- 锁更适合保护写操作的一致性

对于 `Update/Delete` 这种会修改数据的接口，优先使用分布式锁。

### 4.3 锁 key 设计原则

当前写法：

```go
ctx.Locker().AcquireRequestLock(5*time.Second, req.ID)
```

`ctx.Locker()` 会自动加上请求级前缀，`req.ID` 作为资源维度追加到 key 尾部。

因此同一个接口、同一条数据，会天然落到同一把锁上；不同文章不会互相阻塞。

### 4.4 防重注意事项

- 成功获取锁后必须 `defer lock.Release()`
- 锁只包住必要的写操作，不要把无关慢逻辑一起放进去
- 锁的 TTL 不要过短，避免业务没执行完锁先过期
- 锁的 TTL 也不要过长，避免异常时长时间阻塞同资源写入

## 5. 推荐模板

### 5.1 查询接口模板

```go
func (this *Demo) Detail(ctx *api.Context, req *base.BaseId) (*api.Result[dto.Demo], error) {
	cacheKey := fmt.Sprintf("demo:detail:%d", req.ID)
	data, err := api.RememberRedisStruct(ctx.Redis(), cacheKey, 60*time.Second, func() (dto.Demo, error) {
		return service.NewDemo(ctx).GetById(req.ID)
	})
	if err != nil {
		return nil, err
	}
	return api.ResultData(data)
}
```

### 5.2 创建接口模板

```go
func (this *Demo) Create(ctx *api.Context, req *dto.CreateDemo) (*api.Result[any], error) {
	if err := ctx.Limiter().AllowRequestOnce(5*time.Second, "user", userID); err != nil {
		return nil, err
	}

	_, err := service.NewDemo(ctx).Create(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}
```

### 5.3 更新接口模板

```go
func (this *Demo) Update(ctx *api.Context, req *dto.UpdateDemo) (*api.Result[dto.Demo], error) {
	lock, err := ctx.Locker().AcquireRequestLock(5*time.Second, req.ID)
	if err != nil {
		return nil, err
	}
	defer lock.Release()

	data, err := service.NewDemo(ctx).Update(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(data)
}
```

## 6. 结论

把 `internal/api/article.go` 作为前端 `api` 层标准时，核心原则只有三条：

- 查询接口优先用闭包缓存助手，保持 `api` 层短小
- 创建接口优先做限流，防止短时间重复提交
- 更新和删除接口优先做分布式锁，保护同一资源写一致性

如果一个新接口同时涉及“热点查询、重复点击、并发写入”，就分别套用这三种模式，不要把它们混成一种能力。
