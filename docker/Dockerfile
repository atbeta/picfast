# ---- Stage 1: Build Frontend ----
FROM node:24-slim AS frontend-builder

WORKDIR /web
COPY web/package.json web/pnpm-lock.yaml* ./
RUN npm install -g pnpm@10.33.0 && pnpm install --frozen-lockfile

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
    gnupg \
    libvips \
    fontconfig \
    fonts-wqy-microhei \
    ca-certificates \
    tzdata \
    wget \
    && install -d /usr/share/postgresql-common/pgdg \
    && wget -qO- https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.gpg \
    && echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.gpg] https://apt.postgresql.org/pub/repos/apt bookworm-pgdg main" > /etc/apt/sources.list.d/pgdg.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends postgresql-client-16 \
    && rm -rf /var/lib/apt/lists/*

# Apply security patches (e.g. libgnutls30 CVEs) after base installs
RUN apt-get update && apt-get upgrade -y && rm -rf /var/lib/apt/lists/*

RUN groupadd --system --gid 10001 picfast \
    && useradd --system --uid 10001 --gid picfast --home-dir /home/picfast --create-home --shell /usr/sbin/nologin picfast \
    && mkdir -p /app/data/uploads /app/data/thumbnails /app/data/backups

COPY --from=backend-builder /picfast /usr/local/bin/picfast
COPY --from=backend-builder /app/migrations /app/migrations
COPY --from=backend-builder /app/api/openapi.yaml /app/api/openapi.yaml
COPY --from=frontend-builder /web-dist /web-dist

RUN chown -R picfast:picfast /app /web-dist /home/picfast

WORKDIR /app

EXPOSE 8080

ENV PICFAST_SERVER_WEB_DIR=/web-dist \
    MALLOC_ARENA_MAX=2

USER picfast

ENTRYPOINT ["picfast"]
