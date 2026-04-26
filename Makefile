.PHONY: build run generate migrate-up migrate-down docker-up docker-down tidy frontend test

GOPATH := $(shell go env GOPATH)
SQLC := $(GOPATH)/bin/sqlc

frontend:
	cd web && pnpm build

build:
	CGO_ENABLED=0 go build -o ./bin/imgapi ./cmd/imgapi

build-full: frontend build

run: build
	./bin/imgapi

generate: $(SQLC)
	$(SQLC) generate

$(SQLC):
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

migrate-up:
	migrate -path migrations -database "postgres://imgapi:imgapi@localhost:5432/imgapi?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://imgapi:imgapi@localhost:5432/imgapi?sslmode=disable" down 1

tidy:
	go mod tidy

docker-up:
	docker compose -f docker/docker-compose.yml up --build -d

docker-down:
	docker compose -f docker/docker-compose.yml down

docker-logs:
	docker compose -f docker/docker-compose.yml logs -f app

test:
	go test -v -count=1 ./...

docker-test-db:
	docker run -d --name imgapi-test-db -e POSTGRES_USER=imgapi -e POSTGRES_PASSWORD=imgapi -e POSTGRES_DB=imgapi_test -p 5433:5432 postgres:16-alpine

clean:
	rm -rf bin/ data/ web-dist/
