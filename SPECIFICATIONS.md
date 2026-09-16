# GoreeCloud Search — Repository Specifications

## Lifecycle

- Product: GoreeCloud Search
- Version: `0.1.0.dev8`
- Lifecycle: Development
- Stable: No
- License: `AGPL-3.0-or-later`

## Current implementation boundary

The current development candidates provide a query parser/privacy-aware source planner, bounded asynchronous provider execution, a versioned Search ↔ Index contract boundary with pagination, Search-owned normalization/deduplication, content policy, deterministic baseline ranking, Search-local Lens reranking, and a strict portable Lens data contract.

### Query and provider boundaries

The parser supports free text, phrases/exclusions, site/domain/filetype/date/language/region/source/category filters, and `lens:`. Source planning implements Index First, Federated, GoreeCloud Only, External Only, and Offline/Local modes with optional third-party disclosure budgets. The executor may run only plan-admitted providers.

### Lens reranking boundary

A `Lens` is a Search-local deterministic set of domain/filetype/language boost, lower, or exclude rules. Every score adjustment is an inspectable ranking signal and every exclusion is explicit evidence. Unknown requested Lenses fail before provider execution. Search strips `query.filters.lens` from the provider-facing query.

### Portable Lens contract

`goreecloud.search-lens.v1` is the current portable Lens document format. It is JSON and contains exactly: `format_version`, `name`, `description`, and `rules`; each rule contains exactly `target`, `value`, `action`, and `weight`.

Export is deterministic (`sort_keys`, stable floating-point rule weights, UTF-8 text, terminal newline). Import fails closed on incompatible versions, unknown/duplicate fields, non-finite numbers, unsupported targets/actions, invalid types/weights, empty/too-long values, more than 256 rules, or documents larger than 64 KiB. Import/export is an in-memory transformation only: there is no filesystem persistence, remote registry, network fetch, code execution, signature/trust system, or synchronization in this candidate.

### Content-policy and ranking boundaries

Content policy is evaluated after normalization and before ranking. Moderate/Strict SafeSearch fails before provider execution unless a configured hook declares enforcement. Baseline ranking is transparent/deterministic and excludes provider rank, click history, advertising payment, cross-query profiles, and hidden behavioral signals. Lens reranking occurs after the baseline/content-policy stages.

## Not yet implemented

- Authenticated live GoreeCloud Index transport and approved external providers.
- HTTP API and health/readiness endpoints.
- Snippet generation beyond provider-supplied snippets.
- Production SafeSearch/Wardveil classification feeds.
- Lens filesystem persistence, automatic UI import/export, signatures/trust, remote sharing/discovery, synchronization, and hosted registry.
- Private View, Browser/AI integration, Platform-System runtime enforcement, Glaze UI surfaces, persistent history/Sync, and production deployment artifacts.

These remain planned and must not be represented as implemented until code and verification evidence exist.
