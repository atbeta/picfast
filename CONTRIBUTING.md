# Contributing

Thanks for contributing to PicFast.

## Development Setup

1. Install Go 1.26+, Node.js 24+, pnpm, and PostgreSQL 16.
2. Start the database: `make docker-up`
3. Run migrations: `make migrate-up`
4. Seed data (optional): `make seed`
5. Start backend: `go run ./cmd/picfast`
6. Start frontend: `cd web && pnpm install && pnpm dev`

## Before Opening a PR

- Run `make lint`
- Run `go test ./...`
- Run `cd web && pnpm build`
- If you changed sqlc queries or migrations, run `make generate`
- Update `api/openapi.yaml` when APIs change
- Update i18n keys for user-visible text changes (zh-CN + en-US)

## Coding Notes

- Backend: Go + Chi router + sqlc for DB access + pgx/v5
- Frontend: React 19 + TypeScript + Vite + Tailwind CSS v4
- Generated files (`internal/sqlc/*.go`) must not be edited manually — change queries in `internal/sqlc/queries/*.sql` and run `make generate`
- Prefer adding or updating tests when behavior changes
