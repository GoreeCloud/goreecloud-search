# GoreeCloud Search — Feature Roadmap

**Product:** GoreeCloud Search  
**Lifecycle:** Development  
**Repository version:** `0.1.0.dev1`

This is the repository-coupled roadmap source. Planned items remain obligations until implemented and verified or explicitly superseded/cancelled.

## Phase 0 — Native foundation

**Status:** In implementation

- [ ] Merge and verify the native query model/parser.
- [ ] Merge and verify privacy-aware source planning.
- [ ] Merge and verify the replaceable provider contract.
- [ ] Merge and verify baseline CI and test coverage.
- [ ] Reconcile required repository documentation and Platform-System declarations.

The current topic branch is a candidate only; these items remain incomplete until accepted on the authoritative integration line with required verification.

## Phase 1 — GoreeCloud Index path

**Status:** Planned

- Implement the first-party GoreeCloud Index adapter and versioned Search ↔ Index contract.
- Add capability negotiation, timeout/cancellation, pagination, provenance, and degraded-state behavior.
- Verify applicable privacy and identity boundaries.

## Phase 2 — Search result pipeline

**Status:** Planned

- Result normalization, canonicalization, deduplication, source agreement, snippets, deterministic ranking, provenance, and SafeSearch hooks.

## Phase 3 — Federated providers

**Status:** Planned

- Provider SDK/lifecycle, approved external adapters, per-provider controls, disclosure budgets, health, graceful degradation, and opt-in external suggestions.

## Phase 4 — Service API and platform controls

**Status:** Planned

- Versioned HTTP API, health/readiness, Identity, Privacy Shield, Wardveil Security, Mesh, Manager, privacy-safe evidence and observability.

## Phase 5 — Browser and Glaze UI experience

**Status:** Planned

- Browser address-bar/new-tab search, result actions, degraded-state UX, current Stable Glaze UI, accessibility, preferences, Lenses, and Private View.

## Phase 6 — Advanced discovery

**Status:** Planned

- Discussions, Quick Answers, optional GoreeCloud AI answers, private-corpus search, structured verticals, internationalization, and explicit Search → Fetch → Research → Agent/Browser escalation.

## Phase 7 — Production and self-hosting

**Status:** Planned

- OCI/Podman/GoreeCloud Containers/bare-metal deployment, Kubernetes when justified, signed artifacts, SBOM/provenance, and full production/Stable qualification evidence.
