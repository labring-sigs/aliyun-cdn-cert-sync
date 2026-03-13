# certificate-sync Specification

## Purpose

Define requirements for synchronizing cert-manager managed certificates from Kubernetes Secrets to Aliyun certificate storage.

## Requirements

### Requirement: Read certificate material from Kubernetes Secret

The system SHALL read `tls.crt` and `tls.key` from a configured Kubernetes Secret and validate that both values are present before sync.

#### Scenario: Secret contains valid TLS data

- **WHEN** the configured Secret exists and contains both `tls.crt` and `tls.key`
- **THEN** the system SHALL parse and stage certificate material for synchronization

#### Scenario: Secret is incomplete

- **WHEN** either `tls.crt` or `tls.key` is missing
- **THEN** the system SHALL fail the sync attempt and report an actionable error

### Requirement: Upload certificates idempotently to Aliyun

The system SHALL upload certificates to Aliyun CAS and maintain a deterministic mapping between Kubernetes certificate fingerprint and Aliyun certificate identifier. The system SHALL avoid duplicate uploads when an equivalent certificate already exists in CAS.

#### Scenario: Existing CAS certificate matches source fingerprint

- **WHEN** the Kubernetes `tls.crt` fingerprint matches an existing managed CAS certificate
- **THEN** the system SHALL reuse the existing Aliyun certificate identifier and skip creating a new certificate record

#### Scenario: No CAS certificate matches source fingerprint

- **WHEN** no managed CAS certificate matches the Kubernetes certificate fingerprint
- **THEN** the system SHALL create a new CAS certificate record and persist the returned Aliyun certificate identifier for subsequent CDN binding

### Requirement: Retry transient sync failures

The system SHALL classify CAS API failures into retryable and terminal categories and apply bounded retries only to retryable failures.

#### Scenario: CAS request is throttled or times out

- **WHEN** CAS returns throttling, timeout, or transient service errors
- **THEN** the system SHALL retry the request with bounded exponential backoff

#### Scenario: CAS request is invalid

- **WHEN** CAS returns validation, permission, or malformed request errors
- **THEN** the system SHALL fail the current sync attempt without retrying and surface the classified error
