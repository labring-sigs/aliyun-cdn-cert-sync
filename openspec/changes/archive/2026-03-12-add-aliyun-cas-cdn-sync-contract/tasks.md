## 1. Specification Updates

- [x] 1.1 Add certificate-sync deltas for CAS upload identity and idempotency rules.
- [x] 1.2 Add cdn-binding deltas for HTTPS domain update behavior and partial-failure handling.
- [x] 1.3 Add runtime-config deltas for required Aliyun and domain configuration keys.
- [x] 1.4 Add retry semantics deltas distinguishing retryable vs terminal errors.

## 2. Implementation Planning

- [x] 2.1 Map CAS and CDN API responses to internal error categories.
- [x] 2.2 Define reconciliation state fields required to correlate Kubernetes cert and Aliyun cert ID.
- [x] 2.3 Define rollout and rollback notes for production certificate replacement.
