# TODO List

## P0 管理后台登录接入 Google TOTP

### 1. 明确方案边界

- 按 `Google Authenticator/TOTP` 方案实现，不做 `reCAPTCHA`。
- 接入范围限定为管理后台登录接口 `/admin-api/user/login`。
- 复用现有配置项 `totp.enable` 和 `totp.issuer`，不改整体认证中间件链路。

### 2. 调整数据存储

- 不复用 `sys_users.origin_setting` 存放 TOTP 密钥。
- 为 `sys_users` 增加独立字段：`totp_secret`、`totp_enabled`、`totp_bound_at`。
- `totp_secret` 需可逆加密存储，禁止明文落库。
- 手动执行初始化 SQL：[docs/sql/admin_totp_manual.sql](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/sql/admin_totp_manual.sql:1)。

### 3. 补充后台接口

- 增加管理员已登录后的 TOTP 绑定信息获取接口。
- 增加确认绑定接口，校验首次动态码后再启用。
- 增加关闭 TOTP 接口，要求二次校验身份。
- 所有接口保持 `/admin-api` 前缀，并遵循当前项目返回规范：`code=200`、`message` 字段。

### 4. 改造登录流程

- 在用户名密码校验通过后、签发 JWT 前增加 TOTP 校验。
- 扩展登录请求体，增加 `otp` 或等价字段。
- 当 `totp.enable=true` 且用户已绑定时强制校验动态码。
- 当系统开启但用户未绑定时，明确登录期望行为，避免直接锁死现有管理员。

### 5. 处理返回脱敏

- `totp_secret` 禁止出现在登录响应、用户详情响应、用户列表响应中。
- 若用户 DTO 需要返回 TOTP 状态，只返回布尔状态或绑定时间，不返回密钥。
- 复查 `originSetting`、`userInfo`、`login` 等现有返回结构，避免密钥穿透到前端。

### 6. 选型与实现细节

- 优先使用 Go TOTP 库实现标准 6 位动态码校验。
- 支持生成 `otpauth://` URI，并提供二维码展示所需内容。
- 校验时间窗口建议容忍前后 1 个 step，降低时钟偏差导致的误判。

### 7. 补充测试与验证

- 增加未开启 TOTP、已开启未绑定、已开启已绑定的登录测试。
- 增加绑定、确认绑定、关闭绑定接口测试。
- 增加密钥不外泄的响应测试。
- 改完代码后执行 `go build ./...` 验证编译通过。
