# Design: Aliyun CAS and CDN sync contract

## API error classification

The runtime already exposes two internal categories for Aliyun API failures:

- `ErrRetryable`: safe to retry with bounded exponential backoff
- `ErrTerminal`: fail the current sync attempt immediately

The Aliyun adapter should classify CAS and CDN responses into those categories using the following contract.

| API surface | Example condition | Internal category | Reconcile behavior |
| --- | --- | --- | --- |
| CAS `DescribeUserCertificateList` | throttling, upstream timeout, temporary service unavailable, connection reset | `ErrRetryable` | retry list/query operation |
| CAS `DescribeUserCertificateList` | authentication failure, permission denied, invalid request parameters, malformed response contract | `ErrTerminal` | fail sync and surface context |
| CAS `UploadUserCertificate` | throttling, timeout, temporary 5xx, transport interruption after request dispatch | `ErrRetryable` | retry create operation, then re-check by fingerprint |
| CAS `UploadUserCertificate` | invalid PEM, invalid private key, unauthorized caller, unsupported region or action | `ErrTerminal` | fail sync without retry |
| CDN `SetDomainServerCertificate` | throttling, timeout, temporary 5xx, transient network failure | `ErrRetryable` | retry per-domain bind |
| CDN `SetDomainServerCertificate` | invalid domain, invalid certificate identifier, permission denied, malformed payload | `ErrTerminal` | record per-domain failure and continue |

### Classification rules

1. Transport-level failures from `context.DeadlineExceeded`, connection timeouts, temporary DNS resolution failures, and interrupted TLS sessions are treated as retryable.
2. Explicit Aliyun RPC error codes representing throttling, transient service instability, or timeout conditions are treated as retryable.
3. Validation, credential, authorization, and contract-shape failures are terminal because retrying cannot change the outcome.
4. Response parsing failures are terminal when the remote status indicates success but the body shape is incompatible with the expected contract; they indicate an implementation mismatch rather than a transient outage.

## Reconciliation state model

The reconciler needs only a small persisted state footprint to correlate Kubernetes certificates with Aliyun CAS identifiers.

### Required persisted fields

- `fingerprintToCertId`: map from normalized Kubernetes certificate fingerprint to Aliyun CAS certificate identifier

### Derived runtime fields

These values do not need persistence because they can be recalculated each run:

- Kubernetes secret namespace/name from validated config
- current certificate fingerprint derived from `tls.crt`
- `uploaded` flag for the current run report
- per-domain bind success/failure counters for the current run report
- aggregate retry count for the current run report

### State transitions

1. Read the Kubernetes secret and derive the certificate fingerprint.
2. Check persisted `fingerprintToCertId` for an existing Aliyun CAS identifier.
3. If no persisted mapping exists, query CAS by fingerprint.
4. When CAS returns a matching certificate or a new upload succeeds, persist `fingerprintToCertId[fingerprint] = certID`.
5. Use the resolved `certID` for CDN binding in the same sync run and future runs.

This keeps the persisted state deterministic, idempotent, and minimal while satisfying the certificate-sync requirement.

## Rollout and rollback notes

### Production rollout

1. Deploy with validated `aliyun.region`, `aliyun.credentialSource`, and at least one configured CDN domain.
2. Start with a canary deployment or single scheduled run against one low-risk certificate/domain pair.
3. Verify the first successful run records the expected `fingerprintToCertId` mapping in the configured state file.
4. Confirm CDN domain HTTPS configuration references the resolved Aliyun CAS certificate identifier.
5. Expand to the remaining domains after one full sync cycle completes without terminal errors.

### Rollback

1. Stop further sync runs by scaling the deployment down or suspending the CronJob.
2. Rebind affected CDN domains to the previously known good Aliyun certificate identifier.
3. Restore the prior state file if the new mapping should not be reused on the next run.
4. Correct configuration or certificate input issues before re-enabling sync.

### Operational cautions

- Treat state file updates and CDN binding as separate recovery points.
- Preserve the previous Aliyun certificate identifier during rollout so operators can manually rebind domains if needed.
- Do not delete the previous CAS certificate until all target domains have converged and the new certificate is verified.
