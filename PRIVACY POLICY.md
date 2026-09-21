# GoreeCloud Search — Privacy Policy

## Current development implementation

Version `0.1.0.dev5` contains a local query parser, source planner, bounded provider-execution engine, versioned Search ↔ Index contract models with pagination, a transport-injected Index adapter boundary, result normalization/deduplication, and deterministic ranking. The repository code in this version:

- Ships no authenticated live network transport or approved external network provider. The development CLI performs no network access.
- Can execute explicitly injected provider adapters. If an embedding supplies a network-capable adapter, that adapter determines the resulting network disclosure and must remain within the planner-approved source plan.
- Supports an optional per-query disclosure budget that caps the number of distinct third-party providers admitted to the source plan before execution. A zero budget prevents third-party execution entirely for modes that still have an eligible first-party/local path, and makes an external-only request unsatisfiable rather than disclosing the query.
- Does not persist search history.
- Does not contain advertising or tracking code.
- Does not build behavioral profiles.
- Does not require third-party analytics.
- Does not use click history, cross-query behavioral profiles, advertising identifiers, or paid placement for the implemented ranking baseline.

## Source planning

The source planner marks whether an execution plan would disclose a query to a third-party provider and records the selected third-party-provider count, any explicit budget, and the number of otherwise eligible third-party providers omitted by that budget.

`goreecloud_only` and `offline_local` are required to fail closed against third-party query disclosure.

## Future network features

Any live network provider transport, remote suggestion, synchronization, account storage, analytics, AI, Browser, Index, or other network-capable feature must update this policy to match verified behavior and must apply the relevant GoreeCloud Identity, Privacy Shield, Wardveil Security, retention, consent, minimization, and disclosure controls before production acceptance.

## Development warning

This document describes the current repository implementation. It does not claim that a future hosted deployment has the same data flows until that deployment is separately verified.

## Authenticated Index Delegation Candidate — 2026-09-21

The Development HTTP boundary accepts only `query`, `category`, and `limit` in the request body. The GoreeCloud Identity credential and opaque Privacy Shield capability reference use separate headers and are not included in Search results or ordinary request logging. Request syntax is validated before the capability-reference verifier is asked to consume the one-operation reference.

No live authority service, approved external provider, retained search history, production processing, or deployment is enabled by this candidate.
