# GoreeCloud Search Candidate Outcome Observability

## Purpose

This document defines how GoreeCloud Search records the outcome of the controlled first-Stable candidate workflow without relying on a single reporting surface.

## Outcome Sinks

A completed candidate request reports its sanitized outcome through two independent GitHub surfaces:

1. A commit status using the context `goreecloud/first-stable-candidate` on the exact candidate source revision.
2. A top-level pull-request conversation receipt on the first-Stable stabilization pull request.

The commit status links to the exact GitHub Actions workflow run and reports only `success` or `failure`. The pull-request receipt records the request-validation result, candidate image/evidence result, overall outcome, deterministic artifact locator when applicable, and isolated candidate/rollback rehearsal state.

Neither outcome sink substitutes for inspecting the candidate release-evidence artifact and its checksums.

## Independence and Redundancy Requirement

The commit-status publication and pull-request receipt must be attempted independently. Failure to publish one outcome sink must not prevent the workflow from attempting the other.

Each reporting step records its own original outcome while allowing the alternate sink to run. The workflow then enforces redundancy: candidate outcome observability remains satisfied when at least one outcome sink succeeds, but the outcome-reporting job fails closed when both outcome sinks fail.

This prevents a transient issue-comment API failure from making the candidate outcome invisible through commit status, prevents a commit-status API failure from suppressing the durable pull-request receipt, and avoids marking an otherwise successful candidate publication/rehearsal as failed solely because one redundant reporting surface was unavailable.

## Security and Privacy Boundary

Outcome reporting must not include:

- registry credentials;
- GitHub tokens;
- application secrets;
- user search queries;
- provider response content;
- production configuration;
- private runtime data;
- production-cutover authorization;
- Stable-promotion authorization.

The outcome surfaces are status and artifact locators only.

## Release Boundary

A successful commit status or pull-request receipt proves only that the controlled candidate workflow reported success for its own validation, immutable-image publication, registry retrieval, and isolated image-level rehearsal path.

Production cutover, target-host changes, compatibility-name retirement, persistent-data changes, backup restoration, application-level recovery, and Stable promotion remain separate controlled decisions requiring their own evidence.
