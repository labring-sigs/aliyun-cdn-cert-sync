## ADDED Requirements

### Requirement: Build container image in GitHub Actions
The system SHALL define a GitHub Actions workflow that builds the service container image from the repository source using a repository-managed Docker build definition.

#### Scenario: Pull request build validation
- **WHEN** a pull request updates the repository source
- **THEN** GitHub Actions SHALL run the image build workflow and fail the check if the container image build does not succeed

### Requirement: Trigger image builds on mainline changes
The system SHALL run the image build workflow for changes merged to the default branch so the mainline repository state always has a verified container build.

#### Scenario: Default branch update
- **WHEN** code is pushed to the default branch
- **THEN** GitHub Actions SHALL execute the image build workflow for that revision

### Requirement: Produce deterministic CI image tags
The system SHALL assign image tags or labels derived from repository revision metadata so each CI build result can be traced to the source commit without requiring publication to an external registry.

#### Scenario: Workflow builds a revision
- **WHEN** the workflow builds an image for a specific commit
- **THEN** the build configuration SHALL include a deterministic identifier based on that revision in the image tag, label, or equivalent build metadata

### Requirement: Avoid registry publishing in initial workflow
The system SHALL limit the initial GitHub Actions workflow to image build validation and SHALL NOT require registry credentials or publish images as part of this change.

#### Scenario: Workflow runs for an external contribution
- **WHEN** the workflow executes for a pull request without access to repository secrets
- **THEN** the image build job SHALL complete without requiring image push credentials
