# Changelog

All notable changes to Tavrel are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html). While Tavrel is on 0.x, breaking changes may land in minor releases.

## [Unreleased]

## [0.1.0] - 2026-11-23

Initial public release.

### Added

- OTLP/HTTP and OTLP/gRPC ingest for traces and logs.
- SQLite storage with WAL, embedded migrations, and automatic retention.
- Web UI for the dashboard, traces, span trees, logs, spans, service graph, environments, filters, and time ranges.
- Optional authentication for the UI/API (`AUTH_TOKEN`) and separate OTLP ingest authentication (`INGEST_TOKEN`).
- Multi-arch Docker image based on distroless, running as non-root with a read-only root filesystem and a healthcheck.
- User-facing `compose.yaml` and a `compose.dev.yaml` override for local builds.
- `tavrel version` and `tavrel healthcheck` subcommands.

### Changed

- Trace writes are batched into one transaction per OTLP request instead of one transaction per trace.
- SQLite uses `synchronous=NORMAL` in WAL mode, removing a per-commit fsync from the ingest path.
- Defaults: `TASK_QUEUE_SIZE=256` (was 5000) and `OTLP_MAX_REQUEST_SIZE=8 MiB` (was 64 MiB).

### Fixed

- OTLP requests are rejected with 429/Unavailable once the ingest queue is full instead of growing without bound.
