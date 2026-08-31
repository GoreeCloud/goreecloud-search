# GoreeCloud Search Features

This file records Search functionality and implementation state. A listed capability is not a production-readiness or Stable claim unless current runtime and acceptance evidence support it.

## Implemented native foundations

### Native search engine

- GoreeCloud-owned Go search engine under `native/internal/search`.
- Bounded 512-rune query validation.
- Explicit General, Images, Videos, News, and Files category model.
- General-category development execution even with no configured providers.
- Specialized categories fail closed unless an executable provider path exists.
- Concurrent provider execution under a bounded request timeout; the request returns degraded timeout evidence even when an integration adapter fails to honor context cancellation.
- Bounded provider availability/timeout status without returning raw provider error messages.
- Deterministic result ordering and de-duplication.
- HTTP/HTTPS result URL normalization with fragments removed.
- Result URLs with embedded user-info credentials rejected.
- Provider identities normalized and bounded to 128 runes; blank/control-character provider names are not advertised or executed.
- Sanitized provider capability definitions without credentials, endpoints, runtime errors, or mutable controls.

### Native GoreeCloud product experience

- GoreeCloud-owned native service entry point and web-presentation foundation.
- First-party homepage/results presentation work.
- First-party preference state and organization.
- GoreeCloud product identity rather than an upstream-only shell.
- Native-first development direction governed by the latest applicable Stable Glaze UI contract.

### Privacy-first behavior

- No GoreeCloud advertising or sponsored-ranking business model.
- No intended behavioral profiling.
- Explicit query bounds and minimized native state.
- No hidden fallback requirement that bypasses GoreeCloud Search as configured Browser search authority.
- Provider errors reduced to bounded status codes.
- Result URLs containing embedded credentials rejected before presentation.

### GoreeCloud Sync foundations

- Application-owned `search.history` capability contract.
- Exact dataset/schema validation.
- 512-byte record/cursor bounds.
- Authenticated submission requirements.
- Payload-free deletion tombstones.
- Bounded paginated retrieval.
- Client-side Ed25519 record-proof verification against the canonical GoreeCloud Sync vector.

These are source contracts; they do not imply that account search-history synchronization is enabled in a production environment.

## Transitional retained capabilities

The repository still contains the inherited SearXNG-derived runtime for continuity and migration. While it remains in use it may provide broader mature metasearch/provider/category functionality than the current native implementation.

Inherited capability presence is transitional evidence, not the target application architecture. Features must be retained, replaced, improved, or explicitly approved for retirement before the inherited runtime is removed.

## Planned / incomplete native capabilities

- Production-approved native external-provider adapters and credentials integration.
- Accepted native provider coverage for every category selected for release.
- Completed feature-parity migration from the inherited runtime.
- Full native Glaze UI visual/accessibility/device acceptance.
- Complete Privacy Shield, Wardveil Security, Everkeep, GoreeCloud Identity, and GoreeCloud Mesh runtime/evidence integration where applicable.
- Production-approved account search history, saved searches, synchronization, or personalization.
- Governed machine-readable Search API for approved first-party/AI/research consumers.
- Complete monitoring, rate limiting, abuse protection, target-host deployment, restore, rollback, migration, and operational evidence.
- Stable lifecycle promotion of the native application.

## Feature-governance rule

Planning, source implementation, CI validation, runtime integration, production acceptance, migration cutover, and Stable promotion are separate states. Search documentation must preserve those distinctions.
