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
- 上传与缩略图：`./data/uploads`；缩略图随图片存入各存储策略的 `_picfast/thumbnails/`，`./data/thumbnails` 仅作旧版回退（`init-permissions` 自动修正权限）。迁移旧缩略图：`picfast maintenance repair-thumbnails --apply`
- 维护备份：`./data/backups`（管理后台「维护与迁移」页的默认命令会写入这里）

---

## 文件说明

| 文件 | 用途 |
|------|------|
| [`Dockerfile`](Dockerfile) | 多阶段构建镜像（完整版，含 libvips） |
| [`Dockerfile.slim`](Dockerfile.slim) | 精简镜像：纯 Go 构建，无 libvips/CGO，distroless 运行时 |
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

最小示例见 [DOCKER.md](DOCKER.md)；Fly.io / Railway / Render / Koyeb / Zeabur 等 serverless 平台的部署方式见 [picfast.dev/docs/serverless](https://picfast.dev/docs/serverless)。

镜像变体：`latest`/`vX` 为完整镜像（libvips：WebP 输出、文字水印、`pg_dump` 维护备份/恢复）；`slim`/`vX-slim` 为纯 Go 精简镜像（distroless，无 libvips/CGO，镜像更小、冷启动更快），不支持 WebP 输出、文字水印与 `pg_dump` 备份 —— 推荐容器 PaaS 使用。注意：精简镜像重新编码时不保留 EXIF 元数据，且无法对 WebP 源图做动态缩放/压缩（`.webp` 的 `@w_`/`@q_` 变体返回原图；WebP 缩略图仍可生成）。

核心变量：

- `PICFAST_DATABASE_URL` — 数据库连接串（必填）
- `PICFAST_SERVER_BASE_URL` — 站点对外访问地址
- `PICFAST_JWT_SECRET` — JWT 签名密钥（生产必填）
- `PORT` — PaaS 平台注入的监听端口（未设置 `PICFAST_SERVER_PORT` 时生效）
- `PICFAST_SERVER_METRICS_ENABLED` — 设为 `false` 可关闭独立的 `/metrics` 监听端口
- `PICFAST_JOBS_EMBEDDED` — 设为 `false` 关闭进程内后台任务调度（webhook 投递、每小时清理），改用外部定时触发 `POST /internal/jobs/run` 或 `picfast jobs run`
- `PICFAST_JOBS_TOKEN` — `POST /internal/jobs/run` 的 Bearer 令牌；不设置则不注册该端点
- `PICFAST_STORAGE_DEFAULT_TYPE` + `PICFAST_STORAGE_DEFAULT_CONFIG` — 首次启动（数据库尚无任何存储策略）时创建默认存储策略的类型与 JSON 配置。`PICFAST_STORAGE_DEFAULT_CONFIG` **留空**时会回退到本地存储并打印 WARN，**JSON 写错**则启动失败（fail-fast）；库里已有策略时会被忽略并打印 WARN（0.26 及更早版本因迁移预置了一条本地策略，这两个变量实际从未生效过；新建数据库已修正）。已有策略的老库请在后台「存储策略」里**新增** S3/R2 策略并设为分组默认——若改用 `delete from strategies;` 清表重启，注意 `images.strategy_id` 会被置为 NULL（有历史图片时勿用）。示例（S3）：`PICFAST_STORAGE_DEFAULT_TYPE=s3`，`PICFAST_STORAGE_DEFAULT_CONFIG='{"endpoint":"https://s3.example.com","region":"auto","bucket":"pics","access_key":"AK","secret_key":"SK","url":"https://cdn.example.com"}'`；Cloudflare R2（S3 兼容）：`"endpoint":"https://<account>.r2.cloudflarestorage.com"`、`"region":"auto"`、`"use_path_style":true`。

> ⚠️ 容器 PaaS（无持久卷）上默认的 `local` 存储位于临时磁盘，实例重建后上传文件即丢失 — 请用上述变量配置 S3/R2 等持久后端。

健康检查端点（服务根路径，无需鉴权）：

- `GET /healthz` — 存活探针，不访问数据库或磁盘
- `GET /readyz` — 就绪探针，仅检查数据库连通性，失败返回 503

数据库迁移已嵌入二进制，启动时自动执行。设 `PICFAST_DATABASE_AUTO_MIGRATE=false` 可关闭自动迁移，改为在启动前手动运行 `picfast migrate`（例如多副本/容器 PaaS 部署时由 Job 统一执行）。

托管 Postgres 的事务模式连接池（Supabase `:6543`、Neon `-pooler`、pgbouncer）：在 `PICFAST_DATABASE_URL` 查询串追加 `default_query_exec_mode=exec`，或改用直连连接串 —— pgx 默认语句缓存在事务池化下不可靠。保持 `pool_max_conns >= 3`（后台任务持锁占用一条连接，另需连接执行查询）；服务端 `idle_in_transaction_session_timeout` 不宜低于约 60 秒，否则会提前释放任务锁。

后台任务（webhook 投递、过期图片/日志清理）默认在进程内调度。设 `PICFAST_JOBS_EMBEDDED=false` 后可由外部调度器触发：HTTP `POST /internal/jobs/run`（需 `Authorization: Bearer $PICFAST_JOBS_TOKEN`）或 CLI `picfast jobs run`。webhook 投递延迟等于触发频率，建议每分钟一次；多实例并发触发安全（Postgres 咨询锁）。注意限流器按实例计数，多副本时实际限额约为配置值 × 实例数。
