# GoreeCloud Search — Repository Notes

## Current development baseline

The repository's authoritative `main` line was initialized on September 16, 2026. The first implementation slice starts with a GoreeCloud-owned, side-effect-free query and planning core before any live provider or platform-service network integration.

## Design decisions

- Python 3.11+ is used for the initial core; there are no runtime third-party dependencies.
- Provider execution is separated behind a protocol so GoreeCloud Index and optional external sources remain replaceable.
- Privacy-restrictive source modes never silently substitute external providers.
- This version is Development only.

## Open decisions

- Final production service/runtime framework.
- Public source licensing/rights model.
- Approved production provider set.
- Exact current Stable Platform-System contract versions when integrations are implemented.
- Production deployment topology and persistence model.
