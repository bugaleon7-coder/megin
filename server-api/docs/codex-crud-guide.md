# Codex 极简开发约定

本文不是项目设计文档，只是“我和 Codex 怎么协作”的最小约定。

目标：

- 让我提需求时更省话
- 让 Codex 收到 CRUD 需求后直接按统一方式落地

## 1. 最常用提示词

如果我要你写一个模块的 CRUD，直接用下面这句：

```text
按本项目规范，为 xxx 模块实现一套完整 CRUD。
要求：
1. 先判断这个模块属于 /api 还是 /admin-api，再按对应入口实现
2. 按现有最佳实践补齐 handler、dto、service、repository、model、router
3. 请求和响应字段写中文注释
4. 返回结构走项目统一响应
5. 如果改动了 `.go` 文件，执行 go build ./... 验证
参考：
- docs/api-article-best-practice.md
- internal/api/article.go
- internal/admin-api/article.go
```

如果我要你写后台模块，建议直接补一句：

```text
这是后台管理模块，不是 /api，是 /admin-api。
```

如果我要你写前台模块，建议直接补一句：

```text
这是前台业务模块，接口走 /api。
```

## 2. CRUD 需求最好再补充 6 个信息

如果信息齐全，Codex 基本可以直接开工：

```text
模块名：
用途：
是后台还是前台：
需要哪些字段：
列表筛选条件：
是否需要权限控制：
```

最小示例：

```text
按本项目规范，为品牌模块实现一套完整 CRUD。

模块名：brand
用途：后台品牌管理
是后台还是前台：后台，走 /admin-api
需要哪些字段：ID、名称、Logo、排序、状态、备注
列表筛选条件：名称、状态
是否需要权限控制：需要

要求：
1. 补齐分页列表、详情、新增、修改、删除
2. 请求和返回 DTO 都写中文注释
3. 参考 internal/admin-api/article.go
4. 如果改动了 `.go` 文件，执行 go build ./...
```

## 3. 如果我懒得写全，你默认这样理解

当我只说“帮我写 xxx 模块 CRUD”时，Codex 默认按下面规则执行：

1. 先在仓库里找最接近的现有实现，不新造风格
2. 先判断这个需求属于 `/api` 还是 `/admin-api`
3. `/api` 优先参考 [docs/api-article-best-practice.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/api-article-best-practice.md) 和 [internal/api/article.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/api/article.go)
4. `/admin-api` 再参考 [internal/admin-api/article.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/admin-api/article.go)
5. 返回结构默认走统一响应，成功为 `code=200`、字段为 `message`
6. 请求和响应不直接复用数据库 model，而是单独建 DTO
7. 缓存默认不启用，只有你明确要求 Redis 缓存或本地缓存时才加
8. 严禁参考 `system` 模块代码开发新业务，参考代码统一以 `article` 相关实现为准
9. 如果改动了 `.go` 文件，执行 `go build ./...`

如果这 9 条和我的本次需求冲突，以我本次消息为准。

## 4. Codex 写 CRUD 的默认最佳实践

收到 CRUD 需求后，默认按下面方式实现：

### 4.1 目录落点

普通业务模块默认结构：

```text
internal/
├── admin-api/xxx.go
└── module/xxx/
    ├── dto/
    ├── model/
    ├── repository/
    └── service/
```

如果是前台模块，则 handler 落到 `internal/api/xxx.go`。

### 4.2 Handler 层只做 4 件事

1. 接收参数
2. 调用 service
3. 返回统一结果
4. 在必要时接入限流、锁等上下文能力；缓存只有在你明确提出时才接

默认不在 handler 里做复杂业务判断，不直接查库。

### 4.3 默认会补齐的接口

一个标准 CRUD 默认包括：

1. 分页列表 `PageList`，路径 `/pageList`
2. 详情 `Detail`
3. 新增 `Create`
4. 修改 `Update`
5. 删除 `Delete`
6. 不分页列表 `List`，路径 `/list`

如果你不想要详情或删除，需要明确说。

### 4.4 DTO 默认拆分

默认至少拆成这些 DTO：

1. 列表查询 DTO
2. 详情查询 DTO
3. 新增 DTO
4. 更新 DTO
5. 列表项/详情响应 DTO

这样方便后续控字段、加校验、写注释。

### 4.5 注释默认要求

我会默认这样处理注释：

1. handler 方法写中文接口说明
2. 请求 DTO 字段写中文注释
3. 返回 DTO 字段写中文注释
4. 复杂逻辑处补简短中文注释

### 4.6 缓存默认规则

默认不主动加缓存。

只有你明确说下面这些指令时，我才会加缓存：

1. 加 Redis 缓存
2. 加本地缓存
3. 详情接口加缓存
4. 这个接口需要缓存

`internal/api/article.go` 里的缓存写法只是演示能力，不代表普通 CRUD 默认要接缓存。

### 4.7 更新和删除的默认处理

如果仓库里同类模块已有锁或防重复提交写法，优先保持一致。

如果没有特殊约束，默认至少保证：

1. 先校验数据是否存在
2. 再执行更新或删除
3. 不在 handler 中写跨层业务逻辑

### 4.8 权限和中间件

如果你说“这是后台管理模块”，我会主动检查：

1. 路由是否挂到 `/admin-api`
2. 是否需要登录态
3. 是否需要 Casbin 权限控制
4. 是否要对齐现有后台路由注册方式

## 5. 你可以直接复制的几种指令

### 5.1 标准 CRUD

```text
按本项目规范，为 xxx 模块实现一套完整 CRUD。
这是后台模块，接口走 /admin-api。
字段有：xxx、xxx、xxx。
列表支持按 xxx、xxx 筛选。
请求和返回字段写中文注释。
如果改动了 `.go` 文件，执行 go build ./...。
```

### 5.2 只补列表和新增

```text
按本项目规范，为 xxx 模块补列表和新增接口。
这是后台模块，接口走 /admin-api。
参考现有 article 模块写法。
如果改动了 `.go` 文件，执行 go build ./...。
```

### 5.3 按数据库表生成 CRUD

```text
按本项目规范，基于 xxx 表为 xxx 模块实现 CRUD。
这是后台模块，接口走 /admin-api。
你先查看表结构，再补齐 model、dto、service、repository、handler、router。
请求和返回字段写中文注释。
如果改动了 `.go` 文件，执行 go build ./...。
```

### 5.4 修改已有 CRUD

```text
按本项目规范，调整 xxx 模块 CRUD。
只改 xxx，不要改接口路径和返回结构。
先阅读现有实现后再改。
如果改动了 `.go` 文件，执行 go build ./...。
```

## 6. 我不想来回解释时，可以再加这句

```text
不要先给方案，直接先查仓库相似实现，然后改代码，最后把修改结果和验证结果告诉我。
```

这句适合你想让我直接动手，不要先讲一大段分析的时候使用。
