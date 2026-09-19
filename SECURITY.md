# GoreeCloud Search — Security

## Lifecycle

GoreeCloud Search is currently in Development and is not production-qualified.

## Reporting security issues

Do not publish sensitive exploit details, credentials, tokens, private data, or active attack instructions in a public issue.

Use GitHub private vulnerability reporting when it is enabled for this repository, or another approved private GoreeCloud security-reporting channel.

## Current security boundary

Version `0.1.0.dev4` ships a provider-execution framework but no authenticated live network transport. Implemented controls include input validation, deterministic source planning, a fail-closed rule preventing external-provider inclusion in GoreeCloud-only and offline/local modes, execution only of plan-approved providers, bounded concurrency and per-provider timeouts, cancellation propagation, failure isolation, provider-provenance validation, strict Search ↔ Index contract-version/pagination validation, conservative result-URL validation, rejection of candidates from undeclared providers, and deterministic ranking that does not invoke network services.

## Future requirements

Before network-capable or privileged service operation is production-accepted, Search must implement and verify applicable:

- GoreeCloud Identity authentication and authorization.
- Privacy Shield permitted-use enforcement.
- Wardveil Security trust and protection controls.
- Secret separation and secure configuration.
- Rate limiting and abuse resistance.
- Stronger provider sandboxing/isolation plus retry, rate, cost, and abuse controls beyond the implemented timeout/cancellation foundation.
- Dependency and artifact integrity.
- Security testing and evidence.
