# PicFast Docker 部署

本目录包含 Dockerfile、两套 Compose 编排及环境变量模板。

---

## 文件说明

| 文件 | 用途 |
|------|------|
| [`Dockerfile`](Dockerfile) | 多阶段构建镜像 |
| [`docker-compose.dev.yml`](docker-compose.dev.yml) | 本地开发：PostgreSQL + Mailpit + PicFast，映射 `18080→8080` |
| [`docker-compose.traefik.yml`](docker-compose.traefik.yml) | 生产向：对接 Traefik，labels + 外部网络 + 数据落盘 |
| [`.env.dev.example`](.env.dev.example) | 开发环境变量参考（配合 `docker-compose.dev.yml`） |
| [`.env.traefik.example`](.env.traefik.example) | 生产环境变量参考（配合 `docker-compose.traefik.yml`） |

---

## 本地开发（dev）

在仓库**根目录**执行：

```bash
make docker-up
```

等价于 `docker compose -f docker/docker-compose.dev.yml up --build -d`，启动后：

- 应用：[http://localhost:18080](http://localhost:18080)
- Mailpit Web UI：[http://localhost:8025](http://localhost:8025)（SMTP `127.0.0.1:1025`）

可选的覆盖项见 [`.env.dev.example`](.env.dev.example)，写入仓库根目录 `.env` 即可生效。

停止：`make docker-down`。

---

## 生产部署（Traefik）

**前置条件**：已有运行中的 Traefik，且 Docker 网络名与 `.env` 中 `TRAEFIK_DOCKER_NETWORK` 一致（默认 `traefik_monitoring`）。

```bash
cd docker
cp .env.traefik.example .env
# 编辑 .env：替换域名、HTTPS URL、数据库密码、JWT 密钥、邮件等
docker compose -f docker-compose.traefik.yml up -d
```

数据落盘在 `docker/` 目录下：
- Postgres：`./data/postgres`
- 上传与缩略图：`./data/uploads`、`./data/thumbnails`（`init-permissions` 自动修正权限）

镜像标签通过 `PICFAST_IMAGE` 指定，可在 `.env` 中固定版本号。各项配置说明见 [`.env.traefik.example`](.env.traefik.example) 内注释。

---

## 仅用镜像（无 Compose）

最小示例见仓库根目录 [README.md](../README.md) 的 Docker 小节。核心变量：`PICFAST_DATABASE_URL`（必填）、`PICFAST_SERVER_BASE_URL`、`PICFAST_JWT_SECRET`（生产必填）。
