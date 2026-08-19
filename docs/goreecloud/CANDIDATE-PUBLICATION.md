# GoreeCloud Search Candidate Publication Control

## Purpose

This procedure allows the reviewed stabilization line to publish and rehearse an immutable first-Stable candidate **without** merging the draft stabilization pull request into `agent/production-acceptance` merely to obtain release evidence.

It does not weaken the Stable gate. Candidate publication remains an evidence-producing staging action only.

## Explicit request marker

The controlled request is `goreecloud/candidate_request.json`.

A valid request must:

- identify `GoreeCloud Search`;
- use schema version 1;
- use request type `publish-and-rehearse-final-candidate`;
- identify the exact reviewed stabilization base revision;
- include a timezone-aware request timestamp and unique request ID;
- keep `production_cutover_authorized` false;
- keep `stable_release_authorized` false;
- keep `target_host_change_authorized` false;
- contain no credentials, tokens, cookies, environment values, or other reusable secrets.

`goreecloud/candidate_request.py` validates this contract and rejects unknown fields so the marker cannot silently expand into a broader deployment authorization document.

## Reviewed-base binding

For a candidate request pull request, release CI verifies that `reviewed_base_revision` equals the exact base SHA of the pull request.

After the request is squash-merged into `agent/mobile-stabilization`, the stabilization candidate workflow verifies that the new candidate commit's direct parent still equals `reviewed_base_revision`.

This fails closed if unrelated commits are inserted between the reviewed stabilization base and the candidate-request commit.

The resulting merge commit is the exact candidate source revision. It intentionally includes the candidate-publication control change itself.

## Trigger behavior

`.github/workflows/goreecloud-stabilization-candidate.yml` runs only when `goreecloud/candidate_request.json` changes on `agent/mobile-stabilization`.

Ordinary stabilization commits do not publish candidate images.

The trigger first validates the exact request and parent binding, then calls the reusable `.github/workflows/goreecloud-candidate-image.yml` workflow. The existing direct candidate-image triggers remain available for `agent/production-acceptance` and explicit manual workflow dispatch.

## Candidate evidence

The reusable candidate-image workflow must continue to:

1. validate the source-controlled immutable rollback baseline;
2. build the exact candidate revision;
3. verify GoreeCloud Search OCI identity, source, revision, version, and AGPL license metadata;
4. publish the candidate to GHCR under the exact source revision;
5. resolve and record the immutable image digest;
6. pull the candidate back by immutable digest;
7. pull the known-good rollback image by immutable digest;
8. rehearse both images separately in isolated loopback-only containers;
9. generate `candidate-image.txt`, `release-evidence.json`, `release-baseline.json`, and `SHA256SUMS`;
10. retain the evidence as a GitHub Actions artifact.

The generated release evidence must continue to state that production cutover is not authorized.

## What this does not authorize

A successful candidate-request workflow does **not** authorize:

- deployment to `goreecloud-vps-01`;
- replacement of the current production Search runtime;
- Caddy, DNS, AdGuard Home, NetBird, firewall, monitoring, backup, or persistent-data changes;
- Stable promotion;
- compatibility-name retirement;
- rollback-material deletion.

Those remain governed by `TARGET-ACCEPTANCE.md`, `RECOVERY-ACCEPTANCE.md`, `FINAL-ACCEPTANCE.md`, and `STABLE-CUTOVER.md`.

## After publication

Use the exact source revision and immutable digest from `release-evidence.json` for every subsequent acceptance artifact.

The next sequence is:

1. stage the candidate separately from production;
2. collect candidate-bound target-runtime evidence;
3. run the real representative provider suite;
4. complete monitoring, alert-delivery, isolated restore, and rollback evidence;
5. complete physical Android/mobile and desktop Glaze UI 1.1 review;
6. complete actual GoreeCloud Browser runtime integration acceptance;
7. validate the final-candidate evidence manifest;
8. make the separate explicit Stable/cutover decision.

If any mandatory gate fails, retain the current known-good production runtime and rollback material.
