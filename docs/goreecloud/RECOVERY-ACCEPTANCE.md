# GoreeCloud Search Recovery Acceptance

## Purpose

This procedure defines the application-level backup, restore, monitoring, and rollback evidence required before the first Stable GoreeCloud Search cutover can be approved.

It does **not** authorize a production cutover and it does not require modifying the active Search route merely to create evidence. The preferred restore exercise is isolated from the current production service.

## Recovery scope

The GoreeCloud Search deployment is intentionally small. The authoritative application-level material that must remain recoverable includes:

- the reviewed Docker Compose deployment definition;
- the approved GoreeCloud Search `settings.yml`;
- the protected runtime-configuration recovery path, without copying reusable secret values into Git, CI artifacts, or ordinary documentation;
- the current Search-specific Caddy route/backend configuration needed to restore service publication;
- the exact immutable candidate and known-good rollback image identities.

The `/var/cache/searxng` mount in the reviewed production baseline is application cache. It is classified for this acceptance contract as **rebuildable and non-authoritative**. A cache loss must not prevent Search recovery and the cache is not treated as irreplaceable user data.

A whole-VPS provider backup remains useful disaster-recovery protection, but it does not replace this application-level restore test.

## Required companion evidence

A completed recovery artifact must be bound to all three of the following reviewed artifacts:

1. `release-evidence.json` from the exact candidate-image workflow;
2. `target-runtime-evidence.json` proving the target runtime under test is the exact candidate digest/source revision;
3. `goreecloud/release_baseline.json`, which records the known-good production rollback image.

The recovery validator calculates SHA-256 hashes for the supplied companion artifacts. A recovery record copied from another candidate or edited to point at unrelated evidence must fail validation.

## Create the candidate-bound template

After the exact candidate-image and target-runtime evidence artifacts exist, create a recovery template:

```bash
python goreecloud/recovery_evidence.py template \
  --release-evidence release-evidence.json \
  --target-runtime-evidence target-runtime-evidence.json \
  --rollback-baseline goreecloud/release_baseline.json \
  --output recovery-evidence.json
```

The generated file is intentionally incomplete. All acceptance booleans begin false and must remain false until the corresponding target-host work is actually performed and recorded.

The template contains no passwords, tokens, cookies, reusable credentials, protected environment values, or query contents. Do not add them.

## Backup evidence

Before the restore exercise, verify the approved backup system contains the current Search recovery scope. Record only a non-secret backup-system name, an opaque snapshot/reference identifier, and a timezone-aware capture timestamp.

The evidence must confirm that the backup/recovery scope covers:

- the stack definition;
- Search settings;
- the protected runtime-configuration **recovery path**;
- the Search-specific Caddy route/backend configuration.

The evidence must not embed the actual protected runtime value.

## Isolated restore exercise

Perform the representative restore into an isolated target such as a dedicated acceptance directory and loopback-only staging container. Do not overwrite the production stack or production route merely to satisfy the restore gate.

The restore exercise must prove that:

- the stack definition can be restored;
- `settings.yml` can be restored;
- the protected runtime configuration can be recovered through the approved secure source;
- a copy of the Search Caddy route/backend configuration can be restored for validation;
- `docker compose config` or equivalent configuration validation passes;
- the restored application can be recreated;
- the recreated runtime becomes healthy;
- GoreeCloud Search product identity is intact;
- the candidate-bound target acceptance harness passes against the restored runtime.

Use a separate sanitized target-runtime evidence artifact for the restored runtime when practical. Do not record secrets in the recovery manifest.

## Rollback evidence

Rollback evidence must preserve the currently known-good image and enough previous runtime/route configuration to return service to the verified prior state if the candidate fails.

The validator accepts one of two documented rollback modes:

- `production-route-rehearsal` — an explicitly controlled production-route rollback was actually tested and recorded; or
- `equivalent-verified-evidence` — the known-good image was exercised in isolation, previous runtime configuration and Caddy route material remain preserved, and an equivalent reviewed rollback procedure/evidence set exists without changing the active production route solely for rehearsal.

Both modes still require:

- the known-good immutable image to be available;
- previous runtime configuration to remain preserved;
- the previous Search route/backend configuration to remain preserved;
- the rollback procedure to be documented;
- isolated rollback-image rehearsal to have passed.

The choice of equivalent evidence must be evaluated as part of final production acceptance; the validator does not independently authorize it.

## Monitoring evidence

Before cutover approval, recovery evidence must also record that:

- the availability monitor identifies the service as **GoreeCloud Search** at the appropriate point in the migration;
- the approved alert-delivery path was exercised successfully.

Do not rename or repoint the authoritative production monitor early merely to make the evidence pass. Monitoring changes belong to the controlled cutover sequence.

## Validate completed evidence

After the actual restore, monitoring, and rollback-evidence work is complete, validate the filled recovery artifact against the exact companion artifacts:

```bash
python goreecloud/recovery_evidence.py validate \
  --evidence recovery-evidence.json \
  --release-evidence release-evidence.json \
  --target-runtime-evidence target-runtime-evidence.json \
  --rollback-baseline goreecloud/release_baseline.json
```

Validation fails closed when required restore, monitoring, rollback, candidate-binding, or artifact-hash evidence is missing or inconsistent.

A successful validation still prints that production cutover is **not** authorized by the artifact. Stable approval remains a separate decision requiring every release gate, including physical Glaze UI review, representative real-provider checks, Browser runtime integration, private-network behavior, monitoring, recovery, and final target acceptance.

## Evidence handling

Recovery evidence is an operational acceptance record, not a secret store. Keep reusable secrets and private credentials in their approved protected systems. The validator rejects common secret-bearing JSON field names so sensitive values are less likely to be copied into the artifact accidentally.

Retain the final validated recovery evidence with the exact release-candidate evidence set and the related change-log record for the accepted release.
