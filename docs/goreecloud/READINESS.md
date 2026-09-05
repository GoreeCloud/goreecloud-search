# GoreeCloud Search Readiness

This document defines the current readiness boundary for the GoreeCloud-owned native Search application. It separates source implementation, exact-revision validation, build/package validation, packaged-runtime CI acceptance, integration readiness, target-environment acceptance, migration cutover, Release Candidate review, production deployment, and Stable promotion so no lower state is represented as a higher one.

## Current product boundary

GoreeCloud Search is undergoing a native migration.

- Target application: GoreeCloud-owned Go service under `native/`.
- Transitional dependency: the inherited SearXNG-derived tree remains only for continuity, migration, compatibility, feature-preservation decisions, applicable upstream security maintenance, and rollback until controlled retirement.
- Current release lifecycle: **Development**.
- Repository Platform Contract state: **nonconformant**.
- Current declared Stable design-system requirement: **GLAZE UI V1.1 / 1.1.0**, subject to the authoritative current-Stable Glaze governance and acceptance evidence.
- Production approval: **false**.

A successful source build, green pull request, Development artifact, working development page, or healthy transitional production service does not establish Release Candidate, production-ready, production-accepted, or Stable qualification for the native application.

## Readiness state model

Search follows the GoreeCloud production-readiness distinction between at least:

1. source implemented;
2. source validated;
3. build/package validated;
4. integration validated;
5. target-environment validated;
6. security validated;
7. backup/restore validated;
8. upgrade/rollback validated;
9. monitoring/alerting validated;
10. visual/accessibility accepted;
11. production deployed; and
12. production accepted.

The current native artifact work can establish the first three states for a bounded Development scope when its exact-revision checks pass. It does not skip the later states.

## Native HTTP readiness surfaces

### Process health — `GET /healthz`

The native `searchd` process exposes a bounded JSON health response identifying the Search service, native development implementation, and `production_approved: false`.

A successful `/healthz` response proves only that the process can serve that route. It does not prove provider availability, production configuration, private routing, monitoring, recovery, platform integrations, migration parity, or production acceptance.

### Service status — `GET /api/v1/status`

The native status endpoint exposes the GoreeCloud-owned API version, product/service identity, native implementation identity, canonical lifecycle `development`, source-level capability availability, canonical endpoint paths, and `production_approved: false`.

The lifecycle value is intentionally one of GoreeCloud's approved release-lifecycle tags. Informal `pre-stable` wording is not used as the machine-readable lifecycle identity.

`machine_readable_search_api: true` means the native `/api/v1/search` endpoint exists in source. It does not authorize arbitrary production consumers.

### Local application readiness — `GET /api/v1/readiness`

The native readiness endpoint evaluates a deliberately bounded `local_native_application` scope.

Current checks are:

- native engine initialized; and
- General category locally executable under the native engine contract.

When those checks pass, the endpoint returns HTTP 200 with `ready: true`. It explicitly reports `production_approved: false` and identifies important production conditions that are not evaluated, including external providers, production provider credentials, private DNS/reverse proxy, monitoring/alerting, backup/restore/rollback, physical-device acceptance, and production cutover.

Neither `/healthz` nor `/api/v1/readiness`, individually or together, authorizes production deployment.

### Provider capability readiness — `GET /api/v1/providers/definitions`

The native provider-definition endpoint exposes sanitized configured-provider capabilities without credentials, endpoint secrets, raw provider errors, authorization headers, or mutable management controls.

A provider interface or advertised category is not production approval. A real provider is acceptable only after its transport, privacy/data-use/terms boundary, credentials, limits, degradation behavior, timestamp authority where applicable, target-runtime behavior, and representative live queries are separately accepted.

## Native search execution readiness

Native Search currently includes bounded query validation, explicit General/Images/Videos/News/Files category handling, concurrent provider execution, timeout/degradation handling, per-provider processing bounds, result URL sanitization, deterministic local ranking, bounded freshness, conservative local correction suggestions, source provenance, image media/proxy handling, and GoreeCloud-owned HTML/JSON response surfaces.

General retains an empty-provider Development behavior. Specialized categories fail closed when the current engine has no executable provider for the requested category.

Deterministic provider fixtures exercise all five native category contracts in source CI. That proves source/category contract behavior, not real provider or production behavior.

## Native provider runtime boundary

`native/internal/providers` provides the deployment-controlled `goreecloud-http-v1` provider runtime.

The runtime:

- configures zero providers by default;
- loads provider policy only when `GOREECLOUD_SEARCH_PROVIDER_CONFIG_FILE` is explicitly set;
- rejects malformed or unsupported configuration;
- references optional bearer credentials by environment-variable name;
- does not place credentials in sanitized provider definitions;
- applies bounded HTTP transport, response, and result controls;
- provides an explicit `GOREECLOUD_SEARCH_REQUIRE_RELEASE_PROVIDER_COVERAGE` structural preflight; and
- does not select or approve a real provider by itself.

When the structural preflight is enabled with `1` or `true`, provider initialization fails closed unless the configured set has a valid executable provider path for General, Images, Videos, News, and Files. General's provider-free Development fallback is deliberately excluded from provider-backed coverage, and a multi-category declaration does not satisfy specialized coverage unless the provider exposes the required category-aware execution path. Unset, `0`, or `false` preserves the ordinary Development behavior. Any other flag value is invalid and fails initialization.

This structural gate is not live-provider acceptance. It does not prove that an endpoint is reachable, credentials work, results are useful, provider terms/privacy behavior are acceptable, timestamps are authoritative, rate limits are acceptable, or the target host has accepted network behavior.

Production-approved provider selection and credentials integration remain incomplete.

## Native Development artifact readiness

`.github/workflows/goreecloud-native-development-artifact.yml` establishes a build/package validation boundary for the exact native source revision.

The workflow:

- explicitly checks out the pull-request head SHA or exact push SHA;
- verifies exact clean Git state;
- runs native Go tests;
- builds Linux amd64 and arm64 `searchd` binaries with `CGO_ENABLED=0`, `-trimpath`, VCS metadata, and the approved Go 1.25.x CI toolchain;
- builds each target twice and requires byte-identical output;
- verifies the embedded Git revision and `vcs.modified=false` state;
- packages the binary with AGPL license and machine-readable metadata;
- creates deterministic tar/gzip packages;
- emits `SHA256SUMS` and `artifact-provenance.json`;
- extracts the packaged Linux/amd64 runtime and starts that exact binary on loopback;
- validates health, canonical Development status, local readiness, zero-provider sanitized definitions, native homepage/Preferences, and fail-closed Images behavior without a configured provider; and
- validates that the packaged binary accepts structurally complete synthetic all-category provider configuration under the release-coverage preflight and refuses incomplete coverage before listening, without sending a Search request to the synthetic provider endpoint.

The provenance manifest deliberately records:

- lifecycle `development`;
- artifact scope `ci-development-candidate`;
- `production_approved: false`;
- `release_candidate_declared: false`;
- `target_environment_validated: false`;
- `live_provider_acceptance_validated: false`; and
- platform conformance `nonconformant`.

Successful completion may support the statements **source validated**, **build/package validated**, and **packaged runtime accepted on the CI Linux boundary** for that exact revision. It does not establish target-host acceptance, Release Candidate status, live-provider acceptance, or production deployment.

The artifact path is documented in `native/docs/ARTIFACT-PROVENANCE.md`.

## Transitional artifact distinction

Inherited SearXNG container build and provider-acceptance paths remain transitional continuity and migration material. Historical SearXNG Release Candidate publication and RC08/RC09 receipt definitions and their accepted evidence remain preserved in Git history and historical records, but those executable workflow files are intentionally retired from the current native development line.

The historical Release Candidate artifacts must not be represented as native Search build/package or release evidence, and the native line must not reactivate the inherited SearXNG RC publication or receipt paths. The workflow supply-chain guard enforces that retirement and no longer carries their package-write or issue-write exceptions.

The previous known-good transitional deployment remains a rollback dependency until native migration and recovery requirements are satisfied.

## Native presentation and image readiness

The GoreeCloud-owned native browser experience includes homepage, Preferences, general results, image results, full-image/source viewing, keyboard navigation, focus restoration, responsive layouts, and bounded accessibility/resilience fallbacks.

Deterministic browser acceptance validates representative Compact, Medium, Expanded, and Wide layouts, Light/Dark appearances, keyboard focus, interaction-target sizing, overflow safety, Reduced Motion, Increased Contrast, Forced Colors, and reduced-transparency behavior for the tested surfaces.

Those tests are rendered CI evidence. They do not independently prove whole-application current-Stable Glaze conformance, physical-device behavior, or target-production behavior.

## Current Stable Glaze UI readiness

The repository Platform Contract currently declares `glaze-ui==1.1.0` / GLAZE UI V1.1 as the Stable consumer requirement.

Search remains `applicable-migration-required` for Glaze UI because whole-application current-Stable acceptance is incomplete. Final consumer acceptance must use the current authoritative immutable Stable Glaze contract and evidence at the time of release review; a superseded, reset-baseline, Candidate, or RC design-system line cannot satisfy that gate.

Required Search-specific acceptance includes applicable semantic hierarchy, typography, spacing, geometry, focus/state behavior, target-size rules, responsive classes, appearance modes, Reduced Motion, Reduced Transparency, Increased Contrast, Forced Colors, effects-free fallbacks, representative user/error/empty/degraded/security/privacy/recovery surfaces, keyboard/pointer behavior, physical-device/browser coverage where needed, and exact-revision rendered/human evidence.

## Privacy Shield readiness

Search must prove applicable Privacy Shield behavior at the real application/runtime boundary before production acceptance, including privacy-first defaults, no GoreeCloud sponsored ranking, no behavioral-profile ranking dependency, minimized query/history retention, privacy-safe diagnostics/logging, no hidden provider bypass, no unnecessary credential/private-content exposure, and explicit controls before persistent convenience features are enabled.

External providers may observe requests from GoreeCloud infrastructure. Search must not claim anonymity from them.

Current Search platform status remains fail-conservative because accepted producer-authoritative Privacy Shield application/runtime evidence is not configured in the shipped native handler.

## Wardveil Security readiness

Native source now contains bounded Wardveil projection logic and GoreeCloud Mesh evidence-envelope provenance validation. A positive source projection requires producer-authoritative, scoped, current, fresh, internally consistent Wardveil evidence.

The default native handler still has no deployed authoritative Wardveil evidence source. Therefore source integration does not prove the actual Search deployment is protected or production-accepted.

Production security acceptance still requires provider-specific abuse controls, dependency/security maintenance, authorization where applicable, deployment hardening, secret handling, private exposure validation, target runtime evidence, and other Wardveil requirements.

## Everkeep and recovery readiness

Native Search contains a fail-closed continuity projection boundary, but backup/recovery readiness is not established by source logic.

Before production acceptance, evidence must identify and validate applicable:

- exact source and deployable artifact;
- native deployment configuration;
- provider policy/configuration needed to restore service;
- protected secret-recovery path without copying reusable secrets into ordinary evidence;
- private publication configuration;
- durable user-controlled state in the selected release;
- backup coverage;
- isolated or representative restore procedure;
- post-restore integrity/service validation;
- previous known-good rollback target; and
- migration rollback while the transitional runtime remains available.

A CI artifact retention period is not an Everkeep backup.

## GoreeCloud Mesh readiness

Search contains local capability semantics and a fail-closed Mesh evidence-envelope validation primitive. That is source integration only.

Accepted Mesh registration, discovery, producer delivery transport, dependency/event integration, consumer authorization, and runtime acceptance remain incomplete. Mesh may coordinate and validate transport/provenance; it does not create producer-domain privacy, security, recovery, identity, or release authority.

## GoreeCloud Identity readiness

GoreeCloud Identity remains authoritative for account, authentication, authorization, session, service, device, and delegated authority where those responsibilities apply.

Source-level Search preferences/history/sync contracts do not establish accepted account-bound production functionality. If the selected release uses account-bound capabilities, the applicable Identity integration must be implemented and accepted before production approval.

## GoreeCloud Manager readiness

Manager operational visibility is applicable, but Search does not yet have accepted Manager runtime integration. A Manager display or platform declaration cannot substitute for producer-authoritative health, privacy, security, recovery, Identity, or Mesh evidence.

## Provider and category acceptance

Before a native release can be accepted for production, each category selected for that release must have production-approved executable provider coverage and sanitized live evidence for applicable:

- useful result completion;
- result integrity and URL safety;
- category correctness;
- latency and timeout behavior;
- rate limiting, access denial, CAPTCHA, and provider outages where encountered;
- provider response/body/result bounds;
- degradation when one provider fails;
- timestamp-authority correctness where freshness is used;
- absence of raw credentials or private query/result content in release evidence; and
- exact artifact/runtime identity used during the test.

Passing the structural release-provider coverage preflight is a prerequisite-style configuration check only. It does not satisfy any of the live evidence above.

No provider becomes trusted merely because it implements the native interface.

## Monitoring and operational readiness

A production candidate requires privacy-conscious monitoring appropriate to Search, including applicable process/service availability, `/healthz` and meaningful readiness through the intended private route, HTTPS/reverse-proxy success, private DNS, certificate validity, representative search completion, provider failure/timeout trends without unnecessary query retention, latency/resource pressure, supporting-runtime health, and verified actionable alert delivery.

The Development artifact workflow intentionally does not claim these target-environment conditions.

## Private-access and network readiness

The native service defaults to `127.0.0.1:8080`. Direct public application-port exposure is not an approved shortcut.

A target deployment must separately validate only the interfaces, routes, and ports required by the approved architecture, including the intended private reverse-proxy/TLS, private DNS, NetBird/private access or equivalent, firewall/open-port state, and individual-access/attribution boundary.

No source or CI workflow may silently change production DNS, Caddy, firewall, NetBird, or current deployed runtime merely because a package is buildable.

## Migration and cutover readiness

The inherited SearXNG-derived runtime remains a controlled transitional dependency until the native release has sufficient accepted capability and rollback evidence.

Before native production cutover:

1. identify one exact candidate source revision and immutable artifact;
2. complete the required exact-revision source and package gates;
3. establish Release Candidate eligibility for the selected scope rather than relabeling a Development build;
4. complete required live native category/provider acceptance;
5. complete current-Stable Glaze UI whole-application and required device/browser acceptance;
6. complete Privacy Shield and Wardveil runtime/evidence acceptance;
7. complete Everkeep backup/restore/migration/rollback acceptance;
8. validate required Identity, Mesh, Manager, Browser, AI, Sync, and other consumer integrations;
9. validate target-host private networking, DNS, reverse proxy/TLS, monitoring, alerts, resources, and abuse controls;
10. preserve the previous known-good transitional deployment until rollback is proven;
11. perform controlled cutover only after the applicable production authorization;
12. verify the deployed artifact/configuration exactly matches the approved candidate; and
13. re-run representative acceptance after cutover.

If source or relevant configuration changes after final validation, it becomes a new candidate and the applicable validation must be repeated.

## Release Candidate decision

Search remains **Development** after Development artifact validation.

A native Search build may move to **Release Candidate** only when an identifiable exact candidate and artifact are established and the evidence required for final acceptance has reached the level required by GoreeCloud lifecycle/production-readiness governance for the selected release scope. Merely slowing feature development, creating a tarball, publishing an OCI image, or passing CI is insufficient.

A controlled RC replacement cutover can be considered later only with the required migration plan, preserved data/configuration, validated recovery, credible rollback, critical integration acceptance, and absence of release-blocking security/privacy/data-integrity/recovery/compatibility defects.

## Stable and production decision

Search may be described as production-ready or Stable only when all applicable GoreeCloud production-readiness and platform-conformance gates have passed for the exact release, artifact, supported platform, and deployment boundary, and no unresolved release blocker remains.

Stable is not a synonym for working, packaged, green CI, deployed, or long-running.
