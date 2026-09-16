# GoreeCloud Search — Security

## Lifecycle

GoreeCloud Search is in Development and is not production-qualified.

## Reporting security issues

Do not publish sensitive exploit details, credentials, tokens, private data, or active attack instructions in a public issue. Use GitHub private vulnerability reporting when enabled or another approved private GoreeCloud security-reporting channel.

## Current boundary

Version `0.1.0.dev1` performs no provider network execution. Implemented controls are input validation, deterministic source planning, and a fail-closed rule preventing external-provider inclusion in GoreeCloud-only and offline/local modes.

## Future requirements

Before network-capable or privileged operation is production-accepted, Search must implement and verify applicable Identity, Privacy Shield, Wardveil Security, secure configuration/secret separation, rate limiting, provider isolation, dependency/artifact integrity, and security testing/evidence.
