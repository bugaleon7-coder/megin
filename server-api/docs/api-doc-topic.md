# API 文档专题

## 一、文档目的

本文用于说明当前项目 API 文档页面的来源、后端 OpenAPI 文档的生成链路、前端 Knife4j 页面做过的定制修改，以及后续升级时应如何处理，避免再次出现路径拼接错误或静态资源同步错误。

本文只描述 API 文档体系本身，不维护具体业务接口清单。

相关文件：

- 后端路由与 Swagger 配置入口：`internal/router/router.go`
- OpenAPI 生成能力：`pkg/openapi/`
- 当前 Swagger JSON 输出目录：`static/swagger/`
- 当前 Knife4j 静态资源目录：`static/knife/`
- Knife4j 上游仓库：`https://github.com/xiaoymin/knife4j`

## 二、整体结构

当前项目的 API 文档由两部分组成：

1. 后端动态生成 OpenAPI JSON。
2. 前端使用 Knife4j 静态页面渲染 OpenAPI JSON。

请求链路如下：

```text
浏览器访问文档页
-> /api-doc/ 或 /admin-api-doc/ 或 /admin-system-doc/
-> Gin 返回 static/knife 下的静态页面
-> Knife4j 页面请求 swagger-config 或 swagger.json
-> Gin 返回 /v3/api-docs/swagger-config 或 /swagger/*.json
-> Knife4j 渲染接口文档
```

## 三、后端文档来源

### 3.1 文档入口地址

当前项目在 `internal/router/router.go` 中暴露了以下文档入口：

- `/api-doc/`：前端业务 API 文档页面
- `/admin-api-doc/`：后台管理 API 文档页面
- `/admin-system-doc/`：后台系统管理 API 文档页面
- `/v3/api-docs/swagger-config`：Knife4j springdoc 分组配置接口
- `/swagger/api/swagger.json`：前端业务 API OpenAPI 文档
- `/swagger/admin-api/swagger.json`：后台管理 API OpenAPI 文档
- `/swagger/system/swagger.json`：后台系统管理 OpenAPI 文档

### 3.2 OpenAPI JSON 生成方式

项目启动时，如果 `conf.ApiDoc.Enable` 为 `true`，会执行：

1. `genOpenApiDoc(registry)`
2. `staticSwaggerRouter(registry)`
3. `ginRouter.GET("/v3/api-docs/swagger-config", SwaggerConfig)`

其中 `genOpenApiDoc` 会根据路由前缀将接口拆成 3 份文档：

- `/api/*` -> `static/swagger/api/swagger.json`
- `/admin-api/*` 且非 system 路由 -> `static/swagger/admin-api/swagger.json`
- system 路由 -> `static/swagger/system/swagger.json`

system 路由的判定依赖 `internal/router/router.go` 中的 `systemRoutePrefixes`。

### 3.3 swagger-config 的作用

`/v3/api-docs/swagger-config` 不是业务接口，它是给 Knife4j 页面提供“当前要加载哪些文档”的配置接口。

当前后端根据来源页面或 `scope` 参数返回不同配置：

- `mixed`
- `api`
- `admin-api`
- `admin-system`

这样同一套静态页面可以按不同入口展示不同文档集合。

## 四、前端页面来源

### 4.1 当前项目中的静态资源来源

当前项目中的 `static/knife/` 不是手写页面，而是从 Knife4j 前端工程构建同步过来的产物。

上游来源建议统一以官方仓库为准：

```text
https://github.com/xiaoymin/knife4j
```

接手者拿到项目后，应先确认当前接入的是 Knife4j 仓库中的哪一份前端源码、哪个分支或哪个版本，再进行升级和同步，不应依赖某个开发者本机目录。

该工程构建后会产出：

- `dist/doc.html`
- `dist/webjars/**`
- `dist/oauth/**`
- `dist/robots.txt`
- `dist/favicon.ico`

同步到本项目时，当前落地方式是：

- `dist/doc.html` 复制为 `static/knife/index.html`
- `dist/webjars/**` 复制到 `static/knife/webjars/**`
- 其它静态资源同步到 `static/knife/` 下

### 4.2 为什么不要直接改编译产物

`static/knife/` 下大部分 JS/CSS 都是构建结果，直接修改存在几个问题：

- 升级或重新打包后会被覆盖。
- 压缩代码可读性差，不利于定位问题。
- 很难判断修改是否影响其它逻辑。

因此，文档页面行为有问题时，应优先修改 Knife4j 的源码工程，然后重新构建并同步产物，而不是直接改本项目中的编译结果。

## 五、本次做过的修改

### 5.1 问题现象

文档页面请求地址错误，出现了：

```text
/api-doc//v3/api-docs/swagger-config
```

实际希望请求的是：

```text
/v3/api-docs/swagger-config
```

### 5.2 根因

问题出在 Knife4j 前端源码对 springdoc 配置地址的自动拼接逻辑。

源码位置以 Knife4j 前端工程中的 `src/core/Knife4jAsync.js` 为准。

原逻辑会读取当前页面路径：

- 当前页面如果是 `/api-doc/`
- 它会把 `basePath` 识别为 `/api-doc`
- 然后拼出 `/api-doc/v3/api-docs/swagger-config`

而本项目的真实后端接口是站点根路径下的：

- `/v3/api-docs/swagger-config`

因此在本项目中，这段自动追加 `basePath` 的逻辑是不正确的。

### 5.3 实际修改内容

修改位置是 Knife4j 前端工程中的：

```text
src/core/Knife4jAsync.js
```

修改前逻辑含义：

- springdoc 模式下，根据当前页面路径自动拼接 `v3/api-docs/swagger-config`

修改后逻辑含义：

- springdoc 模式下，固定请求根路径 `/v3/api-docs/swagger-config`

可概括为：

```js
this.url = options.url || '/v3/api-docs/swagger-config'
```

这样无论文档页面入口是：

- `/api-doc/`
- `/admin-api-doc/`
- `/admin-system-doc/`

最终请求都会落到统一的后端接口：

- `/v3/api-docs/swagger-config`

### 5.4 这次没有改的内容

本次没有保留后端绕过方案，重点是：

- 不依赖后端去兼容前端错误拼接
- 不长期维护对编译产物的手工补丁
- 把问题收敛到前端源码层解决

## 六、当前约束与注意点

### 6.1 后端路径约束

当前后端文档相关路径约定如下：

- Knife4j 页面入口固定挂在文档目录下，例如 `/api-doc/`
- swagger-config 固定走根路径 `/v3/api-docs/swagger-config`
- swagger.json 固定走根路径 `/swagger/...`

如果以后要把项目部署到带统一前缀的网关路径下，例如：

```text
/shop-api/api-doc/
```

则要重新验证以下 2 件事：

1. 前端是否仍然应该请求绝对路径 `/v3/api-docs/swagger-config`
2. 是否需要根据网关前缀改成带统一前缀的绝对路径

也就是说，本次修复适用于“应用根路径就是服务根路径”的当前部署方式。

### 6.2 前端升级风险点

以后升级 Knife4j 前端源码时，要重点关注：

1. `src/core/Knife4jAsync.js` 是否仍然存在。
2. springdoc 的 `swagger-config` 地址拼接逻辑是否被上游改动。
3. 构建产物入口文件是否仍然是 `dist/doc.html`。
4. 构建后的主包文件名是否发生 hash 变化。
5. `static/knife/index.html` 是否仍需由 `doc.html` 覆盖。

如果上游已经修复这类路径拼接问题，则应尽量回归上游标准实现，不再保留本地补丁。

## 七、升级与同步操作建议

### 7.1 推荐升级流程

升级 Knife4j 前端源码时，建议按下面流程执行：

1. 先备份当前 `static/knife/`。
2. 从 `https://github.com/xiaoymin/knife4j` 获取需要接入的源码版本，或确认团队内部维护的 Knife4j 前端源码副本版本。
3. 检查 `src/core/Knife4jAsync.js` 中 springdoc 的 URL 处理逻辑。
4. 如上游未修复，则重新补上本项目所需的固定根路径逻辑。
5. 执行前端构建。
6. 将 `dist/` 产物同步到本项目 `static/knife/`。
7. 本地启动后验证 3 个入口页面。
8. 最后执行 `go build`，确认后端编译正常。

### 7.2 推荐验证项

升级或同步后至少验证以下地址：

- `http://localhost:端口/api-doc/`
- `http://localhost:端口/admin-api-doc/`
- `http://localhost:端口/admin-system-doc/`
- `http://localhost:端口/v3/api-docs/swagger-config`
- `http://localhost:端口/swagger/api/swagger.json`
- `http://localhost:端口/swagger/admin-api/swagger.json`
- `http://localhost:端口/swagger/system/swagger.json`

重点确认：

- 文档首页能正常打开
- 页面不会再请求 `/api-doc//v3/api-docs/swagger-config`
- 文档分组能正确展示
- system 文档和 admin-api 文档没有串组

### 7.3 推荐同步方式

建议保留“源码改完后重新构建，再整体同步”的方式，不建议在当前项目里单独手改压缩 JS。

可参考的同步思路：

1. 前端源码目录执行 `npm run build`
2. 将 `dist/` 内容同步到 `static/knife/`
3. 将 `dist/doc.html` 覆盖为 `static/knife/index.html`

如果后续需要脚本化，建议在本项目补一个专用同步脚本，例如：

```text
scripts/sync-knife4j.sh
```

这样能把“构建 + 同步 + 入口覆盖”固化下来，减少人工操作差错。

## 八、后续优化建议

当前文档体系能工作，但仍有几个可以继续优化的点：

1. 给 `static/knife/` 的来源和同步步骤补一个脚本，而不是靠人工命令执行。
2. 给 `docs/` 增加一份“部署到带网关前缀场景”的兼容说明。
3. 给 API 文档页面增加一次最小化回归检查，避免以后替换静态资源后再次出现路径错误。
4. 如果以后有条件，优先评估是否能直接接入更标准的 OpenAPI 前端，而不是长期维护外部构建产物副本。

## 九、结论

当前项目 API 文档的真实来源可以概括为：

1. 后端在启动时生成 3 份 OpenAPI JSON。
2. Gin 将这些 JSON 暴露在 `/swagger/` 路径下。
3. Gin 将 Knife4j 静态页面暴露在 `/api-doc/`、`/admin-api-doc/`、`/admin-system-doc/` 下。
4. Knife4j 前端通过 `/v3/api-docs/swagger-config` 与 `/swagger/*.json` 加载文档。
5. 本项目对 Knife4j 前端源码做过 1 个关键定制：springdoc 模式下固定从根路径请求 `swagger-config`，避免错误追加 `api-doc` 前缀。

以后升级时，优先检查 Knife4j 前端工程中的 `src/core/Knife4jAsync.js` 是否仍包含这段逻辑，这是本专题最关键的维护点。
