

<p align="center">
  <img src="web/public/favicon-default.svg" width="128" height="128" alt="PicFast logo">
</p>

<h1 align="center">PicFast</h1>

<p align="center">
  PicFast 是可一键部署的开源图床，提供易用的管理后台和开放 API，<br>轻松对接上传工具、自动化工作流及各类 AI 应用。<br>
  <a href="README.en.md">English</a>
</p>

<p align="center">
  <a href="https://github.com/atbeta/picfast/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/atbeta/picfast/ci.yml?branch=main&label=ci&logo=github" alt="CI"></a>
  <a href="https://github.com/atbeta/picfast/releases/latest"><img src="https://img.shields.io/github/v/release/atbeta/picfast?logo=github&label=release" alt="Latest Release"></a>
  <a href="https://hub.docker.com/r/xbeta/picfast"><img src="https://img.shields.io/docker/pulls/xbeta/picfast?logo=docker&label=pulls" alt="Docker Pulls"></a>
  <a href="https://hub.docker.com/r/xbeta/picfast"><img src="https://img.shields.io/docker/image-size/xbeta/picfast/latest?logo=docker&label=image" alt="Docker Image Size"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-GPL--3.0-blue" alt="License"></a>
</p>

---

## 界面预览

![PicFast 控制台（中文）](.github/assets/console-zh-cn.png)

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.26, Chi Router, pgx/v5, sqlc, JWT |
| 前端 | React 19, TypeScript, Vite, React Router, Tailwind CSS v4 |
| 数据库 | PostgreSQL 16 |
| 存储 | 本地 / S3 兼容 / 七牛 Kodo / 阿里云 OSS / 腾讯云 COS / WebDAV |
| 可观测 | Prometheus、健康检查 |

## 功能

多用户与游客上传、相册管理、多种存储策略与管理后台；ShareX、API Token、MCP 等集成。完整文档见 **[picfast.dev](https://picfast.dev)**。

## 本地开发

### 环境准备

- 安装 Go 1.26+、Node 20+、pnpm、Docker（含 Compose）。
- 首次进入仓库执行：`cd web && pnpm install`。
- 复制配置：`cp .env.example .env`（可选：`cp config.example.yaml config.yaml`）。

数据库迁移在后端启动时**自动执行**，无需手动安装 `golang-migrate`。

### 快速验证

- `make docker-up`
- 访问 `http://localhost:18080`

> 未安装 `make` 时，在仓库根目录使用等价命令：
> `docker compose -f docker/docker-compose.dev.yml up --build -d`

### 功能开发

- `make docker-up`
- `docker compose -f docker/docker-compose.dev.yml stop app`
- `cp .env.example .env`
- `go run ./cmd/picfast`

此时后端由本机 Go 进程接管，后端接口地址为 `http://localhost:8080`，`http://localhost:18080` 不再提供前端页面。

如需界面功能验证或前端开发，再按需启动：

- `cd web && pnpm install && pnpm dev`
- 访问 `http://localhost:5173`

Vite 将 `/api`、`/i`、`/t` 代理到 `VITE_BACKEND_URL`（默认 `http://localhost:8080`，见 `web/vite.config.ts`）。

如需演示数据（额外测试账号与样例内容）再执行：`make seed`。
默认管理员账号：`admin@example.com` / `admin123`，测试用户：`test@example.com` / `password123`。

前后端一体静态托管：`cd web && pnpm build && go run ./cmd/picfast`。


## 部署

镜像：`xbeta/picfast`（[Docker Hub](https://hub.docker.com/r/xbeta/picfast)）。快速试跑：

```bash
docker network create picfast-net

docker run -d --name picfast-db --network picfast-net \
  -e POSTGRES_PASSWORD=devonly \
  -v picfast-pgdata:/var/lib/postgresql/data \
  postgres:16-alpine

docker run -d --name picfast --network picfast-net -p 18080:8080 \
  -e PICFAST_DATABASE_URL='postgres://postgres:devonly@picfast-db:5432/postgres?sslmode=disable' \
  -e PICFAST_JWT_SECRET='change-me-in-production' \
  -e PICFAST_SERVER_BASE_URL='http://localhost:18080' \
  -v picfast-uploads:/app/data/uploads \
  -v picfast-thumbnails:/app/data/thumbnails \
  -v picfast-backups:/app/data/backups \
  xbeta/picfast:latest
```

打开 `http://localhost:18080` 完成初始化向导。无人值守部署时同时设置 `PICFAST_APP_ADMIN_EMAIL` 和 `PICFAST_APP_ADMIN_PASSWORD` 即可跳过向导。

可观测性：Docker Compose 默认让 metrics 在容器内 `:9190` 监听，不发布到宿主机；Prometheus 可在同一 Docker 网络抓取 `app:9190/metrics`。通用 Compose、Traefik 模板及 `.env` 示例详见 **[docker/README.md](docker/README.md)**。

## 常用命令

```bash
make dev                # 打印开发步骤
make test && make lint
make docker-up
make docker-down
```


## 文档

| 说明 | 位置 |
|------|------|
| 官网 | [picfast.dev](https://picfast.dev) |
| Docker 部署 | [docker/README.md](docker/README.md) |
| 环境变量 | [.env.example](.env.example) |
| OpenAPI | 启动后访问 `/docs`、`/openapi.yaml` |
| MCP | [docs/mcp-api.md](docs/mcp-api.md) |
| 备份、巡检与恢复（CLI） | [docs/maintenance.md](docs/maintenance.md) |
| 镜像发布与回滚 | [docs/release-playbook.md](docs/release-playbook.md) |
