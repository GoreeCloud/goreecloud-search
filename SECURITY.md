# GoreeCloud Search — Security

## Lifecycle

GoreeCloud Search is currently in Development and is not production-qualified.

## Reporting security issues

Do not publish sensitive exploit details, credentials, tokens, private data, or active attack instructions in a public issue.

Use GitHub private vulnerability reporting when it is enabled for this repository, or another approved private GoreeCloud security-reporting channel.

## Current security boundary

Version `0.1.0.dev9` ships a provider-execution framework plus one opt-in Brave Web Search API transport for Development use. Implemented controls include input validation, deterministic source planning, fail-closed source-mode privacy rules, pre-execution third-party disclosure budgeting, execution only of plan-approved providers, bounded concurrency/per-provider timeouts, cancellation, failure isolation, provider-provenance validation, strict Search ↔ Index contract/pagination validation, conservative result-URL validation, content-policy provenance/error handling, pre-execution SafeSearch enforcement checks, deterministic ranking, and local Lens handling. The Brave transport uses a fixed HTTPS endpoint, refuses redirects, keeps its API key out of query parameters and result evidence, caps one request at 20 results, limits response bodies to 2 MiB, validates returned result URLs, and treats malformed provider items as degraded input rather than trusted data.

Portable Lens import uses an exact versioned schema. It rejects duplicate JSON keys, unknown fields, unsupported format versions/rule enums, non-finite numeric values, invalid weights, excessive rule counts, and oversized documents. Serialization is deterministic. Import creates only typed Lens configuration; it does not execute code, fetch URLs, open files, contact providers, or grant source-plan authority.

A requested Lens must resolve before provider execution. The selected Lens is removed from the provider-facing query, and Lens rules operate only on normalized Search results after content policy.

## Future requirements

The external provider key must be supplied through protected runtime secret handling and must never be committed. Before network-capable or privileged service operation is production-accepted, Search must implement and verify applicable Identity, Privacy Shield, Wardveil, secret separation, rate limiting/abuse resistance, stronger provider isolation/retry/rate/cost controls, dependency/artifact integrity, signed/trusted Lens distribution if remote sharing is introduced, and security testing/evidence.
