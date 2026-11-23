# tavrel

A small, self-hosted observability platform for OpenTelemetry. tavrel accepts OTLP traces and logs and gives you a UI to view, filter and query them. There is no custom SDK — any OpenTelemetry-instrumented service works by pointing its exporter at tavrel.

tavrel is early software (0.x). Expect breaking changes in minor releases; see [CHANGELOG.md](CHANGELOG.md) and [RELEASING.md](RELEASING.md).

## Quick start

tavrel is distributed as a multi-arch Docker image. The compose file pulls the image; it does not build from source.

```sh
docker compose up -d
```

Open <http://localhost:5678>. Send OTLP telemetry to the host on port 4318 (HTTP) or 4317 (gRPC).

To pin a release, create a `.env` next to `compose.yaml`:

```dotenv
TAVREL_VERSION=0.1.0
```

or run `TAVREL_VERSION=0.1.0 docker compose up -d`. The default, `latest`, tracks the newest stable release.

| Port | Purpose      |
| ---- | ------------ |
| 5678 | Web UI + API |
| 4318 | OTLP/HTTP    |
| 4317 | OTLP/gRPC    |

Telemetry is stored in the `tavrel_tavrel-data` volume at `/data`. The container is otherwise stateless.

### Pointing exporters at tavrel

```sh
OTEL_EXPORTER_OTLP_ENDPOINT=http://<host>:4318
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
```

## Configuration

`compose.yaml` reads these from the environment or a `.env` file. See [.env.example](.env.example).

| Variable | Default | Purpose |
| --- | --- | --- |
| `APP_PORT` | `5678` | UI/API port |
| `OTLP_HTTP_PORT` | `4318` | OTLP/HTTP port |
| `OTLP_GRPC_PORT` | `4317` | OTLP/gRPC port |
| `DB_PATH` | `/data/store.db` | SQLite database path |
| `RETENTION_DAYS` | `30` | Delete telemetry older than this; `0` keeps everything |
| `WORKER_COUNT` | `4` | Ingest workers |
| `TASK_QUEUE_SIZE` | `256` | Pending OTLP requests before the receiver returns 429 |
| `OTLP_MAX_REQUEST_SIZE` | `8388608` | Max decompressed OTLP request body (8 MiB) |
| `AUTH_TOKEN` | empty | Protect the UI and API; empty disables authentication |
| `INGEST_TOKEN` | empty | Require `X-API-Key` / `x-api-key` on OTLP ingest |

`AUTH_TOKEN` and `INGEST_TOKEN` must be 16–512 printable ASCII characters with no whitespace. An invalid value stops startup.

## Authentication

The UI/API and the OTLP receivers have separate tokens.

- **`AUTH_TOKEN`** protects the UI and `/api/v1`. Open the UI and enter the token, or send `Authorization: Bearer <token>` to the API. A successful login sets a seven-day session cookie. Rotating the token invalidates existing sessions.
- **`INGEST_TOKEN`** protects OTLP ingest. HTTP exporters send `X-API-Key`; gRPC exporters send the same value as `x-api-key` metadata. For the OpenTelemetry SDKs:

  ```sh
  OTEL_EXPORTER_OTLP_HEADERS=x-api-key=<token>
  ```

If both are empty, tavrel is open. Keep the ingest ports on a trusted network or set `INGEST_TOKEN` before exposing them.

## Data, backup and retention

The database lives in the `tavrel_tavrel-data` volume. Stop the container before copying it so SQLite checkpoints the WAL:

```sh
docker compose stop tavrel
docker run --rm \
  -v tavrel_tavrel-data:/data \
  -v "$PWD":/backup \
  alpine tar czf "/backup/tavrel-$(date +%F).tgz" -C /data .
docker compose start tavrel
```

Restore by stopping the container, extracting the archive into the volume, and starting it again.

`RETENTION_DAYS` deletes old traces, spans and logs at startup and once per retention period. `0` disables retention. Resources referenced by deleted telemetry are not pruned yet.

`docker compose down -v` deletes the volume and all telemetry. `docker compose down` keeps it.

## Upgrading

```sh
docker compose pull
docker compose up -d
```

Migrations run automatically at startup. Back up the volume before upgrading, especially while tavrel is on 0.x. Pin `TAVREL_VERSION` if you need reproducible installs.

## Reverse proxy

Terminate TLS at your proxy and forward the UI/API to port 5678. Forward OTLP/HTTP to 4318; exporters send to `/v1/traces` and `/v1/logs` under the endpoint you configure. gRPC ingest on 4317 needs HTTP/2 end to end.

Set `X-Forwarded-Proto: https` on proxied requests so session cookies are issued as `Secure`.

## Development

- `make install` — install frontend dependencies
- `make dev` — Vite dev server plus the backend with live reload
- `make build` — build the frontend into the Go binary and compile `tmp/main`
- `make test` — run Go tests

To build and run the image locally:

```sh
docker compose -f compose.yaml -f compose.dev.yaml up -d --build
```

The dev override builds `tavrel:local`; the user-facing `compose.yaml` always pulls the published image. Check the embedded version with:

```sh
docker run --rm tavrel:local version
```

See the `Makefile` for the full command set.

## License

Apache License 2.0. See [LICENSE](LICENSE).
