# GoreeCloud Search Platform Status Boundary

## Purpose

The native Search platform-status contract provides a minimized, machine-readable view of the application’s current platform-integration evidence state without converting source presence, CI success, presentation wiring, or evidence transport into runtime or production acceptance.

The Development endpoint is:

- `GET /api/v1/platform/status`

It is also advertised through `GET /api/v1/status` as the `platform_status` capability and endpoint.

## Authority model

Search does not own Privacy Shield, Wardveil Security, or Everkeep truth. The endpoint identifies the authoritative producer/contract family for each system and exposes only Search’s bounded projection of supplied evidence.

Current Development authority references are:

- Privacy Shield — `GoreeCloud/goreecloud-privacy-shield`, `contracts/privacy-shield.platform-evidence.runtime-acceptance.json`;
- Wardveil Security — `GoreeCloud/goreecloud-wardveil-security`, `contracts/wardveil.status.schema.json`;
- Everkeep — `GoreeCloud/goreecloud-everkeep`, `contracts/continuity.status.schema.json` for the current direct continuity-status presentation path.

Glaze UI may present these states, but Search must not create, strengthen, merge, or silently reinterpret the underlying authority.

## Current Development state

The shipped native handler deliberately uses an unavailable runtime-evidence source. Therefore current application behavior remains fail-conservative even though the source contains deterministic projection and provenance-validation logic:

- Privacy Shield source integration is present, while application authorization evidence remains unavailable and state remains `unknown`.
- Wardveil source integration is present, while deployed authoritative application/service evidence is unavailable and the default state remains `unknown` with `positive_claim: false`.
- Everkeep has a native presentation/projection boundary, while deployed continuity evidence is unavailable and the default state remains `unknown`.
- Every platform system reports `production_accepted: false`.
- The aggregate snapshot reports `production_approved: false`.

This endpoint is an evidence-state presentation boundary, not a readiness or release-authority endpoint.

## Runtime evidence source

`native/internal/platformstate` defines an injectable `RuntimeEvidenceSource`. It exists so a future approved transport/runtime adapter can supply already-collected producer evidence without coupling the HTTP contract to a particular file, network endpoint, GoreeCloud Mesh route, credential mechanism, or test fixture.

The default handler does not configure such a producer. Its source is unavailable by design.

Adding a transport implementation later is a separate change. That implementation must follow the applicable producer, Identity, Mesh, privacy, security, freshness, and deployment contracts and must not use the existence of this interface as evidence that connectivity is accepted.

## Mesh evidence-envelope boundary

Search now models the GoreeCloud Mesh `goreecloud.evidence-envelope.v1` provenance/freshness/minimization contract as an ingress-validation primitive. Mesh remains coordination/transport authority only; a structurally valid envelope cannot strengthen the producer-domain outcome carried by that envelope.

The validator rejects envelopes with any of the following:

- wrong envelope version;
- missing or oversized identity/reference fields;
- producer system, repository, producer contract, or authority domain that does not exactly match the consumer expectation;
- producer revision that is not a canonical lowercase 40-character Git SHA;
- wrong subject kind or Search subject ID;
- wrong assertion family or evidence reference;
- future observation time;
- invalid observation/validity ordering;
- unsupported data class;
- malformed optional SHA-256 payload digest;
- user content; or
- secret material.

An expired envelope can remain structurally valid/auditable, but it cannot satisfy a current-state view. The domain projector is responsible for turning such evidence into `stale` rather than a positive state.

Search does not treat Mesh identity/provenance validation as proof that the producer-domain assertion is true. Wardveil, Privacy Shield, Everkeep, and GoreeCloud Identity retain their independent authorities.

## Privacy Shield projection

The current Privacy Shield runtime-acceptance contract is a contract for minimized Privacy Shield evidence delivery through GoreeCloud Mesh. Its own boundary explicitly says that Mesh transport validity never creates consent, purpose authorization, retention authorization, or deletion authority.

Accordingly:

- invalid contract identity is `unverified`;
- `production_acceptance: false` remains non-accepted transport evidence;
- even a future `production_acceptance: true` on this transport contract may only be surfaced as transport acceptance with Search application authorization still unverified;
- Privacy Shield transport evidence can never set Search `positive_claim` or `production_accepted` by itself.

Search requires a separate authoritative application-specific privacy/authorization boundary before any stronger privacy claim can be made.

## Wardveil Security projection

Wardveil status follows `wardveil.status.schema.json` semantics and is additionally bound by Search to scope ID `goreecloud-search` and the Wardveil Mesh producer profile.

A positive `protected` projection requires all of the following:

- contract version `0.1.0`;
- scope kind `application` or `service`;
- scope ID exactly `goreecloud-search`;
- authority system exactly `wardveil-security` plus a non-empty authority control;
- authoritative authority evidence;
- `state: protected`;
- `protected_by_wardveil: true`;
- `evidence.status: current`;
- a non-future observation time;
- a non-empty producer evidence reference;
- a `valid_until` value later than the evaluation time; and
- a structurally valid Mesh envelope bound to that exact Wardveil record.

For the Wardveil record, the envelope must identify producer `wardveil-security`, repository `GoreeCloud/goreecloud-wardveil-security`, producer contract `contracts/wardveil.status.schema.json`, authority domain `security`, the same Search scope kind/ID, assertion family `security-status`, the same opaque evidence reference, and identical observation/validity times.

Missing envelopes, wrong producer/repository/contract, invalid producer revisions, sensitive envelopes, mismatched evidence references/times, contradictory protected state/claim values, invalid scope, future observation, unknown evidence status, missing authority, or malformed provenance fail closed as unverified. A correctly bound but expired envelope/record remains auditable as stale and cannot remain protected.

A valid Wardveil protected record may set the bounded platform `positive_claim`, but it still does not set Search `production_accepted`; release and production acceptance remain broader gates.

## Everkeep projection

A Search-level continuity-ready projection requires a complete, fresh application evidence set rather than one optimistic Everkeep record.

The current direct continuity-status consumer projection requires exactly one usable Search-scoped record for each of:

- `backup_coverage`;
- `restore_capability`; and
- `recovery_freshness`.

Each required record must identify producer exactly `everkeep`, provide a unique non-empty record ID, target scope `goreecloud-search`, have a non-future observation, provide a verification method, and—when ready—have a future `fresh_until` plus a unique non-empty evidence reference. Duplicate required dimensions, duplicate record IDs, duplicate ready evidence references, wrong producer identity, or incomplete freshness/evidence fail closed as unverified.

Only a complete fresh set can surface `state: ready` with a bounded positive continuity claim. It still cannot set Search `production_accepted` or `production_approved` by itself.

### Everkeep Mesh contract distinction

Everkeep’s current Mesh producer profile lists `contracts/everkeep.evidence.schema.json`, recovery-point, protection-policy, and resource contracts as producer evidence contracts. It does **not** list the generic `contracts/continuity.status.schema.json` currently used by Search’s direct presentation path.

Search therefore must not wrap the current continuity-status records in a Mesh envelope and present that as profile-conformant Everkeep delivery. A future Mesh-backed Everkeep ingestion adapter must consume a producer-profile-authorized Everkeep contract—such as `everkeep.evidence.schema.json`—and map that producer-authoritative evidence into Search’s bounded presentation state without weakening restore/recovery requirements.

## Privacy and minimization

The HTTP response is designed to contain no:

- search query text;
- result content;
- user content;
- credentials or reusable secrets;
- authorization headers;
- raw platform/provider runtime errors; or
- raw producer evidence payloads.

Unrelated query-string input is ignored and must not be echoed into the response.

## Relationship to readiness

`GET /api/v1/readiness` remains scoped to `local_native_application` readiness. Platform-status information does not change that scope and does not make local readiness equivalent to production readiness.

A future authoritative runtime adapter may strengthen a bounded platform state only when the applicable producer contract, provenance, freshness, authority, scope, and Search-side projection rules are satisfied. Missing, malformed, stale, contradictory, unverified, or unavailable evidence must not be upgraded by Search.

## Positive-claim requirements

Search must not claim Privacy Shield production acceptance, Wardveil protection, Everkeep recovery readiness, or equivalent positive platform state merely because:

- platform source files exist;
- a platform repository is healthy;
- Search CI is green;
- a UI section displays the platform name;
- a transport such as GoreeCloud Mesh can carry evidence;
- a local adapter can parse a schema;
- an envelope is structurally valid; or
- deterministic tests can inject a synthetic accepted record.

Any positive state remains limited to what current authoritative application/runtime evidence actually proves. Search release, production, and Stable claims remain separate exact-candidate gates.

## Lifecycle boundary

This Development projection implementation does not establish:

- Privacy Shield runtime authorization or production acceptance;
- deployed Privacy Shield/Identity/Mesh evidence transport;
- deployed Wardveil Security protected state;
- deployed Everkeep backup, restore, rollback, or continuity readiness;
- target-host acceptance;
- production cutover authorization; or
- Stable qualification.

Those remain separate evidence-backed release gates.