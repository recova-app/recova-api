# syntax=docker/dockerfile:1.10

FROM golang:1.25.0-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags='-s -w' -o /out/recova-api ./cmd/api

FROM alpine:3.22

ARG VERSION=dev
ARG VCS_REF=unknown
ARG BUILD_DATE=unknown

LABEL org.opencontainers.image.title="recova-backend-v2" \
      org.opencontainers.image.description="Recova Backend API" \
      org.opencontainers.image.version="$VERSION" \
      org.opencontainers.image.revision="$VCS_REF" \
      org.opencontainers.image.created="$BUILD_DATE"

RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -g 10001 -S recova && \
    adduser -u 10001 -S recova -G recova

WORKDIR /app
COPY --from=builder /out/recova-api /app/recova-api
COPY --from=builder /src/docs/generated/openapi.yaml /app/docs/generated/openapi.yaml

EXPOSE 3000

USER recova:recova

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- "http://127.0.0.1:${PORT:-3000}/health/live" >/dev/null || exit 1

ENTRYPOINT ["/app/recova-api"]
