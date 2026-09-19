# GoreeCloud Search — Current Features

**Lifecycle:** Development  
**Version:** `0.1.0.dev6`

This file records implemented behavior in the current development candidate. It does not claim release, deployment, production acceptance, or Stable qualification.

## Implemented in the current development candidate

- Native Python Search core, typed categories/source modes, and initial query operators.
- Deterministic provider eligibility and Index First/Federated/GoreeCloud Only/External Only/Offline Local planning.
- Explicit third-party disclosure signal plus optional per-query disclosure budget and plan evidence.
- Replaceable provider protocol and bounded execution with timeouts, concurrency limits, cancellation propagation, failure isolation, fallback behavior, and availability states.
- Versioned `goreecloud.search-index.v1` contract, transport-injected first-party Index adapter, capability/page-size negotiation, cursor pagination, and degraded/warning propagation.
- Conservative URL canonicalization, canonical-URL/content-hash deduplication, source agreement, and provider provenance.
- Content-policy hook engine after normalization and before ranking.
- `SafeSearchMode.OFF`, `MODERATE`, and `STRICT` intent.
- Fail-closed pre-execution rejection when Moderate/Strict is requested without a hook that declares SafeSearch enforcement.
- Built-in administrator domain blocklist/allowlist policy with subdomain matching.
- Policy decision provenance and allow/warn/block outcomes.
- Fail-closed handling for broken hooks and spoofed hook provenance.
- Deterministic Search-owned ranking with inspectable signals and human-readable explanations.
- No behavioral-history, advertising-payment, click-profile, or hidden provider-specific ranking boost.
- Development CLI, unit tests, and Python 3.11/3.12 CI.

## Not implemented

No authenticated live network provider, production SafeSearch/content classifier, Wardveil safety/threat feed, user-facing web UI, Browser integration, AI synthesis, persistent history, or production deployment exists yet.
