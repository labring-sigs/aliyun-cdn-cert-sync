## 1. Container Build Definition

- [x] 1.1 Add a repository `Dockerfile` for building the `cdn-cert-sync` service image.
- [x] 1.2 Validate the Docker build context and required files so the image can be built in CI without manual steps.

## 2. GitHub Actions Workflow

- [x] 2.1 Add a workflow under `.github/workflows/` that builds the container image on pull requests.
- [x] 2.2 Extend the workflow to also run on default-branch updates.
- [x] 2.3 Configure deterministic build metadata or tags based on commit information without publishing the image.
