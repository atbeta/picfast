# Contributing

Thanks for contributing to PicFast.

## Development Setup

1. Install Go 1.26+, Node.js 20+, pnpm, PostgreSQL 16, and `golang-migrate`.
2. Start the database with `make docker-up` or a local PostgreSQL instance.
3. Copy `.env.example` to `.env`.
4. Run `make migrate-up`.
5. Run `make seed`.
6. Start the backend with `go run ./cmd/picfast`.
7. Start the frontend with `cd web && pnpm install && pnpm dev`.

## Before Opening a PR

- Run `make lint`
- Run `go test ./...`
- Run `cd web && pnpm build`
- Update docs when behavior, APIs, or setup steps change

## Pull Request Notes

- Keep changes focused and scoped to one concern when possible.
- Include screenshots for UI changes.
- Mention any migration, config, or deployment impact in the PR description.
- If you change APIs, update `api/openapi.yaml`.

## Coding Notes

- The backend is written in Go and uses `sqlc` for database access.
- The frontend is a React + TypeScript + Vite app under `web/`.
- Prefer adding or updating tests when behavior changes.
