# GoreeCloud Search — Privacy Policy

## Current development implementation

Version `0.1.0.dev7` contains a local query parser, source planner, bounded provider-execution engine, versioned Search ↔ Index contract models with pagination, a transport-injected Index adapter boundary, result normalization/deduplication, content-policy hooks, deterministic ranking, and Search-local Lens reranking. The repository code in this version:

- Ships no authenticated live network transport or approved external network provider. The development CLI performs no network access.
- Can execute explicitly injected provider adapters only within the planner-approved source plan.
- Supports an optional per-query disclosure budget that caps distinct third-party providers before execution.
- Applies configured local content-policy hooks after normalization and before ranking.
- Rejects Moderate/Strict SafeSearch before provider execution when no configured hook can truthfully enforce the requested mode.
- Resolves a requested Lens before provider execution and removes the `lens:` selection from the provider-facing parsed query. The selected local Lens name/rules are therefore not disclosed to provider adapters by this implementation.
- Applies Lens boost/lower/exclude rules only after normalized results pass content policy and baseline ranking.
- Does not persist search history or Lens state.
- Does not contain advertising or tracking code and does not build behavioral profiles.
- Does not use click history, cross-query behavioral profiles, advertising identifiers, or paid placement for ranking.

## Source planning

The source planner marks whether an execution plan would disclose a query to a third-party provider and records selected third-party-provider count, explicit budget, and providers omitted by that budget. `goreecloud_only` and `offline_local` fail closed against third-party query disclosure.

## Lenses

The current Lens engine is local configuration supplied to `SearchCore`. Lens rules may alter ordering or exclude results, but every applied boost/lower rule is represented as an inspectable ranking signal and every Lens exclusion is represented in the Lens application report. Lens selection is not a hidden behavioral profile.

Future Lens storage, synchronization, sharing, discovery, or remote retrieval would create additional data flows and must be documented and governed before implementation acceptance.

## Future network features

Any live network provider transport, remote suggestion, synchronization, account storage, analytics, AI, Browser, Index, or other network-capable feature must update this policy to match verified behavior and must apply the relevant GoreeCloud Identity, Privacy Shield, Wardveil Security, retention, consent, minimization, and disclosure controls before production acceptance.

## Development warning

This document describes the current repository implementation. It does not claim that a future hosted deployment has the same data flows until that deployment is separately verified.
