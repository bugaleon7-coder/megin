# 系统管理接口契约

## 一、契约规则

本文记录 system 模块必须提供给当前管理后台前端的接口契约。

统一规则：

- 完整前缀：`/admin-api`
- 成功响应：`code=200`
- 消息字段：`message`
- 除外层 `code` 和 `message` 外，路径、方法、参数和 `data` 结构保持原版兼容
- 管理接口需要 JWT、黑名单和 Casbin 校验
- 标记为“公开”的接口不需要认证，但仍使用 `/admin-api` 前缀

标准分页请求：

```json
{
  "page": 1,
  "pageSize": 10
}
```

标准分页响应 `data`：

```json
{
  "list": [],
  "total": 0,
  "page": 1,
  "pageSize": 10
}
```

## 二、公开接口

| 方法 | 路径 | 请求 | `data` |
|---|---|---|---|
| POST | `/admin-api/user/login` | JSON：用户名、密码、验证码 | `{ user, token, expiresAt }` |
| POST | `/admin-api/user/captcha` | 空 JSON | `{ captchaId, picPath, captchaLength, openCaptcha }` |
| POST | `/admin-api/sysError/createSysError` | JSON：错误信息 | 空 |

`/user/admin_register` 是管理员创建用户接口，不是公开注册接口。

## 三、用户管理

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| POST | `/admin-api/user/admin_register` | JSON | 空 |
| POST | `/admin-api/user/changePassword` | JSON | 空 |
| POST | `/admin-api/user/getUserList` | JSON | PageResult |
| POST | `/admin-api/user/setUserAuthority` | JSON | 空；响应头包含新 Token |
| POST | `/admin-api/user/setUserAuthorities` | JSON | 空 |
| DELETE | `/admin-api/user/deleteUser` | JSON | 空 |
| PUT | `/admin-api/user/setUserInfo` | JSON | 空 |
| PUT | `/admin-api/user/setSelfInfo` | JSON | 空 |
| PUT | `/admin-api/user/setSelfSetting` | JSON | 空 |
| GET | `/admin-api/user/getUserInfo` | Query/空 | `{ userInfo }` |
| POST | `/admin-api/user/resetPassword` | JSON | 空 |

用户 JSON 主键保持 `ID`；用户名字段保持 `userName`。

## 四、角色管理

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| POST | `/admin-api/authority/getAuthorityList` | JSON/空 | 角色树数组 |
| POST | `/admin-api/authority/createAuthority` | JSON | 空 |
| POST | `/admin-api/authority/copyAuthority` | JSON | 空 |
| POST | `/admin-api/authority/deleteAuthority` | JSON | 空 |
| PUT | `/admin-api/authority/updateAuthority` | JSON | 空 |
| POST | `/admin-api/authority/setDataAuthority` | JSON | 空 |
| GET | `/admin-api/authority/getUsersByAuthority` | Query：`authorityId` | 用户 ID 数组 |
| POST | `/admin-api/authority/setRoleUsers` | JSON | 空 |

角色树必须受当前管理员的父子角色权限限制。

## 五、菜单管理

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| POST | `/admin-api/menu/getMenu` | 空 JSON | `{ menus: [...] }` |
| POST | `/admin-api/menu/getMenuList` | JSON/空 | 菜单树数组 |
| POST | `/admin-api/menu/getBaseMenuTree` | 空 JSON | `{ menus: [...] }` |
| POST | `/admin-api/menu/getMenuAuthority` | JSON：`authorityId` | `{ menus: [...] }` |
| POST | `/admin-api/menu/addMenuAuthority` | JSON | 空 |
| POST | `/admin-api/menu/addBaseMenu` | JSON | 空 |
| POST | `/admin-api/menu/deleteBaseMenu` | JSON | 空 |
| POST | `/admin-api/menu/updateBaseMenu` | JSON | 空 |
| POST | `/admin-api/menu/getBaseMenuById` | JSON | 菜单详情 |
| GET | `/admin-api/menu/getMenuRoles` | Query：`menuId` | `{ authorityIds, defaultRouterAuthorityIds }` |
| POST | `/admin-api/menu/setMenuRoles` | JSON | 空 |

动态菜单必须包含原版需要的 `children`、`parameters`、`menuBtn` 和 `btns`。

## 六、API 与 Casbin

### 6.1 API 管理

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| POST | `/admin-api/api/getApiList` | JSON | PageResult |
| POST | `/admin-api/api/createApi` | JSON | 空 |
| POST | `/admin-api/api/getApiById` | JSON | API 详情 |
| POST | `/admin-api/api/updateApi` | JSON | 空 |
| POST | `/admin-api/api/getAllApis` | JSON/空 | `{ apis: [...] }` |
| POST | `/admin-api/api/deleteApi` | JSON | 空 |
| DELETE | `/admin-api/api/deleteApisByIds` | JSON | 空 |
| GET | `/admin-api/api/freshCasbin` | Query/空 | 空 |
| GET | `/admin-api/api/syncApi` | Query/空 | 待同步 API 数据 |
| GET | `/admin-api/api/getApiGroups` | Query/空 | `{ groups, apiGroupMap }` |
| POST | `/admin-api/api/ignoreApi` | JSON | 空 |
| POST | `/admin-api/api/enterSyncApi` | JSON | 空 |
| GET | `/admin-api/api/getApiRoles` | Query：`path`,`method` | 角色 ID 数组 |
| POST | `/admin-api/api/setApiRoles` | JSON | 空 |

前端仍保留的 `/api/setAuthApi` 不是当前原版契约；确认无调用后删除客户端废弃定义。

### 6.2 Casbin

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| POST | `/admin-api/casbin/updateCasbin` | JSON | 空 |
| POST | `/admin-api/casbin/getPolicyPathByAuthorityId` | JSON | `{ paths: [...] }` |

Casbin 内部维护接口不对前端公开，除非在契约中新增并完成安全评审。

## 七、按钮权限

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| POST | `/admin-api/authorityBtn/getAuthorityBtn` | JSON | `{ selected: [...] }` |
| POST | `/admin-api/authorityBtn/setAuthorityBtn` | JSON | 空 |
| POST | `/admin-api/authorityBtn/canRemoveAuthorityBtn` | Query | 布尔结果或原版对象 |

## 八、字典管理

### 8.1 字典

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| POST | `/admin-api/sysDictionary/createSysDictionary` | JSON | 空 |
| DELETE | `/admin-api/sysDictionary/deleteSysDictionary` | JSON | 空 |
| PUT | `/admin-api/sysDictionary/updateSysDictionary` | JSON | 空 |
| GET | `/admin-api/sysDictionary/findSysDictionary` | Query | 字典详情 |
| GET | `/admin-api/sysDictionary/getSysDictionaryList` | Query | 字典树数组 |
| GET | `/admin-api/sysDictionary/exportSysDictionary` | Query | 导出数据 |
| POST | `/admin-api/sysDictionary/importSysDictionary` | JSON | 空 |

### 8.2 字典详情

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| POST | `/admin-api/sysDictionaryDetail/createSysDictionaryDetail` | JSON | 空 |
| DELETE | `/admin-api/sysDictionaryDetail/deleteSysDictionaryDetail` | JSON | 空 |
| PUT | `/admin-api/sysDictionaryDetail/updateSysDictionaryDetail` | JSON | 空 |
| GET | `/admin-api/sysDictionaryDetail/findSysDictionaryDetail` | Query | 详情 |
| GET | `/admin-api/sysDictionaryDetail/getSysDictionaryDetailList` | Query | PageResult |
| GET | `/admin-api/sysDictionaryDetail/getDictionaryTreeList` | Query | 字典详情树 |
| GET | `/admin-api/sysDictionaryDetail/getDictionaryTreeListByType` | Query | 字典详情树 |
| GET | `/admin-api/sysDictionaryDetail/getDictionaryDetailsByParent` | Query | 子节点数组 |
| GET | `/admin-api/sysDictionaryDetail/getDictionaryPath` | Query | 完整路径 |

## 九、操作历史和登录日志

### 9.1 操作历史

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| DELETE | `/admin-api/sysOperationRecord/deleteSysOperationRecord` | JSON | 空 |
| DELETE | `/admin-api/sysOperationRecord/deleteSysOperationRecordByIds` | JSON | 空 |
| GET | `/admin-api/sysOperationRecord/getSysOperationRecordList` | Query | PageResult |

### 9.2 登录日志

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| DELETE | `/admin-api/sysLoginLog/deleteLoginLog` | JSON | 空 |
| DELETE | `/admin-api/sysLoginLog/deleteLoginLogByIds` | JSON | 空 |
| GET | `/admin-api/sysLoginLog/getLoginLogList` | Query | PageResult |
| GET | `/admin-api/sysLoginLog/findLoginLog` | Query | 登录日志详情 |

## 十、参数管理

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| POST | `/admin-api/sysParams/createSysParams` | JSON | 空 |
| DELETE | `/admin-api/sysParams/deleteSysParams` | Query | 空 |
| DELETE | `/admin-api/sysParams/deleteSysParamsByIds` | Query | 空 |
| PUT | `/admin-api/sysParams/updateSysParams` | JSON | 空 |
| GET | `/admin-api/sysParams/findSysParams` | Query | 参数详情 |
| GET | `/admin-api/sysParams/getSysParamsList` | Query | PageResult |
| GET | `/admin-api/sysParams/getSysParam` | Query：`key` | 参数值或详情 |

## 十一、API Token

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| POST | `/admin-api/sysApiToken/createApiToken` | JSON | `{ token }` |
| POST | `/admin-api/sysApiToken/getApiTokenList` | JSON | PageResult |
| POST | `/admin-api/sysApiToken/deleteApiToken` | JSON | 空 |

## 十二、版本管理

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| DELETE | `/admin-api/sysVersion/deleteSysVersion` | Query | 空 |
| DELETE | `/admin-api/sysVersion/deleteSysVersionByIds` | Query | 空 |
| GET | `/admin-api/sysVersion/findSysVersion` | Query | 版本详情 |
| GET | `/admin-api/sysVersion/getSysVersionList` | Query | PageResult |
| POST | `/admin-api/sysVersion/exportVersion` | JSON | 导出结果 |
| GET | `/admin-api/sysVersion/downloadVersionJson` | Query | JSON 文件流 |
| POST | `/admin-api/sysVersion/importVersion` | JSON | 空 |

## 十三、错误日志

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| POST | `/admin-api/sysError/createSysError` | JSON | 空 |
| DELETE | `/admin-api/sysError/deleteSysError` | Query | 空 |
| DELETE | `/admin-api/sysError/deleteSysErrorByIds` | Query | 空 |
| PUT | `/admin-api/sysError/updateSysError` | JSON | 空 |
| GET | `/admin-api/sysError/findSysError` | Query | 错误详情 |
| GET | `/admin-api/sysError/getSysErrorList` | Query | PageResult |
| GET | `/admin-api/sysError/getSysErrorSolution` | Query | 暂时返回“功能未启用” |

前端生成文件中的 `getSysErrorPublic` 当前没有合法原版后端契约，不应据此新增接口。

## 十四、JWT

| 方法 | 路径 | 参数位置 | `data` |
|---|---|---|---|
| POST | `/admin-api/jwt/jsonInBlacklist` | 从请求头读取当前 Token | 空 |

## 十五、契约测试要求

契约测试至少覆盖：

1. 所有表中路径和 HTTP 方法已注册。
2. 公开接口没有认证中间件。
3. 管理接口必须通过 JWT、黑名单和 Casbin。
4. 请求参数来自正确的 Body 或 Query。
5. 分页字段为 `page/pageSize/total/list`。
6. 特殊响应包装如 `{ menus }`、`{ apis }` 不得丢失。
7. 前端 `web/src/api` 中不存在无后端实现的有效调用。
