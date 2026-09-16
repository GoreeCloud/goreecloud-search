# GoreeCloud Search — Feature Roadmap

**Product:** GoreeCloud Search  
**Lifecycle:** Development  
**Repository version:** `0.1.0.dev4`

This is the repository-coupled roadmap source. Planned items remain implementation obligations until implemented and verified or explicitly superseded/cancelled.

## Phase 0 — Native foundation

**Status:** In implementation

- [ ] Merge and verify the native query model/parser.
- [ ] Merge and verify privacy-aware source planning.
- [ ] Merge and verify the replaceable provider contract.
- [ ] Merge and verify baseline CI and test coverage.
- [ ] Reconcile required repository documentation and Platform-System declarations.

The current topic branch is a candidate only; checkboxes remain incomplete until accepted on the authoritative integration line with required verification.

## Phase 1 — GoreeCloud Index path

**Status:** In implementation

- [x] Implement the first-party GoreeCloud Index provider adapter boundary.
- [x] Define the versioned Search ↔ Index request/response contract.
- [x] Add initial category/capability negotiation.
- [ ] Add authenticated live Index transport.
- [x] Add timeout, cancellation, pagination execution, and degraded-state behavior.
- [x] Add Index provenance fields required for later ranking and result explanations.
- [ ] Verify Privacy Shield, Identity, Wardveil Security, and runtime authority boundaries for live Index requests.

Completed checkboxes above describe the stacked development candidate only until that candidate is accepted onto the authoritative integration line.

## Phase 2 — Search result pipeline

**Status:** In implementation

- [x] Add deterministic result normalization.
- [x] Add conservative canonical URL handling.
- [x] Add canonical-URL and content-hash deduplication.
- [x] Add source-agreement counts.
- [ ] Add snippet generation.
- [x] Add initial deterministic ranking with inspectable scoring signals.
- [x] Add provider/result provenance primitives and initial “Why this result?” explanations.
- [ ] Add SafeSearch/content-policy hooks.

Normalization deliberately does not rerank results. Ranking is a separate Search-owned phase and the current candidate uses only inspectable query/result evidence plus a bounded source-agreement signal.

## Phase 3 — Federated providers

**Status:** Planned

- Provider SDK/adapter lifecycle.
- Explicit provider capability declarations.
- Approved external provider integrations.
- [x] Bounded execution foundation with per-provider timeouts and concurrency controls.
- Retry/rate/cost controls beyond the current timeout foundation.
- Query-disclosure budgets.
- Provider health and graceful degradation.
- External suggestions only when explicitly enabled.

## Phase 4 — Service API and platform controls

**Status:** Planned

- Versioned HTTP API.
- Health/readiness endpoints.
- GoreeCloud Identity authentication and requester identity.
- Privacy Shield permitted-use enforcement.
- Wardveil Security trust/risk enforcement.
- GoreeCloud Mesh discovery and integration contracts.
- GoreeCloud Manager administrative state.
- Evidence and observability that preserve privacy.

## Phase 5 — Browser and Glaze UI experience

**Status:** Planned

- GoreeCloud Browser address-bar search.
- New-tab search.
- Search result actions.
- Search availability/degraded-state presentation.
- Glaze UI result experience.
- Accessibility and adaptive layouts.
- Search preferences and Lenses.
- Private View / isolated browsing workflows.

## Phase 6 — Advanced discovery

**Status:** Planned

- Discussions.
- Quick Answers.
- Optional GoreeCloud AI answers with source-backed citations.
- Private-corpus search.
- Structured verticals.
- Internationalization and regional controls.
- Research workspaces with explicit escalation from Search → Fetch → Research → Agent/Browser.

## Phase 7 — Production and self-hosting

**Status:** Planned

- OCI container image.
- Rootless Podman profile.
- GoreeCloud Containers profile.
- Bare-metal service definitions.
- Kubernetes/Helm profile when justified.
- Signed artifacts, SBOM, provenance, and reproducible build metadata.
- Performance, reliability, security, privacy, and recovery qualification.
- Stable-release evidence and documentation reconciliation.
