# GoreeCloud Search — Feature Roadmap

**Product:** GoreeCloud Search  
**Lifecycle:** Development  
**Repository version:** `0.1.0.dev8`

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
- [ ] Add snippet generation.
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
- [ ] Approved external provider integrations.
- [x] Bounded execution foundation with per-provider timeouts and concurrency controls.
- [ ] Retry/rate/cost controls beyond the current timeout foundation.
- [x] Query-disclosure budgets with pre-execution third-party provider caps and plan evidence.
- [ ] Provider health and graceful degradation beyond per-request execution state.
- [ ] External suggestions only when explicitly enabled.

## Phase 4 — Service API and platform controls

**Status:** Planned

- Versioned HTTP API and health/readiness endpoints.
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

**Status:** Planned

- OCI/rootless Podman/GoreeCloud Containers/bare-metal deployment profiles.
- Kubernetes/Helm only when justified.
- Signed artifacts, SBOM, provenance, reproducible build metadata.
- Performance, reliability, security, privacy, recovery, and Stable qualification evidence.
