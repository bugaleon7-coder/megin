# Shop-API 架构与开发规范

## 一、文档定位

本文只描述 Shop-API 中长期稳定的架构边界、演进方向和开发规范，不维护具体系统接口清单，也不记录迁移进度。

- 系统接口契约：`docs/system-api-contract.md`
- 系统重构计划：`docs/system-migration-plan.md`
- 系统功能范围：`docs/img.png`
- 当前仓库约束：`AGENTS.md`

发生冲突时按以下顺序判断：

1. `AGENTS.md`
2. 本文档的架构边界和演进原则
3. 已确认的接口契约
4. 原项目 Vue 接口行为
5. 当前前端实际调用
6. 实际数据库结构
7. 本文档的通用编码规范

## 二、整体架构

### 2.1 架构边界

项目按调用方分为两类入口：

- `/admin-api/*`：管理后台接口，面向运营、管理员和后台 Vue。
- `/api/*`：App 端业务接口，面向 C 端或移动端。

`system` 是 `admin-api` 下的系统管理通用模块，包含用户、角色、菜单、API、Casbin、字典等后台基础能力。它来源于开源项目适配，但本项目只兼容原项目 Vue 侧接口契约，不兼容、不照搬原项目后端代码架构。

`system` 的兼容范围：

- 兼容 URL、HTTP Method、请求参数字段、参数位置和响应 JSON 结构。
- 兼容 Vue 侧依赖的业务行为、错误语义、分页字段和树结构。
- 兼容成功 `code=200`、响应字段 `message` 的本项目外层响应约定。

`system` 不兼容的范围：

- 不保留 `GVA_MODEL`、`GVA_DB`、`global.GVA_*` 等 GVA 后端特定命名和全局状态。
- 不照搬原项目后端包结构、初始化方式、全局变量风格和框架封装。
- 不为了贴近原项目后端实现而破坏本项目的分层、事务、错误处理和测试规范。

命名原则：

- `SysUser`、`SysAuthority`、`SysMenu`、`SysApi` 等 `Sys` 前缀属于“系统管理”领域命名，可以保留。
- `GVA` 前缀或 `global.GVA_*` 这类来源项目标识不得进入本项目业务代码。
- 对外字段名以 Vue 接口契约为准，例如 `authorityId`、`pageSize`、`nickName`。
- 对内代码遵守本项目架构，优先清晰、可测试、可维护。

### 2.2 演进方向

后续开发按模块属性采用不同策略：

- `internal/system` 和 `internal/admin-api/system`：契约兼容优先。主要做稳定性、权限闭环、性能、事务和测试增强，不主动改 URL、字段、响应 `data` 结构或 Vue 依赖行为。
- `internal/admin-api` 的非 system 业务：作为管理后台最佳实践探索区。必须接入后台 JWT、Casbin、数据权限和必要的操作日志。
- `internal/api`：作为 App 端最佳实践探索区。认证、授权、DTO、限流、响应数据裁剪应独立于后台，不复用后台 handler 契约。
- `/api` 的全局限流、指定接口限流、MySQL 规则加载和内存热刷新统一参考 `docs/api-rate-limit-architecture.md`。
- `internal/module/{domain}`：承载普通业务领域能力。admin-api 和 api 可以共用领域 service/repository/model，但不应强行共用请求/响应 DTO。

关于 `domain` 目录的定义：

- `domain` 表示“按业务领域归档”的工程目录，不等同于本项目整体采用严格 DDD。
- 本项目当前采用的是“按领域归档 + 分层实现”的实用架构，而不是先完整引入聚合、领域事件、限界上下文等整套 DDD 战术模式。
- 只有当某个领域复杂到需要更强业务建模时，才在该领域内部逐步引入 DDD 的部分做法。

如果同一领域同时服务后台和 App：

- Handler 必须分开：`internal/admin-api/{domain}.go` 与 `internal/api/{domain}.go`。
- DTO 默认分开：后台 DTO 可包含管理字段，App DTO 只暴露客户端需要的字段。
- Service 可复用，但当后台规则和 App 规则不同，应拆出不同业务方法，复杂流程放入 `biz`。
- Repository 和 Model 可以复用，但 Repository 不承载调用方权限判断。

### 2.3 服务结构

项目使用一个 Gin Engine 承载路由：

```text
gin.Engine
├── /admin-api/*   管理后台接口
├── /api/*         App 端业务接口，不承载 system 模块
├── /swagger/*
├── /api-doc/*
└── /v3/api-docs/*
```

system 模块所有接口都使用 `/admin-api` 前缀。是否认证由中间件决定，不能根据前缀推断：

- 登录、验证码等公开接口：`/admin-api/*`，不挂认证中间件。
- 系统管理接口：`/admin-api/*`，挂 JWT、黑名单和 Casbin 中间件。

### 2.4 目录职责

```text
internal/
├── admin-api/
│   ├── article.go             管理后台业务 Handler 示例
│   └── system/                system 模块 Handler，兼容原项目 Vue 接口契约
├── api/                       App 端 Handler，不承载 system 模块
├── domain/                    普通业务领域目录，收口 article、customer 等模块
│   └── {domain}/
│       ├── biz/               可选，复杂业务流程编排层
│       ├── dto/               领域内部可复用 DTO，不直接等同于 admin/api 对外契约
│       ├── model/             GORM 模型
│       ├── repository/        数据访问
│       └── service/           领域业务逻辑
├── base/                      通用 Model、Service、Repository
├── middleware/                JWT、Casbin、日志等中间件
├── router/                    顶层路由编排
└── system/
    ├── dto/                   system 请求和响应 DTO，字段遵循 Vue 接口契约
    ├── model/                 GORM 模型
    ├── repository/            数据访问
    ├── router/                system 路由注册
    └── service/               业务逻辑

pkg/
├── bootstrap/                 服务启动
├── context/api/               API Context、Result、Error
├── context/router/            泛型路由注册和参数绑定
├── database/                  数据库、Redis
├── errs/                      业务错误类型
├── logger/                    日志
├── openapi/                   OpenAPI 生成
└── validate/                  参数校验扩展
```

### 2.5 依赖方向

```text
router → handler → biz → service → repository → model
```

`biz` 是可选层。简单 CRUD 可以省略：

```text
router → handler → service → repository → model
```

约束：

- Handler 不直接访问数据库。
- Handler 只能调用 `biz` 或本模块 `service`。
- Biz 是应用业务服务，负责完整业务场景和流程编排。
- Biz 可以调用多个 Service，并优先承载跨模块事务边界。
- Service 不依赖 Gin、HTTP Request 或 Response。
- Service 是领域业务服务，负责单一领域能力，原则上不调用其他模块 Service。
- Repository 不包含业务权限判断。
- Model 不依赖 Handler、Service 或 DTO。
- DTO 与 Model 分离，不直接将数据库模型作为管理接口请求体。

补充约束：

- `domain` 是普通业务的总收口目录，避免 `internal/` 根目录无限平铺。
- `system` 保持独立，不放入 `domain`，因为它的定位是后台通用兼容模块，不是普通业务域。
- 当前代码已开始向 `internal/module` 收口；后续不得再新增 `internal/core` 一类语义不清的业务目录。

### 2.6 system 与业务模块关系

`system` 可以向其他后台业务提供通用能力，例如当前登录用户、角色、数据权限、菜单和 Casbin 策略。但业务模块不能为了省事反向污染 system：

- 业务模块可以通过 `biz` 调用 system service 查询角色或数据权限。
- system 不应依赖具体业务模块，例如 customer、article 等。
- system 接口契约由 `docs/system-api-contract.md` 锁定。
- 新业务接口契约不应模仿 GVA 历史风格，应按本项目最佳实践设计。

## 三、统一响应和错误

### 3.1 外层响应

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

系统迁移只允许修改原版外层响应的两项：

- 成功 `code` 从 `0` 改为 `200`。
- `msg` 改为 `message`。

`data` 内部结构必须遵循 `docs/system-api-contract.md`，不得为了适配后端而随意修改前端。

### 3.2 Handler 返回方式

```go
return api.ResultSuccess()
return api.ResultData(data)
return nil, err
```

禁止：

- Handler 自行调用 `ctx.GinCtx.JSON`。
- Service 返回 `api.Result`。
- 捕获错误后返回伪成功。
- 在成功响应中塞入错误字符串。

### 3.3 错误分类

- 参数错误：请求无法绑定或校验失败。
- 业务错误：资源不存在、状态不允许、越权等预期错误。
- 系统错误：数据库、Redis、Casbin 或外部服务异常。

错误必须保留内部日志上下文，但响应不得暴露密码、Token、SQL DSN 或完整堆栈。

## 四、路由与参数绑定

### 4.1 路由注册

```go
func SysUserRouter(routeRegistry *router.RouteRegistry) *router.RouteRegistry {
    adminApiGroup := routeRegistry.Group("admin-api")
    adminApiGroup.Use(middleware.AdminApiAuthTokenRequired())
    adminApiGroup.Use(middleware.CasbinRequired())

    user := &handler.SysUser{}
    router.POST(adminApiGroup, "/user/getUserList", user.GetUserList)
    return routeRegistry
}
```

公开接口使用相同前缀，但不得挂认证中间件：

```go
publicGroup := routeRegistry.Group("admin-api")
router.POST(publicGroup, "/user/login", user.Login)
router.POST(publicGroup, "/user/captcha", user.Captcha)
```

实际中间件函数名以实现代码为准；中间件顺序必须保持：

```text
JWT 解析与黑名单 → Casbin 授权 → OperationRecord → Handler
```

### 4.2 参数来源

| HTTP 方法 | 默认参数来源 | 说明 |
|---|---|---|
| GET | Query | 使用 `form` 标签 |
| POST | JSON Body | 使用 `json` 标签 |
| PUT | JSON Body | 使用 `json` 标签 |
| DELETE | 由接口契约确定 | 必须支持原版 JSON Body 和 Query 两种形式 |

不得使用 HTTP 方法推测并更改原版参数位置。

Binder 必须支持：

- 匿名嵌入结构体
- 指针字段
- 数组和切片
- `time.Time`
- 空请求 DTO
- `binding` 校验

必填 DTO 收到空 Body 时必须返回参数错误；只有明确使用空 DTO 的接口可以接受空 Body。

### 4.3 DTO 标签

```go
type GetUserListReq struct {
    dto.PageQuery
    Username string `json:"username" form:"username" binding:"omitempty"`
    NickName string `json:"nickName" form:"nickName" binding:"omitempty"`
}
```

请求字段名称以接口契约和前端为准，不以 Go 字段名自动推导。

## 五、Handler 规范

### 5.1 位置与包名

- 目录：`internal/admin-api/system/`
- 包名：`system`
- 在 Router 中可以使用导入别名 `handler`

### 5.2 类型和方法签名

```go
// SysUser @Tag 系统用户管理
type SysUser struct{}

// @Summary 获取用户列表
// @Description 分页查询系统用户列表
func (this *SysUser) GetUserList(
    ctx *api.Context,
    req *systemDto.GetUserListReq,
) (*api.Result[dto.PageResult[systemDto.SysUser]], error) {
    result, err := systemService.NewSysUser(ctx).GetUserInfoList(req)
    if err != nil {
        return nil, err
    }
    return api.ResultData(*result)
}
```

Handler 负责：

- 从 `api.Context` 读取当前用户信息。
- 调用 Service。
- 将业务结果封装为统一响应。

Handler 不负责：

- 拼接 SQL。
- 开启业务事务。
- 实现树构建和权限计算。
- 手工重复执行参数校验。

公开接口中的 `ctx.AdminInfo` 可能为 `nil`，不得直接解引用。

## 六、Biz 规范

### 6.1 定位

`biz` 也是一种业务服务，但它在分层上不同于 `service`：

- `biz`：应用业务服务或流程服务，面向一个完整业务场景。
- `service`：领域业务服务或能力服务，面向单一领域能力。

推荐调用方向：

```text
handler → biz → service → repository
```

简单 CRUD 不强制创建 `biz`，可以保持：

```text
handler → service → repository
```

### 6.2 适用场景

出现以下任一情况时，应优先引入 `biz`：

- 一个接口需要编排多个 Service。
- 一个业务动作跨多个模块或多张表，需要统一事务。
- 需要组合权限上下文、数据权限、操作日志、通知等横切流程。
- 存在导入、导出、批量处理、状态流转等复杂流程。
- Service 之间开始互相调用，出现依赖网状化风险。

### 6.3 职责边界

Biz 负责：

- 串联多个 Service 完成一个完整业务动作。
- 定义跨 Service 的事务边界。
- 组织当前用户、角色、数据权限等上下文。
- 协调审计日志、通知、异步任务等流程。
- 将复杂流程从 Handler 和 Service 中剥离出来。

Biz 不负责：

- 直接拼接 SQL。
- 直接访问数据库替代 Repository。
- 承载单表简单 CRUD。
- 返回 `api.Result`。
- 处理 Gin、HTTP Request 或 Response。

### 6.4 与 Service 的关系

- Biz 可以调用多个 Service。
- Service 原则上只调用本模块 Repository。
- Service 不应为了完成跨模块流程而调用其他模块 Service；这类流程应上移到 Biz。
- 事务优先放在 Biz；只有单领域内部事务才放在 Service。
- Handler 只能调用 Biz 或本模块 Service，不能跳过 Biz 直接编排多个 Service。

示例目录：

```text
internal/module/customer/
├── biz/
├── dto/
├── model/
├── repository/
└── service/
```

示例流程：

```text
internal/admin-api/customer.go
  → internal/module/customer/biz
      → internal/module/customer/service
      → internal/system/service
      → internal/module/customer/repository
```

订单跨领域协作的完整示例单独维护在：

- `docs/domain-order-biz-example.md`
- `docs/domain-biz-examples.md`

后续凡是“一个主业务域同时协调多个其他领域”的复杂流程，默认参考该文档执行。

## 七、Service 规范

### 7.1 初始化

```go
type SysUser struct {
    base.Service
    Repo *repo.SysUser
}

func NewSysUser(ctx *api.Context) *SysUser {
    s := &SysUser{}
    s.Initialize(ctx)
    s.Repo = repo.NewSysUser(ctx)
    return s
}
```

### 7.2 职责

Service 负责：

- 业务规则和状态校验。
- 当前角色能否操作目标角色的权限检查。
- 本领域内的多 Repository 编排。
- 本领域内的事务边界。
- Model 与 DTO 转换。

Service 不负责跨模块流程编排；一旦需要多个模块协作，应上移到 Biz。

### 7.3 事务

以下操作必须使用事务：

- 创建用户并分配角色。
- 删除用户并清理关联关系。
- 角色复制、删除及权限变更。
- 菜单与角色、按钮、参数关联变更。
- API 更新并同步 Casbin。
- 字典及详情导入。
- 版本数据导入。

事务函数内必须使用传入的 `tx`，禁止重新从全局数据库获取连接。

禁止先删除旧关联、再在事务外逐条插入新关联。

## 八、Repository 规范

### 8.1 定义

```go
type SysUser struct {
    base.Repository[model.SysUser]
}

func NewSysUser(ctx *api.Context) *SysUser {
    r := &SysUser{}
    r.Initialize(ctx)
    return r
}
```

### 8.2 主键

通用 `GetById`、`DeleteById` 仅适用于主键列为 `id` 的模型。

特殊主键必须实现专用方法，例如：

```go
func (r *SysAuthority) GetByAuthorityID(id uint) (model.SysAuthority, error)
func (r *SysAuthority) DeleteByAuthorityID(id uint) error
```

禁止在通用 Repository 中写死 `id` 后用于所有模型。

### 8.3 查询规则

- 复杂 JOIN 必须明确写出所有参与表。
- 分页 Count 和数据查询必须使用相同过滤条件。
- 排序字段必须使用白名单，禁止直接拼接用户输入。
- 关联预加载只加载接口需要的数据。
- 禁止 N+1 递归查询构建大型树；优先一次查询后在内存构树。

## 九、Model 规范

### 9.1 基础模型

```go
type SystemModel struct {
    ID        uint           `gorm:"primarykey" json:"ID"`
    CreatedAt *time.Time     `gorm:"column:created_at" json:"createdAt"`
    UpdatedAt *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}
```

### 9.2 数据库映射

每个模型必须：

- 实现 `TableName()`。
- 表名和字段名与现有 `go_app_starter` 数据库一致。
- 通过 GORM DryRun 验证 INSERT、UPDATE 和 DELETE SQL。
- 不得因为其他项目存在审计字段就嵌入 `ControlBy`。

只有数据库真实存在 `create_by`、`update_by` 时才能使用 `ControlBy`。

示例：

```go
type SysUser struct {
    base.SystemModel
    UUID        string `gorm:"column:uuid" json:"uuid"`
    Username    string `gorm:"column:username" json:"userName"`
    Password    string `gorm:"column:password" json:"-"`
    AuthorityID uint   `gorm:"column:authority_id" json:"authorityId"`
}

func (SysUser) TableName() string { return "sys_users" }
```

Model 不能直接用作创建或更新请求 DTO，避免客户端修改内部字段。

## 十、分页与树结构

### 10.1 分页契约

```go
type PageQuery struct {
    Page     int `json:"page" form:"page" binding:"required,min=1"`
    PageSize int `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100"`
}

type PageResult[T any] struct {
    List     []T  `json:"list"`
    Total    int64 `json:"total"`
    Page     int `json:"page"`
    PageSize int `json:"pageSize"`
}
```

如原版接口返回完整树而不是分页结果，禁止强行改成 `PageResult`。

### 10.2 树结构

树构建建议：

1. 一次查询取得扁平数据。
2. 建立 `parentID → children` 映射。
3. 使用递归函数填充子节点。
4. 不得先把值复制到根数组后再修改 Map 中的另一份值。
5. 空子节点统一返回空数组或遵循原版结构，不能在同一接口中混用。

角色树、菜单树和字典树必须有两层以上的测试数据。

## 十一、认证、授权与日志

### 11.1 JWT

- 只接受配置指定的签名算法。
- 不允许硬编码 Secret 回退。
- 每次认证检查 JWT 黑名单。
- Token 作废后必须立即拒绝访问。
- 日志不得输出 Token、密码或验证码。

### 11.2 Casbin

- 使用共享、可刷新 Enforcer，不得每个请求重新初始化数据库 Adapter。
- Subject 使用十进制角色 ID，例如 `strconv.FormatUint`。
- API 路径和 HTTP 方法必须与路由契约一致。
- API 更新、删除和角色授权后必须刷新策略。

### 11.3 操作日志

- 只为需要审计的管理操作记录数据库日志。
- 密码、Token、验证码和敏感响应必须脱敏。
- 文件上传和超大 Body 只记录摘要。
- 记录失败不能改变主业务响应，但必须输出内部错误日志。

## 十二、验证要求

每次修改 Go 代码后至少执行：

```bash
go build ./...
```

按修改范围补充：

- 路由契约测试：路径、方法、认证类型。
- Binder 测试：Body、Query、DELETE、嵌套和切片。
- GORM DryRun 测试：表名和字段。
- Service 单元测试：业务校验和事务失败回滚。
- `gva_test` 集成测试：真实 MySQL 行为。
- JWT/Casbin 测试：未登录、黑名单、无权限和有权限。
- 前端冒烟测试：页面加载、查询、创建、更新和删除。

开发库 `go_app_starter` 只用于只读核对；破坏性测试只能使用独立 `go_app_starter_test`。

## 十三、提交前检查

1. 接口是否存在于 `system-api-contract.md`。
2. 路径、方法、请求位置和响应 `data` 是否一致。
3. 是否错误嵌入 `ControlBy`。
4. 是否需要事务。
5. 是否存在越权路径。
6. 是否记录或返回敏感数据。
7. 是否包含对应测试。
8. `go build ./...` 是否通过。
