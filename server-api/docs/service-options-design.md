# Handler Options 调用策略设计

## 1. 目的

本文确定普通业务接口的缓存、限流和分布式锁方案。

这些能力属于接口调用编排，不属于领域 Service。统一执行层放在泛型路由包装器的外层，直接包住完整 Handler 调用；Handler、Service 和 Repository 不需要手动调用缓存、锁或限流方法。

`system` 模块以兼容接口稳定性为优先，不作为本方案的首批改造目标或新代码参考。

## 2. 最终调用链

```text
Gin 请求
  -> 泛型路由包装器绑定并校验 DTO
  -> 创建 api.Context
  -> 按 Handler 标识、ctx 和 req 构造本次 api.Options
  -> opt.Execute(调用完整 Handler)
  -> Handler 调用纯业务 Service
  -> Service 调用 Repository
  -> 统一输出响应
```

`opt.Execute` 位于 Handler 外层，因此：

- 缓存命中时，Handler 和 Service 都不执行。
- 锁覆盖完整的 Handler 业务调用边界。
- 成功后的缓存保存或删除在统一执行层完成。
- Handler 中没有 `opt.Execute`、`AllowRequestOnce`、`AcquireRequestLock` 等样板代码。

## 3. Handler 与路由形式

Handler 保持现有两参数签名，不需要为了 `Options` 增加第三个参数：

```go
func (h *Article) Detail(
	ctx *api.Context,
	req *base.BaseId,
) (*api.Result[dto.Article], error) {
	article, err := service.NewArticle(ctx).GetById(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(article)
}
```

路由注册方式同样不变：

```go
router.GET(noAuthGroup, "/article/detail", article.Detail)
router.POST(apiGroup, "/article/update", article.Update)
```

路由包装器在内部完成：

```go
opt := api.NewOptions(ctx, req, handlerID)
resp, err := opt.Execute(func() (Resp, error) {
	return handler(ctx, req)
})
```

`Execute` 采用包级泛型函数或 `Options` 持有的泛型无关执行器实现。对外 Handler、Service 方法保持现有强类型签名。

如果后续确实有 Handler 需要读取调用策略，可将 `Options` 放入请求级 `api.Context`；不得存入共享 Handler 结构体字段。默认情况下 Handler 不应感知 `Options`。

## 4. Options 的来源

泛型路由包装器按当前 Handler 标识、`api.Context` 和已绑定的 `req` 创建每请求独立的 `Options`：

```go
opt := api.NewOptions(ctx, req, handlerID)
```

Handler 标识使用“接收者类型 + 方法名”，例如：

```text
Article.Detail
Article.Create
Article.Update
Article.Delete
```

策略集中注册，不在每条路由注册代码中显式配置。策略构造器可以安全使用已校验的 `req.ID`、当前登录用户 ID、客户端 IP 等维度生成 key。

示意：

```go
RegisterHandlerOptions[*base.BaseId]("Article.Detail", func(
	ctx *api.Context,
	req *base.BaseId,
) []api.Option {
	return []api.Option{
		api.WithSaveRedisCache(
			fmt.Sprintf("article:detail:%d", req.ID),
			time.Minute,
		),
	}
})
```

未注册策略时返回空 `Options`，`Execute` 直接调用 Handler，行为与当前实现一致。

为让包装器在执行前获得稳定标识，需要把当前路由中用于 OpenAPI 的 Handler 名称提取逻辑收敛为可复用函数；包装器和路由元数据记录共用同一结果。

## 5. Option 语义

### 5.1 防重复提交

```go
api.WithAllowOnce(key, ttl)
```

使用 Redis `SetNX + TTL` 抢占限流标识。成功后不主动释放；TTL 内同 key 的后续请求被拒绝。

适用 key：

```text
article:create:user:{userID}
order:submit:user:{userID}
captcha:send:ip:{clientIP}
```

### 5.2 分布式锁

```go
api.WithLock(key, ttl)
```

使用带 token 校验的 Redis 锁保护完整 Handler 调用。成功或失败均释放锁；锁已被占用时返回“当前数据正在处理中，请稍后再试”。

适用 key：

```text
article:update:{articleID}
article:delete:{articleID}
order:confirm:{orderID}
```

`WithAllowOnce` 与 `WithLock` 的释放语义不同，必须是两个独立 option。

### 5.3 读取并保存缓存

```go
api.WithSaveRedisCache(key, ttl)
api.WithSaveLocalCache(key, ttl)
```

该 option 声明“先读缓存，未命中后执行并保存”：

1. 从指定缓存读取完整 Handler 成功响应。
2. 命中时直接返回响应，不调用 Handler。
3. 未命中时执行 Handler。
4. Handler 成功后回填缓存。

通常配置给 GET 查询 Handler。缓存默认不启用；同一次调用暂不同时配置 Redis 与本地读缓存，避免层级与失效语义不清。

### 5.4 执行成功后删除缓存

```go
api.WithDeleteRedisCache(keys...)
api.WithDeleteLocalCache(keys...)
```

该 option 声明“Handler 成功后删除”：Handler 或 Service 返回错误时不得删除缓存。通常配置给 POST、PUT、DELETE 等写入 Handler。

命名统一使用 `Delete`，与现有缓存访问器 `Delete(key)` 一致：

- 不使用 `Invalidate`，避免术语不一致。
- 不使用 `Clear`，其容易被理解为清空整个缓存池。
- 不使用 `Remove`，其更接近删除业务资源。

## 6. `Execute` 执行顺序

```text
AllowOnce（可选）
  -> 对 SaveCache option 查缓存（可选）
  -> 获取 Lock（可选）
  -> 调用 Handler
  -> SaveCache option 成功后保存完整响应（可选）
  -> DeleteCache option 成功后删除指定缓存（可选）
  -> 释放 Lock（可选）
```

缓存读取在获取锁之前，保证详情等读取接口命中缓存时不承担锁开销。缓存删除发生在锁仍持有的范围内，之后才释放锁，避免后续并发请求读到旧缓存。

缓存、锁、限流发生系统错误时，沿用现有错误处理约定并直接返回错误，不静默降级。

## 7. 分层约束

```text
router（Options 统一执行层） -> handler -> service -> repository -> model
```

- Router 负责绑定、构造请求级 `Context` 和 `Options`、执行完整 Handler。
- Handler 只负责接口参数到 Service 调用及响应数据转换。
- Service 保持纯业务逻辑，不依赖 `Options`、路由模板、HTTP method 或客户端 IP。
- Repository 不感知缓存、锁、限流和 Option。
- 现有 `ctx.Limiter()`、`ctx.Locker()`、`ctx.Redis()`、`ctx.LocalCache()` 是 `Options` 的底层能力，不重复实现 Redis 逻辑。

## 8. 首个落地与验收

第一阶段以 Article 验证：

1. `Article.Detail`：`WithSaveRedisCache(article:detail:{id})`。
2. `Article.Create`：按用户维度 `WithAllowOnce`。
3. `Article.Update`：按文章 ID `WithLock`，成功后 `WithDeleteRedisCache(article:detail:{id})`。
4. `Article.Delete`：按文章 ID `WithLock`，成功后 `WithDeleteRedisCache(article:detail:{id})`。

验收项：

- 路由注册和 Handler 签名保持不变。
- 缓存命中时 Handler、Service 均不执行。
- 写入失败时不删除缓存。
- 同 lock key 并发请求只有一个进入 Handler。
- 同 allow-once key 在 TTL 内第二次请求被拒绝。
- 空 `Options` 行为与当前实现一致。
- 所有 Go 代码改动完成后执行 `go build ./...`。
