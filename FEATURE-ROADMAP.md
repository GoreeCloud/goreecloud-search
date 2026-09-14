# GoreeCloud Search — Feature Roadmap

**Status:** Active roadmap control  
**As of:** 2026-09-14  
**Authoritative project record:** Project Specification — Search  
**Canonical repository:** GoreeCloud/goreecloud-search  
**Drive control:** `GoreeCloud/Feature Roadmap/GoreeCloud Search/FEATURE-ROADMAP.docx`

## Purpose

This file is the repository-side feature roadmap control for GoreeCloud Search. It records current planned and recommended feature work without replacing the authoritative project record, implementation evidence, release gates, or GoreeCloud Tasks Management.

## Current Development checkpoint

Accepted `master` remains the previously merged Development line. Active PR #171 is a newer Development candidate and is not merged, production accepted, or Stable merely because validation passes.

The latest verified implementation-bearing PR #171 checkpoint is `cee201f0faa4d7d94d2ce3076a3e62f044c056ab`. All 14 current workflow families passed on that exact revision: Platform Contract #64, Search Native Foundation #312, foundation #1806, runtime smoke #1642, API v1 service contract #167, platform integrations #1142, upstream container boundary #1018, container build #1638, browser acceptance #1731, native development artifact #214, native Development container #185, native results browser acceptance #251, Integration #694, and Documentation #692.

That Development candidate now:

- preserves GoreeCloud Search as the Internet/web/current-information authority while Index remains universal/local federation authority and Browser retains navigation/tab/page authority;
- supports bounded first-party POST + JSON search transport while retaining GET only as Development/compatibility support;
- publishes query-free `search.query` capability evidence with bounded result/request limits and authenticated-requester expectations;
- explicitly publishes requester-authentication authority `goreecloud-identity`, Bearer scheme, and `Authorization` carrier header while leaving audience, scopes, token format, issuer/JWKS, and concrete application registrations Identity-owned and unspecified;
- advertises `production_accepted=false` and `privacy_authorization_enforcement=not_enforced_development` until real authenticated runtime enforcement exists;
- adds a fail-closed Search-side Privacy Shield authorization gate that requires a real verifier and authenticated requester resolver before required mode can be claimed ready;
- validates Privacy Shield authority responses independently for verification contract version, positive authorization, exact opaque capability reference, processing zone, destination, and retention constraints;
- includes a bounded Privacy Shield v1 HTTP verification client with loopback-only cleartext allowance, HTTPS for non-loopback, no redirects, strict JSON, bounded response size, no-store, and timeout controls;
- consumes the producer-owned GoreeCloud Identity direct-service profile for `goreecloud-search` → `goreecloud-privacy-shield` through an injected credential source, requires exact service/audience/scope metadata, rejects pre-existing Authorization headers, and binds credential use to the exact configured Privacy Shield verification endpoint before Identity is contacted;
- adds a transport-neutral inbound requester-verification seam that requires exactly one canonical Bearer credential and delegates the opaque credential to GoreeCloud Identity;
- requires Identity verification to return both a registered first-party `ApplicationID` and a user-bound `PrincipalID`, validates both as bounded canonical identifiers, and returns only the verified application ID to the Privacy Shield gate as `requester_id`;
- keeps the verified principal independently required as evidence that the credential represents a user-bound native application session, while never treating the Search service identity, a caller-selected requester header, or an unverified principal as the Privacy Shield requester;
- deliberately does not invent Browser/Index audience, scope, JWKS, token-format, issuer, or application-registration values because those remain Identity-owned contracts;
- keeps Privacy Shield signing keys and raw capability-token interpretation outside Search;
- targets current Stable Glaze UI V1.4 / `1.4.0` while retaining Search-local rendered/accessibility/device acceptance gates;
- declares only the seven authoritative Integral Platform Systems in Platform Contract 0.3 and keeps GoreeCloud Sync as a separate application/service capability;
- preserves cancellation, provider degradation, result safety, privacy-safe diagnostics, Development packaging, and native/transitional migration boundaries.

Identity PR #9 and Privacy Shield PR #105 provide matching Development source contracts for the Search → Privacy Shield direct-service path. Their source-level contracts do not establish deployed Identity issuance/JWKS/key custody, deployed authenticated Privacy Shield hosting, concrete Browser/Index requester registrations/runtime issuance, or production acceptance. Search must not switch to required Privacy Shield enforcement until those runtime dependencies and the real authority path are accepted and tested. Query/history synchronization also remains disabled unless explicit user controls and privacy/retention contracts authorize it.

The existing Glaze V1.4 canonical-source reconciliation work remains Development evidence. Canonical closure/activation, runtime reconciliation, rollback, manual accessibility/optical/device qualification, provider/platform production acceptance, Release Candidate, production, and Stable qualification remain independent open gates.

## Roadmap

| ID | Feature / obligation | Priority | Current state |
| --- | --- | --- | --- |
| FR-001 | Reconcile and maintain every current planned or recommended GoreeCloud Search feature from the authoritative project record and verified repository evidence in this roadmap. | High | Ongoing control; synchronized to verified PR #171 implementation checkpoint `cee201f0faa4d7d94d2ce3076a3e62f044c056ab`. |
| FR-002 | Move actionable feature obligations into GoreeCloud Tasks Management when required, preserving priority, dependency, and lifecycle disposition. | High | Ongoing control; GOR-25 tracks the active Search modernization work and exact checkpoints. |
| FR-003 | Do not mark features implemented, complete, cancelled, or superseded without authoritative evidence and synchronized repository/Drive roadmap updates. | High | Ongoing control; Development source, runtime acceptance, production approval, and Stable status remain distinct. |
| FR-004 | Implement and validate the native `tinyfish-search-v1` provider adapter for TinyFish Search, initially limited to General/web and News with deployment-controlled API-key authentication and fail-closed category claims. | High | Development merged; production provider acceptance pending. |
| FR-005 | Add a bounded TinyFish Fetch retrieval stage behind a GoreeCloud-owned retrieval contract for approved URL extraction after discovery. | High | Development merged; production acceptance pending. |
| FR-006 | Route TinyFish Research behind GoreeCloud Search for source-backed research without creating a direct permanent GoreeCloud AI dependency on TinyFish. | High | Development merged; production acceptance pending. |
| FR-007 | Add policy-controlled TinyFish Agent and Browser escalation with separate Browser Context Profiles, approved Vault usage, and explicit Privacy Shield/Wardveil boundaries. | Medium | Development merged; production acceptance pending. |
| FR-008 | Add least-capability routing, metered-cost controls, health/observability, replacement/fallback behavior, and Manager operational visibility for the TinyFish integration. | High | Development merged; production acceptance pending. |
| FR-009 | Add exact-host authenticated-session bindings so recurring TinyFish automation deterministically selects one saved Browser Context Profile and only explicitly scoped Vault credential references for the intended service/account boundary. | High | Development merged; production acceptance pending. |
| FR-010 | Add a fail-closed Browser Context Profile lifecycle and recovery primitive for authorized create/setup/save/cancel operations so recurring TinyFish sessions can be deliberately established or repaired without exposing reusable secrets. | High | Development merged; production acceptance pending. |
| FR-011 | Add privacy-minimized authenticated-session acceptance evidence so exact-host TinyFish bindings distinguish configured state from observed setup, session reuse, Vault repair, current failures, and secret-handling defects. | High | Development merged; production acceptance pending. |
| FR-012 | Bind recurring authenticated TinyFish execution to the managed exact-host profile/Vault registry and automatically feed normalized execution outcomes into privacy-minimized session acceptance evidence without inferring Vault repair from credential availability alone. | High | Development merged; production acceptance pending. |
| FR-013 | Add bounded deployment loading and a read-only privacy-minimized native status surface for authenticated TinyFish automation so operators can distinguish unconfigured, unverified, reusable, blocked, failed, expired, revoked, and disclosure-risk session states without exposing profile or Vault item identifiers. | High | Development merged; production acceptance pending. |
| FR-014 | Require fail-closed GoreeCloud authorization after exact-host profile/Vault binding and request normalization but before any authenticated TinyFish provider execution, so policy evaluates the actual managed session and credential scope crossing the provider boundary. | High | Development merged; production acceptance pending. |
| FR-015 | Add a fail-closed provider-neutral direct Browser session lifecycle and exact TinyFish Browser API transport with bounded inactivity timeouts, sensitive CDP/session handling, conservative uncertain-outcome semantics, explicit termination control, and no guessed provider fields. | Medium | Development merged; production acceptance pending. |
| FR-016 | Add evidence-driven provider health acceptance gates for metered Agent and future Browser escalation, including explicit success/latency thresholds, cancellation/rejection treatment, fail-closed unavailable routing, and a path to bounded recent-window and durable privacy-minimized health evidence. | High | Development merged; production acceptance pending. |
| FR-017 | Add a scoped direct Browser session lease that validates cleanup authority before creation, always attempts termination after confirmed creation, uses an independent bounded cleanup context after caller cancellation, preserves work and cleanup failures together, and never treats uncertain cleanup as success. | Medium | Development merged in PR #144 after exact-head CI passed; production acceptance pending. |
| FR-018 | Define and validate bounded first-party GoreeCloud Index and Browser consumer contracts over the native Search API, including result/request limits, capability evidence, fail-closed preflight, degraded-state preservation, safe executable destinations, and data minimization. | High | Merged foundation plus active PR #171 hardening. Search publishes the intended production POST/JSON, Privacy Shield reference, authenticated-requester requirement, and requester-authentication carrier metadata (Identity authority, Bearer scheme, Authorization header); concrete Identity requester issuance and live consumer transport remain blocked. |
| FR-019 | Complete GoreeCloud Search adoption and product acceptance against current Glaze UI V1.4 / `1.4.0`, including canonical Stable CSS/runtime reconciliation, accessible optical behavior, rollback, representative-browser/device performance, and product release evidence. | High | V1.4 source/provenance foundation exists and PR #171 retains V1.4 as the current consumer target. Canonical closure/activation, runtime reconciliation, manual accessibility/optical/device qualification, production acceptance, and Stable status remain open. |
| FR-020 | Complete durable operation-scoped Privacy Shield authorization and authenticated first-party requester/service identity for Browser/Index → Search and Search → Privacy Shield without interpreting authority tokens inside Search. | High | PR #171 has exact-head-green fail-closed gate/client contracts, endpoint-bound Identity direct-service transport, independent Privacy Shield authority-response validation, explicit requester-authentication carrier metadata, and application-bound Identity requester verification. Search validates both the Identity-verified application and user principal but passes only the verified application ID as Privacy Shield `requester_id`. Deployed Identity issuance, concrete Browser/Index registrations, authenticated Privacy Shield hosting, runtime wiring, denial/consume/replay/revocation acceptance, and `required` enforcement remain blocked. |
| FR-021 | Preserve Search-owned synchronization boundaries for explicitly approved durable Search datasets without synchronizing query/history content by default or misclassifying GoreeCloud Sync as an Integral Platform System. | Medium | Source capability/dataset boundaries exist; accepted Sync runtime registration, authenticated synchronization, reconciliation, conflict/deletion behavior, user controls, and production evidence remain open. |

## Current sequencing recommendation

1. Obtain accepted GoreeCloud Identity runtime issuance for the existing Search → Privacy Shield direct-service contract, plus concrete producer-owned Browser and Index requester registrations/issuance for Search.
2. Connect Search's validated Identity direct-service transport and application-bound requester-verifier seam to the accepted Identity runtime and Privacy Shield authenticated verification host; keep required mode disabled until exact runtime evidence exists.
3. Add handler-level denial, consume/single-use, replay, revocation/expiry, application/requester mismatch, principal/session invalidity, and constraint-mismatch acceptance against the real authenticated verification path before advertising `privacy_authorization_enforcement=required`.
4. Continue Glaze UI V1.4 canonical closure/runtime reconciliation and Search-local rendered/accessibility/device qualification without conflating source provenance with product acceptance.
5. Complete approved provider runtime acceptance, monitoring/recovery, any separately authorized Sync datasets, signing/provenance, Release Candidate, production, and Stable gates.

## Maintenance and synchronization

This roadmap and the corresponding Drive `FEATURE-ROADMAP.docx` must remain materially synchronized with one another and with the authoritative project or service record. Update both copies whenever feature scope, priority, dependency, implementation status, cancellation, supersession, recommendation, or verification state materially changes.

No feature may be represented as complete or Stable solely because it appears in this roadmap. Completion and lifecycle claims require the applicable authoritative implementation, validation, review, release, and production evidence.

## Reconciliation rule

At each material feature change, reconcile this roadmap against the current authoritative project record, repository implementation state, applicable Platform-System requirements, and GoreeCloud Tasks Management. Missing obligations, stale status, duplicated work, roadmap drift, or undocumented disposition changes are defects to correct.
