# Repository Guidelines

## Project Structure & Module Organization

This project syncs TLS certificates from Kubernetes cert-manager to Aliyun and updates CDN certificate bindings. Keep this Go layout:

- `cmd/cdn-cert-sync/`: main entrypoint and CLI flags
- `internal/k8s/`: Kubernetes API and Secret fetch logic
- `internal/aliyun/`: Aliyun SSL/CDN API clients and request mapping
- `internal/sync/`: reconcile workflow and retry/backoff logic
- `configs/`: sample config files (non-secret only)
- `deploy/`: Kubernetes manifests (RBAC, Deployment, CronJob)
- `docs/`: runbooks and troubleshooting notes

Keep provider-specific code inside `internal/aliyun` and orchestration in `internal/sync`.

## Build, Test, and Development Commands

Use standard Go commands during development:

- `go mod tidy`: normalize dependencies.
- `go build ./...`: compile all packages.
- `go test ./...`: run tests across all packages.
- `go test -race ./...`: detect data races in concurrent sync logic.
- `go run ./cmd/cdn-cert-sync --config ./configs/config.yaml`: run locally.

If a `Makefile` is added, keep these as target backends.

## Coding Style & Naming Conventions

Follow idiomatic Go style and keep code easy to review:

- Formatting: run `gofmt ./...` before committing.
- Linting: run `golangci-lint run` if configured.
- Package names: short, lowercase, no underscores (example: `internal/sync`).
- Exported identifiers: `PascalCase`; unexported: `camelCase`.
- Errors: wrap with context (`fmt.Errorf("upload cert: %w", err)`).

YAML files in `deploy/` and `configs/` should use 2-space indentation.

## Testing Guidelines

Testing should cover certificate parsing, API mapping, and reconcile behavior:

- Place tests next to code as `*_test.go`.
- Name tests by behavior (example: `TestSync_ReplacesExpiredCertificate`).
- Mock Aliyun and Kubernetes clients in unit tests; avoid live cloud calls in CI.
- Keep tests deterministic and runnable via `go test ./...`.

Aim for strong coverage in `internal/sync` and error handling paths.

## Commit & Pull Request Guidelines

Use clear, scoped commits and reviewable PRs:

- Commit format: `<scope>: <imperative summary>` (examples: `sync: add cert expiry check`, `aliyun: handle CDN cert bind retry`).
- PRs must include purpose, config changes, and test output (`go test ./...`).
- Link related issue/ticket and note rollback steps for production-impacting changes.

## Security & Configuration Tips

- Never commit PEM content, Kubernetes Secret dumps, or Aliyun AccessKey secrets.
- Use least-privilege RAM policies for SSL upload and CDN domain certificate updates only.
- Redact certificate subject/SAN details and domain names in logs when possible.
- Keep a `.env.example` or `configs/config.example.yaml` with placeholder values only.
