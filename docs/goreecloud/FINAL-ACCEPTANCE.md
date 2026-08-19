# GoreeCloud Search Final-Candidate Acceptance Evidence

## Purpose

This procedure closes the first-Stable evidence gap between deterministic source/runtime checks and the final reviews that must be performed on real devices and the actual GoreeCloud Browser runtime.

The final manifest is an integrity record, not an authorization record. It binds six completed evidence artifacts to one exact immutable GoreeCloud Search candidate and always keeps production cutover unauthorized.

## Evidence boundary

Final-candidate acceptance requires **six independently completed evidence artifacts** for one exact GoreeCloud Search candidate:

1. immutable release evidence;
2. target-runtime identity evidence;
3. completed recovery/monitoring/rollback evidence;
4. runtime-bound real-provider acceptance evidence;
5. completed visual/device review evidence; and
6. completed actual GoreeCloud Browser runtime evidence.

The final schema-version 2 manifest hashes and binds all six JSON artifacts. The visual and Browser JSON artifacts also contain the immutable digest of the retained underlying review/runtime artifact that was actually inspected. A free-text issue URL or evidence-reference string by itself is not sufficient final-candidate evidence.

All six inputs must refer to the same exact Search source revision and immutable GHCR image as the release candidate. Visual or Browser evidence from a development preview, older candidate, mutable tag, or previous production image is rejected.

No evidence artifact may set `production_cutover_authorized` to true.

## Real-provider evidence

Run the representative suite on the target host against the intentionally staged final candidate, **not** against the still-authoritative production hostname before cutover. Candidate source and image values must come from release evidence rather than a mutable tag or operator assumption.

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
- the named container is actually published on that loopback port;
- the container is running and healthy;
- its configured image reference is the exact immutable candidate image;
- its running image ID matches that digest image;
- the candidate OCI title, source, revision, version, and license metadata are valid; and
- the exact runtime identity remains unchanged before and after the provider requests.

The generated artifact records only sanitized runtime binding and category-level acceptance information. It intentionally does **not** retain query text, response content, cookies, credentials, provider tokens, or reusable secrets.

The first-Stable mandatory provider categories are:

- General
- Images
- Videos
- News
- Files

IT and Science remain diagnostics; neither may replace a required category.

## Visual/device evidence artifact

Create `visual-evidence.json` only after the final candidate has been reviewed. It must use this shape:

```json
{
  "schema_version": 1,
  "product": "GoreeCloud Search",
  "generated_at": "2026-08-19T20:00:00Z",
  "candidate": {
    "source_revision": "<40-character-candidate-source-sha>",
    "image": "ghcr.io/goreecloud/goreecloud-search@sha256:<candidate-digest>"
  },
  "glaze_ui_version": "1.1.0",
  "review_artifact": {
    "reference": "<retained-screenshot-or-review-artifact-locator>",
    "digest": "sha256:<immutable-artifact-digest>"
  },
  "reviews": {
    "compact_light": {"passed": true, "evidence_reference": "<reference>"},
    "compact_dark": {"passed": true, "evidence_reference": "<reference>"},
    "expanded_light": {"passed": true, "evidence_reference": "<reference>"},
    "expanded_dark": {"passed": true, "evidence_reference": "<reference>"}
  },
  "physical_android_preferences_review": {
    "passed": true,
    "evidence_reference": "<physical-device-session-reference>"
  },
  "desktop_regression_review": {
    "passed": true,
    "evidence_reference": "<desktop-session-reference>"
  },
  "persisted_theme_preference_review": {
    "passed": true,
    "evidence_reference": "<theme-persistence-session-reference>"
  },
  "scope": {
    "exact_candidate_visual_artifact_verified": true,
    "manual_visual_acceptance_verified": true,
    "production_cutover_authorized": false,
    "statement": "<sanitized review statement>"
  }
}
```

The retained review artifact digest must be the immutable digest of the actual artifact that was reviewed. For a GitHub Actions artifact, use the returned `sha256:<digest>` value. The four required visual cases are not substitutes for physical Android Preferences review, desktop regression review, or persisted theme-preference behavior; all seven review records must be complete.

Do not mark a visual review complete solely because a screenshot workflow generated files. An operator must actually review the exact candidate artifact and record the physical/device/runtime evidence truthfully.

## GoreeCloud Browser runtime evidence artifact

Create `browser-evidence.json` only after acceptance has run on the actual GoreeCloud Browser runtime against the exact Search candidate. It must use this shape:

```json
{
  "schema_version": 1,
  "product": "GoreeCloud Search",
  "generated_at": "2026-08-19T20:05:00Z",
  "search_candidate": {
    "source_revision": "<40-character-search-candidate-sha>",
    "image": "ghcr.io/goreecloud/goreecloud-search@sha256:<candidate-digest>"
  },
  "browser_source_revision": "<40-character-goreecloud-browser-sha>",
  "runtime_artifact": {
    "reference": "<retained-browser-runtime-artifact-locator>",
    "digest": "sha256:<immutable-artifact-digest>"
  },
  "behaviors": {
    "search_only_default_provider": true,
    "address_bar_routed_through_search": true,
    "new_tab_routed_through_search": true,
    "dedicated_search_field_routed_through_search": true,
    "no_external_browser_fallback": true,
    "search_unavailability_state_verified": true,
    "recovery_after_search_reachability_verified": true
  },
  "scope": {
    "actual_browser_runtime_verified": true,
    "search_candidate_runtime_verified": true,
    "production_cutover_authorized": false,
    "statement": "<sanitized runtime-acceptance statement>"
  }
}
```

A source-code policy claim is insufficient. The retained artifact must represent the actual Browser runtime session and must have an immutable digest. Acceptance must prove the Search-only/default-provider policy, address-bar/new-tab/dedicated-field routing, no silent external fallback, the Search-unavailable state, and recovery when Search becomes reachable again.

## Assemble the final manifest

Only after **all six** artifacts are complete, assemble the final manifest:

```bash
python goreecloud/final_acceptance_evidence.py assemble \
  --release-evidence release-evidence.json \
  --target-runtime-evidence target-runtime-evidence.json \
  --recovery-evidence recovery-evidence.json \
  --provider-evidence provider-evidence.json \
  --visual-evidence visual-evidence.json \
  --browser-evidence browser-evidence.json \
  --output final-acceptance-evidence.json
```

Assembly fails closed if any input is incomplete, unsafe, refers to another candidate, lacks its required immutable review/runtime artifact digest, or violates a mandatory acceptance requirement.

The output is schema version 2. Its `artifact_bindings` contains the SHA-256 hash of all six supplied JSON artifacts. Visual and Browser summaries are copied from those verified bound inputs rather than manually edited into the final manifest.

## Validate the completed final manifest

Validate the manifest against the **same exact six files**:

```bash
python goreecloud/final_acceptance_evidence.py validate \
  --release-evidence release-evidence.json \
  --target-runtime-evidence target-runtime-evidence.json \
  --recovery-evidence recovery-evidence.json \
  --provider-evidence provider-evidence.json \
  --visual-evidence visual-evidence.json \
  --browser-evidence browser-evidence.json \
  --evidence final-acceptance-evidence.json
```

Validation fails closed if:

- any supplied artifact refers to another Search candidate;
- any of the six JSON artifacts changed after the manifest was assembled;
- the final manifest is an older unbound schema version;
- provider evidence lacks verified before/after staged runtime identity binding;
- any mandatory provider category lacks a passing result;
- recovery evidence does not prove application-level restore, monitoring/alerting, and rollback evidence;
- the visual review artifact lacks an immutable artifact digest;
- any Compact/Expanded light/dark review is incomplete;
- physical Android Preferences, desktop regression, or persisted-theme review is incomplete;
- the Browser runtime artifact lacks an immutable artifact digest;
- the Browser source identity or exact Search candidate identity is missing/mismatched;
- any required Browser/Search runtime behavior is unverified;
- the final visual or Browser summary differs from its bound evidence artifact;
- the final record contains prohibited sensitive/unnecessary field names; or
- `production_cutover_authorized` is anything other than false.

## Stable decision boundary

A passing schema-version 2 final-candidate manifest means the defined first-Stable evidence set is complete, cryptographically bound, and internally consistent for one exact candidate.

It does **not** independently authorize production cutover, runtime-name retirement, rollback-material deletion, or Stable publication. Those actions still require the explicit GoreeCloud release decision and the applicable `STABLE-CUTOVER.md` procedure.
