# ---- Stage 1: Build Frontend ----
FROM node:22-alpine AS frontend-builder

WORKDIR /web
COPY web/package.json web/pnpm-lock.yaml* ./
RUN corepack enable && corepack prepare pnpm@latest --activate && pnpm install --frozen-lockfile

COPY web/ .
RUN pnpm build

# ---- Stage 2: Build Backend ----
FROM golang:1.26-alpine AS backend-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /picfast ./cmd/picfast

# ---- Stage 3: Runtime ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata

COPY --from=backend-builder /picfast /usr/local/bin/picfast
COPY --from=backend-builder /app/migrations /migrations
COPY --from=frontend-builder /web-dist /web-dist

EXPOSE 8080

ENV PICFAST_SERVER_WEB_DIR=/web-dist

ENTRYPOINT ["picfast"]
