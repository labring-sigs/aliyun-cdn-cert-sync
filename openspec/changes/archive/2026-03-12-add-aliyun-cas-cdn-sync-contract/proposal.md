# Change: Define explicit Aliyun CAS and CDN sync contract

## Why

Current specs describe high-level behavior but do not define concrete Aliyun API-level expectations. This leaves room for implementation ambiguity in certificate upload identity, CDN binding payloads, and retry boundaries.

## What Changes

- Clarify certificate upload target as Aliyun CAS-managed certificate resources.
- Define deterministic certificate identity using certificate fingerprint and Aliyun certificate ID mapping.
- Define CDN domain HTTPS update behavior using the latest successful CAS certificate.
- Define retryable vs non-retryable failure handling at CAS and CDN API boundaries.
- Define required runtime configuration keys for Aliyun region, credential source, and domain list.

## Impact

- Affected specs: `certificate-sync`, `cdn-binding`, `runtime-config`
- Affected code areas:
  - Aliyun API adapter/client package
  - Reconciliation flow and state tracking
  - Startup config validation

