# ---- Stage 1: Build Frontend ----
FROM node:24-slim AS frontend-builder

WORKDIR /web
COPY web/package.json web/pnpm-lock.yaml* ./
RUN corepack enable && corepack prepare pnpm@latest --activate && pnpm install --frozen-lockfile

COPY web/ .
RUN pnpm build && mv dist /web-dist

# ---- Stage 2: Build Backend ----
FROM golang:1.26-bookworm AS backend-builder

# Install libvips development headers for CGO build
RUN apt-get update && apt-get install -y --no-install-recommends \
    libvips-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /picfast ./cmd/picfast

# ---- Stage 3: Runtime ----
FROM debian:12-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    libvips \
    ca-certificates \
    tzdata \
    wget \
    && rm -rf /var/lib/apt/lists/*

COPY --from=backend-builder /picfast /usr/local/bin/picfast
COPY --from=backend-builder /app/migrations /migrations
COPY --from=frontend-builder /web-dist /web-dist

WORKDIR /app

EXPOSE 8080

ENV PICFAST_SERVER_WEB_DIR=/web-dist

ENTRYPOINT ["picfast"]
