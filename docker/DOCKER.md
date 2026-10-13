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

Useful env vars / 常用环境变量:

- `PORT` — listen port injected by PaaS platforms; ignored when `PICFAST_SERVER_PORT` is set / PaaS 注入的监听端口，`PICFAST_SERVER_PORT` 优先
- `PICFAST_SERVER_METRICS_ENABLED=false` — disable the standalone `/metrics` listener / 关闭独立 metrics 监听端口

Health probes at the server root / 根路径健康探针: `GET /healthz` (liveness, no DB/disk access), `GET /readyz` (readiness, DB ping → 503 on failure).

Migrations are embedded and run automatically at startup / 迁移已嵌入二进制并随启动自动执行. Set `PICFAST_DATABASE_AUTO_MIGRATE=false` and run `docker run --rm <image> migrate` first if you prefer to migrate out of band / 如需手动控制可先执行 `migrate` 子命令。

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
