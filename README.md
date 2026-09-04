# go-app-starter

Go 应用开发基础框架，包含管理后台前端和服务端 API。

## 目录结构

```text
go-app-starter/
├── admin-web/  # Vue 管理后台
└── server-api/ # Go API 服务
```

## 快速安装

默认情况下，首次启动服务会在数据库不存在时自动导入初始化 SQL。以下命令假设本机已安装 MySQL 和与 `server-api/go.mod` 一致的 Go 版本；默认开发配置使用 MySQL `root:123456`。

```shell
# 1. 启动服务（默认 mixed 模式，端口 8800）。
# 数据库不存在时，服务会自动导入 server-api/sql/20260904_0001_init.sql；已有数据库不会被清空。
cd server-api
go run main.go -env=dev
```

需要手动建立一个全新库时，可执行：

```shell
mysql -uroot -p123456 < server-api/sql/20260904_0001_init.sql
```

启动后可访问：

- 管理后台：http://localhost:8800/admin/
- 前端业务 API 文档：http://localhost:8800/api-doc/
- 后台 Admin API 文档：http://localhost:8800/admin-api-doc/

默认数据库连接、Redis 地址和端口均可在 `server-api/config/config-dev.yaml` 中按需修改。

### 前端开发（仅修改管理后台时需要）

如需修改 `admin-web` 管理后台页面，首次开发前端时安装依赖：

```shell
cd admin-web
npm install
```

修改 `admin-web` 后，执行构建即可；构建产物会自动输出到 `server-api/static/admin`，无需手动复制。重新启动 Go 服务后，访问 `/admin/` 即可看到更新。

```shell
cd admin-web
npm run build
```
