# aliyun-cdn-cert-sync

Sync Kubernetes cert-manager TLS Secrets to Aliyun CAS and bind certificates to Aliyun CDN domains.

## Run Modes

The binary supports two adapter modes:

- `memory`: local/dev mode with in-memory Kubernetes and Aliyun adapters.
- `api`: live mode using Kubernetes API and Aliyun RPC APIs.

Set mode in config (`runtime.adapterMode`) or env (`CDN_CERT_SYNC_ADAPTER_MODE`).

## Build

Default build (no Kubernetes client-go integration):

```bash
go build ./...
```

Build the container image locally:

```bash
docker build -t aliyun-cdn-cert-sync:local .
```

Build with Kubernetes `client-go` integration:

```bash
go mod tidy
go build -tags clientgo ./...
```

## Test

Unit tests:

```bash
go test ./...
```

Aliyun live integration tests are opt-in and excluded from normal test runs. To use them:

1. Copy `internal/aliyun/testdata/aliyun-live.env.example` to `internal/aliyun/testdata/aliyun-live.env`
2. Fill in real Aliyun credentials, endpoints, and a known certificate fingerprint
3. Optionally set:
   - `ALIYUN_LIVE_CERT_ID` and `ALIYUN_LIVE_CDN_DOMAIN` for the live CDN binding test
   - `ALIYUN_LIVE_UPLOAD_CERT_PATH` and `ALIYUN_LIVE_UPLOAD_KEY_PATH` for the live CAS upload+cleanup test
3. Run:

```bash
go test -tags integration ./internal/aliyun
```

Or use the Makefile shortcut:

```bash
make test-integration
```

These integration tests make real Aliyun API calls. The CDN binding test can update the configured domain's certificate binding, and the upload test creates a real CAS certificate before deleting it during cleanup.

The CDN live binding test only runs if `ALIYUN_LIVE_CERT_ID` and `ALIYUN_LIVE_CDN_DOMAIN` are set. The upload+cleanup CAS test only runs if `ALIYUN_LIVE_UPLOAD_CERT_PATH` and `ALIYUN_LIVE_UPLOAD_KEY_PATH` are set.

## CI Image Build

GitHub Actions builds the container image for every pull request and every push to `main` using `.github/workflows/image-build.yml`.

- The workflow only validates image builds; it does not publish to any registry.
- Image metadata is deterministic and derived from Git refs and `github.sha`.
- No registry credentials or repository secrets are required for the build job.

Local prerequisites for container builds:

- Docker with BuildKit support enabled.
- Repository files present as checked in, including `Dockerfile`, `go.mod`, `cmd/`, `internal/`, and `configs/config.example.yaml`.

## Configure

Start from `configs/config.example.yaml`.

Required in `api` mode:

- `kubernetes.secretNamespace`
- `kubernetes.secretName`
- `aliyun.region`
- `aliyun.credentialSource` (`env` supported)
- `aliyun.casEndpoint`
- `aliyun.cdnEndpoint`
- `aliyun.cdnDomains` (non-empty)
- `sync.stateFile`

When `aliyun.credentialSource=env`, set:

- `CDN_CERT_SYNC_ALIYUN_ACCESS_KEY_ID`
- `CDN_CERT_SYNC_ALIYUN_ACCESS_KEY_SECRET`

Other supported overrides:

- `CDN_CERT_SYNC_K8S_SECRET_NAMESPACE`
- `CDN_CERT_SYNC_K8S_SECRET_NAME`
- `CDN_CERT_SYNC_CDN_DOMAINS` (comma-separated)
- `CDN_CERT_SYNC_MAX_RETRIES`
- `CDN_CERT_SYNC_RETRY_BASE_MILLIS`
- `CDN_CERT_SYNC_STATE_FILE`

## Run

Local dry run:

```bash
go run ./cmd/cdn-cert-sync --adapter-mode memory --config ./configs/config.example.yaml
```

Live run with in-cluster Kubernetes config:

```bash
go run -tags clientgo ./cmd/cdn-cert-sync --adapter-mode api --in-cluster=true --config ./configs/config.example.yaml
```

Live run with kubeconfig:

```bash
go run -tags clientgo ./cmd/cdn-cert-sync --adapter-mode api --in-cluster=false --kubeconfig ~/.kube/config --config ./configs/config.example.yaml
```

## Kubernetes

Sample Kubernetes manifests live under `deploy/`.

- `deploy/cronjob.yaml` runs the sync as a scheduled `CronJob`.
- `deploy/rbac.yaml` creates the service account and Secret-read RBAC.
- `deploy/configmap.yaml` provides a non-secret config file.
- `deploy/secret.example.yaml` shows the expected Aliyun credential Secret shape.

See `deploy/README.md` for deployment steps and customization notes.
