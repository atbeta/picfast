FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /picfast ./cmd/picfast

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /picfast /usr/local/bin/picfast
COPY migrations /migrations

EXPOSE 8080
ENTRYPOINT ["picfast"]
