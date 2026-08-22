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

This command validates evidence only. It does not stage the candidate, perform provider
requests, execute recovery, complete device or Browser review, modify production, authorize
production cutover, or promote GoreeCloud Search to Stable. Missing real evidence must be
created by the applicable target-environment or physical/manual acceptance procedure; it must
never be fabricated to satisfy the audit.
