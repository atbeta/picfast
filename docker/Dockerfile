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
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
RUN go build -ldflags="-s -w -X github.com/atbeta/picfast/internal/version.Version=${VERSION} -X github.com/atbeta/picfast/internal/version.Commit=${COMMIT} -X github.com/atbeta/picfast/internal/version.BuildTime=${BUILD_TIME}" -o /picfast ./cmd/picfast

# ---- Stage 3: Runtime ----
FROM debian:12-slim

# Install runtime dependencies; fonts-wqy-microhei (~5MB) covers CJK watermark rendering.
RUN apt-get update && apt-get install -y --no-install-recommends \
    libvips \
    fontconfig \
    fonts-wqy-microhei \
    ca-certificates \
    tzdata \
    wget \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --system --gid 10001 picfast \
    && useradd --system --uid 10001 --gid picfast --home-dir /home/picfast --create-home --shell /usr/sbin/nologin picfast \
    && mkdir -p /app/data/uploads /app/data/thumbnails

COPY --from=backend-builder /picfast /usr/local/bin/picfast
COPY --from=backend-builder /app/migrations /app/migrations
COPY --from=backend-builder /app/api/openapi.yaml /app/api/openapi.yaml
COPY --from=frontend-builder /web-dist /web-dist

RUN chown -R picfast:picfast /app /web-dist /home/picfast

WORKDIR /app

EXPOSE 8080

ENV PICFAST_SERVER_WEB_DIR=/web-dist

USER picfast

ENTRYPOINT ["picfast"]
