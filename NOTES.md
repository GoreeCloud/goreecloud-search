# GoreeCloud Search — Repository Notes

## Current development baseline

The repository's authoritative `main` line was initialized on September 16, 2026.

The first implementation slice intentionally starts with a GoreeCloud-owned, side-effect-free query and planning core before any live provider or platform-service network integration is introduced.

## Design decisions

- Python 3.11+ is used for the initial core because the slice requires no runtime third-party dependencies and can be validated quickly.
- Provider execution is separated behind a protocol so GoreeCloud Index and optional external sources can be integrated without coupling the parser/ranking core to a specific provider.
- No external provider is silently substituted when a privacy-restrictive source mode is selected.
- This version is Development only.

## Open decisions

- Final production service/runtime framework.
- Public source licensing/rights model.
- Approved production provider set.
- Exact current Stable Platform-System contract versions at the time each integration is implemented.
- Production deployment topology and persistence model.

## Stacked implementation candidate

`feature/index-contract-normalization` is based on `feature/native-search-core-foundation`. It must be retargeted or reconciled after the parent candidate changes or merges. The child candidate adds contract/normalization primitives only; it does not authorize or implement live Index network access.

## Ranking baseline

The initial ranker intentionally avoids provider-specific hidden boosts and behavioral signals. GoreeCloud Index presence is recorded as provenance but contributes zero score. Source agreement is bounded so federation consensus cannot dominate lexical/query-intent relevance.
