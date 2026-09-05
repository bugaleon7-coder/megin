# 前台 API 限流架构设计

## 一、文档目的

本文用于约束 `/api` 前台业务接口的限流架构，覆盖以下能力：

1. 所有 `/api/*` 接口统一执行的全局限流。
2. 指定 HTTP Method 和路由单独配置的接口限流。
3. 按客户端 IP 和登录用户 UID 分别限流。
4. 限流规则保存在 MySQL，运行时规则和令牌桶保存在进程内存。
5. 服务启动加载、后台修改即时刷新和定时兜底刷新。

本文不把 Redis 作为限流依赖。MySQL 只保存限流规则，不参与每次请求的计数。

相关文档：

- `/api` Handler 规范：`docs/api-article-best-practice.md`
- 总体架构规范：`docs/architecture-dev-guide.md`
- 后台配置型 CRUD 规范：`docs/module-crud-conventions.md`

## 二、当前实现

当前项目已经具备完整的进程内令牌桶能力：

- 所有 `/api` 请求先按 IP 限流。
- Token 鉴权成功后继续按 UID 限流。
- 使用 `golang.org/x/time/rate`。
- 规则来自 MySQL，服务启动时加载到内存快照。
- 后台修改后立即刷新当前实例，定时任务按配置周期刷新其他实例。
- IP 和 UID 的实时令牌数量只保存在当前服务进程内。

```text
MySQL 限流规则
    ↓ 启动加载或热刷新
进程内只读规则快照
    ↓
按规则 ID + IP/UID 保存的内存令牌桶
    ↓
Gin /api 限流中间件
```

当前实现中：

- MySQL 是限流规则的唯一数据来源。
- YAML 保留总开关、刷新周期、内存清理周期，以及规则表为空时使用的全局初始规则。
- 请求过程中不查询 MySQL。
- 接口规则优先于全局规则；指定接口存在独立规则时，使用接口规则替代同维度的全局规则。

## 三、核心概念

限流规则拆成两个相互独立的概念：作用范围和限流维度。

### 3.1 作用范围

| 类型 | 含义 |
|---|---|
| `GLOBAL` | 作用于所有 `/api/*` 路由 |
| `ROUTE` | 只作用于指定 HTTP Method 和 Gin 路由模板 |

这里的“全局”表示规则覆盖所有 `/api` 接口，不表示整个服务的所有用户共同争抢一个令牌桶。

如果以后需要限制单个实例的总 QPS，应新增 `INSTANCE` 维度，或者在 Nginx、Ingress、API Gateway 层完成容量保护。

### 3.2 限流维度

| 维度 | Key | 适用范围 |
|---|---|---|
| `IP` | `ctx.ClientIP()` | 公开接口和鉴权接口 |
| `UID` | Token Claims 中的 `UserID` | 仅 Token 鉴权成功的接口 |

UID 维度不保存原始 Token。同一个用户签发或刷新多个 Token 后仍共享同一个 UID 额度。

### 3.3 令牌桶参数

数据库不直接保存小数 `rate`，而是保存更容易理解的三个参数：

| 字段 | 含义 |
|---|---|
| `rate_count` | 每个周期补充的令牌数量 |
| `interval_seconds` | 令牌补充周期，单位秒 |
| `burst` | 令牌桶容量和空闲后的瞬时放行上限 |

运行时计算：

```text
rate = rate_count / interval_seconds
```

每个请求消耗一个令牌。当前令牌不足一个时，请求被拒绝。

示例：

| 业务含义 | rate_count | interval_seconds | burst |
|---|---:|---:|---:|
| 每 5 秒恢复一个额度，不允许突发 | 1 | 5 | 1 |
| 平均每秒 5 个，允许瞬时 5 个 | 5 | 1 | 5 |
| 每 5 秒恢复一个额度，允许空闲后连续 3 个 | 1 | 5 | 3 |

`rate` 控制长期平均速率，`burst` 控制短时间包容性。`burst` 必须是正整数。

## 四、规则覆盖与回退方式

接口规则和全局规则不叠加消费。每个限流维度独立选择一条最终生效的规则：

1. 当前 Method + Gin 路由模板存在启用的接口规则时，使用接口规则。
2. 当前接口没有对应维度的启用规则时，回退使用全局规则。
3. 接口规则和全局规则都不存在时，跳过该维度的限流。

选择逻辑可以概括为：

```text
最终 IP 规则  = 接口 IP 规则  ?? 全局 IP 规则
最终 UID 规则 = 接口 UID 规则 ?? 全局 UID 规则
```

这里的覆盖按维度独立进行。例如某个鉴权接口只配置了接口 IP 规则，没有配置接口 UID 规则，那么它使用接口 IP 规则和全局 UID 规则，而不是同时执行两条 IP 规则。

| 接口类型 | 最终执行的规则 |
|---|---|
| 无 Token 接口 | 接口 IP 规则存在则使用接口 IP，否则使用全局 IP |
| 有 Token 接口 | IP 和 UID 分别优先使用接口规则，没有接口规则的维度回退到全局规则 |

公开接口请求链路：

```text
请求
  → 选择接口 IP 规则或全局 IP 规则
  → 执行最终选中的一条 IP 规则
  → Handler
```

鉴权接口请求链路：

```text
请求
  → 选择接口 IP 规则或全局 IP 规则
  → 执行最终选中的一条 IP 规则
  → Token 鉴权并解析 UID
  → 选择接口 UID 规则或全局 UID 规则
  → 执行最终选中的一条 UID 规则
  → Handler
```

禁用的接口规则不参与选择，等同于该接口没有独立规则，因此会回退到同维度的全局规则。如果将来需要让某个接口完全绕过全局规则，应单独增加“跳过全局限流”语义，不能用禁用状态表达。

每个维度只消费一个令牌桶，不会发生接口规则拒绝后仍消耗全局规则额度的问题。

## 五、接口规则匹配

接口规则使用下面两个值匹配：

```go
ctx.Request.Method
ctx.FullPath()
```

必须使用 Gin 路由模板，不能使用带实际参数的原始 URL。

例如：

```text
实际请求：/api/article/100
路由模板：/api/article/:id
```

数据库应保存：

```text
GET + /api/article/:id
```

第一版只支持 HTTP Method + Gin 路由模板精确匹配，不支持正则表达式、路径通配符和请求参数条件，避免规则优先级不可预测。

## 六、MySQL 表设计

建议使用一张规则表，每个作用范围和限流维度对应一条记录：

```sql
CREATE TABLE api_rate_limit_rules (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '限流规则ID',
    name                VARCHAR(100) NOT NULL COMMENT '规则名称',
    scope_type          TINYINT NOT NULL COMMENT '作用范围：1全局，2指定接口',
    http_method         VARCHAR(10) NOT NULL DEFAULT '*' COMMENT 'HTTP方法，全局规则固定为*',
    route_path          VARCHAR(255) NOT NULL DEFAULT '*' COMMENT 'Gin路由模板，全局规则固定为*',
    dimension           TINYINT NOT NULL COMMENT '限流维度：1 IP，2 UID',
    rate_count          INT NOT NULL COMMENT '每个周期补充的令牌数',
    interval_seconds    INT NOT NULL COMMENT '令牌补充周期秒数',
    burst               INT NOT NULL COMMENT '令牌桶容量和瞬时放行上限',
    status              TINYINT NOT NULL DEFAULT 1 COMMENT '状态：0禁用，1启用',
    remark              VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
    created_by          BIGINT NOT NULL DEFAULT 0 COMMENT '创建管理员ID',
    updated_by          BIGINT NOT NULL DEFAULT 0 COMMENT '最后修改管理员ID',
    created_at          DATETIME NOT NULL COMMENT '创建时间',
    updated_at          DATETIME NOT NULL COMMENT '修改时间',
    UNIQUE KEY uk_scope_route_dimension (
        scope_type,
        http_method,
        route_path,
        dimension
    ),
    KEY idx_status (status)
) COMMENT='前台API限流规则';
```

唯一索引用于保证：

- 全局 IP 规则最多一条。
- 全局 UID 规则最多一条。
- 同一个接口的 IP 规则最多一条。
- 同一个接口的 UID 规则最多一条。

配置数据以启用和禁用为主。删除操作建议硬删除，避免软删除记录与唯一索引产生冲突；如需保留完整变更历史，应增加独立的审计表。

### 6.1 全局 IP 规则示例

```text
scope_type       = 1
http_method      = *
route_path       = *
dimension        = 1
rate_count       = 2
interval_seconds = 1
burst            = 2
status           = 1
```

### 6.2 全局 UID 规则示例

```text
scope_type       = 1
http_method      = *
route_path       = *
dimension        = 2
rate_count       = 1
interval_seconds = 5
burst            = 1
status           = 1
```

### 6.3 默认指定接口 IP 规则示例

```text
name             = 健康检查接口IP限流示例
scope_type       = 2
http_method      = GET
route_path       = /api/health
dimension        = 1
rate_count       = 1
interval_seconds = 5
burst            = 1
status           = 1
```

`GET /api/health` 无需 Token，用于检查前台 API 服务是否正常，也是项目默认的接口级 IP 限流示例。该规则表示每个 IP 每 5 秒补充 1 个令牌，桶容量为 1；它会替代该接口上的全局 IP 规则。

## 七、运行时内存结构

运行时内存分成“规则快照”和“实时令牌桶”两部分。

### 7.1 规则快照

建议结构：

```go
type RouteKey struct {
    Method string
    Path   string
}

type RuleSnapshot struct {
    GlobalIP  *Rule
    GlobalUID *Rule
    Routes    map[RouteKey]RouteRules
}
```

完整加载和校验数据库规则后，使用 `atomic.Pointer` 一次性替换快照：

```text
数据库查询
  → 构建新的完整快照
  → 完整校验
  → 原子替换
```

禁止在刷新过程中逐条修改请求正在读取的 Map，否则请求可能读到一半新、一半旧的规则。

### 7.2 实时令牌桶

令牌桶按“规则 ID + 限流对象”隔离：

```text
规则 ID + IP  → rate.Limiter
规则 ID + UID → rate.Limiter
```

示例：

```text
全局 IP 规则 #1 + 192.168.1.10
登录 IP 规则 #3 + 192.168.1.10
全局 UID 规则 #2 + UID 1001
用户详情 UID 规则 #5 + UID 1001
```

同一个 IP 或 UID 可以因为访问过不同接口而在内存中同时保留多个桶，但一次请求对同一个维度只选择并消费一个桶：优先消费接口规则的桶，没有接口规则时才消费全局规则的桶。

Map 的创建、读取和清理需要并发保护。单个 `rate.Limiter` 本身支持并发调用。

### 7.3 闲置清理

每个桶记录 `lastSeen`：

- 超过 `idle-expiration-seconds` 没有访问后允许删除。
- 请求到达且超过 `cleanup-interval-seconds` 时惰性触发清理。
- 不为每个规则或每个桶启动 goroutine。
- 已删除或禁用规则遗留的桶不再匹配，并通过闲置清理自然释放。

为了限制恶意 IP 造成的内存增长，后续实现还应增加最大桶数量，并正确配置 Gin 可信代理，防止客户端伪造 `X-Forwarded-For`。

## 八、规则变更与令牌桶更新

定时刷新不应清空所有 limiter，否则每次刷新都会让用户重新获得满桶额度。

建议处理方式：

| 变更 | 运行时行为 |
|---|---|
| 规则未变化 | 继续使用原 limiter |
| `rate_count` 或周期变化 | 调用 `SetLimitAt` 更新原 limiter |
| `burst` 变化 | 调用 `SetBurstAt` 更新原 limiter |
| 新增规则 | 第一次命中时创建 limiter |
| 禁用或删除规则 | 不再匹配，旧桶等待闲置清理 |

每条规则可根据下面字段计算配置指纹：

```text
规则ID + rate_count + interval_seconds + burst + status
```

桶中保存最后使用的配置指纹。只有指纹变化时才更新 limiter 参数。

## 九、加载与刷新机制

推荐同时支持启动加载、后台即时刷新和定时兜底刷新。

### 9.1 服务启动加载

```text
初始化 MySQL
  → 创建 LimiterManager
  → 查询全部启用规则
  → 校验并构建快照
  → 注册 /api 路由
  → 启动 HTTP 服务
```

如果限流功能开启，但首次规则加载失败，建议终止服务启动，避免服务在没有预期保护的情况下运行。

限流规则表和默认规则已收敛至 `server-api/sql/20260904_0001_init.sql`；后续字段或数据变更必须通过日期化增量 SQL 迁移，不再由限流运行时代码建表或补数据。

### 9.2 后台修改后即时刷新

后台新增、修改、启停或删除规则时：

```text
后台 Handler
  → system rate-limit biz
  → 数据库事务提交
  → LimiterManager.Reload()
  → 原子替换当前实例的规则快照
```

数据库提交必须先于内存刷新。刷新失败时要明确记录“数据库已保存但运行时刷新失败”，并允许管理员手动重新刷新。

### 9.3 定时兜底刷新

推荐每分钟执行一次完整规则加载：

```text
@every 1m
```

理由：

- 限流规则数量通常很少，全表查询成本低。
- 当前实例可在后台修改后立即生效。
- 其他服务实例最迟一分钟同步。
- 即时刷新失败后可以自动恢复。

刷新过程使用互斥锁或 `singleflight`，避免后台刷新和定时刷新同时执行。

刷新失败时：

- 保留上一版有效快照。
- 禁止替换成空快照。
- 记录错误日志和监控告警。
- 下一周期继续重试。

## 十、后台管理设计

限流规则属于后台系统基础能力，管理端实现统一收口到 `internal/system`；前台请求链路仍由 `/api` 限流中间件执行。

已实现接口：

| 方法 | 路径 | 用途 |
|---|---|---|
| GET | `/admin-api/system/rate-limit/pageList` | 分页查询规则 |
| GET | `/admin-api/system/rate-limit/detail` | 查询规则详情 |
| POST | `/admin-api/system/rate-limit/create` | 新增规则 |
| PUT | `/admin-api/system/rate-limit/update` | 修改规则 |
| PUT | `/admin-api/system/rate-limit/changeStatus` | 启用或禁用规则 |
| DELETE | `/admin-api/system/rate-limit/delete` | 删除规则 |
| POST | `/admin-api/system/rate-limit/refresh` | 手动刷新当前实例规则 |

这些接口挂载后台 Token 和 Casbin，并遵守后台成功响应 `code=200`、消息字段使用 `message` 的约定。接口元数据和超级管理员初始权限见 `server-api/sql/20260904_0001_init.sql`。

推荐目录：

```text
internal/
├── admin-api/system/
│   └── rate_limit.go
├── middleware/
│   └── api_rate_limit.go
└── system/
    ├── biz/
    │   └── rate_limit.go
    ├── dto/
    │   └── rate_limit.go
    ├── model/
    │   └── rate_limit.go
    ├── repository/
    │   └── rate_limit.go
    ├── router/
    │   └── rate_limit.go
    ├── runtime/
    │   └── rate_limit.go
    └── service/
        └── rate_limit.go
```

`biz` 负责“数据库修改成功后刷新运行时规则”的完整流程；`service` 和 `repository` 不依赖 Gin。

## 十一、保存规则时的校验

当前后台保存规则时校验：

1. `rate_count > 0`。
2. `interval_seconds > 0`。
3. `burst > 0` 且为整数。
4. HTTP Method 是有效方法并统一转为大写。
5. 接口规则路径必须以 `/api/` 开头。
6. 全局规则 Method 和 Path 固定为 `*`。
7. 同一 Scope、Method、Path、Dimension 不允许重复。
8. 禁止为 `/admin-api`、Swagger 和静态资源创建规则。

后台页面应从真实 `/api` 路由列表提供选择器，避免手写错误路径。后续如果开放自由录入，需要再把“UID 规则只能选择鉴权接口”和“路由必须存在”做成服务端路由元数据校验。

## 十二、拒绝响应与可观测性

限流拒绝继续使用项目统一响应：

```json
{
  "code": 429,
  "message": "请求过于频繁，请稍后再试",
  "success": false
}
```

HTTP 状态保持项目当前约定的 `200`。

建议同时返回：

```http
Retry-After: 5
```

内部日志至少记录：

- 命中的规则 ID。
- 全局规则或接口规则。
- IP 或 UID 维度。
- HTTP Method。
- Gin 路由模板。
- 预计等待时间。

对外响应不暴露内部规则名称、服务器结构或完整用户信息。

建议增加以下监控指标：

- 各规则允许请求数。
- 各规则拒绝请求数。
- 当前内存令牌桶数量。
- 规则刷新成功和失败次数。
- 当前规则快照版本或更新时间。

## 十三、多实例边界

MySQL 只同步规则，不同步每个 IP 和 UID 的实时令牌数量。

```text
实例 A：独立规则快照 + 独立令牌桶
实例 B：独立规则快照 + 独立令牌桶
实例 C：独立规则快照 + 独立令牌桶
```

后台修改后：

- 接收后台请求的当前实例立即刷新。
- 其他实例通过每分钟轮询最终同步。

同一个用户经过负载均衡访问不同实例时，实际可用额度可能扩大为实例数量倍。这是纯内存限流的固有限制。

如果未来要求严格的集群级额度，应选择以下方案之一：

1. 使用 Redis 等共享存储保存实时限流状态。
2. 在 Nginx、Ingress 或 API Gateway 层统一限流。
3. 使用能够保证同一限流 Key 固定落到同一实例的流量调度策略，但仍需处理扩缩容和故障迁移。

## 十四、当前实施边界

当前已经实现：

- 全局 IP 规则。
- 全局 UID 规则。
- 指定接口 IP 规则。
- 指定接口 UID 规则。
- 精确 Method + 路由模板匹配。
- 启动加载。
- 后台修改后当前实例立即刷新。
- 每分钟定时刷新。
- 原子规则快照。
- 闲置令牌桶清理。

当前暂不实现：

- Redis 分布式计数。
- 正则或通配符路由。
- 按请求参数、设备号或业务资源限流。
- 长周期计费和用户套餐配额。
- 多条同维度规则的优先级编排。

## 十五、结论

推荐最终结构可以概括为：

```text
MySQL 保存规则
  → 启动加载、后台即时刷新、每分钟兜底刷新
  → atomic.Pointer 保存完整只读规则快照
  → 内存按规则 ID + IP/UID 保存令牌桶
  → 每个维度优先使用接口规则，没有接口规则时回退到全局规则
  → 请求链路不查询 MySQL
```

这套设计在保持纯内存限流性能的同时，提供后台动态配置和指定接口差异化限流能力。它适合单实例或允许多实例额度近似的部署场景，不等同于严格的集群级限流。
