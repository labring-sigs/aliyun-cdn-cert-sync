## Why

The repository can build locally, but it lacks a standardized GitHub Actions workflow to build the service container image automatically on changes. Adding a workflow now makes image creation reproducible, reduces manual release steps, and provides a clear CI entry point for future publishing or deployment automation.

## What Changes

- Add a GitHub Actions workflow that checks out the repository and builds the service container image.
- Define the workflow triggers and job structure needed for pull request validation and branch-based builds.
- Establish the required build context, image tagging approach, and failure behavior for automated image builds.
- Document the expected inputs and constraints so future implementation can extend the workflow for pushes to a registry.

## Capabilities

### New Capabilities
- `image-build-pipeline`: Defines automated GitHub workflow behavior for building the project container image in CI.

### Modified Capabilities
- None.

## Impact

- Adds GitHub Actions workflow files under `.github/workflows/`.
- Likely adds or updates a `Dockerfile` and related build context files if they do not already exist.
- May influence release and deployment processes by introducing CI-produced image artifacts.
- No runtime API changes to the Go service are expected.
