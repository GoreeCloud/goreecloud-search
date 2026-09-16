# GoreeCloud Search — Security

## Lifecycle

GoreeCloud Search is currently in Development and is not production-qualified.

## Reporting security issues

Do not publish sensitive exploit details, credentials, tokens, private data, or active attack instructions in a public issue.

Use GitHub private vulnerability reporting when it is enabled for this repository, or another approved private GoreeCloud security-reporting channel.

## Current security boundary

Version `0.1.0.dev7` ships a provider-execution framework but no authenticated live network transport. Implemented controls include input validation, deterministic source planning, fail-closed source-mode privacy rules, pre-execution third-party disclosure budgeting, execution only of plan-approved providers, bounded concurrency/per-provider timeouts, cancellation propagation, failure isolation, provider-provenance validation, strict Search ↔ Index contract/pagination validation, conservative result-URL validation, rejection of candidates from undeclared providers, fail-closed content-policy hook handling, pre-execution SafeSearch enforcement checks, deterministic ranking, and local Lens handling.

A requested Lens must resolve before provider execution. Unknown Lenses fail before provider calls. The selected Lens is removed from the provider-facing query, and Lens rules operate only on normalized Search results after content policy. Lens exclusions and scoring changes remain explicit evidence rather than hidden state.

## Future requirements

Before network-capable or privileged service operation is production-accepted, Search must implement and verify applicable:

- GoreeCloud Identity authentication and authorization.
- Privacy Shield permitted-use enforcement.
- Wardveil Security trust and protection controls, including verified safety/threat classification where applicable.
- Secret separation and secure configuration.
- Rate limiting and abuse resistance.
- Stronger provider sandboxing/isolation plus retry, rate, cost, and abuse controls beyond the implemented timeout/cancellation foundation.
- Dependency and artifact integrity.
- Secure Lens persistence/import/share validation if those capabilities are introduced.
- Security testing and evidence.
