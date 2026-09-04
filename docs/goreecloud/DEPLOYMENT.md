# GoreeCloud Search Deployment Boundary

## Status

Development documentation. Production cutover is not approved by this file.

The current source-versus-production readiness contract is maintained in `docs/goreecloud/READINESS.md`. Native build/package provenance is described in `native/docs/ARTIFACT-PROVENANCE.md`.

## Current runtime model

GoreeCloud Search is in an active native migration.

- **Target application:** the GoreeCloud-owned Go service under `native/`.
- **Current lifecycle:** Development.
- **Transitional dependency:** the inherited SearXNG-derived runtime remains available only for continuity, migration, compatibility, feature-preservation decisions, applicable upstream security maintenance, and rollback until native cutover requirements are satisfied.
- **Production/Stable approval:** not established by source presence, packaging, green CI, or this documentation.

The long-term application architecture is the native GoreeCloud Search service. The inherited SearXNG product architecture is not the permanent GoreeCloud Search runtime.

## Native Development artifact boundary

`.github/workflows/goreecloud-native-development-artifact.yml` produces exact-revision Linux amd64 and arm64 packages for the native `searchd` service.

The Development artifact path:

- explicitly checks out the pull-request head SHA or exact push SHA;
- verifies the clean exact source revision before building;
- builds each supported CI target twice and requires byte-identical binaries;
- preserves Go VCS revision metadata;
- packages the native binary with its license and machine-readable artifact metadata;
- emits SHA-256 checksums and `artifact-provenance.json`;
- launches the packaged amd64 binary on a Linux CI runner; and
- verifies health, status, local readiness, provider-definition, homepage, Preferences, and zero-provider fail-closed behavior from the packaged runtime.

The provenance record deliberately remains `release_lifecycle: development`, `production_approved: false`, `release_candidate_declared: false`, `target_environment_validated: false`, and `platform_conformance: nonconformant`.

This advances native Search to a reviewable build/package evidence boundary. It does **not** prove a target server, container platform, private route, provider, backup/restore path, monitoring system, or production cutover.

## Native runtime configuration

The native service defaults to `127.0.0.1:8080` through `GOREECLOUD_SEARCH_ADDR`. A non-loopback binding is not implied or approved merely because the environment variable can be changed.

External native providers are deployment-controlled through `GOREECLOUD_SEARCH_PROVIDER_CONFIG_FILE`. The source accepts no configured providers by default. A configured provider must satisfy the native provider schema and transport controls, and optional bearer credentials are referenced by environment-variable name rather than embedded in provider configuration.

Secrets and reusable credentials must not be committed, copied into ordinary build artifacts, written into provenance metadata, or exposed through status/provider-definition responses.

The native package intentionally does not include production provider configuration, provider credentials, Caddy configuration, private DNS state, NetBird state, monitoring credentials, backup material, firewall rules, or a production authorization marker.

## Native private-publication boundary

The intended GoreeCloud user path remains a reviewed private-access architecture. The application service should bind only to loopback or an explicitly approved private interface and should be published through the approved reverse-proxy/private-network boundary rather than exposing the application port directly to the public Internet.

Target-host acceptance must separately prove the actual private DNS, TLS/reverse proxy, NetBird/private networking or equivalent access boundary, firewall/open-port state, individual attribution/access requirements, and service identity used by that deployment.

The CI Development artifact does not establish those controls because a GitHub-hosted runner is not the production target environment.

## Transitional SearXNG deployment boundary

The repository still contains inherited SearXNG container workflows and GoreeCloud-owned transitional configuration such as `goreecloud/settings.yml.example` and `goreecloud/compose.yml.example`.

Those files remain relevant only to the transitional incumbent/rollback path while native migration is incomplete. A SearXNG-derived OCI image, historical SearXNG RC publication record, or successful transitional container test is not a native Search release artifact and must not be used as evidence that the native service has been packaged, deployed, or accepted.

Do not remove the transitional runtime from its rollback role until the applicable native migration, recovery, rollback, and cutover requirements are satisfied.

## Automated validation layers

Current repository validation includes both native and transitional gates. Their evidence scopes must not be blurred.

Native source/runtime gates include, as applicable:

- `goreecloud-native-foundation.yml` — native Go tests and source build;
- `goreecloud-native-results-browser-acceptance.yml` — deterministic native results/image rendered acceptance;
- `goreecloud-native-development-artifact.yml` — exact-revision Linux package provenance plus packaged-runtime CI acceptance;
- platform/API/integration workflows that exercise GoreeCloud-owned native contracts.

Transitional gates include retained SearXNG Integration, runtime-smoke, container-build, browser, provider, and compatibility checks. They remain useful for continuity and migration safety but do not establish native release acceptance.

A successful workflow proves only the boundary it actually evaluates.

## Provider and category acceptance

The native provider runtime supports deployment-controlled category-aware providers, but no real external provider is production-approved by this repository state.

Before native production cutover, each category selected for the release must have accepted live execution through an approved provider on the actual or representative target runtime. Acceptance must distinguish application defects from third-party throttling, access denial, CAPTCHA, rate limits, provider outages, malformed responses, timestamp-authority defects, or empty results.

General, Images, Videos, News, and Files remain subject to live-provider acceptance for the selected release scope even though deterministic native test providers already exercise all category contracts in source CI.

## Current Stable Glaze UI boundary

The repository Platform Contract currently requires GLAZE UI V1.1 / 1.1.0. Native rendered tests provide bounded responsive/accessibility evidence, but they do not establish complete current-Stable application acceptance.

Whole-application visual, accessibility, resilience, physical-device/browser, and final current-Stable Glaze acceptance remain separate release gates. A superseded or Candidate Glaze line must not be substituted for the authoritative current-Stable consumer requirement.

## Local Development artifact validation

A reviewed checkout can build Development packages with:

```bash
bash native/scripts/build-development-artifacts.sh "$(git rev-parse HEAD)" /tmp/goreecloud-search-native-artifacts
```

The script requires a canonical exact revision and a clean working tree. It writes Linux amd64/arm64 packages, `SHA256SUMS`, and `artifact-provenance.json` to the selected output directory.

The packages are Development evidence. Do not rename them as a Release Candidate, publish them as Stable, or use them as a production approval record without the separately required lifecycle and acceptance process.

## Release and artifact provenance

Release-critical artifacts must remain traceable to an exact source revision. A moving branch, `latest` tag, or unverified copied binary is not sufficient provenance.

The new native Development package workflow intentionally uploads GitHub Actions artifacts rather than publishing a new production image or release. Registry publication, signing/attestation policy, immutable release identity, and target-host deployment remain later controlled gates once an actual Release Candidate is justified by evidence.

Historical inherited `goreecloud-rc-publication.yml` behavior applies to the transitional SearXNG-derived release line and must not be interpreted as the native Search publication path.

## Monitoring, recovery, and rollback

Before production approval, native Search still requires evidence-backed:

- process/service and private-route monitoring;
- actionable alert delivery without unnecessary query-content collection;
- backup coverage for deployment configuration, provider policy/configuration, and applicable durable Search state;
- isolated or representative restoration;
- post-restore integrity and service validation;
- upgrade behavior;
- rollback to a known-good native or transitional baseline as appropriate; and
- migration rollback while the incumbent remains available.

A package that starts successfully on a CI runner is not recovery evidence.

## Production acceptance

Before replacing the transitional runtime, the exact native deployment must validate at minimum:

1. exact source revision and immutable artifact provenance;
2. package/runtime integrity on the target architecture;
3. health and meaningful readiness behavior;
4. approved live-provider behavior for every selected category;
5. provider failure, timeout, rate-limit, response-bound, and degradation behavior;
6. current Stable Glaze UI whole-application acceptance across required responsive, appearance, keyboard, contrast, reduced-motion, reduced-transparency, forced-colors, and physical-device/browser scopes;
7. Privacy Shield application/runtime acceptance;
8. Wardveil Security application/runtime acceptance;
9. Everkeep backup, restore, migration, recovery, and rollback acceptance;
10. applicable GoreeCloud Identity, Mesh, Manager, Browser, AI, automation, and other consumer integration acceptance;
11. private DNS, reverse proxy/TLS, NetBird/private access, firewall/open-port, and individual-access boundaries;
12. privacy-conscious logging, monitoring, resource behavior, and actionable alert delivery;
13. exact target-host deployment identity and configuration evidence; and
14. representative post-cutover production acceptance with the previous known-good deployment preserved until rollback requirements are satisfied.

No DNS, Caddy, firewall, NetBird, provider credential, monitoring, backup, or current production runtime state should be changed merely because a Development artifact workflow passes.
