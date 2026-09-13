# GoreeCloud Search — Feature Roadmap

**Status:** Active roadmap control  
**As of:** 2026-09-12  
**Authoritative project record:** Project Specification — Search  
**Canonical repository:** GoreeCloud/goreecloud-search  
**Drive control:** `GoreeCloud/Feature Roadmap/GoreeCloud Search/FEATURE-ROADMAP.docx`

## Purpose

This file is the repository-side feature roadmap control for GoreeCloud Search. It records current planned and recommended feature work without replacing the authoritative project record, implementation evidence, release gates, or GoreeCloud Tasks Management.

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
| FR-011 | Add privacy-minimized authenticated-session acceptance evidence so exact-host TinyFish bindings distinguish configured state from observed setup, session reuse, Vault repair, current failures, and secret-handling defects. | High | Development source foundation in progress |

## Maintenance and synchronization

This roadmap and the corresponding Drive `FEATURE-ROADMAP.docx` must remain materially synchronized with one another and with the authoritative project or service record. Update both copies whenever feature scope, priority, dependency, implementation status, cancellation, supersession, recommendation, or verification state materially changes.

No feature may be represented as complete or Stable solely because it appears in this roadmap. Completion and lifecycle claims require the applicable authoritative implementation, validation, review, release, and production evidence.

## Reconciliation rule

At each material feature change, reconcile this roadmap against the current authoritative project record, repository implementation state, applicable platform-system requirements, and GoreeCloud Tasks Management. Missing obligations, stale status, duplicated work, roadmap drift, or undocumented disposition changes are defects to correct.
