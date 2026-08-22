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
release-evidence verification, immutable publication-provenance checks, target-runtime
compatibility checks, recovery binding, deep visual and Browser review checks, and provider-
result integrity validation. The rollback baseline is supporting provenance and is not a seventh
final companion artifact; the final manifest still binds exactly six companion JSON artifacts.

## Frozen Publication Provenance

Candidate #07 release evidence must be the exact `release-evidence.json` retained in GitHub
Actions artifact `goreecloud-search-candidate-b355aafe769176acebfc938b15a6f7b5b9a2db87`,
artifact ID `9382173615`, GitHub artifact digest
`sha256:1decf1341d0b876c3f3c73ed2519f6eadf952981481c5cb7934a4c2cc8ee09f0`.
The exact published `release-evidence.json` bytes have SHA-256
`b0873cb4fbf244a6bcef2024add86b6579e7557a7e77f30a29f0096b2adf6752`. A semantically
similar, regenerated, reformatted, or edited release file is rejected because it no longer proves
that the operator is auditing the exact published candidate artifact.

Completed visual evidence must reference GitHub Actions artifact
`goreecloud-search-candidate-07-visual-evidence`, artifact ID `9382309578`, with immutable
artifact digest
`sha256:0b9fe7a184a6e15f01e53063ba27bec39aa477d2b08a0ca3c769e0c470451be9`.
The workflow artifact's original `visual-evidence.json` manifest has SHA-256
`e6079dcd36f5f4e8139892b98e8d804a48318f493701dd2ab4cda7b368c00979`; that original
manifest remains intentionally incomplete for physical Android, desktop-runtime, persisted-theme,
and final manual acceptance. The completed schema-version 1 visual companion artifact may add
those truthful manual review results, but its `review_artifact.reference` and
`review_artifact.digest` must continue to identify the exact frozen underlying visual artifact.

These provenance checks do not add another final companion artifact or another readiness input.
They strengthen the existing release and visual companion artifacts while preserving the six-
artifact final contract and the seven operator inputs used by the readiness reporter.

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
`valid`. It also performs the complete companion cross-binding and publication-provenance audit
after all seven required inputs validate. The seven required operator inputs are the six final
companion artifacts plus the supporting rollback baseline. The rollback baseline remains
provenance support and never becomes a seventh final companion artifact.

The overall readiness states are:

- `blocked` — one or more required inputs are missing, invalid, dependency-blocked, or fail
  cross-binding/provenance; process exit code `2`.
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
