<p align="center">
  <img src="web/public/favicon-default.svg" width="128" height="128" alt="PicFast logo">
</p>

<h1 align="center">PicFast</h1>

<p align="center">
  PicFast is open-source image hosting you can deploy in one step,<br>with an intuitive admin console and open APIs for upload tools, automation workflows, and AI applications.<br>
  <a href="README.md">中文</a>
</p>

<p align="center">
  <a href="https://github.com/atbeta/picfast/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/atbeta/picfast/ci.yml?branch=main&label=ci&logo=github" alt="CI"></a>
  <a href="https://github.com/atbeta/picfast/releases/latest"><img src="https://img.shields.io/github/v/release/atbeta/picfast?logo=github&label=release" alt="Latest Release"></a>
  <a href="https://hub.docker.com/r/xbeta/picfast"><img src="https://img.shields.io/docker/pulls/xbeta/picfast?logo=docker&label=pulls" alt="Docker Pulls"></a>
  <a href="https://hub.docker.com/r/xbeta/picfast"><img src="https://img.shields.io/docker/image-size/xbeta/picfast/latest?logo=docker&label=image" alt="Docker Image Size"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-GPL--3.0-blue" alt="License"></a>
</p>

---

## UI Preview

![PicFast console (English)](.github/assets/console-en-us.png)

## Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.26, Chi Router, pgx/v5, sqlc, JWT |
| Frontend | React 19, TypeScript, Vite, React Router, Tailwind CSS v4 |
| Database | PostgreSQL 16 |
| Storage | Local / S3-compatible / Kodo / OSS / COS / WebDAV |
| Observability | Prometheus metrics, health check |

## Features

Multi-user and guest uploads, albums, pluggable storage backends, admin dashboard; ShareX, API tokens, and MCP integration. Full docs at **[picfast.dev](https://picfast.dev)**.

## Local Development

### Environment setup

- Install Go 1.26+, Node 20+, pnpm, and Docker (with Compose).
- First time in this repo, run: `cd web && pnpm install`.
- Copy config: `cp .env.example .env` (optional: `cp config.example.yaml config.yaml`).

Migrations run **automatically** at startup; `golang-migrate` is not required for normal local setup.

### Quick verification

- `make docker-up`
- Open `http://localhost:18080`

> If `make` is unavailable, run this equivalent command from the repository root:
> `docker compose -f docker/docker-compose.dev.yml up --build -d`

### Feature development

- `make docker-up`
- `docker compose -f docker/docker-compose.dev.yml stop app`
- `cp .env.example .env`
- `go run ./cmd/picfast`

At this point, the backend is served by your local Go process at `http://localhost:8080`, and `http://localhost:18080` no longer serves the frontend page.

Start frontend dev server only when needed:

- `cd web && pnpm install && pnpm dev`
- Open `http://localhost:5173`

Vite proxies `/api`, `/i`, `/t` to `VITE_BACKEND_URL` (default `http://localhost:8080`; see `web/vite.config.ts`).

Run `make seed` only when you want demo data (extra test users and sample content).

Bundled static serving: `cd web && pnpm build && go run ./cmd/picfast`.


## Deployment

Image: `xbeta/picfast` ([Docker Hub](https://hub.docker.com/r/xbeta/picfast)). Quick start:

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

Open `http://localhost:18080` and follow the setup wizard. For headless deployments, set both `PICFAST_APP_ADMIN_EMAIL` and `PICFAST_APP_ADMIN_PASSWORD` to skip the wizard.

Observability: Docker Compose listens for metrics on container-local `:9190` without publishing it to the host. Prometheus can scrape `app:9190/metrics` from the same Docker network. See **[docker/README.md](docker/README.md)** for the generic Compose file, Traefik template, and `.env` examples.

## Common Commands

```bash
make dev                # print development steps
make test && make lint
make docker-up
make docker-down
```


## Documentation

| Topic | Location |
|-------|----------|
| Website | [picfast.dev](https://picfast.dev) |
| Docker | [docker/README.md](docker/README.md) |
| Environment | [.env.example](.env.example) |
| OpenAPI | `/docs`, `/openapi.yaml` when running |
| MCP | [docs/mcp-api.md](docs/mcp-api.md) |
| Backup, doctor & restore (CLI) | [docs/maintenance.md](docs/maintenance.md) |
| Images & release | [docs/release-playbook.md](docs/release-playbook.md) |
