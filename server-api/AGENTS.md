## 开发引导

本文不再记录 system 重构任务。当前 `system` 模块迁移已完成。

本文件只保留两个作用：

1. 约束当前仓库的通用开发规则
2. 按开发任务性质，指引应该先看哪份文档

## 通用硬约束

### 1. 编译验证

- 只有改动了 `.go` 文件时，才必须执行 `go build ./...`
- 纯文档、纯配置、纯说明类改动，不强制执行编译验证

### 2. 注释要求

- 代码注释使用中文
- 接口请求字段和返回字段必须写清楚中文注释

### 3. 接口约束

- `/api` 与 `/admin-api` 必须按实际业务场景区分，不能混用
- 后台接口成功响应统一为 `code=200`
- 后台接口成功响应字段统一使用 `message`

### 4. 实现原则

- 优先复用仓库已有实现，不新造风格
- 严禁把 `internal/system` 或 `internal/admin-api/system` 中的代码当作新开发参考代码
- `system` 模块是三方迁移过来的兼容接口，后续以稳定维护为主，接口基本不会变化，也不作为新业务开发模板
- `/api` 前台业务接口优先级更高，相关开发规范优先参考 [docs/api-article-best-practice.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/api-article-best-practice.md) 和 [internal/api/article.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/api/article.go)
- `/admin-api` 后台接口开发再参考 [internal/admin-api/article.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/admin-api/article.go)
- 新开发参考代码统一以 `article` 相关实现为标准
- 缓存默认不启用，只有明确要求“加 Redis 缓存”或“加本地缓存”时才接入
- [internal/api/article.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/api/article.go) 里的缓存写法属于演示，不代表普通接口默认必须加缓存
- 如果涉及数据库连接信息，优先查看 [config/config-dev.yaml](/Users/lchb/go_admin/gin-vue-admin/shop-api/config/config-dev.yaml)

## 先看哪份文档

### 1. 第一次接手项目

先看：

1. [docs/new-developer-guide.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/new-developer-guide.md)
2. [docs/architecture-dev-guide.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/architecture-dev-guide.md)

适用场景：

- 新人第一次参与开发
- 需要先了解目录结构、分层边界、启动方式

### 2. 要和 Codex 协作开发

先看：

1. [docs/codex-crud-guide.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/codex-crud-guide.md)

适用场景：

- 想直接给 Codex 下开发指令
- 想让 Codex 按固定格式实现 CRUD

### 3. 要做后台配置型 CRUD

先看：

1. [docs/module-crud-conventions.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/module-crud-conventions.md)
2. [docs/codex-crud-guide.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/codex-crud-guide.md)

适用场景：

- `VIP` 等级配置
- 分类配置
- 品牌配置
- 标签配置
- 各类字典型、配置型、枚举型模块

重点关注：

- 哪些模块默认要 `status`
- 哪些模块建议有 `sort`
- `/list` 和 `/pageList` 怎么区分
- 哪些模块默认不分页，哪些默认分页，哪些需要先确认

### 4. 要做普通业务 CRUD

先看：

1. [docs/codex-crud-guide.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/codex-crud-guide.md)
2. [docs/architecture-dev-guide.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/architecture-dev-guide.md)
3. [internal/admin-api/article.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/admin-api/article.go)

适用场景：

- 非 system 的后台业务模块
- 文章、客户、商品、内容等普通模块开发

### 5. 要做 `/api` 前台业务接口

先看：

1. [docs/api-article-best-practice.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/api-article-best-practice.md)
2. [docs/architecture-dev-guide.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/architecture-dev-guide.md)

适用场景：

- 前台业务接口开发
- 需要了解限流、锁在 `/api` handler 中的标准写法
- 如果你明确要求加缓存，再参考其中的缓存演示写法

### 6. 要做权限、角色、菜单、Casbin、后台鉴权

先看：

1. [docs/auth-casbin-principles.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/auth-casbin-principles.md)
2. [docs/architecture-dev-guide.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/architecture-dev-guide.md)
3. [internal/admin-api/system](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/admin-api/system)
4. [internal/system](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system)

适用场景：

- 角色权限
- 菜单权限
- Casbin 权限控制
- 后台登录态和中间件链路

### 7. 要改 system 模块接口

先看：

1. [docs/system-api-contract.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/system-api-contract.md)
2. [internal/admin-api/system](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/admin-api/system)
3. [internal/system](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system)

适用场景：

- 修改管理员、角色、菜单、API、字典等 system 接口

注意：

- 当前重点不是“继续迁移”
- 而是在现有实现基础上保持契约稳定、按项目规范继续开发
- 这里只能作为兼容接口维护参考，不能作为新模块开发模板

### 8. 要做缓存相关开发

先看：

1. [docs/cache-topic.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/cache-topic.md)
2. [docs/api-article-best-practice.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/api-article-best-practice.md)

适用场景：

- 详情缓存
- 本地缓存
- Redis 缓存
- 缓存 key 设计

### 9. 要做登录、安全、Token 相关开发

先看：

1. [docs/login-security-topic.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/login-security-topic.md)
2. [docs/auth-casbin-principles.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/auth-casbin-principles.md)

适用场景：

- 登录
- Token
- 黑名单
- 登录安全策略

### 10. 要做文档或接口说明

先看：

1. [docs/api-doc-topic.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/api-doc-topic.md)
2. [docs/system-api-contract.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/system-api-contract.md)

适用场景：

- 补接口文档
- 调整接口说明
- 核对接口契约

### 11. 要处理复杂领域编排或业务分层问题

先看：

1. [docs/domain-biz-examples.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/domain-biz-examples.md)
2. [docs/architecture-dev-guide.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/architecture-dev-guide.md)

适用场景：

- 不确定逻辑放 `handler`、`biz`、`service` 还是 `repository`
- 涉及跨模块业务编排

## 最后执行顺序

接到开发任务后，默认按下面顺序处理：

1. 先判断任务属于哪一类
2. 先读对应文档
3. 再找仓库中最相近实现
4. 按现有风格改代码
5. 执行 `go build ./...`
6. 最后再提交
