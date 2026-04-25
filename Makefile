.PHONY: build run generate migrate-up migrate-down docker-up docker-down tidy

GOPATH := $(shell go env GOPATH)
SQLC := $(GOPATH)/bin/sqlc

build:
	CGO_ENABLED=0 go build -o ./bin/imgapi ./cmd/imgapi

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
	go test ./...

clean:
	rm -rf bin/ data/
