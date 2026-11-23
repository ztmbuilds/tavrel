# syntax=docker/dockerfile:1

FROM node:24.22.3-alpine AS web

ENV COREPACK_ENABLE_DOWNLOAD_PROMPT=0

WORKDIR /src

RUN corepack enable

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./web/

WORKDIR /src/web
RUN pnpm install --frozen-lockfile

COPY web ./
RUN pnpm build

FROM golang:1.27.2-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web /src/internal/web/dist ./internal/web/dist

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/tavrel ./cmd/tavrel
RUN mkdir -p /out/data

FROM gcr.io/distroless/static-debian13:nonroot@sha256:aafe993665bf6200db38a02c11a9351e782a84b6a45e7f33b6695638c88e3931

COPY --from=build --chown=65532:65532 /out/data /data
COPY --from=build /out/tavrel /usr/local/bin/tavrel

WORKDIR /data

ENV APP_PORT=5678 \
    OTLP_HTTP_PORT=4318 \
    OTLP_GRPC_PORT=4317 \
    DB_PATH=/data/store.db

USER nonroot:nonroot

EXPOSE 5678 4318 4317

HEALTHCHECK --interval=30s --timeout=3s --start-period=15s --retries=3 \
  CMD ["/usr/local/bin/tavrel", "healthcheck"]

ENTRYPOINT ["/usr/local/bin/tavrel"]
