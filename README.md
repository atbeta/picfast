<p align="center">
  <img src="web/public/favicon-default.svg" width="128" height="128" alt="PicFast logo">
</p>

<h1 align="center">PicFast</h1>

<p align="center">
  PicFast 是可一键部署的开源图床，提供易用的管理后台和开放 API，<br>轻松对接上传工具、自动化工作流及各类 AI 应用。<br>
  <a href="README.en.md">English</a>
</p>

<p align="center">
  <a href="https://github.com/atbeta/picfast/actions/workflows/ci.yml"><img src="https://github.com/atbeta/picfast/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/atbeta/picfast/actions/workflows/release-please.yml"><img src="https://github.com/atbeta/picfast/actions/workflows/release-please.yml/badge.svg" alt="Release Please"></a>
  <a href="https://github.com/atbeta/picfast/actions/workflows/docker-publish.yml"><img src="https://github.com/atbeta/picfast/actions/workflows/docker-publish.yml/badge.svg" alt="Publish Docker Image"></a>
  <a href="https://hub.docker.com/r/xbeta/picfast"><img src="https://img.shields.io/docker/pulls/xbeta/picfast?logo=docker" alt="Docker Pulls"></a>
  <a href="https://hub.docker.com/r/xbeta/picfast"><img src="https://img.shields.io/docker/image-size/xbeta/picfast/latest?logo=docker&label=image" alt="Docker Image"></a>
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

**环境**：Go 1.26+、Node 20+、pnpm、PostgreSQL 16。数据库迁移在后端启动时**自动执行**，无需手动安装 `golang-migrate`。

```bash
make docker-up                    # 启动 PostgreSQL + Mailpit
cp .env.example .env              # 可选：cp config.example.yaml config.yaml
# 可选：make seed
go run ./cmd/picfast
cd web && pnpm install && pnpm dev
```

打开 [http://localhost:5173](http://localhost:5173)。Vite 将 `/api`、`/i`、`/t` 代理到 `VITE_BACKEND_URL`（默认 `http://localhost:8080`，见 `web/vite.config.ts`）。

前后端一体部署：`cd web && pnpm build && go run ./cmd/picfast`。

## Docker

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
  xbeta/picfast:latest
```

打开 `http://localhost:18080` 完成初始化向导。无人值守部署时同时设置 `PICFAST_APP_ADMIN_EMAIL` 和 `PICFAST_APP_ADMIN_PASSWORD` 即可跳过向导。

Compose、Traefik 及 `.env` 模板详见 **[docker/README.md](docker/README.md)**。

## 常用命令

```bash
make dev                # 打印开发步骤
make test && make lint
make docker-up          # docker-compose.dev.yml
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
| 镜像发布与回滚 | [docs/release-playbook.md](docs/release-playbook.md) |
