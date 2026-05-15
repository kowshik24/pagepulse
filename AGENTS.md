# Repository Guidelines

## Project Structure & Module Organization
`pagepulse` is a small Go service with a single executable entrypoint and internal packages:
- `cmd/pagepulse/main.go`: CLI bootstrap and process lifecycle.
- `internal/app`: config parsing and application wiring.
- `internal/metrics`: metric collection and time-series helpers.
- `internal/web`: HTTP server, JSON APIs, SSE stream, and embedded static assets.
- `internal/web/static`: frontend files (`index.html`, `app.js`, `styles.css`) embedded via `go:embed`.

Keep new backend code inside `internal/<domain>` and wire it through `internal/app` instead of adding logic to `main.go`.

## Build, Test, and Development Commands
- `go run ./cmd/pagepulse --port 8421`: run locally (default host `127.0.0.1`).
- `go run ./cmd/pagepulse --public --port 8421`: bind to `0.0.0.0` for remote access.
- `go build -o pagepulse ./cmd/pagepulse`: build the binary.
- `go test ./...`: run all unit tests.
- `go test -race ./...`: run tests with the race detector for concurrency-sensitive changes.
- `gofmt -w .`: format all Go files before committing.

## Coding Style & Naming Conventions
Use standard Go formatting and idioms:
- Formatting: always `gofmt` output (tabs for indentation).
- Package names: short, lowercase, no underscores (for example, `metrics`, `web`).
- Exported identifiers: `PascalCase`; unexported: `camelCase`.
- Errors: return wrapped errors with context (`fmt.Errorf("...: %w", err)`).

For frontend assets in `internal/web/static`, keep filenames lowercase and use clear API-aligned names.

## Testing Guidelines
Use Go’s `testing` package with table-driven tests where practical.
- Test files end with `_test.go`.
- Test functions follow `TestXxxBehavior` naming.
- Keep tests close to the package they validate (`internal/app/config_test.go`, `internal/metrics/collector_test.go`).

Run `go test ./...` before every PR. For changes in collectors, streaming, or shared state, also run `go test -race ./...`.

## Commit & Pull Request Guidelines
Git history is not available in this checkout, so use a consistent convention:
- Commit format: `type(scope): short summary` (for example, `feat(web): add resource endpoint`).
- Keep commits focused; avoid mixing refactors and behavior changes.

PRs should include:
- What changed and why.
- How to validate (exact commands run).
- API/UI impact notes (for `/api/v1/*` or static dashboard changes).
- Screenshots or sample JSON when UI/output changes.
