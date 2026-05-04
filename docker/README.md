# PicFast Docker 部署

本目录包含 Dockerfile、Compose 编排及环境变量模板。推荐直接下载所需文件，无需克隆整个仓库。

---

## 快速开始

```bash
mkdir picfast && cd picfast

# 下载 Compose 文件和 env 模板
wget https://raw.githubusercontent.com/atbeta/picfast/main/docker/docker-compose.yml
wget https://raw.githubusercontent.com/atbeta/picfast/main/docker/.env.example -O .env

# 编辑 .env — 替换域名、密码、JWT 密钥
vim .env

# 启动
docker compose up -d
```

默认会将 PicFast 绑定到 `127.0.0.1:18080`，推荐在前置 Nginx / Caddy / NPM / 宝塔 等反代到该地址后提供 HTTPS。

如需直接暴露给局域网或公网，在 `.env` 中设置 `PICFAST_HTTP_BIND=0.0.0.0`。生产环境仍建议放在反向代理之后处理 HTTPS、访问日志、压缩与安全策略。

启动后打开 `http://localhost:18080`，首次访问跟随引导向导创建管理员。
设置 `PICFAST_APP_ADMIN_EMAIL` + `PICFAST_APP_ADMIN_PASSWORD` 可跳过向导。

数据落盘在当前目录下：
- Postgres：`./data/postgres`
- 上传与缩略图：`./data/uploads`、`./data/thumbnails`（`init-permissions` 自动修正权限）

---

## 文件说明

| 文件 | 用途 |
|------|------|
| [`Dockerfile`](Dockerfile) | 多阶段构建镜像 |
| [`docker-compose.yml`](docker-compose.yml) | **推荐**：PostgreSQL + PicFast，自行接入反代 |
| [`docker-compose.traefik.yml`](docker-compose.traefik.yml) | Traefik 方案（需已运行 Traefik 并配置证书） |
| [`docker-compose.dev.yml`](docker-compose.dev.yml) | 本地开发：PostgreSQL + Mailpit + PicFast |
| [`.env.example`](.env.example) | 环境变量参考（配合 `docker-compose.yml`） |
| [`.env.traefik.example`](.env.traefik.example) | Traefik 环境变量参考 |
| [`.env.dev.example`](.env.dev.example) | 开发环境变量参考 |

---

## Traefik 方案

适用于已有 Traefik 并配置好证书的场景。

```bash
mkdir picfast && cd picfast
wget https://raw.githubusercontent.com/atbeta/picfast/main/docker/docker-compose.traefik.yml
wget https://raw.githubusercontent.com/atbeta/picfast/main/docker/.env.traefik.example -O .env

vim .env  # 替换域名、数据库密码等
docker compose -f docker-compose.traefik.yml up -d
```

确保 `.env` 中 `TRAEFIK_DOCKER_NETWORK` 与 Traefik 所在网络一致，且 Traefik 已提前载入域名对应的证书（如通过 `tls.certificates` 静态文件或 certResolver）。

---

## 本地开发

在仓库**根目录**执行：

```bash
make docker-up
```

等价于 `docker compose -f docker/docker-compose.dev.yml up --build -d`，启动后：

- 应用：[http://localhost:18080](http://localhost:18080)
- Mailpit Web UI：[http://localhost:8025](http://localhost:8025)
- Prometheus metrics：容器内 `http://app:9190/metrics`

可选的覆盖项见 [`.env.dev.example`](.env.dev.example)，写入仓库根目录 `.env` 即可生效。

停止：`make docker-down`。

---

## Prometheus / Grafana

PicFast 在独立端口暴露 `/metrics`。裸机运行时默认监听 `127.0.0.1:9190`；Docker Compose 会覆盖为 `:9190`，不通过 `ports` 发布到宿主机。

Prometheus 与 PicFast 在同一 Docker 网络时抓取目标：

```yaml
scrape_configs:
  - job_name: picfast
    metrics_path: /metrics
    static_configs:
      - targets: ["app:9190"]
```

---

## 仅用镜像（无 Compose）

最小示例见 [DOCKER.md](DOCKER.md)。核心变量：

- `PICFAST_DATABASE_URL` — 数据库连接串（必填）
- `PICFAST_SERVER_BASE_URL` — 站点对外访问地址
- `PICFAST_JWT_SECRET` — JWT 签名密钥（生产必填）
