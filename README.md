# GoreeCloud Search

GoreeCloud Search is the first-party private metasearch and research application for GoreeCloud. The long-term product is original GoreeCloud-owned native software; the inherited SearXNG-derived tree is transitional and is retained only where it is still needed for service continuity, feature preservation, migration, compatibility, and upstream security maintenance.

## Lifecycle status

**Native migration in progress — not Stable.**

The repository contains both the transitional SearXNG-derived implementation and the growing native GoreeCloud Search implementation. A source merge, passing CI run, or existing transitional deployment does not authorize Stable promotion or retirement of the inherited runtime.

Stable remains blocked until the native application has completed applicable functional, migration, accessibility, runtime, recovery, real-environment, security, privacy, and current-Glaze acceptance work.

## Native implementation

The native application lives under `native/` and is implemented as a GoreeCloud-owned Go service.

Current source areas include:

- `native/cmd/searchd` — native service entry point.
- `native/internal/search` — native search application logic.
- `native/internal/preferences` — first-party Search preference state and behavior.
- `native/internal/webui` — native GoreeCloud Search web presentation.
- `native/internal/syncstate` — application-owned GoreeCloud Sync capability, signing, submission, retrieval, deletion, and validation boundaries.
- `native/docs/EXPERIENCE-REVAMP.md` — native experience, feature-preservation, preferences, accessibility, and migration direction.
- `native/docs/SEARCH-INDEX-BROWSER-CONTRACT.md` — authoritative integration contract for Search delegation from GoreeCloud Index and GoreeCloud Browser.

The native Sync client advertises application capability explicitly, requires negotiated schema conformance, bounds record and continuation identifiers, requires authenticated submission, and preserves Privacy Shield data minimization by keeping deletion tombstones free of application payload.

## Search authority

GoreeCloud Search is authoritative for Internet/web/current-information search.

- **GoreeCloud Index** may include Search as an explicitly authorized remote provider while remaining the universal/local federated indexing authority.
- **GoreeCloud Browser** may delegate non-URL search input to Search while retaining navigation, tab, page-lifecycle, and executable-destination authority.
- Neither consumer may treat Search availability as blanket authorization to transmit a query.
- Privacy Shield remains authoritative for applicable remote-query purpose, minimization, destination, and retention decisions.

The initial interoperability surface is capability `search.query`, contract version `1`, endpoint `/api/v1/search`, with a minimized request containing normalized query, category, and bounded result limit.

See [`native/docs/SEARCH-INDEX-BROWSER-CONTRACT.md`](native/docs/SEARCH-INDEX-BROWSER-CONTRACT.md).

## Transitional implementation

The SearXNG-derived source remains a migration dependency, not the target GoreeCloud application architecture. It may continue to provide retained functionality while native equivalents are implemented and accepted.

Inherited user-facing capabilities must be inventoried as `retain`, `replace`, `improve`, or explicitly approved `retire` before the transitional runtime is removed. Upstream copyright, AGPL licensing, attribution, and relevant security/update obligations remain in force while inherited code is present.

## Mandatory platform gates

GoreeCloud Search must continuously conform to current approved contracts for:

- **Glaze UI** — first-party responsive presentation, accessibility, interaction, appearance, and adaptive layout behavior.
- **Wardveil Security** — application security state, safe external-content behavior, diagnostics, and protection integration where applicable.
- **Privacy Shield** — data minimization, privacy-preserving defaults, query/history controls, purpose limitation, and application-owned privacy boundaries.
- **Everkeep** — portability, recovery, backup/restore, migration, and continuity requirements where applicable.

Missing, materially outdated, or unvalidated mandatory integration keeps the application non-Stable.

## Glaze UI requirement

The current official Stable consumer target published by `GoreeCloud/goreecloud-glaze-ui` is **Glaze UI V1.4 / `1.4.0`**.

Search-owned web and native presentation remains migration-required until repository-local V1.4 adoption and acceptance evidence exists. The design-system Stable release does not automatically make Search conformant or Stable.

Any later Glaze UI Stable promotion becomes the new required target automatically under GoreeCloud policy.

## Validation

The repository uses separate validation layers for native development and the transitional compatibility surface. Relevant workflows include:

- GoreeCloud Search Native Foundation
- GoreeCloud foundation
- GoreeCloud runtime smoke
- GoreeCloud container build
- GoreeCloud browser acceptance
- GoreeCloud platform integrations
- GoreeCloud upstream container boundary
- Integration
- Documentation

Passing source validation proves only the revision and scope exercised by those checks. Production provider acceptance, private-access policy, monitoring, recovery, migration, and Stable qualification remain separate evidence requirements.

For the native Go module:

```bash
cd native
go test ./...
go build ./cmd/searchd
```

These commands validate native source locally; they do not perform or authorize a production deployment.

## Product direction

The native rebuild preserves useful search capabilities while replacing inherited product architecture with GoreeCloud-owned behavior. The target includes:

- first-party homepage and result experiences;
- organized preferences;
- provider-adapter boundaries;
- privacy-preserving local controls;
- accessible keyboard/touch behavior;
- provider degradation handling;
- data portability;
- GoreeCloud Browser/OpenSearch integration;
- GoreeCloud Index remote-provider integration;
- controlled migration from the transitional runtime.

Search should not become a general local-device index. Device/application/content federation belongs to GoreeCloud Index, with explicit handoff when Internet/current-information retrieval is required.

## Repository records

- `FEATURES.md` — implemented, candidate, and planned capabilities.
- `BENEFITS.md` — user, administrative, privacy, resilience, and ownership benefits.
- `COMPETITIVE-OBJECTIVES.md` — product benchmarks and differentiators.
- `native/docs/EXPERIENCE-REVAMP.md` — native experience and migration contract.
- `native/docs/SEARCH-INDEX-BROWSER-CONTRACT.md` — cross-product authority and query contract.
- `docs/goreecloud/READINESS.md` — transitional/runtime release-readiness boundaries where still applicable.
- `LICENSE` — repository licensing terms.

## License and upstream provenance

The inherited SearXNG source is licensed under the GNU Affero General Public License v3.0 or later. GoreeCloud preserves required source availability, copyright, attribution, and license obligations while progressively replacing application-defining inherited code with original native GoreeCloud implementation.
