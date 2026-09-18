# GoreeCloud Search — Security

## Lifecycle

GoreeCloud Search is currently in Development and is not production-qualified.

## Reporting security issues

Do not publish sensitive exploit details, credentials, tokens, private data, or active attack instructions in a public issue.

Use GitHub private vulnerability reporting when it is enabled for this repository, or another approved private GoreeCloud security-reporting channel.

## Current security boundary

Version `0.1.0.dev12` ships a provider-execution framework plus one opt-in Brave Web Search API transport for Development use. Implemented controls include input validation, deterministic source planning, fail-closed source-mode privacy rules, pre-execution third-party disclosure budgeting, execution only of plan-approved providers, bounded concurrency/per-provider timeouts, cancellation, failure isolation, provider-provenance validation, strict Search ↔ Index contract/pagination validation, conservative result-URL validation, content-policy provenance/error handling, pre-execution SafeSearch enforcement checks, deterministic ranking, and local Lens handling. The Brave transport uses a fixed HTTPS endpoint, refuses redirects, keeps its API key out of query parameters and result evidence, caps one request at 20 results, limits response bodies to 2 MiB, validates returned result URLs, and treats malformed provider items as degraded input rather than trusted data.

Portable Lens import uses an exact versioned schema. It rejects duplicate JSON keys, unknown fields, unsupported format versions/rule enums, non-finite numeric values, invalid weights, excessive rule counts, and oversized documents. Serialization is deterministic. Import creates only typed Lens configuration; it does not execute code, fetch URLs, open files, contact providers, or grant source-plan authority.

A requested Lens must resolve before provider execution. The selected Lens is removed from the provider-facing query, and Lens rules operate only on normalized Search results after content policy.

## Future requirements

The local Development HTTP command remains structurally fixed to IPv4 loopback. A separate `serve-container` command exists only for the Docker-network deployment candidate: it binds inside the container network on port 8080, validates the HTTP Host header against loopback plus the configured service hostname, is intended to remain unpublished on the Docker host, and is reachable through the existing Caddy `proxy` network path. Both boundaries expose bounded GET health/API/OpenSearch behavior; the human `/search` surface additionally accepts only bounded form-encoded POST requests while API POST and other mutating methods remain rejected. They omit CORS authorization, bound request targets/search queries/result limits/form body size, send restrictive browser-facing response headers, suppress request-target logs, and return generic internal-failure responses. The HTML surface escapes query/provider-controlled text, serves only local CSS, loads no client-side script or remote presentation asset, constrains styles/forms to the same origin through CSP, denies framing, disables unnecessary browser permissions, applies same-origin opener/resource isolation, uses no-referrer/no-store/noindex controls, and advertises OpenSearch only from configured server origin state rather than reflecting an arbitrary Host header. Neither boundary provides GoreeCloud Identity authentication, authorization, production rate limiting, abuse controls, Glaze UI consumer acceptance, or production acceptance.

The external provider key must be supplied through protected runtime secret handling and must never be committed. The VPS candidate supports a protected runtime credential file and fails closed when both direct and file-based values are configured. The Compose example bind-mounts `/srv/docker/secrets/searxng/brave-search-api-key` read-only and requires the host file to be owned by UID/GID 10001 with mode `0400` so the non-root process can read it without broadening filesystem permissions. Its Compose example runs as UID/GID 10001, uses a read-only root filesystem, drops all Linux capabilities, enables no-new-privileges, bounds PIDs, provides only a small noexec/nosuid/nodev temporary filesystem, rotates Docker logs, exposes container port 8080 only to Docker networks, and requires an approved tag@digest image reference. These source controls are candidate evidence only; no current native image from this repository has been published or accepted on `goreecloud-vps-01`.

Before network-capable service operation is production-accepted, Search must implement and verify applicable Identity, Privacy Shield, Wardveil, Policy, Observability, secret separation, rate limiting/abuse resistance, stronger provider isolation/retry/rate/cost controls, dependency/artifact integrity, target-host monitoring and rollback, signed/trusted Lens distribution if remote sharing is introduced, and security testing/evidence.
