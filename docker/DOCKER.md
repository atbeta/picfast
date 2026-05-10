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
