# go-app-starter

# megin

Go 应用开发基础框架，包含管理后台前端和服务端 API。

## 目录结构

```text
go-app-starter/
├── admin-web/  # Vue 管理后台
└── server-api/ # Go API 服务
```

## 快速安装

默认情况下，只需导入初始化 SQL 并启动服务即可。以下命令假设本机已安装 MySQL 和与 `server-api/go.mod` 一致的 Go 版本；默认开发配置使用 MySQL `root:123456`。

```shell
# 1. 创建数据库并导入初始化数据
mysql -uroot -p123456 -e "create database if not exists go_app_starter default charset utf8mb4 collate utf8mb4_unicode_ci;"
mysql -uroot -p123456 go_app_starter < server-api/go_app_starter.sql

# 2. 启动服务（默认 mixed 模式，端口 8800）
cd server-api
go run main.go -env=dev
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
