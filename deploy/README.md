# Kubernetes deployment

This project is best run as a `CronJob` because `cdn-cert-sync` performs a single sync and exits.

Files:

- `deploy/rbac.yaml`: Service account and least-privilege RBAC to read one TLS Secret.
- `deploy/configmap.yaml`: non-secret application config mounted at `/etc/cdn-cert-sync/config.yaml`.
- `deploy/secret.example.yaml`: example Secret manifest for Aliyun credentials.
- `deploy/cronjob.yaml`: scheduled sync job using in-cluster Kubernetes access.
- `deploy/kustomization.yaml`: base manifest bundle.

Usage:

1. Update `deploy/configmap.yaml` with the cert-manager Secret name/namespace, CDN domains, state path, and Aliyun resource group if you use one.
2. Copy `deploy/secret.example.yaml` to a real Secret manifest and fill in Aliyun credentials.
3. Replace the image in `deploy/cronjob.yaml` with your published image reference.
4. Apply the manifests:

   ```bash
   kubectl apply -f deploy/secret.yaml
   kubectl apply -k deploy/
   ```

Notes:

- The included `Role` grants `get` on Secrets in the namespace where you deploy these manifests. If the TLS Secret lives in another namespace, create the RBAC there or change the deployment namespace strategy.
- `emptyDir` stores the sync state per job run. If you need state persisted across runs, replace it with a persistent volume and keep `sync.stateFile` aligned with that mount.
