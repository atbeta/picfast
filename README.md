# PicFast

PicFast 是一个面向个人与团队的现代化图床/图片托管服务，支持游客上传、多用户、多存储策略、管理后台以及 AI / MCP 集成。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.26, Chi Router, pgx/v5, sqlc, JWT |
| 前端 | React 19, TypeScript, Vite, React Router, Tailwind CSS v4 |
| 数据库 | PostgreSQL 16 |
| 存储 | 本地文件系统 / AWS S3 兼容对象存储 |
| 可观测性 | Prometheus metrics, health check |

## 当前能力

- 用户注册、登录、刷新令牌、登出
- 游客上传与认证用户上传
- 图片列表、权限更新、删除
- 相册管理
- 管理后台：用户、分组、存储策略、系统设置、图片管理
- 本地存储与 S3 兼容存储
- 图片压缩、水印、同步缩略图生成
- 审核能力与审核状态回传
- ShareX 配置下载与上传接口
- API Token 与 MCP Server 集成
- 健康检查、Prometheus 指标、pprof 调试端点

## 快速开始

### 前置要求

- Go 1.26+
- Node.js 20+ 与 pnpm
- PostgreSQL 16
- `golang-migrate` CLI

### 1. 启动数据库

```bash
make docker-up
```

或使用本地 PostgreSQL：

```bash
createdb picfast
```

### 2. 配置环境

```bash
cp .env.example .env
```

按需修改数据库、JWT、存储和站点配置。

### 3. 执行迁移并填充开发数据

```bash
make migrate-up
make seed
```

默认会创建：

- 普通测试用户：`test@example.com` / `password123`
- 管理员用户：`admin@example.com` / `admin123`

### 4. 启动前后端

后端：

```bash
go run ./cmd/picfast
```

前端开发服务器：

```bash
cd web
pnpm install
pnpm dev
```

开发时访问 [http://localhost:5173](http://localhost:5173)，Vite 会代理 `/api`、`/i`、`/t` 到后端 `http://localhost:8080`。

### 5. 本地静态托管前端

如果希望由 Go 服务直接托管前端静态资源：

```bash
cd web && pnpm build
go run ./cmd/picfast
```

服务会优先查找：

1. `PICFAST_SERVER_WEB_DIR` / `server.web_dir`
2. 根目录 `web-dist`
3. 前端默认产物目录 `web/dist`

## 常用命令

```bash
# 开发
make dev
make seed

# 构建
make build
make frontend
make build-full

# 数据库
make migrate-up
make migrate-down
make generate

# 质量
make test
make lint
make format

# Docker
make docker-up
make docker-down
make docker-logs
```

## 目录结构

```text
.
├── cmd/                 # 服务入口与开发 seed 脚本
├── internal/
│   ├── config/          # 配置加载
│   ├── domain/          # 领域模型与常量
│   ├── handler/         # HTTP handlers 与 middleware
│   ├── router/          # 路由注册
│   ├── service/         # 上传、删除、审核、缩略图、存储等业务逻辑
│   ├── sqlc/            # sqlc 生成代码
│   └── testutil/        # 测试数据库与测试辅助
├── migrations/          # 数据库迁移
├── api/                 # OpenAPI 描述
├── docker/              # Dockerfile 与 compose
├── web/                 # React 前端
└── web-dist/            # 可选的根级静态构建产物目录
```

## API 概览

| 端点 | 说明 |
|------|------|
| `POST /api/v1/auth/register` | 用户注册 |
| `POST /api/v1/auth/login` | 用户登录 |
| `POST /api/v1/auth/refresh` | 刷新 Access Token |
| `POST /api/v1/upload` | 游客上传 |
| `POST /api/v1/images` | 认证用户上传 |
| `GET /api/v1/images` | 图片列表 |
| `GET /api/v1/albums` | 相册列表 |
| `GET /api/v1/sharex/config` | 下载 ShareX 配置 |
| `GET /i/{key}.{ext}` | 访问图片 |
| `GET /t/{hash}.png` | 访问缩略图 |
| `GET /health` | 健康检查 |
| `GET /metrics` | Prometheus 指标 |

管理员端点前缀：`/api/v1/admin/*`

## 测试与校验

```bash
make test
make lint
```

说明：

- `go test ./...` 会在默认本地 Postgres 上自动创建 `picfast_test` 测试库（如不存在）。
- `make lint` 会执行 `go vet`、前端 ESLint 和 TypeScript 编译检查。

## 部署说明

- Docker 镜像会在构建时编译前端，并通过 `PICFAST_SERVER_WEB_DIR=/web-dist` 交给 Go 服务托管。
- `/docs` 会加载 OpenAPI 文档页，`/openapi.yaml` 提供规范文件下载。
- `/api/v1/admin/debug/pprof/*` 仅管理员可访问。
