# GoreeCloud Search — Feature Roadmap

**Product:** GoreeCloud Search  
**Lifecycle:** Development  
**Repository version:** `0.1.0.dev14`

This is the repository-coupled roadmap source. Planned items remain implementation obligations until implemented and verified or explicitly superseded/cancelled.

## Current verified Development checkpoint

PR #23 is integrated on authoritative `main` at `7d79959e79ad4de13b807973347944bd92563fbe`. Post-merge CI run `35625485445` passed on that exact revision. The source now contains the bounded authenticated Index delegation HTTP boundary, but live GoreeCloud Identity and Privacy Shield verifier transports, approved external provider execution, deployment, representative runtime/client acceptance, Production acceptance, and Stable qualification remain open.

## Phase 0 — Native foundation

**Status:** In implementation

- [ ] Merge and verify the native query model/parser.
- [ ] Merge and verify privacy-aware source planning.
- [ ] Merge and verify the replaceable provider contract.
- [x] Merge and verify baseline CI and test coverage.
- [x] Reconcile required repository Platform Contract declaration to Contract 0.4 with all nine Integral Platform Systems.
  - PR #22 is accepted on authoritative `main`; runtime Platform-System implementation and acceptance remain separate Development/nonconformant obligations.

Completed checkboxes are bound to the verified Development integration line and cited exact-revision evidence. They do not establish live provider authority, deployment, Production acceptance, or Stable qualification.

## Phase 1 — GoreeCloud Index path

**Status:** In implementation

- [x] Implement the first-party GoreeCloud Index provider adapter boundary.
- [x] Define the versioned Search ↔ Index request/response contract.
- [x] Add initial category/capability negotiation.
- [ ] Add authenticated live Index transport and runtime acceptance.
  - [x] Authoritative `main` includes the bounded server-side Index HTTP carrier with injected Identity and Privacy Shield verification; real authority transports, approved provider execution, deployment, and target-runtime acceptance remain open.
- [x] Add timeout, cancellation, pagination execution, and degraded-state behavior.
- [x] Add Index provenance fields required for later ranking and result explanations.
- [x] Add a cycle-safe Index-originated delegation source path that is external-only and cannot re-enter Index or create a fallback stage.
- [ ] Verify Privacy Shield, Identity, Wardveil Security, and runtime authority boundaries for live Index requests.

Completed checkboxes above describe behavior now accepted on the authoritative Development integration line where identified by the current verified checkpoint; they do not establish live authority transport, deployment, Production acceptance, or Stable qualification.

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

**Status:** In implementation

- Provider SDK/adapter lifecycle.
- Explicit provider capability declarations.
- Approved external provider integrations.
- [x] Bounded execution foundation with per-provider timeouts and concurrency controls.
- Retry/rate/cost controls beyond the current timeout foundation.
- [x] Query-disclosure budgets with pre-execution third-party provider caps and plan evidence.
- Provider health and graceful degradation.
- External suggestions only when explicitly enabled.

## Phase 4 — Service API and platform controls

**Status:** Planned

- [x] Add a bounded versioned Index-delegation HTTP source boundary.
- [x] Add source-level health/readiness endpoints; accepted Observability runtime integration remains open.
- [ ] Connect GoreeCloud Identity authentication/requester verification to its authoritative runtime service.
- [ ] Connect Privacy Shield capability-reference verification/consumption to its authoritative runtime service.
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
