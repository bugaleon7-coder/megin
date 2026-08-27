# 后台认证与 Casbin 授权原理

本文说明当前项目 `admin-api` 后台接口的权限验证链路，重点解释：

- JWT 身份认证的原理
- JWT 黑名单的作用
- Casbin 接口级权限控制的原理
- 当前项目如何兼容原项目的稳定实现

## 1. 总体链路

后台受保护接口统一走 `/admin-api/*` 的 protected group。

当前权限验证顺序：

1. 读取 Token
2. 检查 Token 是否在黑名单
3. 解析 JWT，提取用户与角色信息
4. 使用 Casbin 校验当前角色是否有权限访问目标接口
5. 校验通过后进入 Handler

对应实现入口：

- JWT 与 Casbin 串联中间件：
  [internal/middleware/adminapi_auth.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/middleware/adminapi_auth.go:1)
- 后台 group 创建：
  [internal/router/admin_api.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/router/admin_api.go:1)

这条链路的职责划分非常明确：

- JWT 负责确认“你是谁”
- 黑名单负责确认“这个 token 现在还能不能用”
- Casbin 负责确认“你能不能访问这个接口”

## 2. JWT 身份认证原理

JWT 是后台用户登录后的身份凭证。请求进入后台接口时，系统会先读取 Token 并校验其合法性。

### 2.1 Token 读取顺序

Token 提取逻辑在：
[internal/middleware/token.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/middleware/token.go:1)

按以下顺序读取：

1. `Authorization: Bearer <token>`
2. `X-Token` / `x-token` / `Token`
3. Cookie 中的 `x-token`

这样做是为了兼容不同前端或调用方式。

### 2.2 JWT 解析内容

解析成功后，会得到 claims：

- `UserID`
- `Username`
- `RoleId`

定义位置：
[internal/module/common/dto/base.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/module/common/dto/base.go:67)

这里最关键的是 `RoleId`，因为后续 Casbin 授权要依赖当前角色 ID。

### 2.3 JWT 在权限系统中的职责

JWT 只负责身份认证，不负责接口授权。

也就是说，JWT 只能证明：

- 当前请求对应哪个用户
- 当前请求对应哪个角色
- Token 本身签名是否合法、是否过期

JWT 不能决定：

- 当前角色是否有权限访问 `/authority/createAuthority`
- 当前角色是否能调用 `/api/createApi`

这些都属于 Casbin 的职责。

## 3. JWT 黑名单原理

JWT 天生是无状态的。只要签名没过期，从纯 JWT 角度看它就是合法的。

但后台管理系统必须支持：

- 主动退出登录
- 强制踢下线
- 多端登录时让旧 token 失效

因此需要黑名单机制。

### 3.1 黑名单存储

当前项目使用数据库表 `jwt_blacklists` 存储已失效 token。

模型定义：
[internal/system/model/sys_jwt_blacklist.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/model/sys_jwt_blacklist.go:1)

仓储查询：
[internal/system/repository/sys_jwt_blacklist.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/repository/sys_jwt_blacklist.go:1)

服务逻辑：
[internal/system/service/sys_jwt.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service/sys_jwt.go:1)

### 3.2 黑名单接口

后台接口：

- `POST /admin-api/jwt/jsonInBlacklist`

Handler 位置：
[internal/admin-api/system/sys_jwt.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/admin-api/system/sys_jwt.go:1)

这个接口会把当前 token 或指定 token 写入 `jwt_blacklists`。

### 3.3 黑名单校验时机

中间件在解析 JWT 之前先查黑名单。

实现位置：
[internal/middleware/adminapi_auth.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/middleware/adminapi_auth.go:21)

处理顺序采用原项目逻辑：

1. 先取 token
2. 先查黑名单
3. 再解析 JWT

这样可以保证：

- 已失效 token 会被尽早拒绝
- 即使 JWT 本身签名仍然合法，只要被拉黑也不能继续访问

### 3.4 黑名单的意义

黑名单解决的是“主动失效”问题。

示例：

- 用户登录后拿到 token
- 用户退出登录，token 被写入黑名单
- 之后该 token 再访问后台接口，会在中间件阶段被拒绝

因此：

- JWT 负责“这个 token 原本是不是有效”
- 黑名单负责“这个 token 现在是否仍被允许使用”

## 4. Casbin 授权原理

Casbin 负责后台接口级别的 RBAC 授权。

它要回答的问题是：

- 当前角色是否允许访问当前接口

### 4.1 Casbin 三元组

当前项目使用的核心授权三元组是：

- `sub`：访问主体，当前为角色 ID
- `obj`：访问对象，当前为接口路径
- `act`：访问动作，当前为 HTTP Method

例如：

- 角色：`888`
- 路径：`/user/getUserList`
- 方法：`POST`

会形成如下授权判断：

- `sub = "888"`
- `obj = "/user/getUserList"`
- `act = "POST"`

### 4.2 Casbin model

当前项目的 Casbin model 定义在：
[internal/system/service/sys_casbin.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service/sys_casbin.go:1)

核心规则：

```text
r = sub, obj, act
p = sub, obj, act
m = r.sub == p.sub && keyMatch2(r.obj,p.obj) && r.act == p.act
```

含义：

- 请求角色必须等于策略角色
- 请求路径必须匹配策略路径
- 请求方法必须一致

这里使用 `keyMatch2`，保留了路径模式匹配能力，但当前后台大多数接口仍然是固定路径匹配。

## 5. 为什么 Casbin 路径要去掉 `/admin-api` 前缀

这是当前项目兼容原项目时最关键的点之一。

### 5.1 当前数据库中的策略格式

当前库里的 `casbin_rule.v1` 和 `sys_apis.path` 存储的是原项目风格的路径：

- `/user/getUserList`
- `/authority/getAuthorityList`
- `/api/createApi`

不是：

- `/admin-api/user/getUserList`

这和当前接口对外实际路径不同，但和原项目稳定方案一致。

### 5.2 当前项目的处理方式

因此请求真正进入 Casbin 之前，必须先把 `/admin-api` 前缀裁掉。

实现位置：
[internal/system/service/sys_casbin.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service/sys_casbin.go:1)

标准化函数：

- `NormalizeCasbinPath("/admin-api/user/getUserInfo")`
- 结果：`"/user/getUserInfo"`

### 5.3 为什么必须这样做

如果不去掉 `/admin-api`：

- 请求路径会是 `/admin-api/user/getUserInfo`
- 数据库策略路径是 `/user/getUserInfo`
- Casbin 会全部匹配失败

所以这里不是“风格选择”，而是兼容原项目策略数据结构的必要逻辑。

## 6. Casbin 在当前项目中的执行方式

Casbin 授权是在后台认证中间件内部执行的。

入口位置：
[internal/middleware/adminapi_auth.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/middleware/adminapi_auth.go:62)

调用流程：

1. JWT 解析成功后拿到 `claims.RoleId`
2. 读取当前请求路径和 HTTP Method
3. 调用 `EnforceAuthorityPolicy(roleId, path, method)`
4. `SysCasbin` 内部先标准化路径，再调用 `enforcer.Enforce(...)`
5. 返回 `true` 则放行，返回 `false` 则拒绝

拒绝时返回：

- `403`
- `权限不足`

所以当前后台受保护接口的访问条件是：

1. token 存在
2. token 不在黑名单
3. JWT 解析成功
4. 当前角色在 Casbin 中拥有该接口权限

四者缺一不可。

## 7. 当前项目的共享 Enforcer 机制

P1 优化之后，Casbin 不再每次请求都重新初始化。

### 7.1 优化前的问题

旧实现里每次调用都重复：

- 创建 gorm adapter
- 创建 Casbin model
- 创建 enforcer
- `LoadPolicy()`

这会导致：

- 请求级鉴权性能差
- 资源开销大
- 无法稳定作为全局中间件复用

### 7.2 优化后的实现

现在在：
[internal/system/service/sys_casbin.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service/sys_casbin.go:1)

使用了进程级共享 Enforcer：

- 首次访问时初始化一次
- 后续请求复用同一个实例
- 策略变更后显式 `LoadPolicy()`
- 使用读写锁保护并发读写

当前锁策略：

- `Enforce` 和策略查询走读锁
- `UpdateCasbin`、`SetApiAuthorities`、`UpdateCasbinApi`、`FreshCasbin` 走写锁

这样可以保证：

- 普通鉴权请求不重复初始化 Enforcer
- 策略更新时不会和并发鉴权互相踩数据

## 8. Casbin 策略数据从哪里来

当前项目中的策略数据主要来自三类场景：

### 8.1 角色创建时的默认权限

默认策略种子定义在：
[internal/system/dto/common.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/dto/common.go:70)

这些默认策略仍然遵循原项目风格，使用无 `/admin-api` 前缀的路径。

### 8.2 后台 API 权限管理

接口管理相关逻辑位置：

- Handler：
  [internal/admin-api/system/sys_api.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/admin-api/system/sys_api.go:1)
- Service：
  [internal/system/service/sys_casbin.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service/sys_casbin.go:1)

当后台修改某个 API 允许哪些角色访问时，会更新 `casbin_rule` 中对应的记录。

### 8.3 角色复制、角色权限变更

相关逻辑位置：
[internal/system/service/sys_authority.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service/sys_authority.go:1)

角色复制或删除时，会同步维护 Casbin 策略。

## 9. JWT、黑名单、Casbin 三者的职责边界

这三层不能混淆。

### 9.1 JWT

负责：

- 识别当前登录用户
- 识别当前角色
- 校验 token 是否过期、签名是否合法

不负责：

- 判断接口权限

### 9.2 黑名单

负责：

- 主动让某个本来合法的 token 失效

不负责：

- 用户身份识别
- 接口权限判断

### 9.3 Casbin

负责：

- 判断当前角色是否有权限访问当前接口

不负责：

- 判断 token 是否过期
- 判断 token 是否被主动拉黑

## 10. 典型场景说明

### 10.1 token 合法，但角色无权限

结果：

- JWT 通过
- 黑名单通过
- Casbin 拒绝

表现：

- 返回 `403`
- 提示 `权限不足`

### 10.2 token 曾经合法，但已被拉黑

结果：

- 黑名单直接拒绝
- JWT 解析和 Casbin 都不会继续执行

表现：

- 返回 token 失效相关错误

### 10.3 token 合法，角色也有权限

结果：

- JWT 通过
- 黑名单通过
- Casbin 通过
- 进入业务 Handler

## 11. 当前实现与原项目的兼容原则

当前项目在权限链路上遵循的是“兼容原项目稳定逻辑，而不是随意重写”。

具体体现在：

- 黑名单检查顺序与原项目一致：先查黑名单，再解析 JWT。
- Casbin 授权核心模型与原项目一致：`role + path + method`。
- Casbin 路径存储与匹配方式兼容原项目：策略路径不带 `/admin-api` 前缀。
- 后台 protected 路由统一挂 JWT + Casbin，和原项目 PrivateGroup 的处理思路一致。

同时，当前项目在工程实现上做了符合本项目架构的优化：

- 使用统一 `api.Result` 响应结构
- 使用本项目 `admin-api` 前缀对外暴露接口
- 使用共享 Enforcer 代替每次重新初始化
- 使用本项目的 handler/service/repository 分层

因此可以概括为：

- 对外兼容原项目接口行为
- 对内按本项目架构实现

## 12. 相关文件索引

后台认证与授权主入口：

- [internal/middleware/adminapi_auth.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/middleware/adminapi_auth.go:1)

Token 读取与 JWT 解析：

- [internal/middleware/token.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/middleware/token.go:1)
- [internal/module/common/dto/base.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/module/common/dto/base.go:67)

Casbin 核心实现：

- [internal/system/service/sys_casbin.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service/sys_casbin.go:1)
- [internal/system/service/sys_casbin_test.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service/sys_casbin_test.go:1)

JWT 黑名单：

- [internal/system/model/sys_jwt_blacklist.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/model/sys_jwt_blacklist.go:1)
- [internal/system/repository/sys_jwt_blacklist.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/repository/sys_jwt_blacklist.go:1)
- [internal/system/service/sys_jwt.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service/sys_jwt.go:1)
- [internal/admin-api/system/sys_jwt.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/admin-api/system/sys_jwt.go:1)

后台路由分组：

- [internal/router/admin_api.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/router/admin_api.go:1)
- [internal/router/router.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/router/router.go:1)

## 13. 菜单权限、按钮权限、数据权限与 Casbin API 权限的关系

这四类权限都属于“后台权限体系”，但它们控制的对象不同、作用层级不同，不能混为一谈。

可以先给出一句总定义：

- Casbin API 权限：控制“这个角色能不能调用这个接口”
- 菜单权限：控制“这个角色在后台能不能看到和进入这个功能入口”
- 按钮权限：控制“这个角色在某个页面里能不能使用某个具体操作”
- 数据权限：控制“这个角色虽然进入了页面、调到了接口，但最终能看到或操作哪些数据”

它们是互补关系，不是替代关系。

### 13.1 四类权限的控制维度

#### 13.1.1 Casbin API 权限

Casbin 控制的是接口访问权。

判断对象是：

- 角色 ID
- 接口路径
- HTTP Method

它回答的问题是：

- 当前角色能不能调用 `/authority/createAuthority`
- 当前角色能不能访问 `/user/getUserList`

这是最底层、最硬的后端访问控制。

只要 Casbin 拒绝：

- 前端页面是否可见已经不重要
- 按钮是否显示也不重要
- 请求会直接在中间件阶段被拒绝

所以 Casbin 是后台接口安全的最后一道硬校验。

#### 13.1.2 菜单权限

菜单权限控制的是“页面入口可见性”和“功能导航范围”。

它回答的问题是：

- 当前角色左侧菜单里能不能看到“用户管理”
- 当前角色能不能进入“角色管理”页面
- 当前角色默认首页是什么

菜单权限本质上偏前后台协同：

- 后端返回角色能访问的菜单树
- 前端按菜单树渲染导航

所以菜单权限更多解决“功能暴露面”的问题，不直接等于后端安全。

即使某个角色看不到菜单，如果它手工构造请求：

- 仍然要由 Casbin 再做一次接口级兜底校验

换句话说：

- 菜单权限决定“看不看得到入口”
- Casbin 决定“即使进来了，接口能不能调”

#### 13.1.3 按钮权限

按钮权限控制的是页面内部操作粒度。

它回答的问题是：

- 在“用户管理”页面中，当前角色能不能点“新增”
- 能不能点“删除”
- 能不能点“编辑”
- 能不能点某个业务自定义操作

按钮权限通常挂在某个菜单或页面之下，因此它和菜单权限的关系是：

- 先有菜单页面访问权
- 再在页面内部细分按钮操作权

按钮权限主要服务前端界面控制和更细粒度的操作展示。

但它也不是最终安全边界：

- 用户即使前端看不到按钮，仍可能直接发请求
- 后端依然要靠 Casbin 或业务校验兜底

所以按钮权限解决的是“操作入口是否展示”，不是“接口是否绝对安全可访问”。

#### 13.1.4 数据权限

数据权限控制的是数据访问范围。

它回答的问题是：

- 当前角色能看到全部客户，还是只能看到自己部门的客户
- 当前角色能修改哪些记录
- 当前角色能查询哪些组织、用户、订单、客户

数据权限是在“接口可访问”的前提下，再进一步缩小数据作用域。

也就是说：

- Casbin 先决定能不能访问 `/customer/customerList`
- 数据权限再决定能返回哪些 customer

这类权限往往体现在 service/biz 层查询条件里，而不是中间件里。

例如：

- 角色 A 可访问客户列表接口
- 但只能看到自己有 `dataAuthority` 范围内的数据

因此数据权限是业务数据层面的约束，不是路由层面的约束。

### 13.2 四类权限分别在哪一层生效

从系统层次看，可以这样理解：

#### 13.2.1 Casbin API 权限生效在中间件层

位置：

- [internal/middleware/adminapi_auth.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/middleware/adminapi_auth.go:1)

特点：

- 请求到达 Handler 之前执行
- 拒绝则直接返回 `403`
- 不进入业务逻辑

#### 13.2.2 菜单权限生效在菜单树构建与页面导航层

相关逻辑主要在：

- [internal/system/service/sys_menu.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service/sys_menu.go:1)
- [internal/system/repository/sys_base_menu.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/repository/sys_base_menu.go:1)

特点：

- 后端根据角色返回菜单树
- 前端根据菜单树决定页面入口展示

#### 13.2.3 按钮权限生效在页面操作层

相关逻辑主要在：

- [internal/system/service/sys_authority_btn.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service/sys_authority_btn.go:1)
- [internal/system/repository/sys_authority_btn.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/repository/sys_authority_btn.go:1)

特点：

- 返回某角色在某菜单下的按钮权限
- 前端根据按钮权限决定操作按钮是否展示

#### 13.2.4 数据权限生效在 service/biz 查询和业务校验层

典型使用点可以参考：

- [internal/module/customer/service/customer.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/module/customer/service/customer.go:1)
- [internal/system/service/sys_authority.go](/Users/lchb/go_admin/gin-vue-admin/shop-api/internal/system/service/sys_authority.go:1)

特点：

- 不是中间件统一判断
- 通常由业务 service/biz 根据当前角色的数据权限范围拼装查询条件

### 13.3 四类权限之间的执行顺序

实际请求或页面访问中，常见顺序大致如下。

#### 13.3.1 用户进入后台页面

1. 用户登录成功
2. 后端根据角色返回菜单树
3. 前端展示允许访问的菜单
4. 页面内部根据按钮权限决定显示哪些操作按钮

这个阶段主要是：

- 菜单权限
- 按钮权限

#### 13.3.2 用户点击页面并发起接口请求

1. 请求进入后台中间件
2. JWT 校验
3. 黑名单校验
4. Casbin API 权限校验
5. 进入 Handler / Biz / Service
6. 在业务层再做数据权限过滤

这个阶段主要是：

- Casbin API 权限
- 数据权限

所以从执行时机上说：

- 菜单权限和按钮权限偏“前台展示控制”
- Casbin 和数据权限偏“后台真实访问控制”

### 13.4 为什么不能只靠菜单权限或按钮权限

这是权限设计里最容易犯错的地方。

如果只做菜单和按钮权限：

- 前端页面看起来被限制了
- 但接口本身仍可能被直接调用

例如：

- 某角色看不到“删除用户”按钮
- 但它如果直接发 `DELETE /admin-api/user/deleteUser`
- 如果后端没有 Casbin 或业务校验，就仍可能删成功

因此：

- 菜单权限不是安全边界
- 按钮权限也不是安全边界

真正的安全边界必须在后端：

- Casbin 兜住接口访问
- 业务层兜住数据范围

### 13.5 为什么不能只靠 Casbin

反过来也不能只靠 Casbin。

如果只有 Casbin：

- 后端安全性可能是够的
- 但前端用户体验会很差

因为前端仍可能看到大量自己不能使用的菜单和按钮：

- 点进去才发现没权限
- 操作时才弹 `403`

这会导致：

- 页面噪音大
- 用户困惑
- 权限边界不直观

所以：

- Casbin 负责安全
- 菜单/按钮权限负责可用性和界面体验

### 13.6 为什么数据权限不能并入 Casbin

数据权限和 Casbin 的粒度不同。

Casbin 当前做的是：

- 角色是否可访问某接口

而数据权限通常是：

- 角色可访问这个接口，但只能看到某部分数据

例如：

- `/customer/customerList` 这个接口可以访问
- 但角色 A 只能看到自己负责的数据
- 角色 B 可以看到整个部门的数据
- 超级管理员可以看到全部数据

这不是简单的“能访问/不能访问”二元判断，而是“访问后数据范围是多少”的问题。

因此数据权限更适合放在业务 service/biz 层，而不是强行塞进 Casbin 接口校验。

### 13.7 当前项目中的典型配合关系

可以用后台“客户列表”场景理解。

#### 13.7.1 进入页面

- 菜单权限决定是否能看到“客户管理”
- 按钮权限决定是否显示“新增客户”“删除客户”

#### 13.7.2 调接口

- Casbin 决定能否访问客户列表接口、创建接口、删除接口

#### 13.7.3 看数据

- 数据权限决定客户列表最终返回哪些客户

这三层共同作用，才构成完整权限闭环。

### 13.8 一句话总结这四类权限

可以用下面这组问题快速区分：

- Casbin API 权限：你能不能调这个接口？
- 菜单权限：你能不能看到并进入这个功能页面？
- 按钮权限：你能不能在页面里执行这个具体操作？
- 数据权限：你即使能调这个接口，最终能看到和操作哪些数据？

它们之间的关系不是二选一，而是分层协作：

- 菜单权限和按钮权限负责“看见什么”
- Casbin 负责“能调用什么”
- 数据权限负责“能处理什么数据”
