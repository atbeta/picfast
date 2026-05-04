# PicFast Docker 部署

本目录包含 Dockerfile、三套 Compose 编排及环境变量模板。

---

## 文件说明

| 文件 | 用途 |
|------|------|
| [`Dockerfile`](Dockerfile) | 多阶段构建镜像 |
| [`docker-compose.dev.yml`](docker-compose.dev.yml) | 本地开发：PostgreSQL + Mailpit + PicFast，映射 `18080→8080` |
| [`docker-compose.yml`](docker-compose.yml) | 通用生产：PostgreSQL + PicFast，反向代理自行接入 |
| [`docker-compose.traefik.yml`](docker-compose.traefik.yml) | Traefik 模板：labels + 外部网络 + 数据落盘 |
| [`.env.dev.example`](.env.dev.example) | 开发环境变量参考（配合 `docker-compose.dev.yml`） |
| [`.env.example`](.env.example) | 通用生产环境变量参考（配合 `docker-compose.yml`） |
| [`.env.traefik.example`](.env.traefik.example) | Traefik 环境变量参考（配合 `docker-compose.traefik.yml`） |

---

## 本地开发（dev）

在仓库**根目录**执行：

```bash
make docker-up
```

等价于 `docker compose -f docker/docker-compose.dev.yml up --build -d`，启动后：

- 应用：[http://localhost:18080](http://localhost:18080)
- Mailpit Web UI：[http://localhost:8025](http://localhost:8025)（SMTP `127.0.0.1:1025`）
- Prometheus metrics：容器内 `http://app:9190/metrics`，默认不发布到宿主机

可选的覆盖项见 [`.env.dev.example`](.env.dev.example)，写入仓库根目录 `.env` 即可生效。

停止：`make docker-down`。

---

## 生产部署（通用 Compose）

这份模板不包含任何反向代理配置，只把 PicFast 映射到宿主机本地端口，适合让 Nginx、Caddy、Nginx Proxy Manager、宝塔、HAProxy 等自行反代。

```bash
cd docker
cp .env.example .env
# 编辑 .env：替换 HTTPS URL、数据库密码、JWT 密钥、邮件等
docker compose up -d
```

默认会把 PicFast 绑定到 `127.0.0.1:18080`，推荐反代到：

```text
http://127.0.0.1:18080
```

如果确实需要直接暴露给局域网或公网，可在 `.env` 中设置 `PICFAST_HTTP_BIND=0.0.0.0`。生产环境仍建议放在反向代理之后处理 HTTPS、访问日志、压缩与额外安全策略。

数据落盘在 `docker/` 目录下：
- Postgres：`./data/postgres`
- 上传与缩略图：`./data/uploads`、`./data/thumbnails`（`init-permissions` 自动修正权限）

## 生产部署（Traefik）

**前置条件**：已有运行中的 Traefik，且 Docker 网络名与 `.env` 中 `TRAEFIK_DOCKER_NETWORK` 一致（默认 `traefik_monitoring`）。

```bash
cd docker
cp .env.traefik.example .env
# 编辑 .env：替换域名、HTTPS URL、数据库密码、JWT 密钥、邮件等
docker compose -f docker-compose.traefik.yml up -d
```

Traefik 模板同样会把数据落盘在 `docker/` 目录下：
- Postgres：`./data/postgres`
- 上传与缩略图：`./data/uploads`、`./data/thumbnails`（`init-permissions` 自动修正权限）

镜像标签通过 `PICFAST_IMAGE` 指定，可在 `.env` 中固定版本号。各项配置说明见 [`.env.traefik.example`](.env.traefik.example) 内注释。

### Prometheus / Grafana

PicFast 默认在独立 metrics server 上暴露 `/metrics`。裸机运行时默认监听 `127.0.0.1:9190`；Docker Compose 会覆盖为 `:9190`，但不会通过 `ports` 发布到宿主机。

Prometheus 与 PicFast 在同一个 Docker 网络时，可使用以下抓取目标：

```yaml
scrape_configs:
  - job_name: picfast
    metrics_path: /metrics
    static_configs:
      - targets: ["app:9190"]
```

如需从其他容器网络抓取，请让 Prometheus 加入 PicFast 所在网络，或显式调整 `PICFAST_SERVER_METRICS_ADDR` 与网络策略。不要将 `9190` 直接发布到公网。

---

## 仅用镜像（无 Compose）

最小示例见仓库根目录 [README.md](../README.md) 的 Docker 小节。核心变量：`PICFAST_DATABASE_URL`（必填）、`PICFAST_SERVER_BASE_URL`、`PICFAST_JWT_SECRET`（生产必填）。
