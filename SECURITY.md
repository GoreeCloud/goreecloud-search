# GoreeCloud Search — Security

## Lifecycle

GoreeCloud Search is in Development and is not production-qualified.

## Reporting security issues

Do not publish sensitive exploit details, credentials, tokens, private data, or active attack instructions in a public issue. Use an approved private GoreeCloud security-reporting path when available.

## Current security boundary

Version `0.1.0.dev6` ships no authenticated live network transport. Implemented local controls include input validation; deterministic source planning; source-mode privacy invariants; pre-execution third-party disclosure budgeting; execution only of plan-approved providers; bounded concurrency/timeouts; cancellation and failure isolation; provider-provenance validation; strict Search ↔ Index contract/pagination checks; result URL validation; rejection of undeclared providers; fail-closed content-policy hook provenance/error handling; pre-execution SafeSearch enforcement checks; and deterministic non-network ranking.

## Future requirements

Before network-capable or privileged operation is production-accepted, Search must verify applicable GoreeCloud Identity, Privacy Shield, Wardveil Security trust/risk/safety controls, secret separation, rate limiting/abuse resistance, stronger provider isolation and retry/rate/cost policy, dependency/artifact integrity, and security evidence.
