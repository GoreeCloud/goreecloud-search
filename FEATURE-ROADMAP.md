# GoreeCloud Search — Feature Roadmap

**Status:** Active roadmap control  
**As of:** 2026-09-13  
**Authoritative project record:** Project Specification — Search  
**Canonical repository:** GoreeCloud/goreecloud-search  
**Drive control:** `GoreeCloud/Feature Roadmap/GoreeCloud Search/FEATURE-ROADMAP.docx`

## Purpose

This file is the repository-side feature roadmap control for GoreeCloud Search. It records current planned and recommended feature work without replacing the authoritative project record, implementation evidence, release gates, or GoreeCloud Tasks Management.

## Current Development checkpoint

Current `master` includes the bounded first-party GoreeCloud Index consumer contract from merged PR #146 (`67ecb3db638f00125ea8d17aba05bcbaca58bf63`) and the top-result diversity host-key optimization from merged PR #147 (`f1b12bd62a0911879ce5fde2956df9d623f7fad1`). PR #146 exact head passed its full 13-workflow validation matrix before merge. PR #147 exact head `4e70814c49ec9e55b5a5b1be40f24e5dfbc23d46` passed all 12 workflows triggered for the one-file optimization before merge. These remain Development source/build/automated acceptance results; live Index transport acceptance, production provider/runtime acceptance, Release Candidate, production, and Stable gates remain open.

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
| FR-018 | Define and validate a bounded first-party GoreeCloud Index consumer contract over the native Search API, including an additive 1–100 result limit and explicit data-minimization guidance that forbids unrelated local Index context from crossing the Search boundary. | High | Development source foundation merged in PR #146 after exact-head CI; live Index transport/integration and production acceptance remain pending. |

## Current sequencing recommendation

1. Continue validating the current native Search baseline and preserve the bounded Index API contract as first-party consumers evolve.
2. Keep Index integration transport, service discovery, runtime authority, and production acceptance separately gated; do not treat the merged API contract as live end-to-end acceptance.
3. Continue performance work only with behavior-preserving exact-revision evidence, as demonstrated by the host-key optimization in PR #147.
4. Complete current Glaze UI application acceptance and the remaining provider/platform production gates without reusing historical acceptance as current evidence.
5. Advance Release Candidate, production, and Stable status only after the applicable runtime, accessibility, representative-target, rollback, monitoring, and recovery gates pass.

## Maintenance and synchronization

This roadmap and the corresponding Drive `FEATURE-ROADMAP.docx` must remain materially synchronized with one another and with the authoritative project or service record. Update both copies whenever feature scope, priority, dependency, implementation status, cancellation, supersession, recommendation, or verification state materially changes.

No feature may be represented as complete or Stable solely because it appears in this roadmap. Completion and lifecycle claims require the applicable authoritative implementation, validation, review, release, and production evidence.

## Reconciliation rule

At each material feature change, reconcile this roadmap against the current authoritative project record, repository implementation state, applicable platform-system requirements, and GoreeCloud Tasks Management. Missing obligations, stale status, duplicated work, roadmap drift, or undocumented disposition changes are defects to correct.
