# GoreeCloud Search Platform Status Boundary

## Purpose

The native Search platform-status contract provides a minimized, machine-readable view of the application’s current platform-integration evidence state without converting source presence, CI success, or presentation wiring into runtime or production acceptance.

The Development endpoint is:

- `GET /api/v1/platform/status`

It is also advertised through `GET /api/v1/status` as the `platform_status` capability and endpoint.

## Authority model

Search does not own Privacy Shield, Wardveil Security, or Everkeep truth. The endpoint identifies the authoritative producer/contract family for each system and exposes only Search’s bounded evidence state.

Current Development authority references are:

- Privacy Shield — `GoreeCloud/goreecloud-privacy-shield`, `contracts/privacy-shield.platform-evidence.runtime-acceptance.json`;
- Wardveil Security — `GoreeCloud/goreecloud-wardveil-security`, `contracts/wardveil.status.schema.json`;
- Everkeep — `GoreeCloud/goreecloud-everkeep`, `contracts/continuity.status.schema.json`.

Glaze UI may present these states, but Search must not create, strengthen, merge, or silently reinterpret the underlying authority.

## Current Development state

The source snapshot deliberately fails conservative:

- Privacy Shield source integration is present, while application runtime evidence is unavailable and state remains `unknown`.
- Wardveil source integration is present, while runtime evidence is unverified, state remains `unknown`, and `positive_claim` remains false.
- Everkeep currently has only a presentation boundary in native Search source; application runtime evidence is unavailable and state remains `unknown`.
- Every platform system reports `production_accepted: false`.
- The aggregate snapshot reports `production_approved: false`.

This endpoint is therefore an evidence-state presentation boundary, not a readiness or release-authority endpoint.

## Privacy and minimization

The response is designed to contain no:

- search query text;
- result content;
- user content;
- credentials or reusable secrets;
- authorization headers;
- raw platform/provider runtime errors.

Unrelated query-string input is ignored and must not be echoed into the response.

## Relationship to readiness

`GET /api/v1/readiness` remains scoped to `local_native_application` readiness. Platform-status information does not change that scope and does not make local readiness equivalent to production readiness.

A future authoritative runtime adapter may strengthen a platform state only when the applicable producer contract, freshness, authority, scope, and acceptance rules are satisfied. Missing, malformed, stale, unverified, or unavailable evidence must not be upgraded by Search.

## Positive-claim requirements

Search must not claim Privacy Shield production acceptance, Wardveil protection, Everkeep recovery readiness, or equivalent positive platform state merely because:

- platform source files exist;
- a platform repository is healthy;
- Search CI is green;
- a UI section displays the platform name;
- a transport such as GoreeCloud Mesh can carry evidence;
- a local adapter can parse a schema.

Positive claims require current authoritative application/runtime evidence under the relevant platform contract and remain subject to the exact Search release/deployment boundary.

## Lifecycle boundary

This Development contract does not establish:

- Privacy Shield runtime authorization or production acceptance;
- Wardveil Security protected state;
- Everkeep backup, restore, rollback, or continuity readiness;
- target-host acceptance;
- production cutover authorization; or
- Stable qualification.

Those remain separate evidence-backed release gates.