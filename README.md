# GoreeCloud Search

GoreeCloud Search is the first-party private metasearch and research application for GoreeCloud. The long-term product is original GoreeCloud-owned native software; the inherited SearXNG-derived tree is transitional and is retained only where it is still needed for service continuity, feature preservation, migration, compatibility, and upstream security maintenance.

## Lifecycle status

**Native migration in progress — not Stable.**

The repository contains both the transitional SearXNG-derived implementation and the growing native GoreeCloud Search implementation. A source merge, passing CI run, or existing transitional deployment does not authorize Stable promotion or retirement of the inherited runtime.

Stable remains blocked until the native application has completed the applicable functional, migration, accessibility, runtime, recovery, and real-environment acceptance work and has current validated integration with Glaze UI, Wardveil Security, Privacy Shield, and Everkeep.

## Native implementation

The native application lives under `native/` and is implemented as a GoreeCloud-owned Go service.

Current source areas include:

- `native/cmd/searchd` — native service entry point.
- `native/internal/search` — native search application logic.
- `native/internal/preferences` — first-party Search preference state and behavior.
- `native/internal/webui` — native GoreeCloud Search web presentation.
- `native/internal/syncstate` — application-owned GoreeCloud Sync capability, signing, submission, retrieval, deletion, and validation boundaries.
- `native/docs/EXPERIENCE-REVAMP.md` — native experience, feature-preservation, preferences, accessibility, and migration direction.

The native Sync client currently advertises its application capability explicitly, requires exact negotiated schema conformance, bounds record and continuation identifiers, requires authenticated submission, and preserves Privacy Shield data minimization by keeping deletion tombstones free of application payload.

## Transitional implementation

The SearXNG-derived source remains a migration dependency, not the target GoreeCloud application architecture. It may continue to provide retained functionality while native equivalents are implemented and accepted.

Inherited user-facing capabilities must be inventoried as `retain`, `replace`, `improve`, or explicitly approved `retire` before the transitional runtime is removed. Upstream copyright, AGPL licensing, attribution, and relevant security/update obligations remain in force while inherited code is present.

## Mandatory platform gates

GoreeCloud Search must continuously conform to the current approved contracts for:

- **Glaze UI** — first-party responsive presentation, accessibility, interaction, appearance, and adaptive layout behavior.
- **Wardveil Security** — application security state, safe external-content behavior, diagnostics, and protection integration where applicable.
- **Privacy Shield** — data minimization, privacy-preserving defaults, query/history controls, and application-owned privacy boundaries.
- **Everkeep** — portability, recovery, backup/restore, migration, and continuity requirements where applicable.

Missing, materially outdated, or unvalidated mandatory integration keeps the application non-Stable.

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

These commands validate the native source locally; they do not perform or authorize a production deployment.

## Product direction

The native rebuild preserves useful search capabilities while replacing inherited product architecture with GoreeCloud-owned behavior. The target includes first-party homepage and results experiences, organized preferences, provider-adapter boundaries, privacy-preserving local controls, accessible keyboard/touch behavior, provider degradation handling, data portability, GoreeCloud Browser/OpenSearch integration, and controlled migration from the transitional runtime.

See `native/docs/EXPERIENCE-REVAMP.md` for the current native experience and migration contract.

## Repository records

- `FEATURES.md` — implemented, candidate, and planned capabilities.
- `BENEFITS.md` — user, administrative, privacy, resilience, and ownership benefits.
- `COMPETITIVE-OBJECTIVES.md` — product benchmarks and differentiators.
- `docs/goreecloud/READINESS.md` — transitional/runtime release-readiness boundaries where still applicable.
- `LICENSE` — repository licensing terms.

## License and upstream provenance

The inherited SearXNG source is licensed under the GNU Affero General Public License v3.0 or later. GoreeCloud preserves required source availability, copyright, attribution, and license obligations while progressively replacing application-defining inherited code with original native GoreeCloud implementation.
