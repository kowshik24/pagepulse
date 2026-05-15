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
GitHub Actions release pipeline builds cross-platform binaries and publishes them as release assets for `v*` tags.
GitHub Actions auto-tag pipeline can create the next `v*` tag from commit messages on `main`.

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

## Releases
Manual tag flow:
```bash
git tag v0.1.0
git push origin v0.1.0
```

Automatic flow:
- Push commits to `main`.
- Workflow `.github/workflows/auto-tag.yml` calculates the next semantic tag and pushes it.
- Pushed tag triggers `.github/workflows/release.yml`, which publishes binaries/checksums.
- Required once: add repo secret `RELEASE_PAT` (GitHub Personal Access Token with `repo` scope) so auto-created tags can trigger the release workflow.

Commit-message bump rules:
- `feat: ...` -> minor bump
- `BREAKING CHANGE` in commit body -> major bump
- everything else -> patch bump

To skip auto-tag for a commit, include `[skip-tag]` in the commit message.

Release artifacts include:
- `pagepulse_linux_amd64`, `pagepulse_linux_arm64`
- `pagepulse_darwin_amd64`, `pagepulse_darwin_arm64`
- `pagepulse_windows_amd64.exe`, `pagepulse_windows_arm64.exe`
- matching `.sha256` checksum files
