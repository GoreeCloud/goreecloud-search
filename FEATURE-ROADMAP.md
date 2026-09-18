# GoreeCloud Search — Feature Roadmap

**Product:** GoreeCloud Search  
**Lifecycle:** Development  
**Repository version:** `0.1.0.dev11`

This is the repository-coupled roadmap source. Planned items remain implementation obligations until implemented and verified or explicitly superseded/cancelled.

## Phase 0 — Native foundation

**Status:** In implementation

- [ ] Merge and verify the native query model/parser.
- [ ] Merge and verify privacy-aware source planning.
- [ ] Merge and verify the replaceable provider contract.
- [ ] Merge and verify baseline CI and test coverage.
- [ ] Reconcile required repository documentation and Platform-System declarations.

## Phase 1 — GoreeCloud Index path

**Status:** In implementation

- [x] Implement the first-party GoreeCloud Index provider adapter boundary.
- [x] Define the versioned Search ↔ Index request/response contract.
- [x] Add initial category/capability negotiation.
- [ ] Add authenticated live Index transport.
- [x] Add timeout, cancellation, pagination execution, and degraded-state behavior.
- [x] Add Index provenance fields required for later ranking and result explanations.
- [ ] Verify Privacy Shield, Identity, Wardveil Security, and runtime authority boundaries for live Index requests.

Completed checkboxes describe the stacked development candidate only until accepted onto the authoritative integration line.

## Phase 2 — Search result pipeline

**Status:** In implementation

- [x] Add deterministic result normalization.
- [x] Add conservative canonical URL handling.
- [x] Add canonical-URL and content-hash deduplication.
- [x] Add source-agreement counts.
- [x] Add bounded query-aware snippet generation.
- [x] Add initial deterministic ranking with inspectable scoring signals.
- [x] Add provider/result provenance primitives and initial “Why this result?” explanations.
- [x] Add SafeSearch/content-policy hooks and administrator domain policy controls.
- [x] Add initial transparent Search-local Lens reranking/exclusion primitives.
- [x] Add a strict versioned portable Lens import/export data format.
- [ ] Add verified content/safety classification sources for production SafeSearch enforcement.

## Phase 3 — Federated providers

**Status:** In implementation

- [ ] Provider SDK/adapter lifecycle.
- [ ] Explicit provider capability declarations.
- [x] Add the first Development external provider adapter (Brave Web Search API).
- [ ] Complete production acceptance for the approved external provider set.
- [x] Bounded execution foundation with per-provider timeouts and concurrency controls.
- [ ] Retry/rate/cost controls beyond the current timeout foundation.
- [x] Query-disclosure budgets with pre-execution third-party provider caps and plan evidence.
- [ ] Provider health and graceful degradation beyond per-request execution state.
- [ ] External suggestions only when explicitly enabled.

## Phase 4 — Service API and platform controls

**Status:** Planned

- [x] Add a loopback-only Development `/api/v1/search` and `/healthz` HTTP boundary.
- [ ] Add production service authentication/authorization, readiness semantics, rate/abuse controls, and target-runtime acceptance.
- GoreeCloud Identity authentication and requester identity.
- Privacy Shield permitted-use enforcement.
- Wardveil Security trust/risk/safety integration.
- GoreeCloud Mesh discovery and integration contracts.
- GoreeCloud Manager administrative state.
- Privacy-safe evidence and observability.

## Phase 5 — Browser and Glaze UI experience

**Status:** Planned

- GoreeCloud Browser address-bar and new-tab search.
- Result actions and truthful availability/degraded-state presentation.
- Glaze UI accessibility/adaptive layouts.
- Lens creation/editing UI, file I/O, trusted sharing/discovery, optional synchronization, Search preferences, and Private View.

## Phase 6 — Advanced discovery

**Status:** Planned

- Discussions, Quick Answers, optional source-backed GoreeCloud AI answers.
- Private-corpus search, structured verticals, internationalization/regional controls.
- Research workspaces with explicit Search → Fetch → Research → Agent/Browser escalation.

## Phase 7 — Production and self-hosting

**Status:** In implementation

- [x] Add a GoreeCloud-owned non-root Docker image candidate for the native service.
- [x] Add a VPS Compose service candidate that preserves the verified `searxng-core:8080` Caddy backend contract, uses no host port, and requires an exact tag@digest image reference.
- [x] Add protected runtime credential-file input and container health validation.
- [x] Add independent CI that builds and exercises the hardened container without making a live provider request.
- [x] Add a stdout-only, non-mutating VPS preflight collector that reads the authoritative Search Compose identity, active container/image/health/security/network state, protected credential metadata, Caddy route references, private DNS, and HTTPS homepage state without printing secrets or raw configuration.
- [x] Run the initial read-only preflight on `goreecloud-vps-01` and retain/review the resulting Docker/Compose/image/security/network/credential/DNS/HTTPS evidence.
- [x] Re-run the corrected preflight against the current production Caddy root at `/srv/docker/caddy`; the active `Caddyfile` contains both `search.goreecloud.com` and `searxng-core:8080` references, while historical backup files also contain the same strings.
- [x] Verify the intended private Caddy route directly at the approved NetBird address: the corrected collector returned HTTP 200 with TLS verification success for `search.goreecloud.com` forced to `100.71.27.119`.
- [ ] Validate the owner laptop separately through its actual NetBird + AdGuard private-DNS path. The VPS system resolver still returns public addresses and is not a substitute for this client-path check.
- [ ] Build/publish an immutable GHCR image with exact version, revision, architecture, and digest evidence.
- [ ] Perform real-provider and target-runtime validation on `goreecloud-vps-01`.
- [ ] Complete controlled rollback preparation and private Caddy/DNS acceptance before cutover.
- [ ] Retire legacy SearXNG-derived Valkey/config/naming only when separately proven unnecessary.
- [ ] Add other OCI/rootless Podman/GoreeCloud Containers/bare-metal profiles only when justified.
- [ ] Kubernetes/Helm only when justified.
- [ ] Signed artifacts, SBOM, provenance, reproducible build metadata.
- [ ] Performance, reliability, security, privacy, recovery, and Stable qualification evidence.
