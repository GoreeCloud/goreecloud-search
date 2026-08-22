# GoreeCloud Search First-Stable Integrity Audit

The authoritative master-side integrity entrypoint for frozen first-Stable candidate #07 is:

```bash
python goreecloud/first_stable_candidate_07_integrity_audit.py \
  --release-evidence release-evidence.json \
  --target-runtime-evidence target-runtime-evidence.json \
  --recovery-evidence recovery-evidence.json \
  --rollback-baseline goreecloud/release_baseline.json \
  --provider-evidence provider-evidence.json \
  --visual-evidence visual-evidence.json \
  --browser-evidence browser-evidence.json \
  --final-evidence final-evidence.json
```

The command composes the established candidate-#07 audit, rollback-baseline byte pinning,
release-evidence verification, target-runtime compatibility checks, recovery binding, deep visual
and Browser review checks, and provider-result integrity validation. The rollback baseline is
supporting provenance and is not a seventh final companion artifact; the final manifest still
binds exactly six companion JSON artifacts.

The provider integrity layer validates the authentic result structure emitted by frozen
`provider_acceptance.py`. A successful artifact must contain the exact representative suite in
its frozen order, a positive `minimum_results` threshold, successful HTTP 200 responses,
GoreeCloud Search product identity, zero result exit codes, result-card counts meeting the
threshold, and a successful full diagnostic suite. A `passed=true` flag is never accepted when
those underlying fields contradict it.

The target-runtime artifact's optional `providers` field does not satisfy the independent
real-provider evidence gate. A separate candidate-bound `provider-evidence.json` is always
required by the integrity command and by the schema-version 2 final acceptance contract.

## Operator Readiness Report

Use the readiness reporter before final-manifest assembly or release-governance review:

```bash
python goreecloud/first_stable_candidate_07_readiness.py \
  --release-evidence release-evidence.json \
  --target-runtime-evidence target-runtime-evidence.json \
  --recovery-evidence recovery-evidence.json \
  --rollback-baseline goreecloud/release_baseline.json \
  --provider-evidence provider-evidence.json \
  --visual-evidence visual-evidence.json \
  --browser-evidence browser-evidence.json \
  --final-evidence final-evidence.json
```

`--final-evidence` is optional until the six validated companion artifacts are ready for final
manifest assembly. Add `--json` for a machine-readable report.

The reporter classifies each required input as `missing`, `invalid`, `pending_dependency`, or
`valid`. It also performs the complete companion cross-binding audit after all seven required
inputs validate. The seven required operator inputs are the six final companion artifacts plus
the supporting rollback baseline. The rollback baseline remains provenance support and never
becomes a seventh final companion artifact.

The overall readiness states are:

- `blocked` — one or more required inputs are missing, invalid, dependency-blocked, or fail
  cross-binding; process exit code `2`.
- `ready_for_final_manifest` — all six companion artifacts and rollback provenance validate and
  cross-bind, but no final manifest was supplied; process exit code `3` so automation cannot
  confuse manifest-assembly readiness with completed final acceptance.
- `ready_for_governance_review` — the schema-version 2 final manifest is also supplied and
  passes the complete authoritative integrity audit; process exit code `0`.

Even `ready_for_governance_review` is evidence readiness only. It is not production-cutover
approval and is not Stable-release authorization. The report always emits
`production_cutover_authorized=false` and `stable_promotion_authorized=false`.

This tooling validates evidence only. It does not stage the candidate, perform provider
requests, execute recovery, complete device or Browser review, create a final manifest, modify
production, authorize production cutover, or promote GoreeCloud Search to Stable. Missing real
evidence must be created by the applicable target-environment or physical/manual acceptance
procedure; it must never be fabricated to satisfy either the audit or the readiness report.
