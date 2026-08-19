# GoreeCloud Search Final-Candidate Acceptance Evidence

## Purpose

This procedure closes the first-Stable evidence gap between deterministic source/runtime checks and the manual final-candidate reviews that must be performed on real devices and the actual GoreeCloud Browser runtime.

It creates one candidate-bound evidence set without allowing any evidence artifact to authorize production cutover by itself.

## Evidence boundary

Final-candidate acceptance uses five evidence inputs for one exact GoreeCloud Search candidate:

1. immutable release evidence;
2. target-runtime identity evidence;
3. completed recovery/monitoring/rollback evidence;
4. runtime-bound real-provider acceptance evidence;
5. the final manual Glaze UI and GoreeCloud Browser runtime acceptance record.

The final record hashes and binds the first four artifacts. They must all refer to the same exact source revision and immutable GoreeCloud Search image as the release candidate.

Provider evidence is not candidate-bound merely because an operator supplies a candidate SHA and image digest on the command line. The provider requests themselves must execute against the same identity-verified staged candidate runtime.

## Real-provider evidence

Run the representative suite on the target host against the intentionally staged final candidate, **not** against the still-authoritative production hostname before cutover. The candidate source and image values must come from the release evidence, not from a mutable tag or an operator assumption.

The preferred pre-cutover staging topology is the loopback-only instance defined by `TARGET-ACCEPTANCE.md`:

```bash
python goreecloud/provider_acceptance.py \
  --base-url http://127.0.0.1:8888 \
  --container goreecloud-search \
  --suite \
  --expected-source '<40-character-candidate-source-sha>' \
  --expected-image 'ghcr.io/goreecloud/goreecloud-search@sha256:<candidate-digest>' \
  --evidence-json provider-evidence.json
```

When `--evidence-json` is requested, the provider runner fails closed unless it can verify that:

- the base URL is loopback-only;
- the named Docker container is actually published on that loopback port;
- the container is running and healthy;
- its configured image reference is the exact immutable candidate image;
- its running image ID matches that digest image;
- the candidate OCI title, source, revision, version, and license metadata are valid; and
- the exact runtime identity remains unchanged before and after the provider requests.

The generated artifact records the sanitized runtime binding plus category, HTTP status, GoreeCloud Search identity, result-card count, engine-message count, pass/fail state, and the exact candidate identity. It intentionally does **not** retain query text, response content, cookies, credentials, provider tokens, or reusable secrets.

`https://search.goreecloud.com` remains the production path until a separately authorized cutover. It may be used for post-cutover production rechecks, but it must not be used to create pre-cutover final-candidate provider evidence while it still routes to the previous known-good runtime.

The first-Stable required provider categories are:

- General
- Images
- Videos
- News
- Files

IT and Science remain useful diagnostics. A diagnostic provider failure must be investigated and classified, but it does not silently remove any of the five mandatory first-Stable categories.

## Create the final acceptance template

After release, target-runtime, recovery, and provider evidence all refer to the same candidate, create the final record:

```bash
python goreecloud/final_acceptance_evidence.py template \
  --release-evidence release-evidence.json \
  --target-runtime-evidence target-runtime-evidence.json \
  --recovery-evidence recovery-evidence.json \
  --provider-evidence provider-evidence.json \
  --output final-acceptance-evidence.json
```

The final-evidence validator rejects provider artifacts that lack the staged runtime binding, that do not prove runtime identity before and after the provider requests, or that refer to a different source/image.

The generated template is intentionally incomplete. Its manual fields default to false and must not be changed until the corresponding review is actually performed against that same final candidate.

## Required Glaze UI 1.1 review

The completed final record must contain evidence references for all of these exact final-candidate reviews:

- Compact light appearance;
- Compact dark appearance;
- Expanded light appearance;
- Expanded dark appearance;
- physical Android/mobile Preferences review;
- desktop regression/final visual review.

An evidence reference should identify the retained review record, screenshot set, test-session record, issue, or other approved artifact. Do not store credentials, private keys, tokens, cookies, or secret values in the JSON record.

The review must confirm the integrated candidate rather than a development branch preview or an older production image.

## Required GoreeCloud Browser runtime review

Record the exact GoreeCloud Browser source revision used for acceptance and an evidence reference for the runtime session. The review must prove:

- GoreeCloud Search is the only/default browser search provider;
- address-bar searches route through GoreeCloud Search;
- new-tab searches route through GoreeCloud Search;
- the dedicated browser search field routes through GoreeCloud Search;
- no external browser-level fallback is used when Search is unavailable;
- Search unavailability is shown as a GoreeCloud service/connectivity state;
- Browser recovers correctly when GoreeCloud Search becomes reachable again.

This is runtime acceptance. A source-code configuration claim by itself is insufficient.

## Validate the completed final record

After the manual fields are completed truthfully, validate the record against the exact artifacts it binds:

```bash
python goreecloud/final_acceptance_evidence.py validate \
  --release-evidence release-evidence.json \
  --target-runtime-evidence target-runtime-evidence.json \
  --recovery-evidence recovery-evidence.json \
  --provider-evidence provider-evidence.json \
  --evidence final-acceptance-evidence.json
```

Validation fails closed if:

- any bound artifact refers to another candidate;
- any bound artifact has changed since the final template was created;
- provider evidence lacks verified before/after staged runtime identity binding;
- provider evidence does not identify the exact immutable candidate image and source revision;
- any required provider category lacks a passing result;
- the completed recovery artifact does not prove restore, monitoring/alerting, and rollback evidence;
- any of the four required Glaze UI 1.1 appearance reviews is incomplete;
- Android Preferences or desktop regression review is incomplete;
- Browser source identity or runtime evidence reference is missing;
- any required Browser/Search integration behavior is unverified;
- the final record contains prohibited sensitive/unnecessary field names;
- `production_cutover_authorized` is anything other than false.

## Stable decision boundary

A passing final-candidate evidence artifact means the defined first-Stable acceptance evidence set is complete and internally consistent for one exact candidate.

It does **not** independently authorize production cutover, runtime-name retirement, rollback-material deletion, or Stable release publication. Those actions still require the explicit GoreeCloud release decision and the applicable `STABLE-CUTOVER.md` procedure.
