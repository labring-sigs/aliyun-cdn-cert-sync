## MODIFIED Requirements

### Requirement: Apply certificate to target CDN domains

The system SHALL bind each configured CDN domain to the Aliyun certificate identifier produced by the successful CAS reconciliation in the same sync run or the latest known successful run.

#### Scenario: CAS reconciliation returns certificate identifier

- **WHEN** CAS reconciliation produces a valid Aliyun certificate identifier
- **THEN** the system SHALL update HTTPS configuration for each target CDN domain to reference that identifier

#### Scenario: One domain binding fails

- **WHEN** one CDN domain update request fails but others succeed
- **THEN** the system SHALL continue processing remaining domains and report per-domain success/failure results

### Requirement: Ensure eventual consistency for binding

The system SHALL re-evaluate current CDN domain certificate association on each sync cycle and converge domains to the desired Aliyun certificate identifier.

#### Scenario: Domain already references desired certificate

- **WHEN** a CDN domain already references the desired Aliyun certificate identifier
- **THEN** the system SHALL skip update for that domain and record a no-op result

