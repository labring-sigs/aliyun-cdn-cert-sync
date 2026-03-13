# runtime-config Specification

## Purpose

Define runtime configuration and security requirements for the certificate sync service.

## Requirements

### Requirement: Load required configuration at startup

The system SHALL require explicit startup configuration for Kubernetes Secret reference, Aliyun region, Aliyun credential source, and one or more CDN domains.

#### Scenario: Required Aliyun configuration missing

- **WHEN** Aliyun region or credential source is missing
- **THEN** startup validation SHALL fail and the process SHALL exit with a clear error

#### Scenario: No CDN domains configured

- **WHEN** the configured CDN domain list is empty
- **THEN** startup validation SHALL fail and prevent running a sync loop

### Requirement: Define credential source precedence

The system SHALL use a deterministic credential source precedence order and log only the selected source type.

#### Scenario: Multiple credential sources are available

- **WHEN** environment credentials and instance role credentials are both available
- **THEN** the system SHALL select the configured source according to documented precedence and avoid logging secret values

### Requirement: Protect secrets in logs

The system SHALL prevent leaking sensitive values in logs and error messages.

#### Scenario: Sync request fails

- **WHEN** errors are logged during certificate sync
- **THEN** logs SHALL exclude private key material, full certificate PEM bodies, and raw credential values

### Requirement: Expose operational signals

The system SHALL emit metrics and structured logs for sync outcome observability.

#### Scenario: Sync completes

- **WHEN** a sync loop finishes
- **THEN** the system SHALL record success/failure status, retry counts, and affected domain count
