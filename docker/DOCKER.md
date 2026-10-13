# PicFast

Source & full docs: https://github.com/atbeta/picfast / 源码与完整文档：https://github.com/atbeta/picfast

## Quick Deploy / 快速部署

```bash
docker network create picfast-net

docker run -d \
  --name picfast-db \
  --network picfast-net \
  -e POSTGRES_USER=picfast \
  -e POSTGRES_PASSWORD=picfast \
  -e POSTGRES_DB=picfast \
  -v picfast-pgdata:/var/lib/postgresql/data \
  postgres:16-alpine

docker run -d \
  --name picfast \
  --network picfast-net \
  -p 18080:8080 \
  -e PICFAST_DATABASE_URL='postgres://picfast:picfast@picfast-db:5432/picfast?sslmode=disable' \
  -e PICFAST_JWT_SECRET='replace-with-a-strong-secret' \
  -e PICFAST_SERVER_BASE_URL='http://localhost:18080' \
  -v picfast-uploads:/app/data/uploads \
  -v picfast-thumbnails:/app/data/thumbnails \
  -v picfast-backups:/app/data/backups \
  xbeta/picfast:latest
```

Open http://localhost:18080 — follow the setup wizard / 打开后跟随引导向导创建管理员。

Image variants / 镜像变体: `latest`/`vX` is the full image (libvips — WebP output, text watermark, `pg_dump`-based maintenance backup/restore) / 完整镜像（含 libvips）; `slim`/`vX-slim` is a pure-Go distroless build — smaller, faster cold start, no libvips/CGO, recommended for container PaaS / 纯 Go 精简镜像，推荐容器 PaaS. Slim has no WebP output, no text watermark, and no `pg_dump` backup/restore / slim 不支持 WebP 输出、文字水印与 pg_dump 备份. Caveat: re-encoded images do not keep EXIF metadata on slim even when stripping is off / 精简镜像即使未开启“去除 EXIF”，重新编码也不保留 EXIF。

Useful env vars / 常用环境变量:

- `PORT` — listen port injected by PaaS platforms; ignored when `PICFAST_SERVER_PORT` is set / PaaS 注入的监听端口，`PICFAST_SERVER_PORT` 优先
- `PICFAST_SERVER_METRICS_ENABLED=false` — disable the standalone `/metrics` listener / 关闭独立 metrics 监听端口

Health probes at the server root / 根路径健康探针: `GET /healthz` (liveness, no DB/disk access), `GET /readyz` (readiness, DB ping → 503 on failure).

Migrations are embedded and run automatically at startup / 迁移已嵌入二进制并随启动自动执行. Set `PICFAST_DATABASE_AUTO_MIGRATE=false` and run `docker run --rm <image> migrate` first if you prefer to migrate out of band / 如需手动控制可先执行 `migrate` 子命令。

Managed Postgres behind transaction-mode poolers (Supabase `:6543`, Neon `-pooler`, pgbouncer) / 事务模式连接池的托管 Postgres: append `default_query_exec_mode=exec` to the `PICFAST_DATABASE_URL` query string, or use the direct/session connection string — pgx's default statement cache is unreliable behind transaction poolers / 在连接串追加该参数或改用直连串。Keep `pool_max_conns >= 3` (a running job holds one connection for its advisory lock while querying on others) / 保持 `pool_max_conns >= 3`（任务持锁占用一条连接，另需连接执行查询）; a server-side `idle_in_transaction_session_timeout` below ~60s can release job locks early / 服务端该超时低于约 60 秒会提前释放任务锁。

Background jobs (webhook dispatch, hourly cleanup) run embedded by default / 后台任务（webhook 投递、每小时清理）默认在进程内运行. With `PICFAST_JOBS_EMBEDDED=false`, trigger them externally: `POST /internal/jobs/run` with `Authorization: Bearer $PICFAST_JOBS_TOKEN`, or `picfast jobs run` / 设为 false 后由外部调度器触发上述端点或 CLI. Webhook latency equals the schedule frequency — every minute recommended / webhook 延迟等于触发频率，建议每分钟一次. Concurrent triggers across instances are safe (Postgres advisory locks) / 多实例并发触发安全（咨询锁）. Note the rate limiter counts per instance / 限流器按实例计数.

Thumbnails are stored with each image in its storage strategy under `_picfast/thumbnails/`; `storage.thumbnail_dir` is a read-only legacy fallback only / 缩略图随图片保存在各存储策略的 `_picfast/thumbnails/` 下，`storage.thumbnail_dir` 仅作只读旧版回退。Migrate existing on-disk files with `picfast maintenance repair-thumbnails --apply` / 可执行 `picfast maintenance repair-thumbnails --apply` 迁移旧文件。

First-boot storage strategy / 首次启动的默认存储策略: `PICFAST_STORAGE_DEFAULT_TYPE` + `PICFAST_STORAGE_DEFAULT_CONFIG` (JSON) create the default strategy only when none exists yet / 仅在数据库尚无存储策略时生效. S3 example / S3 示例: `{"endpoint":"https://s3.example.com","region":"auto","bucket":"pics","access_key":"AK","secret_key":"SK","url":"https://cdn.example.com"}`. Cloudflare R2 (S3-compatible) / R2 示例: `"endpoint":"https://<account>.r2.cloudflarestorage.com"`, `"region":"auto"`, `"use_path_style":true`. ⚠️ On container PaaS the default `local` storage is ephemeral — configure a durable backend or uploads are lost on redeploy / 容器 PaaS 上默认 local 存储位于临时磁盘，重部署即丢失，请配置持久后端。

## Docker Compose

```bash
mkdir picfast && cd picfast
wget https://raw.githubusercontent.com/atbeta/picfast/main/docker/docker-compose.yml
wget https://raw.githubusercontent.com/atbeta/picfast/main/docker/.env.example -O .env
vim .env   # set BASE_URL, JWT_SECRET, DB password
docker compose up -d
```

For details: https://github.com/atbeta/picfast/tree/main/docker / 详细说明见 README。

## Architectures / 架构

`linux/amd64` `linux/arm64`
