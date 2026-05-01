# PicFast

[![CI](https://github.com/atbeta/picfast/actions/workflows/ci.yml/badge.svg)](https://github.com/atbeta/picfast/actions/workflows/ci.yml)
[![Release Please](https://github.com/atbeta/picfast/actions/workflows/release-please.yml/badge.svg)](https://github.com/atbeta/picfast/actions/workflows/release-please.yml)
[![Publish Docker Image](https://github.com/atbeta/picfast/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/atbeta/picfast/actions/workflows/docker-publish.yml)
[![Docker Pulls](https://img.shields.io/docker/pulls/xbeta/picfast?logo=docker)](https://hub.docker.com/r/xbeta/picfast)
[![Docker Image](https://img.shields.io/docker/image-size/xbeta/picfast/latest?logo=docker&label=image)](https://hub.docker.com/r/xbeta/picfast)

[中文](README.md)

PicFast is a modern self-hosted image hosting platform for individuals and teams. It provides guest uploads, authenticated uploads, admin management, and flexible storage backends.

## UI Preview

| Chinese UI | English UI |
|---|---|
| ![PicFast Console Chinese UI](.github/assets/console-zh-cn.png) | ![PicFast Console English UI](.github/assets/console-en-us.png) |

## Technology Stack

| Layer | Technology |
|------|------|
| Backend | Go 1.26, Chi Router, pgx/v5, sqlc, JWT |
| Frontend | React 19, TypeScript, Vite, React Router, Tailwind CSS v4 |
| Database | PostgreSQL 16 |
| Storage | Local filesystem / S3-compatible / Qiniu Kodo / Alibaba OSS / Tencent COS / WebDAV |
| Observability | Prometheus metrics, health check |

## Key Capabilities

- User auth: register, login, token refresh, logout
- Guest upload and authenticated upload
- Image list, permission update, deletion
- Album management
- Admin console: users, groups, storage strategies, settings, image management
- Storage backends: local, S3-compatible, Kodo, OSS, COS, WebDAV
- Image compression, watermarking, thumbnail generation
- Moderation and moderation callbacks
- ShareX config endpoint and upload integration
- API token and MCP server integration
- Health check, Prometheus metrics, pprof endpoints

## Quick Start

### Prerequisites

- Go 1.26+
- Node.js 20+ and pnpm
- PostgreSQL 16
- `golang-migrate` CLI

### 1) Start Database

```bash
make docker-up
```

Or use a local PostgreSQL instance:

```bash
createdb picfast
```

### 2) Configure Environment

```bash
cp .env.example .env
```

PicFast supports two config modes:

- Environment variables (`PICFAST_*` in `.env`)
- YAML config file (`config.yaml`)

```bash
cp config.example.yaml config.yaml
```

Environment variables override `config.yaml`.

### 3) Run Migrations and Seed

```bash
make migrate-up
make seed
```

### 4) Start Backend and Frontend

Backend:

```bash
go run ./cmd/picfast
```

Frontend:

```bash
cd web
pnpm install
pnpm dev
```

Open [http://localhost:5173](http://localhost:5173). Vite proxies `/api`, `/i`, `/t` to backend `http://localhost:8080`.

## Docker Deployment

Recommended image:

- `xbeta/picfast`

Common tags:

- `xbeta/picfast:latest` — latest stable build
- `xbeta/picfast:vX.Y.Z` — fixed release tag
- `xbeta/picfast:sha-<commit>` — commit trace tag

### 3-Minute Local Start

```bash
# 1) Start PostgreSQL
docker run -d \
  --name picfast-db \
  -e POSTGRES_USER=picfast \
  -e POSTGRES_PASSWORD=picfast \
  -e POSTGRES_DB=picfast \
  -p 5432:5432 \
  postgres:16-alpine

# 2) Start PicFast
docker run -d \
  --name picfast \
  -p 8080:8080 \
  -e PICFAST_DATABASE_URL='postgres://picfast:picfast@host.docker.internal:5432/picfast?sslmode=disable' \
  -e PICFAST_JWT_SECRET='replace-with-a-strong-secret' \
  -e PICFAST_SERVER_BASE_URL='http://localhost:8080' \
  -e PICFAST_APP_ADMIN_EMAIL='admin@example.com' \
  -e PICFAST_APP_ADMIN_PASSWORD='change-this-password' \
  -v picfast-uploads:/app/data/uploads \
  -v picfast-thumbnails:/app/data/thumbnails \
  xbeta/picfast:latest
```

Open `http://localhost:8080`.

### Pull Image

```bash
docker pull xbeta/picfast:latest
```

### Compose Templates

- [`docker/docker-compose.yml`](docker/docker-compose.yml): local development
- [`docker/docker-compose.traefik.yml`](docker/docker-compose.traefik.yml): production-like Traefik setup

## API Overview

| Endpoint | Description |
|------|------|
| `POST /api/v1/auth/register` | User registration |
| `POST /api/v1/auth/login` | User login |
| `POST /api/v1/auth/refresh` | Refresh access token |
| `POST /api/v1/upload` | Guest upload |
| `POST /api/v1/images` | Authenticated upload |
| `GET /api/v1/images` | List images |
| `GET /api/v1/albums` | List albums |
| `GET /api/v1/sharex/config` | Download ShareX config |
| `GET /i/{key}.{ext}` | Access image |
| `GET /t/{hash}.png` | Access thumbnail |
| `GET /health` | Health check |
| `GET /metrics` | Prometheus metrics |

Admin endpoints prefix: `/api/v1/admin/*`.

## Testing

```bash
make test
make lint
```

## Release and Image Policy

See `docs/release-playbook.md` for:

- Local `dev` image workflow
- CI `release` image workflow
- rollback guidance
