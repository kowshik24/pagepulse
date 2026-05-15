# PagePulse

PagePulse is a lightweight Go system telemetry dashboard. It exposes live CPU, memory, disk, and network metrics through JSON APIs and Server-Sent Events (SSE), then renders them in an embedded web UI.

## Features
- Live dashboard cards for CPU, memory, disk, and network throughput.
- SSE stream at `/api/v1/stream` for near real-time updates.
- JSON summary endpoint at `/api/v1/summary`.
- Static resource catalog endpoint at `/api/v1/resources`.
- Built-in dark/light theme toggle (persisted in browser storage).

## Run Locally
```bash
go run ./cmd/pagepulse
```

Common options:
```bash
go run ./cmd/pagepulse --host 127.0.0.1 --port 8421 --sample-interval 1s
go run ./cmd/pagepulse --public --port 8421
```

Open: `http://127.0.0.1:8421`

## Development & Quality
```bash
go test ./...
go test -race ./...
gofmt -w .
go build -o pagepulse ./cmd/pagepulse
```

Build with explicit version metadata:
```bash
go build -ldflags "-X 'pagepulse/internal/buildinfo.Version=v0.1.0' -X 'pagepulse/internal/buildinfo.Commit=$(git rev-parse --short HEAD)' -X 'pagepulse/internal/buildinfo.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" -o pagepulse ./cmd/pagepulse
```

GitHub Actions CI runs formatting checks, unit tests, and race tests on pushes and PRs.

## API Endpoints
- `GET /api/v1/summary`: current telemetry + trend arrays.
- `GET /api/v1/resources`: host/OS/cpu cores + filtered disk/interface catalog.
- `GET /api/v1/version`: running binary metadata (version, commit, build time, Go version).
- `GET /api/v1/stream`: SSE `summary` events.

## Current Project Status
Implemented (done):
- Core collector with periodic sampling and in-memory trend history.
- Dashboard UI with theme toggle and connection state badge.
- SSE streaming with automatic client reconnect.
- Resource filtering to avoid noisy internal mounts/tunnel interfaces.
- Unit tests for config parsing, history cap, filters, and HTTP behavior.
- CI workflow for format, test, and race checks.

Remaining / potential enhancements:
- Frontend unit tests (JS rendering/format helpers).
- Optional disk/interface filters configurable via flags.
- Export snapshots (CSV/JSON) and alert thresholds.
