# GoreeCloud Search — Feature Roadmap

**Status:** Active roadmap control  
**As of:** 2026-09-13  
**Authoritative project record:** Project Specification — Search  
**Canonical repository:** GoreeCloud/goreecloud-search  
**Drive control:** `GoreeCloud/Feature Roadmap/GoreeCloud Search/FEATURE-ROADMAP.docx`

## Purpose

This file is the repository-side feature roadmap control for GoreeCloud Search. It records current planned and recommended feature work without replacing the authoritative project record, implementation evidence, release gates, or GoreeCloud Tasks Management.

## Current Development checkpoint

Current `master` includes the bounded first-party GoreeCloud Index consumer contract from merged PR #146 (`67ecb3db638f00125ea8d17aba05bcbaca58bf63`), the top-result diversity host-key optimization from PR #147 (`f1b12bd62a0911879ce5fde2956df9d623f7fad1`), published `search.query` capability evidence with `max_results=100` from PR #149 (`92bcce41b5b094e3b98be630cdfd1b939a2f0378`), first-party Index capability-preflight guidance from PR #150 (`9aa6315e58b26eb17d7c92982a1d863b41b30a8b`), and the GLAZE UI V1.4 / `1.4.0` native web source-migration foundation from PR #151 (`0982c9002d199be0f5fe6496b648b26c3b701e1f`).

The Index-facing Search boundary now publishes an authoritative/current `search.query` capability record with API contract version `1`, canonical endpoint `/api/v1/search`, and maximum results `100`. The documented consumer preflight requires Index to fail closed when that capability is absent, stale/non-current, non-authoritative, contract-incompatible, endpoint-incompatible, or otherwise outside the published result bound. Compatibility evidence does not grant authorization or production acceptance, and the contract still forbids unrelated local Index context from crossing the Search boundary.

PR #151 moved the served native Search web runtime off its stale local Glaze UI V1.1 marker and onto a repository-local GLAZE UI V1.4 / `1.4.0` source projection. Home, Preferences, and results now expose the V1.4 runtime marker and bounded `adaptive-optical` state; the appearance bootstrap no longer resets the DOM to V1.1; Forced Colors and Reduced Transparency use solid accessible fallbacks; and no wallpaper, camera, browsing-content, telemetry, personal-data, or remote-context collection was introduced. Exact-head validation passed all 13 triggered workflow families, including supply-chain guard, native foundation, browser acceptance, native results/image browser acceptance, V1.4 whole-shell acceptance, RTL/2×-scale resilience, deterministic 200% text stress, Integration, Documentation, container, and Development artifact gates. These are Development automated/source/rendering results only. Full canonical V1.4 Stable entrypoint/runtime reconciliation, rollback verification, manual accessibility/optical review, representative physical-device performance, platform/provider production acceptance, Release Candidate, production, and Stable qualification remain open.

## Roadmap

| ID | Feature / obligation | Priority | Current state |
| --- | --- | --- | --- |
| FR-001 | Reconcile and maintain every current planned or recommended GoreeCloud Search feature from the authoritative project record and verified repository evidence in this roadmap. | High | Ongoing control |
| FR-002 | Move actionable feature obligations into GoreeCloud Tasks Management when required, preserving priority, dependency, and lifecycle disposition. | High | Ongoing control |
| FR-003 | Do not mark features implemented, complete, cancelled, or superseded without authoritative evidence and synchronized repository/Drive roadmap updates. | High | Ongoing control |
| FR-004 | Implement and validate the native `tinyfish-search-v1` provider adapter for TinyFish Search, initially limited to General/web and News with deployment-controlled API-key authentication and fail-closed category claims. | High | Development merged; production acceptance pending |
| FR-005 | Add a bounded TinyFish Fetch retrieval stage behind a GoreeCloud-owned retrieval contract for approved URL extraction after discovery. | High | Development merged; production acceptance pending |
| FR-006 | Route TinyFish Research behind GoreeCloud Search for source-backed research without creating a direct permanent GoreeCloud AI dependency on TinyFish. | High | Development merged; production acceptance pending |
| FR-007 | Add policy-controlled TinyFish Agent and Browser escalation with separate Browser Context Profiles, approved Vault usage, and explicit Privacy Shield/Wardveil boundaries. | Medium | Development merged; production acceptance pending |
| FR-008 | Add least-capability routing, metered-cost controls, health/observability, replacement/fallback behavior, and Manager operational visibility for the TinyFish integration. | High | Development merged; production acceptance pending |
| FR-009 | Add exact-host authenticated-session bindings so recurring TinyFish automation deterministically selects one saved Browser Context Profile and only explicitly scoped Vault credential references for the intended service/account boundary. | High | Development merged; production acceptance pending |
| FR-010 | Add a fail-closed Browser Context Profile lifecycle and recovery primitive for authorized create/setup/save/cancel operations so recurring TinyFish sessions can be deliberately established or repaired without exposing reusable secrets. | High | Development merged; production acceptance pending |
| FR-011 | Add privacy-minimized authenticated-session acceptance evidence so exact-host TinyFish bindings distinguish configured state from observed setup, session reuse, Vault repair, current failures, and secret-handling defects. | High | Development merged; production acceptance pending |
| FR-012 | Bind recurring authenticated TinyFish execution to the managed exact-host profile/Vault registry and automatically feed normalized execution outcomes into privacy-minimized session acceptance evidence without inferring Vault repair from credential availability alone. | High | Development merged; production acceptance pending |
| FR-013 | Add bounded deployment loading and a read-only privacy-minimized native status surface for authenticated TinyFish automation so operators can distinguish unconfigured, unverified, reusable, blocked, failed, expired, revoked, and disclosure-risk session states without exposing profile or Vault item identifiers. | High | Development merged; production acceptance pending |
| FR-014 | Require fail-closed GoreeCloud authorization after exact-host profile/Vault binding and request normalization but before any authenticated TinyFish provider execution, so policy evaluates the actual managed session and credential scope crossing the provider boundary. | High | Development merged; production acceptance pending |
| FR-015 | Add a fail-closed provider-neutral direct Browser session lifecycle and exact TinyFish Browser API transport with bounded inactivity timeouts, sensitive CDP/session handling, conservative uncertain-outcome semantics, explicit termination control, and no guessed provider fields. | Medium | Development merged; production acceptance pending |
| FR-016 | Add evidence-driven provider health acceptance gates for metered Agent and future Browser escalation, including explicit success/latency thresholds, cancellation/rejection treatment, fail-closed unavailable routing, and a path to bounded recent-window and durable privacy-minimized health evidence. | High | Development merged; production acceptance pending |
| FR-017 | Add a scoped direct Browser session lease that validates cleanup authority before creation, always attempts termination after confirmed creation, uses an independent bounded cleanup context after caller cancellation, preserves work and cleanup failures together, and never treats uncertain cleanup as success. | Medium | Development merged in PR #144 after exact-head CI passed; production acceptance pending. |
| FR-018 | Define and validate a bounded first-party GoreeCloud Index consumer contract over the native Search API, including additive result limits, explicit capability evidence, fail-closed consumer preflight, degraded-state preservation, and data-minimization guidance that forbids unrelated local Index context from crossing the Search boundary. | High | Development source contract is merged: PR #146 added the 1–100 result limit and minimized consumer contract; PR #149 publishes `search.query` API-v1 capability evidence with `/api/v1/search` and `max_results=100`; PR #150 documents the fail-closed Index preflight. Live Index transport/integration and production acceptance remain pending. |
| FR-019 | Complete GoreeCloud Search adoption and product acceptance against the current GLAZE UI V1.4 / `1.4.0` Stable authority, including canonical Stable CSS/runtime reconciliation, accessible optical behavior, rollback, representative-browser/device performance, and product release evidence. | High | V1.4 native web source foundation is merged in PR #151. Exact-head automated browser evidence covers home/preferences/results/images, light/dark/deep-dark, V1.4 whole-shell behavior, RTL/2×-scale resilience, and deterministic 200% text stress. Full canonical entrypoint/runtime reconciliation, rollback, manual accessibility/optical/device qualification, production acceptance, and Stable status remain open. |

## Current sequencing recommendation

1. Continue from the validated current native Search baseline and preserve the explicit `search.query` capability, API-v1 endpoint/result-bound contract, and privacy-minimized Index preflight as first-party integration work advances.
2. Keep Index live transport, service discovery, authorization, DNS/TLS/proxy behavior, runtime authority, and production acceptance separately gated; do not infer them from compatibility or capability evidence.
3. Complete GLAZE UI V1.4 consumer adoption beyond the merged source projection by reconciling the canonical Stable CSS/runtime entrypoints, verifying rollback, and completing remaining manual accessibility/optical, representative-device, performance, and product acceptance gates without fabricating V1.4.1 human-only evidence.
4. Continue behavior-preserving performance work only with exact-revision evidence, and preserve Search ownership of Internet/current-information discovery rather than duplicating remote-provider logic in Index.
5. Complete remaining provider/platform runtime acceptance, monitoring/recovery, signing/provenance, Release Candidate, production, and Stable gates before any lifecycle advancement.

## Maintenance and synchronization

This roadmap and the corresponding Drive `FEATURE-ROADMAP.docx` must remain materially synchronized with one another and with the authoritative project or service record. Update both copies whenever feature scope, priority, dependency, implementation status, cancellation, supersession, recommendation, or verification state materially changes.

No feature may be represented as complete or Stable solely because it appears in this roadmap. Completion and lifecycle claims require the applicable authoritative implementation, validation, review, release, and production evidence.

## Reconciliation rule

At each material feature change, reconcile this roadmap against the current authoritative project record, repository implementation state, applicable platform-system requirements, and GoreeCloud Tasks Management. Missing obligations, stale status, duplicated work, roadmap drift, or undocumented disposition changes are defects to correct.
