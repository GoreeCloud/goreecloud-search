# GoreeCloud Search — Repository Specifications

## Lifecycle

- Product: GoreeCloud Search
- Version: `0.1.0.dev3`
- Lifecycle: Development
- Stable: No

## Current implementation boundary

The current native development candidates remain intentionally narrow. They provide a side-effect-free query parser and source planner plus a versioned Search ↔ Index contract boundary, Search-owned normalization/deduplication primitives, and an initial deterministic ranking layer that can be embedded by later API, Browser, Index transport, and AI integrations.

### Query model

The parser currently supports:

- Free-text terms.
- Double-quoted phrases.
- Excluded terms prefixed with `-`.
- `site:`
- `-domain:`
- `filetype:`
- `ext:`
- `before:YYYY-MM-DD`
- `after:YYYY-MM-DD`
- `language:`
- `region:`
- `source:`
- `category:`
- `lens:`

Unsupported syntax must remain treated as ordinary query text or be rejected explicitly rather than causing hidden network behavior.

### Source modes

The source planner implements these policy modes:

- `index_first`
- `federated`
- `goreecloud_only`
- `external_only`
- `offline_local`

The planner never executes a provider. It returns an ordered plan.

### Privacy invariant

`goreecloud_only` and `offline_local` must never produce a plan that discloses the query to a third-party provider.

### Provider boundary

Provider implementations must conform to the repository's typed provider descriptor and search protocol. Provider adapters must remain replaceable and must declare supported categories and origin.

## Not yet implemented

- Authenticated live GoreeCloud Index transport and runtime integration.
- External search-provider adapters.
- HTTP API.
- Snippet generation and advanced ranking signals beyond the current deterministic baseline.
- SafeSearch enforcement.
- Private View.
- GoreeCloud Browser integration.
- GoreeCloud AI answer integration.
- Identity-authenticated service requests.
- Privacy Shield runtime authorization.
- Wardveil Security runtime checks.
- GoreeCloud Mesh discovery/events.
- GoreeCloud Manager administration.
- User/admin Glaze UI surfaces.
- Persistent search history or GoreeCloud Sync.
- Production deployment artifacts.

These remain planned and must not be represented as implemented until code and verification evidence exist.

## Search ↔ Index contract boundary

The current development candidate defines `goreecloud.search-index.v1` as the first versioned in-process contract model between Search and a future authenticated GoreeCloud Index transport. Search owns query planning, normalization, deduplication, source agreement, and later user-facing ranking. Index supplies document candidates and index-specific provenance. The current adapter is transport-injected and does not establish live connectivity, authentication, authorization, privacy acceptance, or production runtime integration.

## Ranking boundary

The current development candidate ranks normalized results using transparent deterministic signals derived from the parsed query and normalized result evidence: quoted-phrase matches, title/snippet term matches, explicit site/filetype/language matches, and a bounded source-agreement bonus. GoreeCloud Index presence is exposed as provenance with zero ranking weight. Provider rank, click history, advertising payment, cross-query profiles, and hidden behavioral signals are not used by this baseline ranker.
