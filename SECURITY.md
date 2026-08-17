# GoreeCloud Search Security Policy

GoreeCloud Search is a GoreeCloud-maintained fork of SearXNG. Security reports should be routed according to where the affected behavior lives so GoreeCloud-specific changes are not incorrectly treated as upstream issues and upstream vulnerabilities still reach the SearXNG security team.

## GoreeCloud Search issues

Use the GoreeCloud Search repository for vulnerabilities involving GoreeCloud-maintained code or configuration, including Glaze UI changes, GoreeCloud templates and browser metadata, GoreeCloud runtime defaults, deployment examples, CI/acceptance tooling, container metadata, or GoreeCloud-specific integration behavior.

For a sensitive vulnerability, use GitHub's private vulnerability-reporting option on the repository Security tab when it is available. Do not publish exploit details, secrets, private infrastructure information, or a working proof of concept in a public issue. If private reporting is not available, open only a minimal public issue requesting a private reporting channel and omit sensitive technical details until a private channel is established.

For non-sensitive hardening suggestions or security documentation corrections, a normal repository issue is appropriate.

A useful report includes the affected GoreeCloud Search revision or image identifier, the affected component, reproduction conditions, expected and observed behavior, impact, and any safe mitigation or rollback information.

## Upstream SearXNG issues

If the vulnerability is reproducible in unchanged upstream SearXNG and is not caused by a GoreeCloud-specific patch or deployment choice, follow the upstream SearXNG security policy and contact `security@searxng.org` privately.

When the correct ownership is unclear, begin with GoreeCloud Search and identify whether the affected path differs from upstream. Maintainers can then coordinate or redirect the report without requiring the reporter to disclose the vulnerability publicly.

## Production information

Security reports must not include GoreeCloud production secrets, private keys, tokens, private DNS data, internal addresses, user information, or other sensitive infrastructure details. Use sanitized examples whenever practical.

Security is part of GoreeCloud Search release readiness. A confirmed issue that materially affects confidentiality, integrity, authorization, safe operation, or recovery must be resolved or explicitly accepted before the affected release is promoted.