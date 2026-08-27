# Shop-API 新手引导

本文给第一次参与 `Shop-API` 开发的同学使用，目标只有一个：尽快进入“能跑、能改、能按规范提交”的状态。

如果本文和仓库其他文档有冲突，按下面顺序判断：

1. [AGENTS.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/AGENTS.md)
2. [docs/architecture-dev-guide.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/architecture-dev-guide.md)
3. 具体模块契约文档
4. 现有稳定实现

## 1. 先理解项目是怎么分层的

这个项目不是把所有接口都平铺在 `internal/` 下面，而是按“调用入口 + 业务领域 + 分层职责”组织代码。

- `/admin-api/*`：后台管理接口，给管理端用。
- `/api/*`：前台业务接口，给 App 或客户端用。
- `internal/system`：后台系统管理通用模块，主要承载用户、角色、菜单、权限等能力。
- `internal/module`：普通业务域目录，例如文章、客户、用户等业务。

核心依赖方向：

```text
router -> handler -> biz -> service -> repository -> model
```

需要记住的规则：

- `handler` 只处理请求、调用服务、返回结果。
- `service` 写业务逻辑，不直接依赖 Gin。
- `repository` 只做数据访问，不写权限判断。
- 不要在 `handler` 里直接查数据库。
- 不要把数据库模型直接当接口入参和出参。

建议先看这几份文件建立直觉：

- [docs/architecture-dev-guide.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/architecture-dev-guide.md)
- [docs/api-article-best-practice.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/api-article-best-practice.md)
- [internal/admin-api/article.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/admin-api/article.go)
- [internal/router/admin_api.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/router/admin_api.go)

## 2. 第一天先把项目跑起来

### 2.1 环境要求

- Go `1.18+`
- MySQL
- Redis

### 2.2 初始化数据库

```shell
mysql -uroot -p123456 -e "create database if not exists gva default charset utf8mb4 collate utf8mb4_unicode_ci;"
mysql -uroot -p123456 gva < docs/sql/schema.sql
```

### 2.3 配置文件位置

配置文件都在 `config/` 目录：

- [config/config-dev.yaml](/Users/lchb/go_admin/gin-vue-admin/shop-api/config/config-dev.yaml)
- [config/config-test.yaml](/Users/lchb/go_admin/gin-vue-admin/shop-api/config/config-test.yaml)
- [config/config-prod.yaml](/Users/lchb/go_admin/gin-vue-admin/shop-api/config/config-prod.yaml)

默认本地开发使用：

```shell
config/config-dev.yaml
```

### 2.4 本地启动方式

推荐直接用 Go 命令启动，排查问题更直接：

```shell
go run main.go -env=dev
go run cmd/api/main.go -env=dev
go run cmd/admin-api/main.go -env=dev
```

说明：

- `main.go`：混合模式，同时提供 `/api` 和 `/admin-api`
- `cmd/api/main.go`：只启动前台接口
- `cmd/admin-api/main.go`：只启动后台接口

也可以用脚本：

```shell
sh server.sh run --env=dev
```

### 2.5 启动后先访问这两个地址

混合模式默认看下面两个文档入口：

- 前台接口文档：`http://localhost:8800/api-doc/`
- 后台接口文档：`http://localhost:8800/admin-api-doc/`

如果文档打不开，优先检查：

- 配置里的端口是否被占用
- MySQL 和 Redis 是否可连接
- 启动日志里是否有初始化失败信息

## 3. 新需求应该写到哪里

### 3.1 先判断是后台还是前台

- 管理后台功能，写到 `/admin-api`
- App 或客户端功能，写到 `/api`

这是最容易犯错的地方。不要把后台接口写到 `internal/api`，也不要把前台接口写到 `internal/admin-api`。

### 3.2 再判断是不是系统模块

如果需求属于这些范围，优先放到 `system`：

- 管理员管理
- 角色权限
- 菜单管理
- API 权限
- Casbin 权限策略
- 字典、系统配置、登录日志等后台基础设施

如果是普通业务，例如文章、客户、商品、订单，优先放到 `internal/module/{domain}`，再由对应 `handler` 暴露接口。

### 3.3 一个新业务模块的常见落点

```text
internal/
├── admin-api/demo.go           后台 Handler
├── api/demo.go                 前台 Handler
└── module/demo/
    ├── dto/
    ├── model/
    ├── repository/
    └── service/
```

## 4. 写接口时必须遵守的规范

### 4.1 所有后台接口必须以 `admin-api` 开头

这是当前项目明确约束。后台新接口不允许挂到其他前缀。

### 4.2 返回结构必须走统一响应

成功响应统一为：

```json
{
  "code": 200,
  "message": "成功",
  "data": {},
  "trace": [],
  "trace_id": "",
  "success": true
}
```

注意：

- 成功码是 `200`，不是旧项目里的 `0`
- 字段名是 `message`，不是 `msg`

`handler` 层统一这样返回：

```go
return api.ResultSuccess()
return api.ResultData(data)
return nil, err
```

禁止做法：

- 直接 `ctx.GinCtx.JSON(...)`
- `service` 返回 HTTP 响应结构
- 出错后返回伪成功

### 4.3 请求和响应 DTO 要单独定义

请求参数和响应结果要放到 `dto` 中，不要直接复用数据库模型。

这样做有三个原因：

- 接口字段和数据库字段不是一回事
- 后续加校验、裁剪字段更容易
- 可以避免把数据库内部字段直接暴露出去

### 4.4 注释必须写清楚字段含义

仓库已有明确要求：接口请求和返回字段必须写明。

新代码至少做到：

- 结构体中文注释说明用途
- 每个关键字段说明业务含义
- handler 方法写清楚接口作用

### 4.5 参考现成最佳实践，不要凭感觉写

优先参考：

- [internal/admin-api/article.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/admin-api/article.go)
- [docs/api-article-best-practice.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/api-article-best-practice.md)

如果是 system 模块，优先参考：

- [internal/system/router](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/router)
- [internal/system/service](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service)
- [internal/admin-api/system](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/admin-api/system)

## 5. 路由、中间件和权限怎么处理

### 5.1 路由不要手写散落注册

路由统一收口在：

- [internal/router/admin_api.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/router/admin_api.go)
- [internal/router/api.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/router/api.go)
- [internal/system/router](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/router)

新增接口时先看当前模块的路由注册方式，保持一致。

### 5.2 公开接口和受保护接口要分组

后台接口虽然都走 `/admin-api`，但不是所有接口都必须鉴权。

常见区分：

- 登录、验证码：公开接口
- 用户、角色、菜单、权限管理：鉴权接口

### 5.3 后台接口要特别注意权限链路

后台接口通常要经过：

```text
JWT 解析/黑名单 -> Casbin 鉴权 -> 业务 Handler
```

你需要确认三件事：

1. 这个接口是否要求登录
2. 这个接口是否要做 Casbin 权限控制
3. 这个接口是否会影响操作日志或权限数据

如果权限没接对，接口“能调通”也不算完成。

## 6. 写代码时的实用约束

### 6.1 先找已有实现，再写新代码

建议顺序：

1. 先 `rg` 搜同类接口
2. 找一个最接近的 handler
3. 顺着看 service、repository、dto、router
4. 在原有模式上扩展

不要一上来就新造一套写法。

### 6.2 尽量复用项目已有基础能力

优先使用现成封装：

- [pkg/context/api](/Users/lchb/go_admin/gin-vue-admin/shop-api/pkg/context/api)
- [pkg/context/router](/Users/lchb/go_admin/gin-vue-admin/shop-api/pkg/context/router)
- [pkg/errs](/Users/lchb/go_admin/gin-vue-admin/shop-api/pkg/errs)
- [pkg/validate](/Users/lchb/go_admin/gin-vue-admin/shop-api/pkg/validate)
- [pkg/cache](/Users/lchb/go_admin/gin-vue-admin/shop-api/pkg/cache)

不要在业务代码里重复造轮子，例如重复封装响应、重复绑定参数、重复写缓存基础逻辑。

### 6.3 查询、创建、更新、删除的常见分工

- 查询接口：handler 收参，service 组装结果，repository 查数据
- 创建接口：service 校验业务条件，repository 落库
- 更新接口：先确认资源是否存在，再做状态校验和更新
- 删除接口：先确认是否允许删除，再执行删除

### 6.4 能写测试就不要裸提代码

优先补这几类测试：

- service 业务测试
- router/contract 契约测试
- 关键 handler 的请求测试

可以先参考：

- [internal/system/router/contract_test.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/router/contract_test.go)
- [internal/system/service](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service) 下的测试文件

## 7. system 模块开发要特别注意什么

如果你参与的是 system 迁移或扩展开发，额外注意以下约束：

- 只兼容原项目 Vue 依赖的接口契约，不照搬原项目后端架构
- `GVA_MODEL` 要替换为当前项目自己的 `SystemModel`
- 允许调整内部实现，但不要改前端依赖的 URL、字段名、参数位置和 `data` 结构
- 成功响应外层只能改 `code=200` 和 `message`
- 所有 system 后台接口仍然必须使用 `/admin-api`

简单说就是：

“内部按本项目规范重写，对外保持兼容。”

## 8. 提交前必须自检什么

这是新同学最容易漏掉的一步。

### 8.1 最少自检清单

1. 能启动服务
2. 变更接口能实际调通
3. Swagger 或接口文档可访问
4. 相关权限链路正常
5. `go build` 能通过

### 8.2 必跑命令

仓库已有明确要求：改完代码必须做编译验证。

至少执行：

```shell
go build ./...
```

如果只想验证后台入口，也至少补一次：

```shell
go build ./cmd/admin-api
```

### 8.3 提交前再检查一遍这些问题

- 有没有把后台接口写成 `/api`
- 有没有直接返回旧格式 `msg`
- 有没有在 handler 里写复杂业务
- 有没有漏加鉴权或 Casbin
- DTO 注释和字段说明是否写清楚
- 是否误把数据库模型当接口出参

## 9. 推荐阅读顺序

如果你只有半天上手时间，按这个顺序看：

1. [README.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/README.md)
2. [AGENTS.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/AGENTS.md)
3. [docs/architecture-dev-guide.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/architecture-dev-guide.md)
4. [docs/api-article-best-practice.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/api-article-best-practice.md)
5. [internal/admin-api/article.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/admin-api/article.go)
6. 你当前要改的目标模块

如果你要做 system 迁移，再加看：

1. [docs/system-api-contract.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/system-api-contract.md)
2. [docs/system-migration-plan.md](/Users/lchb/go_admin/gin-vue-admin/shop-api/docs/system-migration-plan.md)
3. [internal/admin-api/system](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/admin-api/system)
4. [internal/system](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system)

## 10. 一句话工作流

进入一个需求后，建议固定使用下面这个节奏：

1. 先判断入口：`/api` 还是 `/admin-api`
2. 找同类实现，沿现有分层扩展
3. 补 DTO、service、repository、router
4. 处理鉴权、Casbin、日志等中间件
5. 本地调通后执行 `go build ./...`
6. 再提交代码

这套流程执行稳定之后，基本就不会偏离项目规范。
