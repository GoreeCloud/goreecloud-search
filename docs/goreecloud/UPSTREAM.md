# Upstream Maintenance Record

## Initial baseline

GoreeCloud Search was forked from `searxng/searxng`.

Initial GoreeCloud development baseline:

`b2da6b90f2f8446557c91f67d6be5064ab785ecd`

Baseline date observed: 2026-08-16.

The baseline is identified by commit SHA because the upstream repository does not currently provide a conventional GitHub release/tag sequence suitable as the GoreeCloud source-of-truth version boundary.

## Upstream policy

GoreeCloud-specific changes should remain reviewable and as isolated as practical. Relevant upstream security fixes, provider-engine fixes, dependency fixes, and compatibility changes should be evaluated regularly.

Upstream commits must not be merged automatically into the GoreeCloud release line. Each upstream update requires review of:

- security impact;
- search-provider and engine behavior;
- dependency changes;
- configuration migrations;
- UI conflicts with Glaze UI;
- container and deployment impact;
- tests and build behavior;
- rollback requirements.

## Provenance requirements

The repository must preserve upstream licensing and copyright information. GoreeCloud modifications must not conceal the SearXNG origin of the fork.

When the baseline changes, this document should record the previous baseline, new baseline, reason for the update, validation performed, and any GoreeCloud conflicts or adaptations.
