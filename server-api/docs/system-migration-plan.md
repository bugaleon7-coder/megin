# 系统管理模块重构与修复计划

## 一、目标与范围

将原项目 `/Users/lchb/go_admin/gin-vue-admin/server` 中的系统管理模块迁移到当前项目，并按当前项目分层规范重写。

迁移范围以 `docs/img.png` 为准：

- 角色管理
- 菜单管理
- API 管理
- 用户管理
- 字典管理
- 操作历史
- 参数管理
- API Token
- 登录日志
- 版本管理
- 错误日志

代码目标位置：

```text
internal/admin-api/system/   Handler
internal/system/dto/         请求和响应 DTO
internal/system/model/       GORM 模型
internal/system/repository/  数据访问
internal/system/service/     业务逻辑
internal/system/router/      路由注册
```

迁移原则：

1. 所有接口统一使用 `/admin-api` 前缀。
2. 外层成功状态由原版 `code=0` 改为 `code=200`。
3. 外层响应字段由 `msg` 改为 `message`。
4. 除上述两项外，请求参数、HTTP 方法、路由路径和 `data` 结构保持原版兼容。
5. Handler 开发规范以 `internal/admin-api/article.go` 为准。
6. 每次修改 Go 代码后必须执行 `go build ./...`。

## 二、已确认的技术决策

以下决策已经确认，后续实现不再保留其他兼容分支：

1. **软删除**：将 `SystemModel.DeletedAt` 改为 `gorm.DeletedAt`，恢复 GORM 和原项目的标准软删除语义。
2. **接口兼容**：只保留原版路径和 HTTP 方法，不保留当前错误接口的临时别名。
3. **错误日志 AI 处理**：暂时禁用 AI 处理接口，明确返回“功能未启用”，不保留一分钟后伪处理完成逻辑。
4. **集成测试数据库**：使用独立的 `gva_test` 数据库，不在现有 `gva` 数据库执行破坏性测试。

执行约束：

- `gva` 仅用于只读核对现有表结构和数据契约。
- `gva_test` 从必要的系统表结构和最小测试数据初始化。
- 测试不得依赖开发库中的可变业务数据。
- 删除当前非原版路由前，必须先通过前端 API 契约清单确认没有合法调用方。

## 三、事实来源优先级

当前文档存在历史内容，重构时按以下优先级确定真实需求：

1. 最新 `AGENTS.md`
2. 原项目后端代码
3. 当前前端 `web/src/api` 和页面实际用法
4. `gva` 数据库实际表结构
5. 本文档及其他辅助文档

已知文档问题：

- `docs/sql/schema.sql` 当前为空文件。
- `architecture-dev-guide.md` 仍包含 `/api` 公开分组，与最新要求不一致。
- 旧迁移方案使用 Unix 时间和 `ModelTime`，与当前 `SystemModel` 不一致。
- 旧迁移方案没有完整覆盖 API Token 和版本管理。

## 四、当前问题清单

### 4.1 P0：会直接造成接口失败或安全失效

1. 多个模型错误嵌入 `ControlBy`，但数据库没有 `create_by`、`update_by` 字段。
2. 前端分页参数是 `page/pageSize`，后端 DTO 使用 `page_no/page_size`。
3. 前端分页响应需要 `list/total/page/pageSize`，当前返回 `list/total_size/page_no/page_size`。
4. 大量接口的路径、HTTP 方法、请求位置和响应结构与前端不一致。
5. 当前认证只有 JWT 校验，没有接入 Casbin 权限中间件。
6. Casbin 角色 ID 使用 `string(rune(id))`，会把数字转换成 Unicode 字符。
7. JWT 黑名单没有在认证中间件中检查，退出登录和 Token 作废无效。
8. 菜单树和字典树使用值拷贝构造，子节点可能丢失。
9. `SysAuthority` 的主键是 `authority_id`，通用 Repository 却固定按 `id` 查询。
10. `GetUserAuthorities` 查询缺少 `sys_authorities` 表 JOIN。
11. GET Handler 返回错误后仍可能再次写入成功响应。
12. Query Binder 不支持嵌套结构、切片、指针和时间类型。
13. 空 JSON Body 会绕过 `required` 校验。

### 4.2 P1：业务逻辑与原版不一致

1. 验证码接口只是占位实现，登录没有验证码验证和失败次数限制。
2. 登录日志和操作日志只有查询服务，没有完整记录入口。
3. 用户注册、角色分配、菜单权限等多步骤操作没有事务保护。
4. 角色父子关系、严格角色树和跨级操作校验缺失。
5. API 更新、删除时没有同步 Casbin 策略。
6. 菜单新增、删除、更新遗漏名称唯一、父菜单、首页占用等校验。
7. 用户删除没有禁止删除自己，也没有完整清理角色关联。
8. 用户切换角色没有检查是否拥有该角色，也没有签发新 Token。
9. 字典导入没有完整导入详情，且没有事务保护。
10. 错误日志处理只等待一分钟后标记完成，没有生成解决方案。
11. API Token 模块未迁移。
12. 版本管理模块未迁移。

### 4.3 数据库映射问题

实际数据库表中没有 `ControlBy` 字段，以下模型需要重点核对：

- `SysUser`
- `SysApi`
- `SysDictionary`
- `SysDictionaryDetail`
- `SysParams`
- `SysLoginLog`
- `SysOperationRecord`
- `SysBaseMenuBtn`
- `JwtBlacklist`

另外，JWT 黑名单实际表名是 `jwt_blacklists`，当前模型表名需要修正。

`SystemModel.DeletedAt` 使用 `*time.Time`，需要明确并测试其软删除行为是否与原版 `gorm.DeletedAt` 一致。

## 五、接口兼容性重点

### 5.1 用户管理

- 恢复 `/user/admin_register`。
- 修正 `deleteUser`、`setUserInfo`、`setSelfInfo` 的 HTTP 方法。
- 恢复 `/user/setSelfSetting`。
- 保持 `getUserList` 的分页请求和响应与前端一致。

### 5.2 角色管理

- 修正 `updateAuthority` 的 HTTP 方法。
- 恢复 `/authority/getUsersByAuthority`。
- 角色列表返回前端需要的树形数组，而不是不兼容的分页结构。

### 5.3 菜单管理

- 保持 `getMenuList`、`getBaseMenuTree`、`getBaseMenuById` 的原版方法。
- 恢复 `getMenuRoles` 和 `setMenuRoles`。
- `getBaseMenuTree` 和 `getMenuAuthority` 返回 `{ menus: [...] }`。

### 5.4 API 与 Casbin

- 修正 `getApiById`、`getAllApis`、`deleteApisByIds` 的方法。
- 恢复 `syncApi`、`ignoreApi`、`enterSyncApi`。
- 恢复 `getApiRoles`、`setApiRoles`。
- 保持 `getAllApis` 返回 `{ apis: [...] }`。
- 修正 `getPolicyPathByAuthorityId` 的方法。

### 5.5 字典、日志和参数

恢复前端正在使用的原版路径：

- `findSysDictionary`
- `getSysDictionaryList`
- `findSysDictionaryDetail`
- `getSysDictionaryDetailList`
- `getDictionaryTreeList`
- `getDictionaryTreeListByType`
- `getDictionaryDetailsByParent`
- `getDictionaryPath`
- `getSysOperationRecordList`
- `deleteSysOperationRecordByIds`
- `getLoginLogList`
- `findLoginLog`
- `deleteLoginLogByIds`
- `findSysParams`
- `getSysParamsList`
- `getSysParam`

### 5.6 缺失模块

API Token：

- `createApiToken`
- `getApiTokenList`
- `deleteApiToken`

版本管理：

- `deleteSysVersion`
- `deleteSysVersionByIds`
- `findSysVersion`
- `getSysVersionList`
- `exportVersion`
- `downloadVersionJson`
- `importVersion`

## 六、实施计划

### 阶段一：冻结接口契约

1. 从 `web/src/api` 自动提取路径、HTTP 方法和参数位置。
2. 对照原项目 Router、Handler 和 Service。
3. 建立完整接口兼容矩阵。
4. 明确每个接口的请求 DTO、响应 `data` 结构和认证要求。
5. 增加自动化路由契约测试，防止再次漏注册或改错方法。

验收标准：前端使用的每个系统接口都能在后端找到完全匹配的路由。

### 阶段二：修复公共基础设施

1. 使用 Gin 标准 Binder 修复 GET Query 参数绑定。
2. DELETE 同时兼容原版 JSON Body 和 Query 参数。
3. 修复 Handler 错误后重复写响应。
4. 恢复 `page/pageSize/total/list` 契约。
5. 为无参数请求建立空 DTO，禁止必填请求通过空 Body 绕过校验。
6. 修复事务辅助方法和错误传播。
7. 明确并恢复软删除语义。
8. 增加 GORM DryRun 和数据库字段一致性测试。

验收标准：公共 Binder、分页、错误响应、事务和模型映射测试全部通过。

### 阶段三：认证与权限

1. JWT 中间件限制签名算法。
2. 移除硬编码 JWT Secret 回退。
3. 接入 JWT 黑名单检查。
4. 恢复 Token 刷新及多点登录逻辑。
5. 接入 Casbin 权限中间件。
6. 修复 Casbin Authority ID 转换。
7. 恢复父子角色和严格角色树越权检查。
8. 对请求日志中的密码、Token 等敏感字段脱敏。

验收标准：未授权角色不能访问受限接口，退出或作废的 Token 不能继续使用。

### 阶段四：用户和角色管理

1. 恢复管理员注册、验证码和登录日志。
2. 用户注册和角色关联使用同一事务。
3. 删除用户时禁止删除自己，并清理角色关联。
4. 恢复 `SetSelfSetting`。
5. 切换角色时检查角色归属、默认路由并返回新 Token。
6. 修复角色主键查询。
7. 恢复角色树、数据权限、菜单权限和用户关联逻辑。
8. 角色复制、删除和权限变更全部事务化。

验收标准：用户和角色页面全部功能可用，越权测试通过，无残留关联数据。

### 阶段五：菜单、按钮和 API 权限

1. 重写菜单树构建，支持任意层级。
2. 正确加载菜单 Parameters、MenuBtn 和按钮权限。
3. 恢复菜单名称唯一、父菜单、首页占用和叶子菜单校验。
4. 删除菜单时清理角色菜单、按钮和参数关联。
5. 恢复菜单与角色双向分配接口。
6. API 创建和更新增加唯一性校验。
7. API 更新、删除同步 Casbin。
8. 恢复 API 同步、忽略和角色分配功能。

验收标准：动态菜单、按钮权限、菜单分配和 API 权限页面行为与原版一致。

### 阶段六：字典、参数和日志

1. 移除错误的 `ControlBy` 映射。
2. 恢复原字典和字典详情接口契约。
3. 修复字典树、Level 和 Path 计算。
4. 字典导入导出包含详情并使用事务。
5. 参数增加 Key 唯一校验。
6. 实现 OperationRecord 中间件，并控制记录范围和敏感数据。
7. 登录成功、失败、冻结和验证码错误均记录登录日志。

验收标准：字典树和分页正确，操作历史及登录日志能持续产生数据。

### 阶段七：补齐缺失模块

API Token：

1. 完成模型、DTO、Repository、Service、Handler 和 Router。
2. 签发前校验用户是否拥有指定角色。
3. 作废 Token 时加入 JWT 黑名单。
4. 完成分页和状态筛选。

版本管理：

1. 完成版本列表、详情、删除、批量删除。
2. 完成菜单、API 和字典数据导出。
3. 完成版本 JSON 下载和导入。
4. 导入过程使用事务并处理冲突。

错误日志：

1. 保持 `sys_error` 表结构。
2. 恢复实际解决方案生成。
3. 正确区分处理中、处理完成和处理失败。
4. 避免使用请求生命周期 Context 执行后台任务。

### 阶段八：测试、文档与验收

每个模块必须执行：

1. `go build ./...`
2. 路由契约测试
3. GORM DryRun 字段测试
4. 数据库事务集成测试
5. JWT/Casbin 越权测试
6. 前端核心页面冒烟测试

最终工作：

1. 更新 OpenAPI/Swagger。
2. 更新 `architecture-dev-guide.md`。
3. 从实际数据库生成 `docs/sql/schema.sql`。
4. 删除过时的接口和迁移说明。

## 七、推荐实施顺序

```text
公共契约
  → 数据模型
  → JWT/Casbin
  → 用户/角色
  → 菜单/API
  → 字典/参数
  → 操作日志/登录日志
  → API Token/版本管理
  → 错误日志
  → 全量验收
```

在公共契约、模型映射和权限中间件完成前，不继续批量新增业务代码，避免在错误基础上扩大返工范围。
