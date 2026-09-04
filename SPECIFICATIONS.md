# GoreeCloud Search Specifications

## Product and lifecycle

GoreeCloud Search is the first-party GoreeCloud private metasearch and research application.

- Lifecycle: native migration in progress — pre-Stable.
- Canonical repository: `GoreeCloud/goreecloud-search`.
- Native implementation: `native/` Go service and GoreeCloud-owned web experience.
- Transitional implementation: inherited SearXNG-derived tree retained only for continuity, feature preservation, migration, compatibility, and applicable upstream security maintenance.
- Production/Stable approval: not established by source presence or CI success.

The target product architecture is GoreeCloud-owned native software. Transitional SearXNG code is not the permanent application architecture.

## Native search engine contract

The native engine currently provides source-level application logic for:

- bounded query validation with a 512-rune maximum;
- explicit categories: General, Images, Videos, News, and Files;
- General-category empty-provider development behavior;
- fail-closed specialized-category execution unless a configured provider has an executable category path;
- concurrent provider execution under one bounded request context;
- per-provider availability/timeout status without exposing provider error strings to the response;
- a deterministic 512-result processing ceiling per provider per request before native URL sanitization and ranking; an oversized provider result slice is copied to the bounded working set, provider status records the processed count, and `truncated=true` discloses that the ceiling was applied;
- HTTP/HTTPS result URL validation with fragment removal;
- rejection of result URLs containing embedded user-info credentials;
- provider identity normalization with a 128-rune maximum and rejection of blank or control-character names before advertisement or execution;
- sanitized provider-definition exposure that does not publish credentials, endpoints, mutable controls, or runtime errors;
- GoreeCloud-owned request-local result ranking rather than direct cross-provider trust in arbitrary provider score scales;
- Unicode-aware query/result token normalization for deterministic relevance scoring;
- local query-intent parsing that separates `site:`, `filetype:`/`ext:`, quoted phrases, domain-directed targets, and temporal modifiers from ordinary relevance tokens without rewriting the submitted provider query;
- strong title relevance and exact-title signals, with lower-weight snippet and URL token coverage;
- bounded quoted-phrase boosts when an explicitly quoted multi-word phrase remains contiguous in a result title, snippet, or URL text;
- bounded one-edit or adjacent-transposition tolerance for tokens of at least five runes, with fuzzy title/snippet/URL contributions kept materially below exact relevance signals;
- conservative user-visible query correction derived only from result-title evidence already present in the current request: Search proposes at most one unquoted, non-operator, non-domain token change, requires at least two independent normalized providers to support the same one-edit/transposition alternative, fails closed on ambiguity, and never rewrites the query submitted to providers;
- no fuzzy matching for shorter tokens, reducing accidental matches on common short words;
- bounded positive preference for explicit `site:` host/subdomain matches and bounded demotion of results that violate an explicit site target;
- bounded positive preference for domain-directed navigational results whose hostname matches the requested domain;
- bounded positive preference for `filetype:`/`ext:` URL extensions and bounded demotion when an explicit requested extension does not match;
- actual domain-target recognition rather than a generic dotted-token heuristic, preventing dotted versions such as `1.5.0` from disabling hostname diversity;
- provider-supplied integer score retained only as bounded supporting evidence with a maximum contribution of 300 ranking points;
- explicit provider-level `PublishedAtProvider` authority for publication/update timestamps: adapters may opt in only when `Result.PublishedAt` is copied from a trustworthy upstream field with publication semantics rather than inferred from snippets, URLs, crawl time, or provider score;
- stripping of untrusted, zero, pre-Unix, or more-than-24-hours-future publication timestamps before aggregation; retained timestamp output records the authoritative provider in `published_at_source`;
- request-local freshness scoring only for explicit temporal intent or the News category;
- clear unquoted modifiers `latest`, `recent`, `recently`, `today`, `breaking`, and `newest`, plus the phrases `this week` and `this month`, activate freshness while being omitted from ordinary lexical relevance so they do not dilute the actual subject terms;
- leading unquoted `current` activates freshness and is omitted from lexical relevance, while noun/non-leading uses such as `electric current` remain ordinary lexical content and do not activate freshness;
- content-bearing unquoted terms `news`, `updated`, and `updates` may activate freshness but remain lexical terms because they can describe the subject itself;
- quoted temporal wording remains literal phrase/text relevance and does not independently activate freshness, so a query such as `"latest goreecloud" search` keeps `latest` as requested content rather than silently interpreting it as a recency modifier;
- a freshness contribution bounded to at most 1,200 ranking points, with declining buckets through 90 days and a lower implicit weight for News-category recency when the query itself has no temporal wording;
- no freshness bias for ordinary General searches that do not express temporal intent;
- temporal-only modifier queries may have no lexical tokens and can still order accepted timestamp-bearing candidates by the bounded freshness signal rather than manufacturing synthetic relevance terms;
- exact normalized-URL clustering that preserves deterministic source-agreement evidence rather than discarding duplicate-provider consensus;
- source-agreement bonus bounded to 900 ranking points, preventing consensus from becoming unlimited ranking authority;
- selection of the most query-relevant duplicate title/snippet as the displayed representative while preserving deterministic tie breaks;
- first-viewport hostname diversity of at most two results per hostname when other ranked hosts are available;
- no hostname-diversity override for explicit `site:` or valid domain-directed queries, where concentration is likely intentional;
- final deterministic score/URL/provider ordering before the bounded diversity pass;
- native result metadata for deterministic `source_count` and sorted `sources` provenance.

The native ranking and correction paths are deterministic and request-local. They do not use click history, behavioral profiles, advertising signals, sponsored placement, remote spelling/correction lookups, remote freshness lookups, or remote ranking telemetry. Search does not silently rewrite or replace the query sent to providers. A user-visible correction is only an explicit optional link derived from bounded provider-title evidence already returned for the same request. Freshness uses only timestamp metadata that passed the explicit provider-authority boundary; Search does not infer publication time from result text or URL structure. Temporal-modifier separation is also request-local parsing: it changes only GoreeCloud-owned reranking interpretation and does not alter the original query submitted to configured providers.

The provider result ceiling bounds Search-owned post-provider work; it cannot prevent an adapter or external provider implementation from allocating its own oversized response before returning. Production provider adapters therefore remain responsible for their own transport/body/result bounds in addition to this engine-level ceiling.

The native tree now also contains the GoreeCloud-owned deployment-controlled provider runtime under `native/internal/providers`. Source-level implementation includes the schema-versioned `goreecloud-http-v1` category-aware contract, an explicit provider-config-file activation boundary, zero configured providers by default, environment-only references for optional bearer credentials, fail-closed malformed or incomplete configured-provider startup, HTTPS/public-address-only outbound enforcement with connection-time DNS revalidation, redirect rejection, strict JSON response parsing, bounded provider response/result processing, and sanitized capability exposure that does not reveal endpoints or credentials. This source capability does not select or approve a real external provider.

No external provider is production-approved merely because the native provider interfaces or runtime exist. Provider selection, credentials, privacy policy and data use, terms, health, rate limiting, degradation behavior, abuse controls, timestamp-authority review, live-provider behavior, and target-runtime evidence remain separate acceptance work.

## Native presentation and preferences

The native tree contains GoreeCloud-owned homepage/results/presentation work under `native/internal/webui` and first-party preference state under `native/internal/preferences`.

The native results surface now implements source-level scan-first presentation with:

- a compact persistent query field and category navigation;
- restrained list-based result composition rather than a wall of equally elevated cards;
- title-first hierarchy, subordinate URL/snippet treatment, and no user-visible internal numeric ranking score;
- source-agreement disclosure for clustered multi-provider results;
- human-readable publication dates only when `PublishedAt` survived the authoritative provider boundary, rendered through semantic `<time datetime>` markup rather than inferred or synthetic dates;
- an explicit optional “Search instead for” correction link only when the bounded local provider-agreement rule produces a single unambiguous alternative;
- a ranking explanation that discloses trustworthy freshness as a conditional signal when the query/category requests recency;
- a separate source-health surface for available/degraded provider state, including visible “limit applied” disclosure when a provider exceeds the Search-owned result-processing ceiling;
- explicit local-ranking/privacy explanation without claiming anonymity from external providers;
- adaptive desktop and narrow-window composition;
- a 48px minimum interaction-target floor for the tested native results controls across Compact, Medium, Expanded, and Wide browser acceptance, plus visible focus and reduced-motion, increased-contrast, forced-colors, and reduced-transparency fallbacks;
- script-free result rendering through Go `html/template`, preserving automatic escaping of query/provider/title/snippet/source content.

The native Images slice additionally contains bounded image media metadata, same-origin mediated image loading, a responsive image grid, full-image/source viewing, keyboard previous/next/Escape behavior, focus restoration, and deterministic browser acceptance across representative responsive and accessibility modes. Source integration of these capabilities does not establish live-provider or production deployment acceptance.

Search-owned surfaces must use the latest approved Stable Glaze UI contract when production acceptance is evaluated. The current required application target is GLAZE UI V1.1 / 1.1.0. Reset-baseline, superseded, Candidate, RC, or earlier product-version evidence remains historical migration or audit evidence and does not establish current consumer conformance. Source structure, CSS implementation, unit tests, or bounded rendered acceptance alone do not establish whole-application visual/accessibility/device or production conformance.

## Sync boundary

Search owns the semantics of `search.history`; GoreeCloud Sync coordinates authorized replication.

Current native Sync source includes capability/schema negotiation, bounded record/cursor identifiers, authenticated submission, exact envelope validation, payload-free deletion tombstones, retrieval pagination, and client-side Ed25519 record-proof preflight against the canonical Sync vector.

This does not make account history synchronization production-ready. Production identity/session authority, deployed transport, privacy controls, recovery, and end-to-end acceptance remain required.

## Privacy boundary

Privacy Shield is authoritative for Search data-use governance.

Search requirements include:

- no GoreeCloud advertising or sponsored-result ranking;
- no behavioral profiling business model;
- minimized persistent query/history state;
- explicit user controls before account history or personalization is enabled;
- no hidden provider fallback that bypasses the configured Search authority;
- bounded Search-owned per-provider result processing for native ranking/resource control;
- no result URL user-info credentials entering the native response surface;
- provider errors represented by bounded status codes rather than raw error text that may contain secrets;
- no click-history or behavioral-profile dependency in native relevance ranking;
- no external lookup requirement for the implemented intent parsing, typo-tolerant ranking, explicit local correction suggestion, or freshness scoring signals;
- no inferred publication time from result text/URLs and no use of provider timestamps unless the adapter explicitly satisfies the publication-metadata authority contract;
- no user-visible internal ranking score that could be mistaken for a provider trust or quality guarantee.

External providers may observe requests from GoreeCloud infrastructure. Search must not claim anonymity from external providers.

## Security boundary

Wardveil Security is authoritative for GoreeCloud security acceptance. Search must fail closed on malformed native input/provider identities at enforced boundaries and must keep provider secrets outside source and user-visible diagnostics.

Production acceptance still requires applicable Wardveil runtime/evidence integration, provider-specific abuse controls, deployment hardening, and operational security validation. Generic provider-transport safeguards do not independently establish Wardveil acceptance for a selected provider or deployment.

## Continuity boundary

Everkeep is authoritative for backup, recovery, rollback, preservation, and continuity. Search source, configuration, provider policy, user-controlled state, and native migration require recoverable, documented paths before Stable promotion.

## Identity and Mesh boundaries

GoreeCloud Identity is authoritative for account/session identity and authorization. GoreeCloud Mesh is authoritative for cross-application capability coordination. Search must not create parallel identity or platform-coordination authority.

## Transitional runtime

The inherited SearXNG-derived implementation remains a temporary migration dependency. While present:

- required AGPL/source/attribution obligations remain in force;
- security-maintenance updates may be applied when needed;
- inherited user-facing features must be classified retain, replace, improve, or explicitly approved retire before removal;
- new GoreeCloud-owned product behavior should be implemented in native code unless a temporary compatibility change is necessary and documented.

## Production and Stable blockers

Stable remains blocked by at least:

- production-approved native provider adapters, provider selection, and credentials/secrets integration where required;
- complete live native category/provider coverage required for the selected release;
- production-reviewed provider timestamp authorities and live-provider acceptance for result classes where freshness is required; the source-level freshness contract/ranker alone is not production evidence;
- accepted GLAZE UI V1.1 / 1.1.0 whole-application native visual/accessibility/device evidence beyond the bounded results-renderer acceptance already present;
- applicable Wardveil Security and Privacy Shield runtime/evidence integration;
- Everkeep-backed backup/restore/migration/rollback acceptance;
- GoreeCloud Identity and Mesh integration where the release uses account-bound capabilities;
- migration parity and controlled cutover from the transitional runtime;
- monitoring, alerting, provider-specific rate-limit/degradation/abuse/resource behavior, live target-runtime validation, and rollback evidence;
- supported Browser/device/runtime acceptance;
- exact-release provenance and production approval.

A lower lifecycle state must never be represented as a higher one.