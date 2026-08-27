# 管理后台登录安全专题

## 1. 背景

当前管理后台已接入用户名密码、验证码、Google TOTP。

当账号密码或 Google TOTP 连续失败过多时，如果系统不做额外限制，会暴露出以下风险：

- 被撞库或暴力枚举密码
- 被穷举 Google TOTP 动态码
- 攻击者通过错误提示推断当前认证阶段
- 高危账号被持续探测，影响后台整体安全性

因此登录链路需要增加“失败过多后的安全处理”能力，并且该能力必须由后端统一控制。

## 2. 设计目标

- 不允许在 TOTP 已启用时绕过动态码校验
- 对密码阶段和 TOTP 阶段分别做失败控制
- 同时覆盖账号维度和 IP 维度
- 以临时锁定为主，避免轻易永久锁号
- 前端提示模糊化，后端日志精确化
- 支持管理员审计和手工解锁

## 3. 总体方案

### 3.1 认证阶段拆分计数

登录失败不能只维护一个总计数，建议拆成三类：

- `password_fail_count`：用户名密码失败次数
- `totp_fail_count`：密码通过后，Google TOTP 失败次数
- `captcha_fail_count`：验证码失败次数

这样可以避免“只是 OTP 输入错误”也把整个账号当成密码攻击处理。

### 3.2 双维度限速

同时按账号和按 IP 控制：

- 单账号失败过多，防止针对某个管理员持续撞库
- 单 IP 失败过多，防止单点暴力探测

建议阈值：

- 单账号 15 分钟内密码失败超过 5 次，锁定 15 分钟
- 单账号在“密码已通过”后，TOTP 连续失败超过 5 次，锁定 10 分钟
- 单 IP 5 分钟内总失败次数超过 20 次，锁定 10 到 30 分钟
- 单 IP 对同一账号的 TOTP 尝试超过 10 次，临时封禁

### 3.3 临时锁定优先

优先采用自动过期的临时锁，而不是永久锁号：

- 攻击停止后可自动恢复
- 减少人工介入成本
- 降低误伤管理员的运维成本

建议维护的状态：

- `lock_type`
- `locked_until`
- `password_fail_count`
- `totp_fail_count`
- `last_fail_at`

### 3.4 TOTP 不允许降级

当系统 `totp.enable=true` 且管理员已绑定 TOTP 时：

- 未提供 OTP，登录必须失败
- OTP 错误，登录必须失败
- 不能因为密码正确就签发 JWT

也就是说，TOTP 已启用后，动态码属于必经认证步骤，不允许降级回“只校验密码”。

### 3.5 提升挑战强度

达到阈值后，除了直接拒绝，还可以逐级提高挑战强度：

- 强制启用验证码
- 增加重试等待时间
- 采用指数退避策略

建议：

- 第一次达到阈值：等待 30 秒
- 第二次：等待 60 秒
- 第三次：等待 5 分钟
- 多次持续触发：等待 15 分钟或更长

对高危后台环境，还可以进一步叠加：

- 仅允许固定办公 IP
- 仅允许 VPN 来源访问

## 4. 前后端职责边界

### 4.1 后端职责

后端必须负责：

- 判断系统是否开启 TOTP
- 判断当前用户是否已绑定 TOTP
- 判断是否必须输入 OTP
- 判断是否达到失败阈值
- 判断是否进入锁定期
- 统一记录审计日志
- 统一执行解锁和封禁策略

### 4.2 前端职责

前端只负责：

- 在登录前读取最小公开配置，如 `totp.enable`
- 根据后端配置决定是否展示 OTP 输入框
- 展示后端返回的受限状态

前端不能决定：

- 某个用户是否可以跳过 OTP
- 某个账号是否已被锁定
- 是否解除锁定

## 5. 错误提示和日志策略

### 5.1 前端提示

对外提示应尽量模糊，避免帮助攻击者判断进度：

- `用户名、密码或验证码错误`
- `登录受限，请稍后再试`

不要把以下信息直接暴露给前端：

- 是密码错了还是 OTP 错了
- 当前账号是否已绑定 TOTP
- 当前账号还剩几次重试机会

### 5.2 后端日志

后端日志必须精确记录：

- 用户名
- 用户 ID（若已识别）
- IP
- User-Agent
- 失败阶段：密码、验证码、TOTP
- 失败原因枚举
- 是否触发锁定
- 锁定截止时间

## 6. 存储实现建议

推荐优先使用 Redis 存储失败计数和锁定状态。

原因：

- 失败计数天然适合 TTL
- 多实例部署时容易共享状态
- 临时锁定到期可自动失效
- 不污染主业务表结构

建议的 Key 设计：

- `admin:login:fail:user:password:{username}`
- `admin:login:fail:user:totp:{username}`
- `admin:login:fail:ip:{ip}`
- `admin:login:lock:user:{username}`
- `admin:login:lock:ip:{ip}`

成功登录后：

- 清理账号相关密码失败计数
- 清理账号相关 TOTP 失败计数
- 视策略决定是否清理 IP 失败计数

## 7. 解锁机制

建议提供后台高权限解锁能力：

- 管理员可以手工解除账号临时锁
- 管理员可以解除 IP 临时封禁
- 解锁动作必须写审计日志

不建议依赖手工改数据库解锁。

## 8. 推荐落地顺序

### 8.1 第一阶段

- 接入 Redis 失败计数
- 密码失败和 TOTP 失败分开统计
- 超阈值后返回统一受限错误
- 成功登录后清理计数

### 8.2 第二阶段

- 增加账号锁定和 IP 封禁
- 增加指数退避
- 增加登录安全审计日志

### 8.3 第三阶段

- 增加后台手工解锁能力
- 增加安全告警
- 高危环境增加固定 IP 或 VPN 限制

## 9. 本项目建议的最小可实施版本

建议先按以下标准落地：

- 单账号密码失败 5 次，锁 15 分钟
- 单账号 TOTP 失败 5 次，锁 10 分钟
- 单 IP 总失败 20 次，锁 30 分钟
- 所有状态存入 Redis，并使用 TTL 自动过期
- 前端统一提示“登录受限，请稍后再试”

### 9.1 建议的代码调用方式

项目内可统一通过 `pkg/cache` 封装失败计数：

```go
ctx := context.Background()
redisStore := cache.NewRedisStore(config.GetRedis().GetDB())
counter := cache.NewCounter(redisStore, redisStore, "admin:login:fail")

result, err := counter.HitAndCheck(ctx, "user:password:admin", 5, 15*time.Minute)
if err != nil {
	return err
}
if result.Limited {
	return errs.NewBusinessError(403, "登录受限，请稍后再试")
}
```

本地开发或单测场景可以使用本地缓存版本：

```go
localStore := cache.NewLocalStore(1024 * 1024)
counter := cache.NewCounter(nil, localStore, "admin:login:fail")
```

### 9.2 本地缓存与 Redis 性能对比

为避免只停留在经验判断，已在项目内增加 1000 次固定次数 benchmark，对比 `LocalStore` 与 `RedisStore` 的基础读写性能。

测试环境：

- `darwin/arm64`
- `Apple M4`
- `go test ./pkg/cache -run '^$' -bench ... -benchtime=1000x -benchmem`

结构体场景，包含 `SetStruct/GetStruct` 的 JSON 编解码：

- `LocalStore SetStruct`: `1454 ns/op`
- `LocalStore GetStruct`: `1362 ns/op`
- `RedisStore SetStruct`: `49009 ns/op`
- `RedisStore GetStruct`: `15087 ns/op`

字符串场景，使用 `SetString/GetString`，不包含结构体序列化：

- `LocalStore SetString`: `980.5 ns/op`
- `LocalStore GetString`: `508.2 ns/op`
- `RedisStore SetString`: `34914 ns/op`
- `RedisStore GetString`: `14231 ns/op`

结论：

- 在当前实现下，本地缓存明显快于 Redis
- `SetStruct/GetStruct` 的 JSON 成本存在，但不是主导项
- 主导差异仍然是本地内存访问和 Redis 网络往返、协议处理之间的差异
- 登录失败计数、锁定状态这类跨实例共享数据，仍然应优先放 Redis
- 配置类、热点只读数据、单机临时态更适合放本地缓存

## 10. 结论

管理后台登录安全不能只停留在“接入 TOTP”。

真正安全的实现必须同时具备：

- 后端统一开关控制
- TOTP 必经校验
- 失败分阶段计数
- 账号和 IP 双维度限速
- 临时锁定与自动恢复
- 审计日志和管理员解锁机制

以上方案作为本项目管理后台登录链路的安全基线，后续相关改造应以本专题为准。
