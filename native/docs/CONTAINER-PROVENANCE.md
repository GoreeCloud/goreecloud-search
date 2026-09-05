# GoreeCloud Search Native Development Container Provenance

## Status

Development evidence only. This document does not authorize registry publication, target-host deployment, production cutover, or Stable promotion.

The native Search container path packages the already-built GoreeCloud-owned `searchd` binary into a minimal runtime image. It is separate from the retained SearXNG-derived container path, which remains transitional continuity and rollback material while native migration is incomplete.

## Runtime base

The Development container uses the immutable runtime base:

`gcr.io/distroless/static-debian13:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6`

The base is pinned by digest in `native/container/Containerfile`. The image contains no compiler or package-manager step in the GoreeCloud runtime definition, and the GoreeCloud process runs explicitly as UID/GID `65532:65532`.

A future base update must be a reviewed source change. It must re-run the exact-head native container, native source, supply-chain, and applicable repository acceptance gates. A mutable tag alone is not acceptable release provenance.

## Build boundary

`.github/workflows/goreecloud-native-container-development.yml`:

- checks out the exact pull-request head or push revision;
- tests the native Go source before container packaging;
- invokes the existing reproducible Development package builder and reuses its exact-source Linux amd64 binary;
- verifies the binary's embedded Go VCS revision and unmodified-source state;
- pulls the exact digest-pinned distroless runtime base;
- builds a local Development image with source revision and Development version labels;
- verifies the OCI/GoreeCloud metadata, non-root user, immutable base declaration, and absence of secret-like labels/environment configuration;
- verifies that the resulting image has no `/bin/sh` runtime shell;
- runs the image read-only, with all Linux capabilities dropped, `no-new-privileges`, a bounded PID limit, and host publication restricted to loopback;
- validates health, bounded local readiness, status/build provenance, provider definitions, Home, Preferences, and zero-provider fail-closed behavior;
- validates the release-provider structural coverage startup gate using read-only synthetic provider configuration without contacting an external provider;
- saves the local image as an OCI archive, emits SHA-256 checksums and sanitized `native-container-provenance.json`, and uploads them as a retained GitHub Actions Development artifact.

The workflow has `contents: read` only. It does not authenticate to a container registry and has no package-write permission.

## Network boundary

The native binary itself continues to default to `127.0.0.1:8080` outside container-specific orchestration.

For isolated container acceptance, the workflow sets `GOREECLOUD_SEARCH_ADDR=0.0.0.0:8080` inside the container because container port forwarding requires the process to listen on the container network namespace. The host publication remains explicitly loopback-only (`127.0.0.1`). This container-internal bind is not permission to publish the application port publicly.

A production deployment must separately prove its approved private DNS, reverse-proxy/TLS, NetBird or equivalent private-network boundary, firewall/open-port state, identity/access model, and target-host configuration.

## Artifact identity

The Development image carries bounded OCI metadata for:

- GoreeCloud Search product/component identity;
- source repository;
- exact source revision;
- Development version identity;
- AGPL-3.0-or-later license;
- `org.goreecloud.lifecycle=development`;
- `org.goreecloud.production-approved=false`; and
- `org.goreecloud.release-candidate=false`.

The retained provenance JSON records the exact source revision, exact immutable runtime-base reference, local image ID, OCI archive SHA-256, and the completed isolated Development-runtime checks. It must retain:

- `release_lifecycle: development`;
- `production_approved: false`;
- `release_candidate_declared: false`;
- `registry_published: false`;
- `target_environment_validated: false`;
- `live_provider_acceptance_validated: false`; and
- `platform_conformance: nonconformant`.

A local Podman image ID or CI OCI archive is not a production registry digest and must not be represented as one.

## Secret and configuration boundary

The container image intentionally excludes production provider configuration, provider credentials, bearer tokens, Caddy configuration, private DNS state, NetBird state, firewall state, monitoring credentials, backup material, production environment files, and production authorization markers.

Provider credentials, when later used, remain environment/secret-manager references at the runtime boundary. They must not be copied into the image, labels, provenance records, logs, or retained CI artifacts.

## Provider and TLS boundary

The Development container gate does not contact a real external Search provider. Structural provider coverage uses a synthetic HTTPS endpoint only to prove startup/category wiring and does not issue a query to that endpoint.

Live HTTPS connectivity, provider certificate behavior, credentials, privacy/data-use terms, rate limits, abuse controls, timestamp authority, outage behavior, response quality, and General/Images/Videos/News/Files acceptance remain separate provider and target-environment gates.

## Production and release boundary

This container work establishes a native **Development build/package/runtime evidence layer**. It does not establish:

- a Release Candidate;
- a published registry artifact;
- a production-approved immutable image;
- target-host validation;
- production provider approval;
- whole-application Glaze UI acceptance;
- Privacy Shield, Wardveil Security, Everkeep, Identity, Mesh, or Manager production acceptance;
- monitoring, alert delivery, restore, recovery, migration, or rollback acceptance;
- deployment to `search.goreecloud.com`; or
- Stable qualification.

Those states remain separately governed and must be backed by exact evidence for the actual candidate and target runtime.
