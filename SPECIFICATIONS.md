# GoreeCloud Search — Repository Specifications

## Lifecycle

- Product: GoreeCloud Search
- Version: `0.1.0.dev1`
- Lifecycle: Development
- Stable: No

## Current implementation boundary

The current native implementation is intentionally narrow. It provides a side-effect-free query parser and source planner that later API, Browser, Index, and AI integrations can consume.

### Query model

Supported today:

- Free-text terms and double-quoted phrases.
- Excluded terms prefixed with `-`.
- `site:`, `-domain:`, `filetype:`, `ext:`, `before:`, `after:`, `language:`, `region:`, `source:`, `category:`, and `lens:`.

### Source modes

The planner implements `index_first`, `federated`, `goreecloud_only`, `external_only`, and `offline_local`.

The planner returns an ordered plan and never executes a provider.

### Privacy invariant

`goreecloud_only` and `offline_local` must never produce a plan that discloses a query to a third-party provider.

### Provider boundary

Future providers must use the repository's typed descriptor and provider protocol, remain replaceable, and declare supported categories and origin.

## Not yet implemented

Live GoreeCloud Index access, external provider adapters, HTTP APIs, normalization, deduplication, ranking, SafeSearch enforcement, Private View, Browser integration, AI synthesis, Identity, Privacy Shield, Wardveil, Mesh, Manager, Glaze UI, Sync, persistence, and production deployment are not implemented and must not be represented as complete.
