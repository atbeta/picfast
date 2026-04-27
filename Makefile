.PHONY: build run dev generate migrate-up migrate-down docker-up docker-down tidy frontend test format lint openapi-lint seed clean

GOPATH := $(shell go env GOPATH)
SQLC := $(GOPATH)/bin/sqlc

## Development

dev:
	@echo "Starting development environment..."
	@echo "1. Ensure Postgres is running (make docker-up or use local Postgres)"
	@echo "2. Run migrations (make migrate-up)"
	@echo "3. Seed data (make seed)"
	@echo "4. Start backend:  go run ./cmd/picfast"
	@echo "5. Start frontend: cd web && pnpm dev"

seed:
	go run ./cmd/seed

## Build

frontend:
	cd web && pnpm build

build:
	go build -o ./bin/picfast ./cmd/picfast

build-full: frontend build

run: build
	./bin/picfast

## Code generation

generate: $(SQLC)
	$(SQLC) generate

$(SQLC):
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

## Database

migrate-up:
	migrate -path migrations -database "postgres://picfast:picfast@localhost:5432/picfast?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://picfast:picfast@localhost:5432/picfast?sslmode=disable" down 1

## Quality

format:
	gofmt -w ./internal ./cmd
	cd web && pnpm exec prettier --write "src/**/*.{ts,vue,css}"

lint:
	go vet ./...
	cd web && pnpm exec vue-tsc --noEmit

openapi-lint:
	pnpm --package=@redocly/cli dlx redocly lint --config .redocly.yaml api/openapi.yaml

## Dependencies

tidy:
	go mod tidy

## Docker

docker-up:
	docker compose -f docker/docker-compose.yml up --build -d

docker-down:
	docker compose -f docker/docker-compose.yml down

docker-logs:
	docker compose -f docker/docker-compose.yml logs -f app

## Testing

test:
	go test -v -count=1 ./...

docker-test-db:
	docker run -d --name picfast-test-db -e POSTGRES_USER=picfast -e POSTGRES_PASSWORD=picfast -e POSTGRES_DB=picfast_test -p 5433:5432 postgres:16-alpine

## Cleanup

clean:
	rm -rf bin/ data/ web-dist/
