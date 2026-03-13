## Context

The project currently documents local Go builds and runtime configuration, but it does not include a repository-native CI workflow for container image builds. There is also no existing `Dockerfile` in the repository, so the change needs to define both the GitHub Actions workflow responsibilities and the expected image build entrypoint. The workflow should fit the current Go project layout, remain simple to review, and avoid introducing registry publishing before the build contract is established.

## Goals / Non-Goals

**Goals:**
- Add a GitHub Actions workflow that automatically builds the service container image from repository contents.
- Ensure the workflow runs in common validation scenarios such as pull requests and protected branch updates.
- Define a deterministic image tagging approach suitable for CI verification and future publishing.
- Fail fast when the image build cannot complete successfully.

**Non-Goals:**
- Publishing images to Docker Hub, GHCR, or Aliyun Container Registry in this change.
- Adding deployment automation or release orchestration.
- Reworking application runtime behavior unrelated to containerization.

## Decisions

- Use GitHub Actions as the CI entry point because the change is explicitly about repository-hosted workflow automation and GitHub is the source-control platform.
- Add a dedicated workflow file under `.github/workflows/` rather than embedding image build steps into unrelated checks, so image build behavior is isolated and easier to extend later.
- Define the workflow around standard repository events such as `pull_request` and pushes to the main branch, because these cover both change validation and mainline build verification.
- Use `docker build` against a repository `Dockerfile` as the build mechanism because it mirrors how downstream environments consume the service and provides a portable container contract.
- Use non-publishing tags derived from commit SHA and branch/ref metadata so builds are traceable without requiring registry credentials.
- Keep credentials out of scope for the initial workflow so the automation remains safe for forks and simple validation paths.

## Risks / Trade-offs

- [No existing Dockerfile] → The workflow cannot succeed until a supported image build definition is added; implementation should add or validate the Dockerfile alongside the workflow.
- [Longer CI runtime] → Building a container image increases validation time; keep the workflow focused on build verification only.
- [Future publishing mismatch] → A local-only tagging strategy may need revision when registry pushes are added later; document the tag format clearly so it can evolve predictably.
- [GitHub-hosted runner differences] → Builds may behave differently from local environments; prefer a straightforward Docker build with minimal runner assumptions.

## Migration Plan

- Add the workflow file and any required Docker build assets in the same change.
- Validate the workflow on a feature branch and confirm it builds successfully in pull requests.
- Merge to the default branch so subsequent changes automatically receive image build validation.
- If the workflow causes unexpected failures, disable or revert the workflow file without affecting runtime service behavior.

## Open Questions

- Should the initial workflow build on every branch push or only on pull requests and the default branch?
- Should implementation include build cache optimization now, or defer until the basic workflow is proven?
- Which container base image should the Dockerfile standardize on for production use?
