# GoreeCloud Search — Security

## Lifecycle

GoreeCloud Search is currently in Development and is not production-qualified.

## Reporting security issues

Do not publish sensitive exploit details, credentials, tokens, private data, or active attack instructions in a public issue.

Use GitHub private vulnerability reporting when it is enabled for this repository, or another approved private GoreeCloud security-reporting channel.

## Current security boundary

Version `0.1.0.dev3` performs no provider network execution. The primary implemented controls are input validation, deterministic source planning, a fail-closed rule preventing external-provider inclusion in GoreeCloud-only and offline/local modes, strict Search ↔ Index contract-version validation, conservative result-URL validation, rejection of candidates from undeclared providers, and deterministic ranking that does not execute code or invoke network services.

## Future requirements

Before network-capable or privileged service operation is production-accepted, Search must implement and verify applicable:

- GoreeCloud Identity authentication and authorization.
- Privacy Shield permitted-use enforcement.
- Wardveil Security trust and protection controls.
- Secret separation and secure configuration.
- Rate limiting and abuse resistance.
- Provider isolation and timeout/cancellation behavior.
- Dependency and artifact integrity.
- Security testing and evidence.
